package gateway

import (
	_ "embed"
	"net/http"
)

//go:embed playground/index.html
var playgroundHTML []byte

// servePlayground serves the embedded WebChat Playground HTML page.
func (s *Server) servePlayground(w http.ResponseWriter, _ *http.Request) {
	w.Header().Set("Content-Type", "text/html; charset=utf-8")
	w.Header().Set("Cache-Control", "no-cache")
	_, _ = w.Write(playgroundHTML)
}
