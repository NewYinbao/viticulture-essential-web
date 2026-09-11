package sharing

import (
	"bufio"
	"context"
	"fmt"
	"io"
	"net/http"
	"os"
	"os/exec"
	"path/filepath"
	"regexp"
	"sync"
	"time"
)

type Tunnel struct {
	cmd     *exec.Cmd
	Done    chan error
	once    sync.Once
	release func()
	dir     string
}

func (t *Tunnel) Stop() {
	t.once.Do(func() { _ = t.cmd.Process.Kill(); <-t.Done; t.release(); _ = os.RemoveAll(t.dir) })
}

var tunnelURL = regexp.MustCompile(`https://[a-z0-9-]+\.trycloudflare\.com`)

func Start(ctx context.Context, executable, origin, key, scratch string) (*Tunnel, string, error) {
	abs, err := filepath.Abs(executable)
	if err != nil {
		return nil, "", err
	}
	if _, err = os.Stat(abs); err != nil {
		return nil, "", fmt.Errorf("cloudflared missing: run scripts/install-cloudflared.ps1: %w", err)
	}
	if err = os.MkdirAll(scratch, 0700); err != nil {
		return nil, "", err
	}
	dir, err := os.MkdirTemp(scratch, "tunnel-")
	if err != nil {
		return nil, "", err
	}
	dir, err = filepath.Abs(dir)
	if err != nil {
		return nil, "", err
	}
	config := filepath.Join(dir, "config.yml")
	// Explicit empty config; never read or rename user Cloudflare credentials.
	if err = os.WriteFile(config, []byte("{}"), 0600); err != nil {
		os.RemoveAll(dir)
		return nil, "", err
	}
	cmd := exec.Command(abs, "tunnel", "--config", config, "--no-autoupdate", "--url", origin, "--protocol", "http2")
	cmd.Dir = dir
	prepare(cmd)
	reader, writer := io.Pipe()
	cmd.Stdout = writer
	cmd.Stderr = writer
	if err = cmd.Start(); err != nil {
		reader.Close()
		writer.Close()
		os.RemoveAll(dir)
		return nil, "", err
	}
	release, err := contain(cmd)
	if err != nil {
		cmd.Process.Kill()
		cmd.Wait()
		reader.Close()
		writer.Close()
		os.RemoveAll(dir)
		return nil, "", err
	}
	t := &Tunnel{cmd: cmd, Done: make(chan error, 1), release: release, dir: dir}
	urls := make(chan string, 1)
	go func() {
		scanner := bufio.NewScanner(reader)
		for scanner.Scan() {
			if u := tunnelURL.FindString(scanner.Text()); u != "" {
				select {
				case urls <- u:
				default:
				}
			}
		}
		reader.Close()
	}()
	go func() { err := cmd.Wait(); writer.Close(); t.Done <- err; close(t.Done) }()
	timer := time.NewTimer(120 * time.Second)
	defer timer.Stop()
	var url string
	select {
	case url = <-urls:
	case <-ctx.Done():
		t.Stop()
		return nil, "", ctx.Err()
	case <-timer.C:
		t.Stop()
		return nil, "", fmt.Errorf("Quick Tunnel URL timed out (check network/config)")
	case err = <-t.Done:
		t.Stop()
		return nil, "", fmt.Errorf("cloudflared exited before URL: %v", err)
	}
	client := &http.Client{Timeout: 5 * time.Second}
	lastError := "not attempted"
	for {
		req, _ := http.NewRequestWithContext(ctx, "GET", url+"/api/transport", nil)
		req.AddCookie(&http.Cookie{Name: "viticulture_gate", Value: key})
		res, e := client.Do(req)
		if e == nil {
			b, _ := io.ReadAll(io.LimitReader(res.Body, 128))
			res.Body.Close()
			lastError = fmt.Sprintf("HTTP %d", res.StatusCode)
			if res.StatusCode == 200 && string(b) == `{"poll":true}` {
				return t, url, nil
			}
		}
		if e != nil {
			lastError = e.Error()
		}
		select {
		case <-ctx.Done():
			t.Stop()
			return nil, "", ctx.Err()
		case <-timer.C:
			t.Stop()
			return nil, "", fmt.Errorf("Quick Tunnel allocated but public route not ready: %s", lastError)
		case err = <-t.Done:
			t.Stop()
			return nil, "", fmt.Errorf("cloudflared exited: %v", err)
		case <-time.After(time.Second):
		}
	}
}
