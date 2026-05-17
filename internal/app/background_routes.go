package app

import (
	"net/http"
	"strings"
	"time"

	"github.com/go-chi/chi/v5"

	"github.com/langoai/lango/internal/background"
	"github.com/langoai/lango/internal/gateway"
)

func registerBackgroundRoutes(r chi.Router, app *App, auth *gateway.AuthManager) {
	r.Route("/api/bg", func(r chi.Router) {
		r.Use(gateway.RequireAuth(auth))
		r.Get("/tasks", backgroundListHandler(app))
		r.Get("/tasks/{id}", backgroundStatusHandler(app))
		r.Get("/tasks/{id}/result", backgroundResultHandler(app))
		r.Post("/tasks/{id}/cancel", backgroundCancelHandler(app))
	})
}

type backgroundTaskDTO struct {
	ID            string `json:"id"`
	Status        string `json:"status"`
	Prompt        string `json:"prompt"`
	OriginChannel string `json:"originChannel,omitempty"`
	OriginSession string `json:"originSession,omitempty"`
	StartedAt     string `json:"startedAt,omitempty"`
	CompletedAt   string `json:"completedAt,omitempty"`
	Duration      string `json:"duration,omitempty"`
	Error         string `json:"error,omitempty"`
	Result        string `json:"result,omitempty"`
}

func backgroundListHandler(app *App) http.HandlerFunc {
	return func(w http.ResponseWriter, _ *http.Request) {
		mgr, ok := backgroundRouteManager(w, app)
		if !ok {
			return
		}

		snaps := mgr.List()
		tasks := make([]backgroundTaskDTO, 0, len(snaps))
		for _, snap := range snaps {
			tasks = append(tasks, backgroundTaskDTOFromSnapshot(snap))
		}

		w.Header().Set("Content-Type", "application/json")
		writeJSON(w, map[string]any{"tasks": tasks})
	}
}

func backgroundStatusHandler(app *App) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		mgr, ok := backgroundRouteManager(w, app)
		if !ok {
			return
		}

		snap, err := mgr.Status(chi.URLParam(r, "id"))
		if err != nil {
			writeBackgroundRouteError(w, http.StatusNotFound, err.Error())
			return
		}

		w.Header().Set("Content-Type", "application/json")
		writeJSON(w, map[string]any{"task": backgroundTaskDTOFromSnapshot(*snap)})
	}
}

func backgroundResultHandler(app *App) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		mgr, ok := backgroundRouteManager(w, app)
		if !ok {
			return
		}

		id := chi.URLParam(r, "id")
		snap, err := mgr.Status(id)
		if err != nil {
			writeBackgroundRouteError(w, http.StatusNotFound, err.Error())
			return
		}
		if snap.Status != background.Done {
			writeBackgroundRouteError(w, http.StatusConflict, "task result is not available until task is done")
			return
		}

		result, err := mgr.Result(id)
		if err != nil {
			writeBackgroundRouteError(w, http.StatusConflict, err.Error())
			return
		}

		w.Header().Set("Content-Type", "application/json")
		writeJSON(w, map[string]any{"result": result})
	}
}

func backgroundCancelHandler(app *App) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		mgr, ok := backgroundRouteManager(w, app)
		if !ok {
			return
		}

		id := chi.URLParam(r, "id")
		if err := mgr.Cancel(id); err != nil {
			if strings.Contains(err.Error(), "not found") {
				writeBackgroundRouteError(w, http.StatusNotFound, err.Error())
				return
			}
			writeBackgroundRouteError(w, http.StatusConflict, err.Error())
			return
		}

		w.Header().Set("Content-Type", "application/json")
		writeJSON(w, map[string]any{"id": id, "cancelled": true})
	}
}

func backgroundRouteManager(w http.ResponseWriter, app *App) (*background.Manager, bool) {
	if app == nil || app.BackgroundManager == nil {
		writeBackgroundRouteError(w, http.StatusServiceUnavailable, "background manager unavailable")
		return nil, false
	}
	return app.BackgroundManager, true
}

func backgroundTaskDTOFromSnapshot(snap background.TaskSnapshot) backgroundTaskDTO {
	dto := backgroundTaskDTO{
		ID:            snap.ID,
		Status:        snap.Status.String(),
		Prompt:        snap.Prompt,
		OriginChannel: snap.OriginChannel,
		OriginSession: snap.OriginSession,
		Error:         snap.Error,
		Result:        snap.Result,
	}

	if !snap.StartedAt.IsZero() {
		dto.StartedAt = snap.StartedAt.Format(time.RFC3339)
	}
	if !snap.CompletedAt.IsZero() {
		dto.CompletedAt = snap.CompletedAt.Format(time.RFC3339)
	}
	if !snap.StartedAt.IsZero() && !snap.CompletedAt.IsZero() {
		dto.Duration = snap.CompletedAt.Sub(snap.StartedAt).String()
	}

	return dto
}

func writeBackgroundRouteError(w http.ResponseWriter, status int, message string) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(status)
	writeJSON(w, map[string]string{"error": message})
}
