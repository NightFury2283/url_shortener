package save

import (
	"bytes"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"
	"url-shortener/internal/lib/api/response"
	"url-shortener/internal/storage"

	"log/slog"
	"os"
)

type MockURLSaver struct {
	SaveURLFunc       func(urlToSave string, alias string) (int64, error)
	GetURLFunc        func(alias string) (string, error)
	GetAliasByURLFunc func(url string) (string, error)
}

func (m *MockURLSaver) SaveURL(urlToSave string, alias string) (int64, error) {
	return m.SaveURLFunc(urlToSave, alias)
}

func (m *MockURLSaver) GetURL(alias string) (string, error) {
	if m.GetURLFunc != nil {
		return m.GetURLFunc(alias)
	}
	return "", nil
}

func (m *MockURLSaver) GetAliasByURL(url string) (string, error) {
	if m.GetAliasByURLFunc != nil {
		return m.GetAliasByURLFunc(url)
	}
	return "", nil
}

func TestSaveHandler_SuccessWithAlias(t *testing.T) {
	mockSaver := &MockURLSaver{
		SaveURLFunc: func(urlToSave string, alias string) (int64, error) {
			if urlToSave != "https://example.com" {
				t.Errorf("expected url=https://example.com, got %s", urlToSave)
			}
			if alias != "myalias" {
				t.Errorf("expected alias=myalias, got %s", alias)
			}
			return 1, nil
		},
	}

	logger := slog.New(slog.NewTextHandler(os.Stdout, nil))
	handler := New(logger, mockSaver)

	reqBody := storage.Request{
		URL:   "https://example.com",
		Alias: "myalias",
	}

	body, _ := json.Marshal(reqBody)
	req := httptest.NewRequest(http.MethodPost, "/save", bytes.NewReader(body))
	w := httptest.NewRecorder()

	handler.ServeHTTP(w, req)

	if w.Code != http.StatusOK {
		t.Errorf("expected status 200, got %d", w.Code)
	}

	var resp storage.Response
	err := json.Unmarshal(w.Body.Bytes(), &resp)
	if err != nil {
		t.Fatalf("failed to unmarshal response: %v", err)
	}

	if resp.Status != response.StatusOK {
		t.Errorf("expected status OK, got %s", resp.Status)
	}

	if resp.Alias != "myalias" {
		t.Errorf("expected alias=myalias, got %s", resp.Alias)
	}
}

func TestSaveHandler_SuccessWithRandomAlias(t *testing.T) {
	var capturedAlias string

	mockSaver := &MockURLSaver{
		SaveURLFunc: func(urlToSave string, alias string) (int64, error) {
			capturedAlias = alias
			if len(alias) != aliasLength {
				t.Errorf("expected alias length=%d, got %d", aliasLength, len(alias))
			}
			return 1, nil
		},
	}

	logger := slog.New(slog.NewTextHandler(os.Stdout, nil))
	handler := New(logger, mockSaver)

	reqBody := storage.Request{
		URL: "https://example.com",
		// Alias not provided, should be generated randomly
	}

	body, _ := json.Marshal(reqBody)
	req := httptest.NewRequest(http.MethodPost, "/save", bytes.NewReader(body))
	w := httptest.NewRecorder()

	handler.ServeHTTP(w, req)

	if w.Code != http.StatusOK {
		t.Errorf("expected status 200, got %d", w.Code)
	}

	var resp storage.Response
	json.Unmarshal(w.Body.Bytes(), &resp)

	if resp.Alias == "" {
		t.Error("expected non-empty alias in response")
	}

	if capturedAlias == "" {
		t.Error("expected non-empty alias passed to SaveURL")
	}
}

func TestSaveHandler_InvalidJSON(t *testing.T) {
	mockSaver := &MockURLSaver{
		SaveURLFunc: func(urlToSave string, alias string) (int64, error) {
			t.Error("SaveURL should not be called for invalid JSON")
			return 0, nil
		},
	}

	logger := slog.New(slog.NewTextHandler(os.Stdout, nil))
	handler := New(logger, mockSaver)

	req := httptest.NewRequest(http.MethodPost, "/save", bytes.NewReader([]byte("invalid json")))
	w := httptest.NewRecorder()

	handler.ServeHTTP(w, req)

	if w.Code != http.StatusOK {
		t.Errorf("expected status 200, got %d", w.Code)
	}

	var resp response.Response
	json.Unmarshal(w.Body.Bytes(), &resp)

	if resp.Status != response.StatusError {
		t.Errorf("expected status Error, got %s", resp.Status)
	}

	if resp.Error == "" {
		t.Error("expected error message")
	}
}

