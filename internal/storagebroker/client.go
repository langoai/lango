package storagebroker

import (
	"bufio"
	"context"
	"encoding/json"
	"fmt"
	"io"
	"os"
	"os/exec"
	"sync"
	"time"
)

// API is the broker client contract used by bootstrap and higher layers.
type API interface {
	Health(ctx context.Context) (HealthResult, error)
	OpenDB(ctx context.Context, req OpenDBRequest) (OpenDBResult, error)
	DBStatusSummary(ctx context.Context, req DBStatusSummaryRequest) (DBStatusSummaryResult, error)
	EncryptPayload(ctx context.Context, plaintext []byte) (EncryptPayloadResult, error)
	DecryptPayload(ctx context.Context, ciphertext, nonce []byte, keyVersion int) (DecryptPayloadResult, error)
	LoadSecurityState(ctx context.Context) (LoadSecurityStateResult, error)
	StoreSalt(ctx context.Context, salt []byte) error
	StoreChecksum(ctx context.Context, checksum []byte) error
	Close(ctx context.Context) error
}

// Client manages a long-lived storage broker child process.
type Client struct {
	cmd    *exec.Cmd
	stdin  io.WriteCloser
	stdout io.ReadCloser

	mu      sync.Mutex
	nextID  uint64
	closed  bool
	pending map[uint64]chan Response
}

// Start launches the current executable in broker mode.
func Start(ctx context.Context) (*Client, error) {
	selfPath, err := os.Executable()
	if err != nil {
		return nil, fmt.Errorf("resolve broker executable: %w", err)
	}

	cmd := exec.CommandContext(ctx, selfPath, brokerFlag)
	cmd.ExtraFiles = nil
	stdin, err := cmd.StdinPipe()
	if err != nil {
		return nil, fmt.Errorf("broker stdin pipe: %w", err)
	}
	stdout, err := cmd.StdoutPipe()
	if err != nil {
		return nil, fmt.Errorf("broker stdout pipe: %w", err)
	}
	markFilesCloseOnExec(stdin, stdout)
	cmd.Stderr = os.Stderr

	if err := cmd.Start(); err != nil {
		return nil, fmt.Errorf("start storage broker: %w", err)
	}

	c := &Client{
		cmd:     cmd,
		stdin:   stdin,
		stdout:  stdout,
		pending: make(map[uint64]chan Response),
	}
	go c.readLoop()
	return c, nil
}

func (c *Client) Health(ctx context.Context) (HealthResult, error) {
	var result HealthResult
	if err := c.call(ctx, methodHealth, nil, &result); err != nil {
		return HealthResult{}, err
	}
	return result, nil
}

func (c *Client) OpenDB(ctx context.Context, req OpenDBRequest) (OpenDBResult, error) {
	var result OpenDBResult
	if err := c.call(ctx, methodOpenDB, req, &result); err != nil {
		return OpenDBResult{}, err
	}
	return result, nil
}

func (c *Client) DBStatusSummary(ctx context.Context, req DBStatusSummaryRequest) (DBStatusSummaryResult, error) {
	var result DBStatusSummaryResult
	if err := c.call(ctx, methodDBStatus, req, &result); err != nil {
		return DBStatusSummaryResult{}, err
	}
	return result, nil
}

func (c *Client) EncryptPayload(ctx context.Context, plaintext []byte) (EncryptPayloadResult, error) {
	var result EncryptPayloadResult
	if err := c.call(ctx, methodEncryptPayload, EncryptPayloadRequest{Plaintext: plaintext}, &result); err != nil {
		return EncryptPayloadResult{}, err
	}
	return result, nil
}

func (c *Client) DecryptPayload(ctx context.Context, ciphertext, nonce []byte, keyVersion int) (DecryptPayloadResult, error) {
	var result DecryptPayloadResult
	if err := c.call(ctx, methodDecryptPayload, DecryptPayloadRequest{
		Ciphertext: ciphertext,
		Nonce:      nonce,
		KeyVersion: keyVersion,
	}, &result); err != nil {
		return DecryptPayloadResult{}, err
	}
	return result, nil
}

