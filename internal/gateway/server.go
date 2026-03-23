package gateway

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"net/http"
	"strings"
	"sync"
	"time"

	"go.uber.org/zap"

	"github.com/go-chi/chi/v5"
	"github.com/go-chi/chi/v5/middleware"
	"github.com/gorilla/websocket"
	"github.com/langoai/lango/internal/adk"
	"github.com/langoai/lango/internal/approval"
	"github.com/langoai/lango/internal/config"
	"github.com/langoai/lango/internal/deadline"
	"github.com/langoai/lango/internal/gatekeeper"
	"github.com/langoai/lango/internal/logging"
	"github.com/langoai/lango/internal/runledger"
	"github.com/langoai/lango/internal/security"
	"github.com/langoai/lango/internal/session"
)

func logger() *zap.SugaredLogger { return logging.Gateway() }

// emptyResponseFallback is returned to the user when the agent succeeds
// but produces no visible text (e.g. Gemini thought-only responses).
const emptyResponseFallback = "I processed your message but couldn't formulate a visible response. Could you try rephrasing your question?"

// TurnCallback is called after each agent turn completes (for buffer triggers, etc).
type TurnCallback func(sessionKey string)

// Server represents the gateway server
type Server struct {
	config             Config
	agent              *adk.Agent
	provider           *security.RPCProvider
	auth               *AuthManager
	store              session.Store
	runLedgerStore     runledger.RunLedgerStore
	router             chi.Router
	httpServer         *http.Server
	upgrader           websocket.Upgrader
	clients            map[string]*Client
	clientsMu          sync.RWMutex
	handlers           map[string]RPCHandler
	handlersMu         sync.RWMutex
	pendingApprovals   map[string]chan approval.ApprovalResponse
	pendingApprovalsMu sync.Mutex
	turnCallbacks      []TurnCallback
	sanitizer          *gatekeeper.Sanitizer
	shutdownCtx        context.Context
	shutdownCancel     context.CancelFunc
}

// Config holds gateway server configuration
type Config struct {
	Host             string
	Port             int
	HTTPEnabled      bool
	WebSocketEnabled bool
	AllowedOrigins   []string
	ApprovalTimeout  time.Duration
	RequestTimeout   time.Duration
	IdleTimeout      time.Duration // inactivity timeout (0 = disabled)
	MaxTimeout       time.Duration // absolute hard ceiling
	RunLedger        config.RunLedgerConfig
}

// Client represents a connected WebSocket client
type Client struct {
	ID         string
	Type       string // "ui" or "companion"
	Conn       *websocket.Conn
	Server     *Server
	Send       chan []byte
	SessionKey string
	closed     bool
	closeMu    sync.Mutex
}

// RPCRequest represents an incoming RPC request
type RPCRequest struct {
	ID     string          `json:"id"`
	Method string          `json:"method"`
	Params json.RawMessage `json:"params,omitempty"`
}

// RPCResponse represents an RPC response
type RPCResponse struct {
	ID     string      `json:"id"`
	Result interface{} `json:"result,omitempty"`
	Error  *RPCError   `json:"error,omitempty"`
}

// RPCError represents an RPC error
type RPCError struct {
	Code    int    `json:"code"`
	Message string `json:"message"`
}

// RPCHandler is a function that handles an RPC method.
// The client parameter provides the calling client's context (session, type, etc).
type RPCHandler func(client *Client, params json.RawMessage) (interface{}, error)

