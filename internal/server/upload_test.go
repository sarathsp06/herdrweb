package server

import (
	"bytes"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"os"
	"path/filepath"
	"testing"
)

// pngBytes is the 8-byte PNG signature plus filler; http.DetectContentType only
// needs the signature to classify it as image/png.
var pngBytes = append([]byte("\x89PNG\r\n\x1a\n"), bytes.Repeat([]byte{0x00}, 16)...)

func TestHandleUploadWritesImage(t *testing.T) {
	h := &Hub{}
	rec := httptest.NewRecorder()
	h.handleUpload(rec, httptest.NewRequest(http.MethodPost, "/api/upload", bytes.NewReader(pngBytes)))

	if rec.Code != http.StatusOK {
		t.Fatalf("status = %d, want 200 (%s)", rec.Code, rec.Body.String())
	}
	var body struct {
		Path string `json:"path"`
	}
	if err := json.Unmarshal(rec.Body.Bytes(), &body); err != nil {
		t.Fatalf("decode: %v", err)
	}
	if body.Path == "" {
		t.Fatal("empty path in response")
	}
	t.Cleanup(func() { _ = os.Remove(body.Path) })

	if filepath.Ext(body.Path) != ".png" {
		t.Fatalf("extension = %q, want .png", filepath.Ext(body.Path))
	}
	got, err := os.ReadFile(body.Path)
	if err != nil {
		t.Fatalf("read written file: %v", err)
	}
	if !bytes.Equal(got, pngBytes) {
		t.Fatal("written bytes differ from upload")
	}
}

func TestHandleUploadRejectsNonImage(t *testing.T) {
	h := &Hub{}
	rec := httptest.NewRecorder()
	h.handleUpload(rec, httptest.NewRequest(http.MethodPost, "/api/upload", bytes.NewReader([]byte("this is plain text, not an image"))))
	if rec.Code != http.StatusUnsupportedMediaType {
		t.Fatalf("status = %d, want 415", rec.Code)
	}
}

func TestHandleUploadRejectsEmpty(t *testing.T) {
	h := &Hub{}
	rec := httptest.NewRecorder()
	h.handleUpload(rec, httptest.NewRequest(http.MethodPost, "/api/upload", bytes.NewReader(nil)))
	if rec.Code != http.StatusBadRequest {
		t.Fatalf("status = %d, want 400", rec.Code)
	}
}

func TestHandleUploadRejectsGet(t *testing.T) {
	h := &Hub{}
	rec := httptest.NewRecorder()
	h.handleUpload(rec, httptest.NewRequest(http.MethodGet, "/api/upload", nil))
	if rec.Code != http.StatusMethodNotAllowed {
		t.Fatalf("status = %d, want 405", rec.Code)
	}
}
