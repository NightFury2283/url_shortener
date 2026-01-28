package save_test

import (
	"bytes"
	"encoding/json"
	"log/slog"
	"net/http"
	"net/http/httptest"
	"os"
	"testing"
	save "url-shortener/internal/http-server/handlers/url/save"
	"url-shortener/internal/storage"

	"github.com/stretchr/testify/mock"
	"github.com/stretchr/testify/require"
)

type MockURLSaver struct {
	mock.Mock
}

func (m *MockURLSaver) SaveURL(urlToSave string, alias string) (int64, error) {
	args := m.Called(urlToSave, alias)
	return args.Get(0).(int64), args.Error(1)
}

func (m *MockURLSaver) GetURL(alias string) (string, error) {
	args := m.Called(alias)
	return args.String(0), args.Error(1)
}

func (m *MockURLSaver) GetAliasByURL(url string) (string, error) {
	args := m.Called(url)
	return args.String(0), args.Error(1)
}

func TestSaveHandler(t *testing.T) {
	cases := []struct {
		name      string
		alias     string
		url       string
		respError string
		mockSetup func(m *MockURLSaver)
	}{
		{
			name:  "Success",
			alias: "test_alias",
			url:   "https://google.com",
			mockSetup: func(m *MockURLSaver) {
				m.On("GetAliasByURL", "https://google.com").Return("", storage.ErrUrlNotFound)
				m.On("GetURL", "test_alias").Return("", storage.ErrUrlNotFound)
				m.On("SaveURL", "https://google.com", "test_alias").Return(int64(1), nil)
			},
		},
		{
			name:  "URL Already Exists",
			alias: "new_alias",
			url:   "https://google.com",
			mockSetup: func(m *MockURLSaver) {
				m.On("GetAliasByURL", "https://google.com").Return("existing_alias", nil)
			},
		},
		{
			name:      "Alias Conflict (Save Error)",
			alias:     "test_alias",
			url:       "https://google.com",
			respError: "url with this alias already exists",
			mockSetup: func(m *MockURLSaver) {
				m.On("GetAliasByURL", "https://google.com").Return("", storage.ErrUrlNotFound)
				m.On("GetURL", "test_alias").Return("", storage.ErrUrlNotFound)
				m.On("SaveURL", "https://google.com", "test_alias").Return(int64(0), storage.ErrURLExists)
			},
		},
	}

	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			urlSaverMock := new(MockURLSaver)
			if tc.mockSetup != nil {
				tc.mockSetup(urlSaverMock)
			}

			logger := slog.New(slog.NewTextHandler(os.Stdout, nil))
			handler := save.New(logger, urlSaverMock)

			input := storage.Request{
				URL:   tc.url,
				Alias: tc.alias,
			}

			body, _ := json.Marshal(input)
			req, _ := http.NewRequest(http.MethodPost, "/save", bytes.NewReader(body))
			rr := httptest.NewRecorder()

			handler.ServeHTTP(rr, req)

			if tc.respError == "" {
				require.Equal(t, http.StatusOK, rr.Code)
			} else {
				require.Contains(t, rr.Body.String(), tc.respError)
			}
		})
	}
}