// New creates a new gateway server
func New(cfg Config, agent *adk.Agent, provider *security.RPCProvider, store session.Store, auth *AuthManager) *Server {
	originChecker := makeOriginChecker(cfg.AllowedOrigins)
	shutdownCtx, shutdownCancel := context.WithCancel(context.Background())

	s := &Server{
		config:           cfg,
		agent:            agent,
		provider:         provider,
		auth:             auth,
		store:            store,
		router:           chi.NewRouter(),
		clients:          make(map[string]*Client),
		handlers:         make(map[string]RPCHandler),
		pendingApprovals: make(map[string]chan approval.ApprovalResponse),
		shutdownCtx:      shutdownCtx,
		shutdownCancel:   shutdownCancel,
		upgrader: websocket.Upgrader{
			ReadBufferSize:  1024,
			WriteBufferSize: 1024,
			CheckOrigin:     originChecker,
		},
	}
	s.setupRoutes()

	// Register RPC handlers
	s.RegisterHandler("chat.message", s.handleChatMessage)
	s.RegisterHandler("sign.response", s.handleSignResponse)
	s.RegisterHandler("encrypt.response", s.handleEncryptResponse)
	s.RegisterHandler("decrypt.response", s.handleDecryptResponse)
	s.RegisterHandler("companion.hello", s.handleCompanionHello)
	s.RegisterHandler("approval.response", s.handleApprovalResponse)

	// Wire up provider sender
	if s.provider != nil {
		s.provider.SetSender(func(event string, payload interface{}) error {
			// Routes signing/decryption requests to companions
			if strings.HasPrefix(event, "sign.") || strings.HasPrefix(event, "encrypt.") || strings.HasPrefix(event, "decrypt.") {
				s.BroadcastToCompanions(event, payload)
			} else {
				s.Broadcast(event, payload)
			}
			return nil
		})
	}

	return s
}

