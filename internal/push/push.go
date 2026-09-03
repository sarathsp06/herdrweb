// Package push delivers Web Push notifications (VAPID) to subscribed browsers so
// the operator is alerted when an agent needs them even while the tab is closed.
//
// The manager owns a persistent VAPID key pair and the set of browser push
// subscriptions, both stored next to the Herdr config. The bridge exposes the
// VAPID public key to the UI, records subscriptions the browser posts back, and
// fans a notification to every subscription when a pane transitions to blocked.
package push

import (
	"context"
	"encoding/json"
	"log"
	"net/http"
	"os"
	"path/filepath"
	"sync"

	webpush "github.com/SherClockHolmes/webpush-go"
)

// subscriber is the VAPID `sub` claim. Push services require a contact URI but
// do not validate it for a loopback/tailnet single-operator deployment.
const subscriber = "mailto:herdr-web@localhost"

// Notification is the payload delivered to the service worker.
type Notification struct {
	Title string `json:"title"`
	Body  string `json:"body"`
	URL   string `json:"url"`
}

// Manager holds the VAPID keys and current subscriptions.
type Manager struct {
	keyPath string
	subPath string

	pub  string
	priv string

	mu   sync.Mutex
	subs []webpush.Subscription
}

type vapidKeys struct {
	Public  string `json:"public"`
	Private string `json:"private"`
}

// New loads (or generates) the VAPID key pair and loads persisted subscriptions
// from dir. It never fails on a missing subscriptions file.
func New(dir string) (*Manager, error) {
	m := &Manager{
		keyPath: filepath.Join(dir, "webpush.json"),
		subPath: filepath.Join(dir, "push-subs.json"),
	}
	if err := m.loadOrCreateKeys(); err != nil {
		return nil, err
	}
	m.loadSubs()
	return m, nil
}

func (m *Manager) loadOrCreateKeys() error {
	if data, err := os.ReadFile(m.keyPath); err == nil {
		var k vapidKeys
		if json.Unmarshal(data, &k) == nil && k.Public != "" && k.Private != "" {
			m.pub, m.priv = k.Public, k.Private
			return nil
		}
	}
	priv, pub, err := webpush.GenerateVAPIDKeys()
	if err != nil {
		return err
	}
	m.pub, m.priv = pub, priv
	return writeJSON(m.keyPath, vapidKeys{Public: pub, Private: priv})
}

func (m *Manager) loadSubs() {
	data, err := os.ReadFile(m.subPath)
	if err != nil {
		return
	}
	var subs []webpush.Subscription
	if json.Unmarshal(data, &subs) == nil {
		m.subs = subs
	}
}

// PublicKey returns the base64url VAPID public key the browser subscribes with.
func (m *Manager) PublicKey() string { return m.pub }

// Add records a browser subscription, replacing any with the same endpoint, and
// persists the set.
func (m *Manager) Add(sub webpush.Subscription) {
	if sub.Endpoint == "" {
		return
	}
	m.mu.Lock()
	next := m.subs[:0:0]
	for _, s := range m.subs {
		if s.Endpoint != sub.Endpoint {
			next = append(next, s)
		}
	}
	m.subs = append(next, sub)
	snapshot := append([]webpush.Subscription(nil), m.subs...)
	m.mu.Unlock()
	if err := writeJSON(m.subPath, snapshot); err != nil {
		log.Printf("push: persist subscriptions: %v", err)
	}
}

// Notify fans n to every subscription, pruning those the push service reports
// as gone (404/410). It is a no-op when there are no subscriptions.
func (m *Manager) Notify(ctx context.Context, n Notification) {
	m.mu.Lock()
	subs := append([]webpush.Subscription(nil), m.subs...)
	m.mu.Unlock()
	if len(subs) == 0 {
		return
	}
	payload, err := json.Marshal(n)
	if err != nil {
		return
	}
	var dead []string
	for i := range subs {
		sub := subs[i]
		resp, err := webpush.SendNotificationWithContext(ctx, payload, &sub, &webpush.Options{
			Subscriber:      subscriber,
			VAPIDPublicKey:  m.pub,
			VAPIDPrivateKey: m.priv,
			TTL:             60,
		})
		if err != nil {
			log.Printf("push: send: %v", err)
			continue
		}
		if resp.StatusCode == http.StatusNotFound || resp.StatusCode == http.StatusGone {
			dead = append(dead, sub.Endpoint)
		}
		resp.Body.Close()
	}
	if len(dead) > 0 {
		m.prune(dead)
	}
}

func (m *Manager) prune(endpoints []string) {
	gone := make(map[string]struct{}, len(endpoints))
	for _, e := range endpoints {
		gone[e] = struct{}{}
	}
	m.mu.Lock()
	next := m.subs[:0:0]
	for _, s := range m.subs {
		if _, ok := gone[s.Endpoint]; !ok {
			next = append(next, s)
		}
	}
	m.subs = next
	snapshot := append([]webpush.Subscription(nil), m.subs...)
	m.mu.Unlock()
	if err := writeJSON(m.subPath, snapshot); err != nil {
		log.Printf("push: persist subscriptions: %v", err)
	}
}

func writeJSON(path string, v any) error {
	data, err := json.MarshalIndent(v, "", "  ")
	if err != nil {
		return err
	}
	if err := os.MkdirAll(filepath.Dir(path), 0o755); err != nil {
		return err
	}
	tmp := path + ".tmp"
	if err := os.WriteFile(tmp, data, 0o600); err != nil {
		return err
	}
	return os.Rename(tmp, path)
}

// Subscription is re-exported so callers decode the browser payload without
// importing the vendored package directly.
type Subscription = webpush.Subscription
