package security

import (
	"context"
	"fmt"
	"sync"
	"time"

	"github.com/google/uuid"
	"github.com/langowarny/lango/internal/logging"
	"github.com/langowarny/lango/internal/types"
)

var rpcLogger = logging.SubsystemSugar("security")

const rpcTimeout = 30 * time.Second

// SignRequest represents the payload sent to the signer provider.
type SignRequest struct {
	ID      string `json:"id"`
	KeyID   string `json:"keyId"`
	Payload []byte `json:"payload"`
}

// SignResponse represents the payload received from the signer provider.
type SignResponse struct {
	ID        string `json:"id"`
	Signature []byte `json:"signature"`
	Error     string `json:"error,omitempty"`
}

// EncryptRequest represents the payload for encryption.
type EncryptRequest struct {
	ID        string `json:"id"`
	KeyID     string `json:"keyId"`
	Plaintext []byte `json:"plaintext"`
}

// EncryptResponse represents the payload for encryption response.
type EncryptResponse struct {
	ID         string `json:"id"`
	Ciphertext []byte `json:"ciphertext"`
	Error      string `json:"error,omitempty"`
}

// DecryptRequest represents the payload for decryption.
type DecryptRequest struct {
	ID         string `json:"id"`
	KeyID      string `json:"keyId"`
	Ciphertext []byte `json:"ciphertext"`
}

// DecryptResponse represents the payload for decryption response.
type DecryptResponse struct {
	ID        string `json:"id"`
	Plaintext []byte `json:"plaintext"`
	Error     string `json:"error,omitempty"`
}

// RPCProvider implements CryptoProvider using an asynchronous RPC mechanism.
type RPCProvider struct {
	sender         types.RPCSenderFunc
	signPending    sync.Map // map[string]chan SignResponse
	encryptPending sync.Map // map[string]chan EncryptResponse
	decryptPending sync.Map // map[string]chan DecryptResponse
}

// NewRPCProvider creates a new RPCProvider.
func NewRPCProvider() *RPCProvider {
	return &RPCProvider{}
}

// SetSender configures the function used to send requests.
func (s *RPCProvider) SetSender(sender types.RPCSenderFunc) {
	s.sender = sender
}

// Sign implements the CryptoProvider interface.
func (s *RPCProvider) Sign(ctx context.Context, keyID string, payload []byte) ([]byte, error) {
	if s.sender == nil {
		return nil, fmt.Errorf("provider sender not configured")
	}

	reqID := uuid.New().String()
	respChan := make(chan SignResponse, 1)
	s.signPending.Store(reqID, respChan)
	defer s.signPending.Delete(reqID)

	req := SignRequest{
		ID:      reqID,
		KeyID:   keyID,
		Payload: payload,
	}

	rpcLogger.Infow("requesting signature", "reqId", reqID, "keyId", keyID)
	if err := s.sender("sign.request", req); err != nil {
		return nil, fmt.Errorf("send sign request: %w", err)
	}

	select {
	case resp := <-respChan:
		if resp.Error != "" {
			return nil, fmt.Errorf("remote signing error: %s", resp.Error)
		}
		return resp.Signature, nil
	case <-ctx.Done():
		return nil, ctx.Err()
	case <-time.After(rpcTimeout):
		return nil, fmt.Errorf("signing request timed out")
	}
}

// Encrypt implements the CryptoProvider interface.
func (s *RPCProvider) Encrypt(ctx context.Context, keyID string, plaintext []byte) ([]byte, error) {
	if s.sender == nil {
		return nil, fmt.Errorf("provider sender not configured")
	}

	reqID := uuid.New().String()
	respChan := make(chan EncryptResponse, 1)
	s.encryptPending.Store(reqID, respChan)
	defer s.encryptPending.Delete(reqID)

	req := EncryptRequest{
		ID:        reqID,
		KeyID:     keyID,
		Plaintext: plaintext,
	}

	rpcLogger.Infow("requesting encryption", "reqId", reqID, "keyId", keyID)
	if err := s.sender("encrypt.request", req); err != nil {
		return nil, fmt.Errorf("send encrypt request: %w", err)
	}

	select {
	case resp := <-respChan:
		if resp.Error != "" {
			return nil, fmt.Errorf("remote encryption error: %s", resp.Error)
		}
		return resp.Ciphertext, nil
	case <-ctx.Done():
		return nil, ctx.Err()
	case <-time.After(rpcTimeout):
		return nil, fmt.Errorf("encryption request timed out")
	}
}

// Decrypt implements the CryptoProvider interface.
func (s *RPCProvider) Decrypt(ctx context.Context, keyID string, ciphertext []byte) ([]byte, error) {
	if s.sender == nil {
		return nil, fmt.Errorf("provider sender not configured")
	}

	reqID := uuid.New().String()
	respChan := make(chan DecryptResponse, 1)
	s.decryptPending.Store(reqID, respChan)
	defer s.decryptPending.Delete(reqID)

	req := DecryptRequest{
		ID:         reqID,
		KeyID:      keyID,
		Ciphertext: ciphertext,
	}

	rpcLogger.Infow("requesting decryption", "reqId", reqID, "keyId", keyID)
	if err := s.sender("decrypt.request", req); err != nil {
		return nil, fmt.Errorf("send decrypt request: %w", err)
	}

	select {
	case resp := <-respChan:
		if resp.Error != "" {
			return nil, fmt.Errorf("remote decryption error: %s", resp.Error)
		}
		return resp.Plaintext, nil
	case <-ctx.Done():
		return nil, ctx.Err()
	case <-time.After(rpcTimeout):
		return nil, fmt.Errorf("decryption request timed out")
	}
}

// HandleSignResponse processes an incoming sign response.
func (s *RPCProvider) HandleSignResponse(resp SignResponse) error {
	ch, ok := s.signPending.Load(resp.ID)
	if !ok {
		rpcLogger.Warnw("received sign response for unknown request", "reqId", resp.ID)
		return nil
	}
	responseChan := ch.(chan SignResponse)
	select {
	case responseChan <- resp:
	default:
		rpcLogger.Warnw("sign response channel full", "reqId", resp.ID)
	}
	return nil
}

// HandleEncryptResponse processes an incoming encrypt response.
func (s *RPCProvider) HandleEncryptResponse(resp EncryptResponse) error {
	ch, ok := s.encryptPending.Load(resp.ID)
	if !ok {
		rpcLogger.Warnw("received encrypt response for unknown request", "reqId", resp.ID)
		return nil
	}
	responseChan := ch.(chan EncryptResponse)
	select {
	case responseChan <- resp:
	default:
		rpcLogger.Warnw("encrypt response channel full", "reqId", resp.ID)
	}
	return nil
}

// HandleDecryptResponse processes an incoming decrypt response.
func (s *RPCProvider) HandleDecryptResponse(resp DecryptResponse) error {
	ch, ok := s.decryptPending.Load(resp.ID)
	if !ok {
		rpcLogger.Warnw("received decrypt response for unknown request", "reqId", resp.ID)
		return nil
	}
	responseChan := ch.(chan DecryptResponse)
	select {
	case responseChan <- resp:
	default:
		rpcLogger.Warnw("decrypt response channel full", "reqId", resp.ID)
	}
	return nil
}
