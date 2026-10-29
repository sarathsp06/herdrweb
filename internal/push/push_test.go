package push

import (
	"context"
	"crypto/ecdh"
	"crypto/rand"
	"encoding/base64"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"net/url"
	"strings"
	"testing"

	webpush "github.com/SherClockHolmes/webpush-go"
)

// TestNotifySendsASubjectApplePushAccepts proves the VAPID JWT the push
// service receives carries a `sub` Apple accepts: a single https: URL, or a
// mailto: on a real domain. Apple answers anything else (a doubled
// "mailto:mailto:", @localhost, reserved TLDs) with 403 BadJwtToken, so iOS
// never receives a notification while Chrome and Firefox work fine.
func TestNotifySendsASubjectApplePushAccepts(t *testing.T) {
	var auth string
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		auth = r.Header.Get("Authorization")
		w.WriteHeader(http.StatusCreated)
	}))
	defer srv.Close()

	m, err := New(t.TempDir())
	if err != nil {
		t.Fatal(err)
	}
	m.Add(fakeBrowserSubscription(t, srv.URL))
	if res := m.Notify(context.Background(), Notification{Title: "t"}); res.Sent != 1 {
		t.Fatalf("Notify result %+v, want 1 sent", res)
	}

	sub := jwtSubject(t, auth)
	u, err := url.Parse(sub)
	if err != nil {
		t.Fatalf("sub %q is not a URI: %v", sub, err)
	}
	var domain string
	switch u.Scheme {
	case "https":
		domain = u.Hostname()
	case "mailto":
		_, domain, _ = strings.Cut(u.Opaque, "@")
	default:
		t.Fatalf("sub %q: scheme must be https or mailto", sub)
	}
	if strings.Contains(u.Opaque, ":") || domain == "" || domain == "localhost" || !strings.Contains(domain, ".") {
		t.Fatalf("sub %q does not name a real domain", sub)
	}
	for _, reserved := range []string{".local", ".invalid", ".test", ".example", ".localhost"} {
		if strings.HasSuffix(domain, reserved) {
			t.Fatalf("sub %q uses reserved TLD %s", sub, reserved)
		}
	}
}

// fakeBrowserSubscription returns a subscription with real P-256 keys, as a
// browser would post it, pointed at endpoint.
func fakeBrowserSubscription(t *testing.T, endpoint string) webpush.Subscription {
	t.Helper()
	key, err := ecdh.P256().GenerateKey(rand.Reader)
	if err != nil {
		t.Fatal(err)
	}
	secret := make([]byte, 16)
	if _, err := rand.Read(secret); err != nil {
		t.Fatal(err)
	}
	enc := base64.RawURLEncoding
	return webpush.Subscription{
		Endpoint: endpoint,
		Keys:     webpush.Keys{P256dh: enc.EncodeToString(key.PublicKey().Bytes()), Auth: enc.EncodeToString(secret)},
	}
}

// jwtSubject extracts the `sub` claim from a "vapid t=<jwt>, k=<key>" header.
func jwtSubject(t *testing.T, header string) string {
	t.Helper()
	token, ok := strings.CutPrefix(header, "vapid t=")
	if !ok {
		t.Fatalf("Authorization %q is not a VAPID header", header)
	}
	token, _, _ = strings.Cut(token, ",")
	parts := strings.Split(token, ".")
	if len(parts) != 3 {
		t.Fatalf("malformed JWT %q", token)
	}
	payload, err := base64.RawURLEncoding.DecodeString(parts[1])
	if err != nil {
		t.Fatal(err)
	}
	var claims struct{ Sub string }
	if err := json.Unmarshal(payload, &claims); err != nil {
		t.Fatal(err)
	}
	return claims.Sub
}