func (c *Client) LoadSecurityState(ctx context.Context) (LoadSecurityStateResult, error) {
	var result LoadSecurityStateResult
	if err := c.call(ctx, methodLoadSecurityState, LoadSecurityStateRequest{}, &result); err != nil {
		return LoadSecurityStateResult{}, err
	}
	return result, nil
}

func (c *Client) StoreSalt(ctx context.Context, salt []byte) error {
	return c.call(ctx, methodStoreSalt, StoreSaltRequest{Salt: salt}, nil)
}

func (c *Client) StoreChecksum(ctx context.Context, checksum []byte) error {
	return c.call(ctx, methodStoreChecksum, StoreChecksumRequest{Checksum: checksum}, nil)
}

func (c *Client) Close(ctx context.Context) error {
	c.mu.Lock()
	if c.closed {
		c.mu.Unlock()
		return nil
	}
	c.closed = true
	c.mu.Unlock()

	var result ShutdownResult
	_ = c.call(ctx, methodShutdown, nil, &result)
	if err := c.stdin.Close(); err != nil {
		return err
	}
	return c.cmd.Wait()
}

func (c *Client) call(ctx context.Context, method string, payload interface{}, out interface{}) error {
	id, ch, err := c.reserve()
	if err != nil {
		return err
	}

	req := Request{ID: id, Method: method}
	if deadline, ok := ctx.Deadline(); ok {
		req.DeadlineMS = max(int64(deadline.Sub(now()).Milliseconds()), 1)
	}
	if payload != nil {
		raw, err := json.Marshal(payload)
		if err != nil {
			c.release(id)
			return fmt.Errorf("marshal broker payload: %w", err)
		}
		req.Payload = raw
	}

	if err := json.NewEncoder(c.stdin).Encode(req); err != nil {
		c.release(id)
		return fmt.Errorf("write broker request: %w", err)
	}

	select {
	case resp, ok := <-ch:
		if !ok {
			return fmt.Errorf("broker response channel closed")
		}
		if !resp.OK {
			return fmt.Errorf("%s", resp.Error)
		}
		if out != nil && len(resp.Result) > 0 {
			if err := json.Unmarshal(resp.Result, out); err != nil {
				return fmt.Errorf("decode broker result: %w", err)
			}
		}
		return nil
	case <-ctx.Done():
		c.release(id)
		return ctx.Err()
	}
}

func (c *Client) reserve() (uint64, chan Response, error) {
	c.mu.Lock()
	defer c.mu.Unlock()
	if c.closed {
		return 0, nil, fmt.Errorf("broker client closed")
	}
	c.nextID++
	id := c.nextID
	ch := make(chan Response, 1)
	c.pending[id] = ch
	return id, ch, nil
}

func (c *Client) release(id uint64) {
	c.mu.Lock()
	defer c.mu.Unlock()
	if ch, ok := c.pending[id]; ok {
		delete(c.pending, id)
		close(ch)
	}
}

func (c *Client) readLoop() {
	scanner := bufio.NewScanner(c.stdout)
	for scanner.Scan() {
		var resp Response
		if err := json.Unmarshal(scanner.Bytes(), &resp); err != nil {
			continue
		}
		c.mu.Lock()
		ch, ok := c.pending[resp.ID]
		if ok {
			delete(c.pending, resp.ID)
		}
		c.mu.Unlock()
		if ok {
			ch <- resp
			close(ch)
		}
	}

	c.mu.Lock()
	defer c.mu.Unlock()
	for id, ch := range c.pending {
		delete(c.pending, id)
		close(ch)
	}
}

func now() time.Time {
	return time.Now()
}

func max(a, b int64) int64 {
	if a > b {
		return a
	}
	return b
}
