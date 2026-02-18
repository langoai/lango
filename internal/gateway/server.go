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
	"github.com/langowarny/lango/internal/adk"
	"github.com/langowarny/lango/internal/logging"
	"github.com/langowarny/lango/internal/security"
	"github.com/langowarny/lango/internal/session"
)

func logger() *zap.SugaredLogger { return logging.Gateway() }

// TurnCallback is called after each agent turn completes (for buffer triggers, etc).
type TurnCallback func(sessionKey string)

// Server represents the gateway server
type Server struct {
	config             Config
	agent              *adk.Agent
	provider           *security.RPCProvider
	auth               *AuthManager
	store              session.Store
	router             chi.Router
	httpServer         *http.Server
	upgrader           websocket.Upgrader
	clients            map[string]*Client
	clientsMu          sync.RWMutex
	handlers           map[string]RPCHandler
	handlersMu         sync.RWMutex
	pendingApprovals   map[string]chan bool
	pendingApprovalsMu sync.Mutex
	turnCallbacks      []TurnCallback
}

// Config holds gateway server configuration
type Config struct {
	Host             string
	Port             int
	HTTPEnabled      bool
	WebSocketEnabled bool
	AllowedOrigins   []string
	ApprovalTimeout  time.Duration
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

	s := &Server{
		config:           cfg,
		agent:            agent,
		provider:         provider,
		auth:             auth,
		store:            store,
		router:           chi.NewRouter(),
		clients:          make(map[string]*Client),
		handlers:         make(map[string]RPCHandler),
		pendingApprovals: make(map[string]chan bool),
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
		Message    string `json:"message"`
		SessionKey string `json:"sessionKey"`
	}
	if err := json.Unmarshal(params, &req); err != nil {
		return nil, fmt.Errorf("invalid params: %w", err)
	}

	if req.Message == "" {
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

	if s.agent == nil {
		return nil, fmt.Errorf("agent not configured")
	}

	// Notify UI that agent is thinking
	s.BroadcastToSession(sessionKey, "agent.thinking", map[string]string{
		"sessionKey": sessionKey,
	})

	ctx := session.WithSessionKey(context.Background(), sessionKey)
	response, err := s.agent.RunAndCollect(ctx, sessionKey, req.Message)

	// Fire turn-complete callbacks (buffer triggers, etc.) regardless of error.
	for _, cb := range s.turnCallbacks {
		cb(sessionKey)
	}

	// Notify UI that agent is done
	s.BroadcastToSession(sessionKey, "agent.done", map[string]string{
		"sessionKey": sessionKey,
	})

	if err != nil {
		return nil, err
	}

	return map[string]string{
		"response": response,
	}, nil
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

// RequestApproval broadcasts an approval request to companions and waits for response
func (s *Server) RequestApproval(ctx context.Context, message string) (bool, error) {
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
		return false, fmt.Errorf("no companion connected")
	}

	// 2. Create approval request
	id := fmt.Sprintf("req-%d", time.Now().UnixNano())
	respChan := make(chan bool, 1)

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
	case approved := <-respChan:
		return approved, nil
	case <-ctx.Done():
		return false, ctx.Err()
	case <-time.After(timeout):
		return false, fmt.Errorf("approval timeout")
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
		RequestID string `json:"requestId"`
		Approved  bool   `json:"approved"`
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
		// Non-blocking send
		select {
		case ch <- req.Approved:
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
		r.Use(requireAuth(s.auth))

		if s.config.HTTPEnabled {
			r.Get("/status", s.handleStatus)
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
	json.NewEncoder(w).Encode(map[string]string{
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
	json.NewEncoder(w).Encode(map[string]interface{}{
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

	// Bind authenticated session to client (empty if no auth)
	sessionKey := SessionFromContext(r.Context())

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

		result, err := handler(c, req.Params)
		if err != nil {
			c.sendError(req.ID, -32000, err.Error())
			continue
		}

		c.sendResult(req.ID, result)
	}
}

// writePump writes messages to WebSocket
func (c *Client) writePump() {
	ticker := time.NewTicker(30 * time.Second)
	defer func() {
		ticker.Stop()
		c.Close()
	}()

	for {
		select {
		case message, ok := <-c.Send:
			if !ok {
				c.Conn.WriteMessage(websocket.CloseMessage, []byte{})
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
