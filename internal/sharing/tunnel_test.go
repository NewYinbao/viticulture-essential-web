package sharing

import (
	"context"
	"errors"
	"fmt"
	"io"
	"net/http"
	"os"
	"path/filepath"
	"strings"
	"testing"
	"time"
)

// The test executable doubles as a fake cloudflared; no network or installed binary.
func TestMain(m *testing.M) {
	if len(os.Args) > 1 && os.Args[1] == "tunnel" {
		switch os.Getenv("VITICULTURE_FAKE_TUNNEL") {
		case "exit":
			os.Exit(7)
		case "url":
			fmt.Println("https://unit-test.trycloudflare.com")
		}
		for {
			time.Sleep(time.Hour)
		}
	}
	os.Exit(m.Run())
}

type fakeTransport func(*http.Request) (*http.Response, error)

func (f fakeTransport) RoundTrip(r *http.Request) (*http.Response, error) { return f(r) }
func emptyScratch(t *testing.T, dir string) {
	t.Helper()
	entries, e := os.ReadDir(dir)
	if e != nil || len(entries) != 0 {
		t.Fatalf("scratch not cleaned: %v %v", entries, e)
	}
}
func TestTunnelMissingAndStartFailure(t *testing.T) {
	scratch := t.TempDir()
	if _, _, e := Start(context.Background(), filepath.Join(scratch, "missing"), "http://127.0.0.1:1", "key", scratch); e == nil || !strings.Contains(e.Error(), "cloudflared missing") {
		t.Fatalf("missing: %v", e)
	}
	// A directory exists but cannot execute; this also checks config cleanup.
	exe := t.TempDir()
	if _, _, e := Start(context.Background(), exe, "http://127.0.0.1:1", "key", scratch); e == nil {
		t.Fatal("directory executed")
	}
	emptyScratch(t, scratch)
}
func TestTunnelExitAndCancellation(t *testing.T) {
	exe, e := os.Executable()
	if e != nil {
		t.Fatal(e)
	}
	for _, mode := range []string{"exit", "wait"} {
		t.Run(mode, func(t *testing.T) {
			t.Setenv("VITICULTURE_FAKE_TUNNEL", mode)
			scratch := t.TempDir()
			ctx, cancel := context.WithTimeout(context.Background(), 300*time.Millisecond)
			defer cancel()
			tunnel, url, e := Start(ctx, exe, "http://127.0.0.1:1", "key", scratch)
			if e == nil || tunnel != nil || url != "" {
				t.Fatalf("unexpected success: %v", e)
			}
			if mode == "wait" && !errors.Is(e, context.DeadlineExceeded) {
				t.Fatal(e)
			}
			if mode == "exit" && !strings.Contains(e.Error(), "exited") {
				t.Fatal(e)
			}
			emptyScratch(t, scratch)
		})
	}
}
func TestTunnelReadinessAndIdempotentStop(t *testing.T) {
	t.Setenv("VITICULTURE_FAKE_TUNNEL", "url")
	exe, e := os.Executable()
	if e != nil {
		t.Fatal(e)
	}
	old := http.DefaultTransport
	defer func() { http.DefaultTransport = old }()
	http.DefaultTransport = fakeTransport(func(r *http.Request) (*http.Response, error) {
		c, e := r.Cookie("viticulture_gate")
		if e != nil || c.Value != "test-key" || r.URL.String() != "https://unit-test.trycloudflare.com/api/transport" {
			t.Errorf("bad readiness request: %v", r)
		}
		return &http.Response{StatusCode: 200, Body: io.NopCloser(strings.NewReader(`{"poll":true}`)), Header: make(http.Header)}, nil
	})
	scratch := t.TempDir()
	ctx, cancel := context.WithTimeout(context.Background(), 2*time.Second)
	defer cancel()
	tunnel, url, e := Start(ctx, exe, "http://127.0.0.1:1", "test-key", scratch)
	if e != nil {
		t.Fatal(e)
	}
	defer tunnel.Stop()
	if url != "https://unit-test.trycloudflare.com" {
		t.Fatal(url)
	}
	if b, e := os.ReadFile(filepath.Join(tunnel.dir, "config.yml")); e != nil || string(b) != "{}" {
		t.Fatalf("isolated config: %q %v", b, e)
	}
	tunnel.Stop()
	tunnel.Stop()
	emptyScratch(t, scratch)
	if tunnel.cmd.ProcessState == nil || !tunnel.cmd.ProcessState.Exited() && tunnel.cmd.ProcessState.Success() {
		t.Fatal("child not reaped")
	}
}
func TestTunnelRejectsUnreadyPublicRoute(t *testing.T) {
	t.Setenv("VITICULTURE_FAKE_TUNNEL", "url")
	exe, e := os.Executable()
	if e != nil {
		t.Fatal(e)
	}
	for _, tc := range []struct {
		name   string
		status int
		body   string
	}{{"gate", 401, `{"poll":true}`}, {"wrong-body", 200, `{"poll":false}`}} {
		t.Run(tc.name, func(t *testing.T) {
			old := http.DefaultTransport
			defer func() { http.DefaultTransport = old }()
			http.DefaultTransport = fakeTransport(func(*http.Request) (*http.Response, error) {
				return &http.Response{StatusCode: tc.status, Body: io.NopCloser(strings.NewReader(tc.body)), Header: make(http.Header)}, nil
			})
			ctx, cancel := context.WithTimeout(context.Background(), 300*time.Millisecond)
			defer cancel()
			scratch := t.TempDir()
			tunnel, _, e := Start(ctx, exe, "http://127.0.0.1:1", "key", scratch)
			if !errors.Is(e, context.DeadlineExceeded) || tunnel != nil {
				t.Fatalf("unready route accepted: %v", e)
			}
			emptyScratch(t, scratch)
		})
	}
}