// handleChatMessage processes chat messages via Agent
func (s *Server) handleChatMessage(client *Client, params json.RawMessage) (interface{}, error) {
	var req struct {
		Message       string `json:"message"`
		SessionKey    string `json:"sessionKey"`
		ResumeRunID   string `json:"resumeRunId"`
		ConfirmResume bool   `json:"confirmResume"`
	}
	if err := json.Unmarshal(params, &req); err != nil {
		return nil, fmt.Errorf("invalid params: %w", err)
	}

	if req.Message == "" && !(req.ConfirmResume && req.ResumeRunID != "") {
		return nil, fmt.Errorf("message is required")
	}

	// Determine session key:
	// - Authenticated client: always use their authenticated session key
	// - Unauthenticated (auth disabled): use provided key or "default"
	sessionKey := "default"
	if client.SessionKey != "" {
		// Authenticated user — force their own session
		sessionKey = client.SessionKey
	} else if req.SessionKey != "" {
		// No auth — allow client-specified key
		sessionKey = req.SessionKey
	}

	if s.runLedgerStore != nil {
		rm := runledger.NewResumeManager(s.runLedgerStore, s.config.RunLedger.StaleTTL)

		if req.ConfirmResume && req.ResumeRunID != "" {
			resumeCtx, cancel := s.newResumeContext()
			defer cancel()

			if _, err := rm.Resume(resumeCtx, req.ResumeRunID, "user"); err != nil {
				return nil, fmt.Errorf("resume run: %w", err)
			}
			s.BroadcastToSession(sessionKey, "agent.resume_confirmed", map[string]string{
				"sessionKey": sessionKey,
				"runId":      req.ResumeRunID,
			})
			return map[string]interface{}{
				"resumed": true,
				"runId":   req.ResumeRunID,
			}, nil
		}

		if runledger.DetectResumeIntent(req.Message) {
			resumeCtx, cancel := s.newResumeContext()
			defer cancel()

			candidates, err := rm.FindCandidates(resumeCtx, sessionKey)
			if err != nil {
				return nil, fmt.Errorf("find resume candidates: %w", err)
			}
			if len(candidates) > 0 {
				payload := map[string]interface{}{
					"sessionKey":  sessionKey,
					"candidates":  candidates,
					"message":     "Resume candidates found. Confirm one to continue.",
					"requiresAck": true,
				}
				s.BroadcastToSession(sessionKey, "agent.resume_required", payload)
				return map[string]interface{}{
					"resumeRequired": true,
					"candidates":     candidates,
				}, nil
			}
		}
	}

	if s.agent == nil {
		return nil, ErrAgentNotReady
	}

	// Notify UI that agent is thinking
	s.BroadcastToSession(sessionKey, "agent.thinking", map[string]string{
		"sessionKey": sessionKey,
	})

	var (
		ctx         context.Context
		cancel      context.CancelFunc
		extDeadline *deadline.ExtendableDeadline
		runOpts     []adk.RunOption
	)

	idleTimeout := s.config.IdleTimeout
	hardCeiling := s.config.MaxTimeout
	if hardCeiling <= 0 {
		hardCeiling = s.config.RequestTimeout
	}
	if hardCeiling <= 0 {
		hardCeiling = 5 * time.Minute
	}

	if idleTimeout > 0 {
		ctx, extDeadline = deadline.New(s.shutdownCtx, idleTimeout, hardCeiling)
		cancel = extDeadline.Stop
		runOpts = append(runOpts, adk.WithOnActivity(extDeadline.Extend))
	} else {
		ctx, cancel = context.WithTimeout(s.shutdownCtx, hardCeiling)
	}
	defer cancel()

	// Warn UI when approaching timeout (80%).
	warnTimer := time.AfterFunc(time.Duration(float64(hardCeiling)*0.8), func() {
		logger().Warnw("agent request approaching timeout",
			"session", sessionKey,
			"timeout", hardCeiling.String())
		s.BroadcastToSession(sessionKey, "agent.warning", map[string]string{
			"sessionKey": sessionKey,
			"message":    "Request is taking longer than expected",
			"type":       "approaching_timeout",
		})
	})
	defer warnTimer.Stop()

	// Start periodic progress broadcast every 15s.
	progressStart := time.Now()
	progressDone := make(chan struct{})
	var progressOnce sync.Once
	go func() {
		ticker := time.NewTicker(15 * time.Second)
		defer ticker.Stop()
		for {
			select {
			case <-progressDone:
				return
			case <-ticker.C:
				elapsed := time.Since(progressStart).Truncate(time.Second)
				s.BroadcastToSession(sessionKey, "agent.progress", map[string]string{
					"sessionKey": sessionKey,
					"elapsed":    elapsed.String(),
					"message":    fmt.Sprintf("Thinking... (%s)", elapsed),
				})
			}
		}
	}()
	stopProgress := func() { progressOnce.Do(func() { close(progressDone) }) }

	ctx = session.WithSessionKey(ctx, sessionKey)
	response, err := s.agent.RunStreaming(ctx, sessionKey, req.Message, func(chunk string) {
		if s.sanitizer != nil && s.sanitizer.Enabled() {
			chunk = s.sanitizer.Sanitize(chunk)
		}
		if chunk == "" {
			return
		}
		s.BroadcastToSession(sessionKey, "agent.chunk", map[string]string{
			"sessionKey": sessionKey,
			"chunk":      chunk,
		})
	}, runOpts...)

	// Stop progress updates now that the agent has finished.
	stopProgress()

	// Fire turn-complete callbacks (buffer triggers, etc.) regardless of error.
	for _, cb := range s.turnCallbacks {
		cb(sessionKey)
	}

	// Guard against empty responses (e.g. Gemini thought-only output).
	if err == nil && response == "" {
		response = emptyResponseFallback
		logger().Warnw("empty agent response, using fallback",
			"session", sessionKey)
	}

	// Apply response sanitization.
	if err == nil && s.sanitizer != nil && s.sanitizer.Enabled() {
		response = s.sanitizer.Sanitize(response)
	}

	if err != nil {
		// Classify the error for UI display.
		errType := "unknown"
		errCode := ""
		hint := ""
		userMsg := err.Error()

		var agentErr *adk.AgentError
		if errors.As(err, &agentErr) {
			errType = string(agentErr.Code)
			errCode = string(agentErr.Code)
			userMsg = agentErr.UserMessage()
		}

		if ctx.Err() != nil {
			errType = string(deadline.ReasonMaxTimeout)
			if extDeadline != nil {
				switch extDeadline.Reason() {
				case deadline.ReasonIdle:
					errType = string(deadline.ReasonIdle)
					errCode = string(adk.ErrIdleTimeout)
				case deadline.ReasonMaxTimeout:
					errType = string(deadline.ReasonMaxTimeout)
				}
			}
			// Annotate session to prevent error leak.
			if s.store != nil {
				_ = s.store.AnnotateTimeout(sessionKey, "")
			}
		}

		// Notify UI of the error so it can stop thinking indicators
		// and display a user-visible error message.
		s.BroadcastToSession(sessionKey, "agent.error", map[string]string{
			"sessionKey": sessionKey,
			"error":      userMsg,
			"type":       errType,
			"code":       errCode,
			"hint":       hint,
		})
		return nil, err
	}

	// Notify UI that agent completed successfully.
	s.BroadcastToSession(sessionKey, "agent.done", map[string]string{
		"sessionKey": sessionKey,
	})

	return map[string]string{
		"response": response,
	}, nil
}

