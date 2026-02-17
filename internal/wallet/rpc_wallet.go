package wallet

import (
	"context"
	"fmt"
	"math/big"
	"sync"
	"time"

	"github.com/google/uuid"
)

// SenderFunc sends a message to the companion app via WebSocket.
type SenderFunc func(msgType string, payload interface{}) error

// SignTxRequest is sent to the companion for transaction signing.
type SignTxRequest struct {
	RequestID string `json:"requestId"`
	RawTx     []byte `json:"rawTx"`
}

// SignTxResponse is received from the companion after signing.
type SignTxResponse struct {
	RequestID string `json:"requestId"`
	Signature []byte `json:"signature,omitempty"`
	Error     string `json:"error,omitempty"`
}

// SignMsgRequest is sent to the companion for message signing.
type SignMsgRequest struct {
	RequestID string `json:"requestId"`
	Message   []byte `json:"message"`
}

// SignMsgResponse is received from the companion after message signing.
type SignMsgResponse struct {
	RequestID string `json:"requestId"`
	Signature []byte `json:"signature,omitempty"`
	Error     string `json:"error,omitempty"`
}

// AddressRequest is sent to the companion to retrieve the wallet address.
type AddressRequest struct {
	RequestID string `json:"requestId"`
}

// AddressResponse is received from the companion with the wallet address.
type AddressResponse struct {
	RequestID string `json:"requestId"`
	Address   string `json:"address,omitempty"`
	Error     string `json:"error,omitempty"`
}

// RPCWallet delegates wallet operations to a companion app via WebSocket RPC.
// Mirrors the security.RPCProvider pattern: correlation IDs + channel-based response.
type RPCWallet struct {
	sender  SenderFunc
	timeout time.Duration

	mu             sync.Mutex
	pendingSignTx  map[string]chan SignTxResponse
	pendingSignMsg map[string]chan SignMsgResponse
	pendingAddr    map[string]chan AddressResponse
}

// NewRPCWallet creates an RPC-based wallet provider.
func NewRPCWallet() *RPCWallet {
	return &RPCWallet{
		timeout:        30 * time.Second,
		pendingSignTx:  make(map[string]chan SignTxResponse),
		pendingSignMsg: make(map[string]chan SignMsgResponse),
		pendingAddr:    make(map[string]chan AddressResponse),
	}
}

// SetSender configures the transport for sending requests to the companion.
func (w *RPCWallet) SetSender(fn SenderFunc) {
	w.sender = fn
}

// Address requests the wallet address from the companion.
func (w *RPCWallet) Address(ctx context.Context) (string, error) {
	if w.sender == nil {
		return "", fmt.Errorf("RPC wallet: no sender configured")
	}

	reqID := uuid.New().String()
	ch := make(chan AddressResponse, 1)

	w.mu.Lock()
	w.pendingAddr[reqID] = ch
	w.mu.Unlock()

	defer func() {
		w.mu.Lock()
		delete(w.pendingAddr, reqID)
		w.mu.Unlock()
	}()

	if err := w.sender("wallet.address.request", AddressRequest{RequestID: reqID}); err != nil {
		return "", fmt.Errorf("send address request: %w", err)
	}

	select {
	case resp := <-ch:
		if resp.Error != "" {
			return "", fmt.Errorf("companion address error: %s", resp.Error)
		}
		return resp.Address, nil
	case <-time.After(w.timeout):
		return "", fmt.Errorf("wallet address request timed out")
	case <-ctx.Done():
		return "", ctx.Err()
	}
}

// Balance is not supported via RPC — returns an error directing to local wallet.
func (w *RPCWallet) Balance(_ context.Context) (*big.Int, error) {
	return nil, fmt.Errorf("RPC wallet: balance query not supported (use local RPC node)")
}

// SignTransaction sends a signing request to the companion and waits for the response.
func (w *RPCWallet) SignTransaction(ctx context.Context, rawTx []byte) ([]byte, error) {
	if w.sender == nil {
		return nil, fmt.Errorf("RPC wallet: no sender configured")
	}

	reqID := uuid.New().String()
	ch := make(chan SignTxResponse, 1)

	w.mu.Lock()
	w.pendingSignTx[reqID] = ch
	w.mu.Unlock()

	defer func() {
		w.mu.Lock()
		delete(w.pendingSignTx, reqID)
		w.mu.Unlock()
	}()

	if err := w.sender("wallet.sign_tx.request", SignTxRequest{
		RequestID: reqID,
		RawTx:     rawTx,
	}); err != nil {
		return nil, fmt.Errorf("send sign_tx request: %w", err)
	}

	select {
	case resp := <-ch:
		if resp.Error != "" {
			return nil, fmt.Errorf("companion sign error: %s", resp.Error)
		}
		return resp.Signature, nil
	case <-time.After(w.timeout):
		return nil, fmt.Errorf("sign_tx request timed out")
	case <-ctx.Done():
		return nil, ctx.Err()
	}
}

// SignMessage sends a message signing request to the companion.
func (w *RPCWallet) SignMessage(ctx context.Context, message []byte) ([]byte, error) {
	if w.sender == nil {
		return nil, fmt.Errorf("RPC wallet: no sender configured")
	}

	reqID := uuid.New().String()
	ch := make(chan SignMsgResponse, 1)

	w.mu.Lock()
	w.pendingSignMsg[reqID] = ch
	w.mu.Unlock()

	defer func() {
		w.mu.Lock()
		delete(w.pendingSignMsg, reqID)
		w.mu.Unlock()
	}()

	if err := w.sender("wallet.sign_msg.request", SignMsgRequest{
		RequestID: reqID,
		Message:   message,
	}); err != nil {
		return nil, fmt.Errorf("send sign_msg request: %w", err)
	}

	select {
	case resp := <-ch:
		if resp.Error != "" {
			return nil, fmt.Errorf("companion sign error: %s", resp.Error)
		}
		return resp.Signature, nil
	case <-time.After(w.timeout):
		return nil, fmt.Errorf("sign_msg request timed out")
	case <-ctx.Done():
		return nil, ctx.Err()
	}
}

// HandleSignTxResponse dispatches a signing response from the companion.
func (w *RPCWallet) HandleSignTxResponse(resp SignTxResponse) {
	w.mu.Lock()
	ch, ok := w.pendingSignTx[resp.RequestID]
	w.mu.Unlock()

	if ok {
		ch <- resp
	}
}

// HandleSignMsgResponse dispatches a message signing response from the companion.
func (w *RPCWallet) HandleSignMsgResponse(resp SignMsgResponse) {
	w.mu.Lock()
	ch, ok := w.pendingSignMsg[resp.RequestID]
	w.mu.Unlock()

	if ok {
		ch <- resp
	}
}

// HandleAddressResponse dispatches an address response from the companion.
func (w *RPCWallet) HandleAddressResponse(resp AddressResponse) {
	w.mu.Lock()
	ch, ok := w.pendingAddr[resp.RequestID]
	w.mu.Unlock()

	if ok {
		ch <- resp
	}
}
