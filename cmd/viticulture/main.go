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
)

func main() {
	addr := flag.String("addr", "127.0.0.1:3010", "监听地址，局域网使用0.0.0.0:3010")
	data := flag.String("data", "ee-data", "存档目录")
	flag.Parse()
	a, e := server.New(*data)
	if e != nil {
		log.Fatal(e)
	}
	srv := &http.Server{Addr: *addr, Handler: a.Handler(), ReadHeaderTimeout: 5 * time.Second, ReadTimeout: 15 * time.Second, IdleTimeout: 60 * time.Second, MaxHeaderBytes: 16384}
	fmt.Println("葡萄酒庄园 · Essential Edition")
	fmt.Printf("监听：http://%s\n", *addr)
	_, port, _ := net.SplitHostPort(*addr)
	ips, _ := net.InterfaceAddrs()
	for _, ip := range ips {
		if n, ok := ip.(*net.IPNet); ok && !n.IP.IsLoopback() && n.IP.To4() != nil {
			fmt.Printf("网卡候选地址 http://%s:%s（需监听0.0.0.0且防火墙允许）\n", n.IP, port)
		}
	}
	abs, _ := filepath.Abs(*data)
	fmt.Println("存档：", abs)
	fmt.Println("仅限可信局域网，不要映射到公网。关闭此窗口停止服务器。")
	ctx, stop := signal.NotifyContext(context.Background(), os.Interrupt, syscall.SIGTERM)
	defer stop()
	go func() {
		<-ctx.Done()
		c, cancel := context.WithTimeout(context.Background(), 3*time.Second)
		defer cancel()
		_ = srv.Shutdown(c)
	}()
	if e := srv.ListenAndServe(); e != nil && e != http.ErrServerClosed {
		log.Fatal(e)
	}
}