func (s *Server) newResumeContext() (context.Context, context.CancelFunc) {
	timeout := s.config.MaxTimeout
	if timeout <= 0 {
		timeout = s.config.RequestTimeout
	}
	if timeout <= 0 {
		timeout = 5 * time.Minute
	}
	return context.WithTimeout(s.shutdownCtx, timeout)
}

// BroadcastToSession sends an event to all UI clients belonging to a specific session.
// When the session key is empty (no auth), it broadcasts to all UI clients.
func (s *Server) BroadcastToSession(sessionKey, event string, payload interface{}) {
	msg, _ := json.Marshal(map[string]interface{}{
		"type":    "event",
		"event":   event,
		"payload": payload,
	})

	s.clientsMu.RLock()
	defer s.clientsMu.RUnlock()

	for _, client := range s.clients {
		if client.Type != "ui" {
			continue
		}
		// If authenticated, scope to the session; otherwise broadcast to all UI clients
		if sessionKey != "" && client.SessionKey != "" && client.SessionKey != sessionKey {
			continue
		}
		select {
		case client.Send <- msg:
		default:
			// Client buffer full, skip
		}
	}
}

// handleSignResponse proxies signature responses to the RPCProvider
func (s *Server) handleSignResponse(_ *Client, params json.RawMessage) (interface{}, error) {
	if s.provider == nil {
		return nil, fmt.Errorf("provider not configured")
	}

	var resp security.SignResponse
	if err := json.Unmarshal(params, &resp); err != nil {
		return nil, fmt.Errorf("invalid params: %w", err)
	}

	if err := s.provider.HandleSignResponse(resp); err != nil {
		return nil, err
	}

	return map[string]string{"status": "ok"}, nil
}

// handleEncryptResponse proxies encryption responses to the RPCProvider
func (s *Server) handleEncryptResponse(_ *Client, params json.RawMessage) (interface{}, error) {
	if s.provider == nil {
		return nil, fmt.Errorf("provider not configured")
	}

	var resp security.EncryptResponse
	if err := json.Unmarshal(params, &resp); err != nil {
		return nil, fmt.Errorf("invalid params: %w", err)
	}

	if err := s.provider.HandleEncryptResponse(resp); err != nil {
		return nil, err
	}

	return map[string]string{"status": "ok"}, nil
}

// handleDecryptResponse proxies decryption responses to the RPCProvider
func (s *Server) handleDecryptResponse(_ *Client, params json.RawMessage) (interface{}, error) {
	if s.provider == nil {
		return nil, fmt.Errorf("provider not configured")
	}

	var resp security.DecryptResponse
	if err := json.Unmarshal(params, &resp); err != nil {
		return nil, fmt.Errorf("invalid params: %w", err)
	}

	if err := s.provider.HandleDecryptResponse(resp); err != nil {
		return nil, err
	}

	return map[string]string{"status": "ok"}, nil
}

// RequestApproval broadcasts an approval request to companions and waits for response.
func (s *Server) RequestApproval(ctx context.Context, message string) (approval.ApprovalResponse, error) {
	// 1. Check if any companions connected
	s.clientsMu.RLock()
	hasCompanion := false
	for _, c := range s.clients {
		if c.Type == "companion" {
			hasCompanion = true
			break
		}
	}
	s.clientsMu.RUnlock()

	if !hasCompanion {
		return approval.ApprovalResponse{}, ErrNoCompanion
	}

	// 2. Create approval request
	id := fmt.Sprintf("req-%d", time.Now().UnixNano())
	respChan := make(chan approval.ApprovalResponse, 1)

	s.pendingApprovalsMu.Lock()
	s.pendingApprovals[id] = respChan
	s.pendingApprovalsMu.Unlock()

	defer func() {
		s.pendingApprovalsMu.Lock()
		delete(s.pendingApprovals, id)
		s.pendingApprovalsMu.Unlock()
	}()

	// 3. Broadcast request
	req := map[string]string{
		"id":      id,
		"message": message,
	}
	s.BroadcastToCompanions("approval.request", req)

	// 4. Wait for response or timeout
	timeout := s.config.ApprovalTimeout
	if timeout <= 0 {
		timeout = 30 * time.Second
	}

	select {
	case resp := <-respChan:
		return resp, nil
	case <-ctx.Done():
		return approval.ApprovalResponse{}, ctx.Err()
	case <-time.After(timeout):
		return approval.ApprovalResponse{}, ErrApprovalTimeout
	}
}

