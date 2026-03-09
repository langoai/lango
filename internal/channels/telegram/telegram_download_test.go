package telegram

import (
	"fmt"
	"net/http"
	"net/http/httptest"
	"testing"

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
