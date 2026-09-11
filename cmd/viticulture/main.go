package main

import (
	"context"
	"flag"
	"fmt"
	"log"
	"net"
	"net/http"
	"os"
	"os/signal"
	"path/filepath"
	"syscall"
	"time"
	"vineyard/internal/server"
	"vineyard/internal/sharing"
)

func main() {
	if err := run(); err != nil {
		log.Print(err)
		os.Exit(1)
	}
}
func run() error {
	addr := flag.String("addr", "127.0.0.1:3010", "监听地址，局域网使用0.0.0.0:3010")
	data := flag.String("data", "ee-data", "存档目录")
	public := flag.Bool("public", false, "Enable temporary Cloudflare public link with random shared key")
	cloudflared := flag.String("cloudflared", "dist/tools/cloudflared.exe", "cloudflared executable")
	flag.Parse()
	listener, e := net.Listen("tcp", *addr)
	if e != nil {
		return fmt.Errorf("port unavailable (no existing process touched): %w", e)
	}
	defer listener.Close()
	a, e := server.New(*data)
	if e != nil {
		return e
	}
	a.WriteEnrollmentCodes(os.Stdout)
	handler := a.Handler()
	key := ""
	if *public {
		key, e = sharing.NewKey()
		if e != nil {
			return e
		}
		handler = sharing.Gate(handler, key)
		fmt.Println("WARNING: PUBLIC access through Cloudflare. Share URL and key only with trusted friends. No SLA; not hardened Internet hosting.")
		fmt.Println("Access key (changes every launch):", key)
	}
	srv := &http.Server{Addr: *addr, Handler: handler, ReadHeaderTimeout: 5 * time.Second, ReadTimeout: 15 * time.Second, IdleTimeout: 60 * time.Second, MaxHeaderBytes: 16384}
	fmt.Println("葡萄酒庄园 · Essential Edition")
	fmt.Printf("监听：http://%s\n", listener.Addr())
	_, port, _ := net.SplitHostPort(listener.Addr().String())
	ips, _ := net.InterfaceAddrs()
	for _, ip := range ips {
		if n, ok := ip.(*net.IPNet); ok && !n.IP.IsLoopback() && n.IP.To4() != nil {
			fmt.Printf("网卡候选地址 http://%s:%s（需监听0.0.0.0且防火墙允许）\n", n.IP, port)
		}
	}
	abs, _ := filepath.Abs(*data)
	fmt.Println("存档：", abs)
	fmt.Println("关闭此窗口停止服务器及本次隧道。默认仅局域网；不修改防火墙。")
	ctx, stop := signal.NotifyContext(context.Background(), os.Interrupt, syscall.SIGTERM)
	defer stop()
	go func() {
		<-ctx.Done()
		c, cancel := context.WithTimeout(context.Background(), 3*time.Second)
		defer cancel()
		_ = srv.Shutdown(c)
	}()
	served := make(chan error, 1)
	go func() { served <- srv.Serve(listener) }()
	defer srv.Close()
	if *public {
		host, _, _ := net.SplitHostPort(listener.Addr().String())
		if host == "" || host == "0.0.0.0" {
			host = "127.0.0.1"
		}
		if host == "::" {
			host = "::1"
		}
		tunnel, url, err := sharing.Start(ctx, *cloudflared, "http://"+net.JoinHostPort(host, port), key, "artifacts/tunnels")
		if err != nil {
			return fmt.Errorf("PUBLIC START FAILED; server stopped: %w", err)
		}
		defer tunnel.Stop()
		fmt.Println("Temporary HTTPS URL:", url)
		fmt.Println("Public route verified; polling enabled, not SSE. This readiness check is not a gameplay self-test.")
		select {
		case err := <-tunnel.Done:
			return fmt.Errorf("PUBLIC TUNNEL STOPPED; server stopped: %v", err)
		case err := <-served:
			if err != nil && err != http.ErrServerClosed {
				return err
			}
			return nil
		case <-ctx.Done():
			return nil
		}
	}
	e = <-served
	if e != nil && e != http.ErrServerClosed {
		return e
	}
	return nil
}
