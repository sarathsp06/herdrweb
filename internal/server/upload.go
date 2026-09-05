package server

import (
	"crypto/rand"
	"encoding/hex"
	"fmt"
	"io"
	"net/http"
	"os"
	"path/filepath"
	"time"
)

// maxUploadBytes caps a single image upload. A phone screenshot or photo is well
// under this; the limit just bounds per-request memory and disk.
const maxUploadBytes = 25 << 20 // 25 MiB

// uploadExt maps a sniffed image content type to a file extension. Only these
// raster formats are accepted; the request body is sniffed, not trusted.
var uploadExt = map[string]string{
	"image/png":  ".png",
	"image/jpeg": ".jpg",
	"image/gif":  ".gif",
	"image/webp": ".webp",
}

// uploadDir is where uploaded images land. The bridge shares a host with the
// agents it drives, so an absolute path here is readable by the agent the
// operator references it to.
func uploadDir() string {
	return filepath.Join(os.TempDir(), "herdr-uploads")
}

// handleUpload accepts a raw image body, writes it to a host-local file, and
// returns its absolute path for the composer to drop into a prompt. Coding
// agents read images by path, so this is how a pasted screenshot or a phone
// photo reaches the agent — there is no image-input RPC to proxy.
func (h *Hub) handleUpload(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		http.Error(w, "method not allowed", http.StatusMethodNotAllowed)
		return
	}
	r.Body = http.MaxBytesReader(w, r.Body, maxUploadBytes)
	data, err := io.ReadAll(r.Body)
	if err != nil {
		http.Error(w, "upload too large", http.StatusRequestEntityTooLarge)
		return
	}
	if len(data) == 0 {
		http.Error(w, "empty upload", http.StatusBadRequest)
		return
	}
	ext, ok := uploadExt[http.DetectContentType(data)]
	if !ok {
		http.Error(w, "unsupported image type", http.StatusUnsupportedMediaType)
		return
	}
	dir := uploadDir()
	if err := os.MkdirAll(dir, 0o700); err != nil {
		http.Error(w, "cannot create upload dir", http.StatusInternalServerError)
		return
	}
	pruneUploads(dir)
	path := filepath.Join(dir, fmt.Sprintf("%d-%s%s", time.Now().UnixNano(), randHex(4), ext))
	if err := os.WriteFile(path, data, 0o600); err != nil {
		http.Error(w, "cannot write upload", http.StatusInternalServerError)
		return
	}
	writeJSON(w, map[string]any{"path": path})
}

// pruneUploads best-effort removes upload files older than a day so the temp
// dir never grows without bound. Errors are ignored; this is housekeeping.
func pruneUploads(dir string) {
	entries, err := os.ReadDir(dir)
	if err != nil {
		return
	}
	cutoff := time.Now().Add(-24 * time.Hour)
	for _, e := range entries {
		if e.IsDir() {
			continue
		}
		info, err := e.Info()
		if err != nil || info.ModTime().After(cutoff) {
			continue
		}
		_ = os.Remove(filepath.Join(dir, e.Name()))
	}
}

func randHex(n int) string {
	b := make([]byte, n)
	if _, err := rand.Read(b); err != nil {
		return "0000"
	}
	return hex.EncodeToString(b)
}