// handleCompanionHello processes companion hello message
func (s *Server) handleCompanionHello(_ *Client, params json.RawMessage) (interface{}, error) {
	var req struct {
		DeviceID  string `json:"deviceId"`
		PublicKey string `json:"publicKey"`
	}
	if err := json.Unmarshal(params, &req); err != nil {
		return nil, fmt.Errorf("invalid params: %w", err)
	}

	// Logic to store device capabilities or register pubkey for encryption can go here
	logger().Infow("companion hello received", "deviceId", req.DeviceID)

	return map[string]string{"status": "ok"}, nil
}

// handleApprovalResponse processes approval response from companion
func (s *Server) handleApprovalResponse(_ *Client, params json.RawMessage) (interface{}, error) {
	var req struct {
		RequestID   string `json:"requestId"`
		Approved    bool   `json:"approved"`
		AlwaysAllow bool   `json:"alwaysAllow"`
	}
	if err := json.Unmarshal(params, &req); err != nil {
		return nil, fmt.Errorf("invalid params: %w", err)
	}

	s.pendingApprovalsMu.Lock()
	ch, exists := s.pendingApprovals[req.RequestID]
	if exists {
		delete(s.pendingApprovals, req.RequestID)
	}
	s.pendingApprovalsMu.Unlock()

	if exists {
		resp := approval.ApprovalResponse{
			Approved:    req.Approved,
			AlwaysAllow: req.AlwaysAllow,
		}
		// Non-blocking send
		select {
		case ch <- resp:
		default:
		}
	}

	return map[string]string{"status": "ok"}, nil
}

// setupRoutes configures HTTP routes
func (s *Server) setupRoutes() {
	s.router.Use(middleware.Logger)
	s.router.Use(middleware.Recoverer)
	s.router.Use(middleware.RequestID)

	// Public routes — no auth required
	if s.config.HTTPEnabled {
		s.router.Get("/health", s.handleHealth)
	}

	// Auth routes — public, with rate limiting
	if s.auth != nil {
		s.auth.RegisterRoutes(s.router)
	}

	// Protected routes — require auth when OIDC is configured
	s.router.Group(func(r chi.Router) {
		r.Use(RequireAuth(s.auth))

		if s.config.HTTPEnabled {
			r.Get("/status", s.handleStatus)
			r.Get("/playground", s.servePlayground)
		}
		if s.config.WebSocketEnabled {
			r.Get("/ws", s.handleWebSocket)
		}
	})

	// Companion endpoint — separate group, no OIDC auth, origin restriction only
	if s.config.WebSocketEnabled && s.provider != nil {
		s.router.Get("/companion", s.handleCompanionWebSocket)
	}
}

// Router returns the underlying chi.Router for mounting additional routes.
func (s *Server) Router() chi.Router {
	return s.router
}

// SetAgent sets the agent on the server (used for deferred wiring).
func (s *Server) SetAgent(agent *adk.Agent) {
	s.agent = agent
}

// SetSanitizer sets the response sanitizer for output gatekeeper filtering.
func (s *Server) SetSanitizer(san *gatekeeper.Sanitizer) {
	s.sanitizer = san
}

// SetRunLedgerStore wires optional RunLedger access for resume and authoritative flows.
func (s *Server) SetRunLedgerStore(store runledger.RunLedgerStore) {
	s.runLedgerStore = store
}

// OnTurnComplete registers a callback that fires after each agent turn.
func (s *Server) OnTurnComplete(cb TurnCallback) {
	s.turnCallbacks = append(s.turnCallbacks, cb)
}

