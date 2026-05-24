package telegram

import (
	"errors"
	"fmt"
	"io"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
	"time"

	tgbotapi "github.com/go-telegram-bot-api/telegram-bot-api/v5"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// downloadMockBot extends MockBotAPI with a configurable GetFile.
type downloadMockBot struct {
	MockBotAPI
	GetFileFunc func(config tgbotapi.FileConfig) (tgbotapi.File, error)
}

func (m *downloadMockBot) GetFile(config tgbotapi.FileConfig) (tgbotapi.File, error) {
	if m.GetFileFunc != nil {
		return m.GetFileFunc(config)
	}
	return tgbotapi.File{}, nil
}

// redirectTransport rewrites every request URL to target the given test server.
type redirectTransport struct {
	targetURL string
}

func (t *redirectTransport) RoundTrip(req *http.Request) (*http.Response, error) {
	// Rewrite the URL to point to our test server, preserving the path.
	req.URL.Scheme = "http"
	req.URL.Host = t.targetURL
	return http.DefaultTransport.RoundTrip(req)
}

type inspectTransport struct {
	t *testing.T
}

func (t *inspectTransport) RoundTrip(req *http.Request) (*http.Response, error) {
	t.t.Helper()

	assert.Equal(t.t, http.MethodGet, req.Method)
	assert.Contains(t.t, req.URL.Path, "documents/test-file.pdf")

	deadline, ok := req.Context().Deadline()
	if !ok {
		t.t.Fatal("download request missing deadline")
	}
	remaining := time.Until(deadline)
	assert.LessOrEqual(t.t, remaining, downloadTimeout)
	assert.Greater(t.t, remaining, 25*time.Second)

	return &http.Response{
		StatusCode: http.StatusOK,
		Body:       io.NopCloser(strings.NewReader("image-bytes")),
		Header:     make(http.Header),
	}, nil
}

type errReadCloser struct {
	err error
}

func (e *errReadCloser) Read(_ []byte) (int, error) {
	return 0, e.err
}

func (e *errReadCloser) Close() error { return nil }

func TestDownloadFile(t *testing.T) {
	t.Parallel()

	tests := []struct {
		give        string
		giveFileID  string
		giveHandler http.HandlerFunc
		giveGetFile func(config tgbotapi.FileConfig) (tgbotapi.File, error)
		wantData    []byte
		wantErr     string
	}{
		{
			give:       "success",
			giveFileID: "file-123",
			giveHandler: http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
				w.WriteHeader(http.StatusOK)
				_, _ = w.Write([]byte("image-bytes"))
			}),
			wantData: []byte("image-bytes"),
		},
		{
			give:       "HTTP error",
			giveFileID: "file-404",
			giveHandler: http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
				w.WriteHeader(http.StatusNotFound)
			}),
			wantErr: "download file: HTTP 404",
		},
		{
			give:       "empty body",
			giveFileID: "file-empty",
			giveHandler: http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
				w.WriteHeader(http.StatusOK)
				// write nothing
			}),
			wantErr: "download file: empty response body",
		},
		{
			give:       "read body error",
			giveFileID: "file-read-error",
			giveHandler: http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
				w.WriteHeader(http.StatusOK)
				hj, ok := w.(http.Hijacker)
				if !ok {
					t.Fatal("test server does not support hijacking")
				}
				conn, buf, err := hj.Hijack()
				require.NoError(t, err)
				defer conn.Close()
				_, err = buf.WriteString("HTTP/1.1 200 OK\r\nContent-Length: 4\r\n\r\n")
				require.NoError(t, err)
				require.NoError(t, buf.Flush())
			}),
			wantErr: "read file body",
		},
		{
			give:       "GetFile API error",
			giveFileID: "file-bad",
			giveGetFile: func(config tgbotapi.FileConfig) (tgbotapi.File, error) {
				return tgbotapi.File{}, fmt.Errorf("telegram: file not found")
			},
			wantErr: "get file: telegram: file not found",
		},
	}

	for _, tt := range tests {
		t.Run(tt.give, func(t *testing.T) {
			t.Parallel()

			var srv *httptest.Server
			if tt.giveHandler != nil {
				srv = httptest.NewServer(tt.giveHandler)
				defer srv.Close()
			} else {
				// dummy server that won't be called
				srv = httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
					w.WriteHeader(http.StatusInternalServerError)
				}))
				defer srv.Close()
			}

			serverHost := srv.Listener.Addr().String()

			mockBot := &downloadMockBot{}
			if tt.giveGetFile != nil {
				mockBot.GetFileFunc = tt.giveGetFile
			} else {
				mockBot.GetFileFunc = func(config tgbotapi.FileConfig) (tgbotapi.File, error) {
					return tgbotapi.File{
						FileID:   config.FileID,
						FilePath: "documents/test-file.pdf",
					}, nil
				}
			}

			ch := &Channel{
				config: Config{
					BotToken: "TEST_TOKEN",
					HTTPClient: &http.Client{
						Transport: &redirectTransport{targetURL: serverHost},
					},
				},
				bot:      mockBot,
				stopChan: make(chan struct{}),
			}

			data, err := ch.DownloadFile(tt.giveFileID)

			if tt.wantErr != "" {
				require.Error(t, err)
				assert.Contains(t, err.Error(), tt.wantErr)
				assert.Nil(t, data)
			} else {
				require.NoError(t, err)
				assert.Equal(t, tt.wantData, data)
			}
		})
	}
}

func TestDownloadFile_PreservesReadBodyCause(t *testing.T) {
	t.Parallel()

	mockBot := &downloadMockBot{
		GetFileFunc: func(config tgbotapi.FileConfig) (tgbotapi.File, error) {
			return tgbotapi.File{
				FileID:   config.FileID,
				FilePath: "documents/test-file.pdf",
			}, nil
		},
	}

	readErr := errors.New("stream interrupted")
	ch := &Channel{
		config: Config{
			BotToken: "TEST_TOKEN",
			HTTPClient: &http.Client{
				Transport: roundTripFunc(func(req *http.Request) (*http.Response, error) {
					return &http.Response{
						StatusCode: http.StatusOK,
						Body:       &errReadCloser{err: readErr},
						Header:     make(http.Header),
					}, nil
				}),
			},
		},
		bot:      mockBot,
		stopChan: make(chan struct{}),
	}

	data, err := ch.DownloadFile("file-read-cause")
	require.Error(t, err)
	assert.Nil(t, data)
	assert.Contains(t, err.Error(), "read file body")
	assert.Contains(t, err.Error(), "stream interrupted")
	assert.NotContains(t, err.Error(), "empty response body")
}

type roundTripFunc func(*http.Request) (*http.Response, error)

func (f roundTripFunc) RoundTrip(req *http.Request) (*http.Response, error) {
	return f(req)
}

func TestDownloadFile_UsesGETAndTimeoutContext(t *testing.T) {
	t.Parallel()

	mockBot := &downloadMockBot{
		GetFileFunc: func(config tgbotapi.FileConfig) (tgbotapi.File, error) {
			return tgbotapi.File{
				FileID:   config.FileID,
				FilePath: "documents/test-file.pdf",
			}, nil
		},
	}

	ch := &Channel{
		config: Config{
			BotToken: "TEST_TOKEN",
			HTTPClient: &http.Client{
				Transport: &inspectTransport{t: t},
			},
		},
		bot:      mockBot,
		stopChan: make(chan struct{}),
	}

	data, err := ch.DownloadFile("file-ctx")
	require.NoError(t, err)
	assert.Equal(t, []byte("image-bytes"), data)
}
