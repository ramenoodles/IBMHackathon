package main

import (
	"context"
	"fmt"
	"grepwrapper/internal/config"
	"grepwrapper/internal/httpapi"
	"grepwrapper/internal/workspace"
	"log"
	"net/http"
	"os"
	"os/signal"
	"syscall"
	"time"
)

func main() {
	cfg := config.FromEnvironment()
	m, e := workspace.NewManager()
	if e != nil {
		log.Fatal(e)
	}
	defer m.Close()
	h := httpapi.New(m, cfg.RGBinary, cfg.WatsonxModel, cfg.WatsonxAPIKey != "" && cfg.WatsonxProjectID != "" && cfg.WatsonxModel != "")
	srv := &http.Server{Addr: cfg.Host + ":" + cfg.Port, Handler: timeout(h.Handler()), ReadHeaderTimeout: 10 * time.Second, WriteTimeout: 120 * time.Second, IdleTimeout: 60 * time.Second}
	ctx, stop := signal.NotifyContext(context.Background(), os.Interrupt, syscall.SIGTERM)
	defer stop()
	go func() {
		<-ctx.Done()
		shutdownCtx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
		defer cancel()
		_ = srv.Shutdown(shutdownCtx)
	}()
	log.Printf("grepwrapper API listening on %s", srv.Addr)
	if e := srv.ListenAndServe(); e != nil && e != http.ErrServerClosed {
		log.Fatal(e)
	}
}
func timeout(next http.Handler) http.Handler {
	return http.TimeoutHandler(next, 120*time.Second, fmt.Sprintf(`{"error":{"code":"timeout","message":"request timed out"}}`))
}