// RegisterHandler registers an RPC method handler
func (s *Server) RegisterHandler(method string, handler RPCHandler) {
	s.handlersMu.Lock()
	defer s.handlersMu.Unlock()
	s.handlers[method] = handler
}

// Start starts the gateway server
func (s *Server) Start() error {
	addr := fmt.Sprintf("%s:%d", s.config.Host, s.config.Port)
	s.httpServer = &http.Server{
		Addr:         addr,
		Handler:      s.router,
		ReadTimeout:  15 * time.Second,
		WriteTimeout: 15 * time.Second,
		IdleTimeout:  60 * time.Second,
	}

	logger().Infow("gateway server is listening", "address", addr, "http", s.config.HTTPEnabled, "ws", s.config.WebSocketEnabled)
	err := s.httpServer.ListenAndServe()
	if errors.Is(err, http.ErrServerClosed) {
		return nil
	}
	return err
}

// Shutdown gracefully shuts down the server
func (s *Server) Shutdown(ctx context.Context) error {
	// Cancel all in-flight request contexts so agent runs stop immediately.
	s.shutdownCancel()

	// Close all WebSocket connections
	s.clientsMu.Lock()
	for _, client := range s.clients {
		client.Close()
	}
	s.clientsMu.Unlock()

	return s.httpServer.Shutdown(ctx)
}

// handleHealth returns health status
func (s *Server) handleHealth(w http.ResponseWriter, _ *http.Request) {
	w.Header().Set("Content-Type", "application/json")
	_ = json.NewEncoder(w).Encode(map[string]string{
		"status": "ok",
		"time":   time.Now().Format(time.RFC3339),
	})
}

// handleStatus returns server status
func (s *Server) handleStatus(w http.ResponseWriter, _ *http.Request) {
	s.clientsMu.RLock()
	clientCount := len(s.clients)
	s.clientsMu.RUnlock()

	w.Header().Set("Content-Type", "application/json")
	_ = json.NewEncoder(w).Encode(map[string]interface{}{
		"status":      "running",
		"clients":     clientCount,
		"wsEnabled":   s.config.WebSocketEnabled,
		"httpEnabled": s.config.HTTPEnabled,
	})
}

// handleWebSocket handles WebSocket upgrade and connection for UI clients
func (s *Server) handleWebSocket(w http.ResponseWriter, r *http.Request) {
	s.handleWebSocketConnection(w, r, "ui")
}

// handleCompanionWebSocket handles WebSocket upgrade and connection for companion apps
func (s *Server) handleCompanionWebSocket(w http.ResponseWriter, r *http.Request) {
	s.handleWebSocketConnection(w, r, "companion")
}

// handleWebSocketConnection handles generic WebSocket upgrade
func (s *Server) handleWebSocketConnection(w http.ResponseWriter, r *http.Request, clientType string) {
	conn, err := s.upgrader.Upgrade(w, r, nil)
	if err != nil {
		logger().Errorw("websocket upgrade failed", "error", err)
		return
	}

	clientID := fmt.Sprintf("%s-%d", clientType, time.Now().UnixNano())

	// Bind authenticated session to client; isolate unauthenticated clients
	// by assigning the unique clientID as their session key.
	sessionKey := SessionFromContext(r.Context())
	if sessionKey == "" {
		sessionKey = clientID
	}

	client := &Client{
		ID:         clientID,
		Type:       clientType,
		Conn:       conn,
		Server:     s,
		Send:       make(chan []byte, 256),
		SessionKey: sessionKey,
	}

	s.clientsMu.Lock()
	s.clients[clientID] = client
	s.clientsMu.Unlock()

	logger().Infow("client connected", "clientId", clientID, "authenticated", sessionKey != "")

	// Start read/write pumps
	go client.writePump()
	go client.readPump()
}

// Broadcast sends a message to all connected clients (defaulting to UI)
func (s *Server) Broadcast(event string, payload interface{}) {
	s.broadcastToType(event, payload, "ui")
}

// BroadcastToCompanions sends a message to all connected companions
func (s *Server) BroadcastToCompanions(event string, payload interface{}) {
	s.broadcastToType(event, payload, "companion")
}