func TestSaveHandler_MissingURL(t *testing.T) {
	mockSaver := &MockURLSaver{
		SaveURLFunc: func(urlToSave string, alias string) (int64, error) {
			t.Error("SaveURL should not be called for invalid request")
			return 0, nil
		},
	}

	logger := slog.New(slog.NewTextHandler(os.Stdout, nil))
	handler := New(logger, mockSaver)

	reqBody := storage.Request{
		URL: "", // Missing URL
	}

	body, _ := json.Marshal(reqBody)
	req := httptest.NewRequest(http.MethodPost, "/save", bytes.NewReader(body))
	w := httptest.NewRecorder()

	handler.ServeHTTP(w, req)

	var resp response.Response
	json.Unmarshal(w.Body.Bytes(), &resp)

	if resp.Status != response.StatusError {
		t.Errorf("expected status Error, got %s", resp.Status)
	}
}

func TestSaveHandler_InvalidURL(t *testing.T) {
	mockSaver := &MockURLSaver{
		SaveURLFunc: func(urlToSave string, alias string) (int64, error) {
			t.Error("SaveURL should not be called for invalid URL")
			return 0, nil
		},
	}

	logger := slog.New(slog.NewTextHandler(os.Stdout, nil))
	handler := New(logger, mockSaver)

	reqBody := storage.Request{
		URL: "not a valid url", // Invalid URL
	}

	body, _ := json.Marshal(reqBody)
	req := httptest.NewRequest(http.MethodPost, "/save", bytes.NewReader(body))
	w := httptest.NewRecorder()

	handler.ServeHTTP(w, req)

	var resp response.Response
	json.Unmarshal(w.Body.Bytes(), &resp)

	if resp.Status != response.StatusError {
		t.Errorf("expected status Error, got %s", resp.Status)
	}

	if resp.Error == "" {
		t.Error("expected error message")
	}
}

func TestSaveHandler_URLAlreadyExists(t *testing.T) {
	mockSaver := &MockURLSaver{
		SaveURLFunc: func(urlToSave string, alias string) (int64, error) {
			return 0, storage.ErrURLExists
		},
	}

	logger := slog.New(slog.NewTextHandler(os.Stdout, nil))
	handler := New(logger, mockSaver)

	reqBody := storage.Request{
		URL:   "https://example.com",
		Alias: "existingalias",
	}

	body, _ := json.Marshal(reqBody)
	req := httptest.NewRequest(http.MethodPost, "/save", bytes.NewReader(body))
	w := httptest.NewRecorder()

	handler.ServeHTTP(w, req)

	if w.Code != http.StatusOK {
		t.Errorf("expected status 200, got %d", w.Code)
	}

	var resp response.Response
	json.Unmarshal(w.Body.Bytes(), &resp)

	if resp.Status != response.StatusError {
		t.Errorf("expected status Error, got %s", resp.Status)
	}

	if resp.Error == "" {
		t.Error("expected error message")
	}
}

func TestSaveHandler_StorageError(t *testing.T) {
	mockSaver := &MockURLSaver{
		SaveURLFunc: func(urlToSave string, alias string) (int64, error) {
			return 0, storage.ErrUrlNotFound // Any error other than ErrURLExists
		},
	}

	logger := slog.New(slog.NewTextHandler(os.Stdout, nil))
	handler := New(logger, mockSaver)

	reqBody := storage.Request{
		URL:   "https://example.com",
		Alias: "myalias",
	}

	body, _ := json.Marshal(reqBody)
	req := httptest.NewRequest(http.MethodPost, "/save", bytes.NewReader(body))
	w := httptest.NewRecorder()

	handler.ServeHTTP(w, req)

	var resp response.Response
	json.Unmarshal(w.Body.Bytes(), &resp)

	if resp.Status != response.StatusError {
		t.Errorf("expected status Error, got %s", resp.Status)
	}
}
