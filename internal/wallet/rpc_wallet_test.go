package wallet

import (
	"context"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestNewRPCWallet_Defaults(t *testing.T) {
	w := NewRPCWallet()

	assert.NotNil(t, w.pendingSignTx)
	assert.NotNil(t, w.pendingSignMsg)
	assert.NotNil(t, w.pendingAddr)
	assert.Equal(t, 30*time.Second, w.timeout)
	assert.Nil(t, w.sender)
}

func TestRPCWallet_Address_NoSender(t *testing.T) {
	w := NewRPCWallet()
	ctx := context.Background()

	_, err := w.Address(ctx)
	require.Error(t, err)
	assert.Contains(t, err.Error(), "no sender configured")
}

func TestRPCWallet_Address_Success(t *testing.T) {
	w := NewRPCWallet()
	w.timeout = 2 * time.Second

	var capturedEvent string
	var capturedPayload interface{}
	w.SetSender(func(event string, payload interface{}) error {
		capturedEvent = event
		capturedPayload = payload
		// Simulate async companion response.
		go func() {
			req := capturedPayload.(AddressRequest)
			w.HandleAddressResponse(AddressResponse{
				RequestID: req.RequestID,
				Address:   "0xABCDEF",
			})
		}()
		return nil
	})

	ctx := context.Background()
	addr, err := w.Address(ctx)
	require.NoError(t, err)
	assert.Equal(t, "0xABCDEF", addr)
	assert.Equal(t, "wallet.address.request", capturedEvent)
}

func TestRPCWallet_Address_CompanionError(t *testing.T) {
	w := NewRPCWallet()
	w.timeout = 2 * time.Second

	w.SetSender(func(event string, payload interface{}) error {
		go func() {
			req := payload.(AddressRequest)
			w.HandleAddressResponse(AddressResponse{
				RequestID: req.RequestID,
				Error:     "wallet locked",
			})
		}()
		return nil
	})

	ctx := context.Background()
	_, err := w.Address(ctx)
	require.Error(t, err)
	assert.Contains(t, err.Error(), "companion address error")
	assert.Contains(t, err.Error(), "wallet locked")
}

func TestRPCWallet_Address_Timeout(t *testing.T) {
	w := NewRPCWallet()
	w.timeout = 50 * time.Millisecond

	w.SetSender(func(_ string, _ interface{}) error {
		// Do not send any response to trigger timeout.
		return nil
	})

	ctx := context.Background()
	_, err := w.Address(ctx)
	require.Error(t, err)
	assert.Contains(t, err.Error(), "timed out")
}

func TestRPCWallet_Address_ContextCanceled(t *testing.T) {
	w := NewRPCWallet()
	w.timeout = 5 * time.Second

	w.SetSender(func(_ string, _ interface{}) error {
		return nil
	})

	ctx, cancel := context.WithCancel(context.Background())
	cancel() // Cancel immediately.

	_, err := w.Address(ctx)
	require.Error(t, err)
	assert.ErrorIs(t, err, context.Canceled)
}

func TestRPCWallet_Balance_NotSupported(t *testing.T) {
	w := NewRPCWallet()
	ctx := context.Background()

	_, err := w.Balance(ctx)
	require.Error(t, err)
	assert.Contains(t, err.Error(), "not supported")
}

func TestRPCWallet_SignTransaction_NoSender(t *testing.T) {
	w := NewRPCWallet()
	ctx := context.Background()

	_, err := w.SignTransaction(ctx, []byte("raw-tx"))
	require.Error(t, err)
	assert.Contains(t, err.Error(), "no sender configured")
}

func TestRPCWallet_SignTransaction_Success(t *testing.T) {
	w := NewRPCWallet()
	w.timeout = 2 * time.Second

	expectedSig := []byte{0x01, 0x02, 0x03}
	w.SetSender(func(event string, payload interface{}) error {
		assert.Equal(t, "wallet.sign_tx.request", event)
		go func() {
			req := payload.(SignTxRequest)
			w.HandleSignTxResponse(SignTxResponse{
				RequestID: req.RequestID,
				Signature: expectedSig,
			})
		}()
		return nil
	})

	ctx := context.Background()
	sig, err := w.SignTransaction(ctx, []byte("raw-tx-data"))
	require.NoError(t, err)
	assert.Equal(t, expectedSig, sig)
}

func TestRPCWallet_SignTransaction_CompanionError(t *testing.T) {
	w := NewRPCWallet()
	w.timeout = 2 * time.Second

	w.SetSender(func(_ string, payload interface{}) error {
		go func() {
			req := payload.(SignTxRequest)
			w.HandleSignTxResponse(SignTxResponse{
				RequestID: req.RequestID,
				Error:     "user rejected",
			})
		}()
		return nil
	})

	ctx := context.Background()
	_, err := w.SignTransaction(ctx, []byte("raw-tx-data"))
	require.Error(t, err)
	assert.Contains(t, err.Error(), "companion sign error")
	assert.Contains(t, err.Error(), "user rejected")
}

func TestRPCWallet_SignTransaction_Timeout(t *testing.T) {
	w := NewRPCWallet()
	w.timeout = 50 * time.Millisecond

	w.SetSender(func(_ string, _ interface{}) error {
		return nil
	})

	ctx := context.Background()
	_, err := w.SignTransaction(ctx, []byte("raw-tx"))
	require.Error(t, err)
	assert.Contains(t, err.Error(), "timed out")
}

func TestRPCWallet_SignTransaction_SenderError(t *testing.T) {
	w := NewRPCWallet()
	w.timeout = 2 * time.Second

	w.SetSender(func(_ string, _ interface{}) error {
		return assert.AnError
	})

	ctx := context.Background()
	_, err := w.SignTransaction(ctx, []byte("raw-tx"))
	require.Error(t, err)
	assert.Contains(t, err.Error(), "send sign_tx request")
}

func TestRPCWallet_SignMessage_NoSender(t *testing.T) {
	w := NewRPCWallet()
	ctx := context.Background()

	_, err := w.SignMessage(ctx, []byte("msg"))
	require.Error(t, err)
	assert.Contains(t, err.Error(), "no sender configured")
}

func TestRPCWallet_SignMessage_Success(t *testing.T) {
	w := NewRPCWallet()
	w.timeout = 2 * time.Second

	expectedSig := []byte{0xAA, 0xBB, 0xCC}
	w.SetSender(func(event string, payload interface{}) error {
		assert.Equal(t, "wallet.sign_msg.request", event)
		go func() {
			req := payload.(SignMsgRequest)
			w.HandleSignMsgResponse(SignMsgResponse{
				RequestID: req.RequestID,
				Signature: expectedSig,
			})
		}()
		return nil
	})

	ctx := context.Background()
	sig, err := w.SignMessage(ctx, []byte("hello"))
	require.NoError(t, err)
	assert.Equal(t, expectedSig, sig)
}

func TestRPCWallet_SignMessage_CompanionError(t *testing.T) {
	w := NewRPCWallet()
	w.timeout = 2 * time.Second

	w.SetSender(func(_ string, payload interface{}) error {
		go func() {
			req := payload.(SignMsgRequest)
			w.HandleSignMsgResponse(SignMsgResponse{
				RequestID: req.RequestID,
				Error:     "invalid key",
			})
		}()
		return nil
	})

	ctx := context.Background()
	_, err := w.SignMessage(ctx, []byte("hello"))
	require.Error(t, err)
	assert.Contains(t, err.Error(), "companion sign error")
	assert.Contains(t, err.Error(), "invalid key")
}

func TestRPCWallet_SignMessage_Timeout(t *testing.T) {
	w := NewRPCWallet()
	w.timeout = 50 * time.Millisecond

	w.SetSender(func(_ string, _ interface{}) error {
		return nil
	})

	ctx := context.Background()
	_, err := w.SignMessage(ctx, []byte("hello"))
	require.Error(t, err)
	assert.Contains(t, err.Error(), "timed out")
}

func TestRPCWallet_SignMessage_ContextCanceled(t *testing.T) {
	w := NewRPCWallet()
	w.timeout = 5 * time.Second

	w.SetSender(func(_ string, _ interface{}) error {
		return nil
	})

	ctx, cancel := context.WithCancel(context.Background())
	cancel()

	_, err := w.SignMessage(ctx, []byte("hello"))
	require.Error(t, err)
	assert.ErrorIs(t, err, context.Canceled)
}

func TestRPCWallet_PublicKey_NotSupported(t *testing.T) {
	w := NewRPCWallet()
	ctx := context.Background()

	_, err := w.PublicKey(ctx)
	require.Error(t, err)
	assert.Contains(t, err.Error(), "not supported")
}

func TestRPCWallet_HandleSignTxResponse_UnknownRequestID(t *testing.T) {
	w := NewRPCWallet()

	// Should not panic when dispatching a response for an unknown request ID.
	w.HandleSignTxResponse(SignTxResponse{
		RequestID: "unknown-id",
		Signature: []byte("sig"),
	})
}

func TestRPCWallet_HandleSignMsgResponse_UnknownRequestID(t *testing.T) {
	w := NewRPCWallet()

	// Should not panic when dispatching a response for an unknown request ID.
	w.HandleSignMsgResponse(SignMsgResponse{
		RequestID: "unknown-id",
		Signature: []byte("sig"),
	})
}

func TestRPCWallet_HandleAddressResponse_UnknownRequestID(t *testing.T) {
	w := NewRPCWallet()

	// Should not panic when dispatching a response for an unknown request ID.
	w.HandleAddressResponse(AddressResponse{
		RequestID: "unknown-id",
		Address:   "0x1234",
	})
}

func TestRPCWallet_PendingCleanup_AfterResponse(t *testing.T) {
	w := NewRPCWallet()
	w.timeout = 2 * time.Second

	w.SetSender(func(_ string, payload interface{}) error {
		go func() {
			req := payload.(AddressRequest)
			w.HandleAddressResponse(AddressResponse{
				RequestID: req.RequestID,
				Address:   "0xABC",
			})
		}()
		return nil
	})

	ctx := context.Background()
	_, err := w.Address(ctx)
	require.NoError(t, err)

	// After the call, the pending map should be cleaned up.
	w.mu.Lock()
	assert.Empty(t, w.pendingAddr)
	w.mu.Unlock()
}

func TestRPCWallet_PendingCleanup_AfterTimeout(t *testing.T) {
	w := NewRPCWallet()
	w.timeout = 50 * time.Millisecond

	w.SetSender(func(_ string, _ interface{}) error {
		return nil
	})

	ctx := context.Background()
	_, _ = w.SignTransaction(ctx, []byte("raw-tx"))

	// After timeout, the pending map should be cleaned up.
	w.mu.Lock()
	assert.Empty(t, w.pendingSignTx)
	w.mu.Unlock()
}

func TestRPCWallet_SetSender(t *testing.T) {
	w := NewRPCWallet()
	assert.Nil(t, w.sender)

	called := false
	w.SetSender(func(_ string, _ interface{}) error {
		called = true
		return nil
	})

	assert.NotNil(t, w.sender)
	_ = w.sender("test", nil)
	assert.True(t, called)
}

func TestRPCWallet_SignTransaction_ContextCanceled(t *testing.T) {
	w := NewRPCWallet()
	w.timeout = 5 * time.Second

	w.SetSender(func(_ string, _ interface{}) error {
		return nil
	})

	ctx, cancel := context.WithCancel(context.Background())
	cancel()

	_, err := w.SignTransaction(ctx, []byte("tx"))
	require.Error(t, err)
	assert.ErrorIs(t, err, context.Canceled)
}

func TestRPCWallet_SignMessage_SenderError(t *testing.T) {
	w := NewRPCWallet()
	w.timeout = 2 * time.Second

	w.SetSender(func(_ string, _ interface{}) error {
		return assert.AnError
	})

	ctx := context.Background()
	_, err := w.SignMessage(ctx, []byte("msg"))
	require.Error(t, err)
	assert.Contains(t, err.Error(), "send sign_msg request")
}

func TestRPCWallet_Address_SenderError(t *testing.T) {
	w := NewRPCWallet()
	w.timeout = 2 * time.Second

	w.SetSender(func(_ string, _ interface{}) error {
		return assert.AnError
	})

	ctx := context.Background()
	_, err := w.Address(ctx)
	require.Error(t, err)
	assert.Contains(t, err.Error(), "send address request")
}