func (s *Server) broadcastToType(event string, payload interface{}, targetType string) {
	msg, _ := json.Marshal(map[string]interface{}{
		"type":    "event",
		"event":   event,
		"payload": payload,
	})

	s.clientsMu.RLock()
	defer s.clientsMu.RUnlock()

	for _, client := range s.clients {
		if client.Type == targetType || targetType == "all" {
			select {
			case client.Send <- msg:
			default:
				// Client buffer full, skip
			}
		}
	}
}

// readPump reads messages from WebSocket
func (c *Client) readPump() {
	defer func() {
		if r := recover(); r != nil {
			logger().Errorw("readPump panic recovered", "clientId", c.ID, "panic", r)
		}
		c.Server.removeClient(c.ID)
		c.Close()
	}()

	for {
		_, message, err := c.Conn.ReadMessage()
		if err != nil {
			if websocket.IsUnexpectedCloseError(err, websocket.CloseGoingAway, websocket.CloseAbnormalClosure) {
				logger().Warnw("websocket read error", "clientId", c.ID, "error", err)
			}
			break
		}

		// Parse RPC request
		var req RPCRequest
		if err := json.Unmarshal(message, &req); err != nil {
			c.sendError(req.ID, -32700, "parse error")
			continue
		}

		// Handle request
		c.Server.handlersMu.RLock()
		handler, exists := c.Server.handlers[req.Method]
		c.Server.handlersMu.RUnlock()

		if !exists {
			c.sendError(req.ID, -32601, fmt.Sprintf("method not found: %s", req.Method))
			continue
		}

		c.handleRPC(req, handler)
	}
}

// handleRPC executes an RPC handler with panic recovery so a single handler
// panic does not tear down the entire readPump.
func (c *Client) handleRPC(req RPCRequest, handler RPCHandler) {
	defer func() {
		if r := recover(); r != nil {
			logger().Errorw("RPC handler panic recovered", "clientId", c.ID, "method", req.Method, "panic", r)
			c.sendError(req.ID, -32000, fmt.Sprintf("internal error: %v", r))
		}
	}()

	result, err := handler(c, req.Params)
	if err != nil {
		c.sendError(req.ID, -32000, err.Error())
		return
	}
	c.sendResult(req.ID, result)
}

// writePump writes messages to WebSocket
func (c *Client) writePump() {
	ticker := time.NewTicker(30 * time.Second)
	defer func() {
		if r := recover(); r != nil {
			logger().Errorw("writePump panic recovered", "clientId", c.ID, "panic", r)
		}
		ticker.Stop()
		c.Close()
	}()

	for {
		select {
		case message, ok := <-c.Send:
			if !ok {
				_ = c.Conn.WriteMessage(websocket.CloseMessage, []byte{})
				return
			}
			if err := c.Conn.WriteMessage(websocket.TextMessage, message); err != nil {
				return
			}
		case <-ticker.C:
			if err := c.Conn.WriteMessage(websocket.PingMessage, nil); err != nil {
				return
			}
		}
	}
}

func (c *Client) sendResult(id string, result interface{}) {
	resp := RPCResponse{ID: id, Result: result}
	data, _ := json.Marshal(resp)
	c.Send <- data
}

func (c *Client) sendError(id string, code int, message string) {
	resp := RPCResponse{ID: id, Error: &RPCError{Code: code, Message: message}}
	data, _ := json.Marshal(resp)
	c.Send <- data
}

// Close closes the client connection
func (c *Client) Close() {
	c.closeMu.Lock()
	defer c.closeMu.Unlock()

	if !c.closed {
		c.closed = true
		close(c.Send)
		c.Conn.Close()
	}
}

// HasCompanions returns true if at least one companion client is connected.
func (s *Server) HasCompanions() bool {
	s.clientsMu.RLock()
	defer s.clientsMu.RUnlock()

	for _, c := range s.clients {
		if c.Type == "companion" {
			return true
		}
	}
	return false
}

func (s *Server) removeClient(id string) {
	s.clientsMu.Lock()
	defer s.clientsMu.Unlock()

	if client, exists := s.clients[id]; exists {
		logger().Infow("client disconnected", "clientId", id)
		delete(s.clients, id)
		_ = client
	}
}
