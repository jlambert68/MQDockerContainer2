package main

import (
	"context"
	"log"
	"net"
	"net/http"
	"os"
	"os/signal"
	"syscall"
	"time"

	"google.golang.org/grpc"

	// 🔁 CHANGE THIS MODULE PREFIX to match your go.mod
	"mq-gateway/api/grpcsrv"
	"mq-gateway/api/proto/mqpb"
	"mq-gateway/api/rest"
	"mq-gateway/internal/mqcore"
)

func getenv(key, def string) string {
	if v := os.Getenv(key); v != "" {
		return v
	}
	return def
}

func main() {
	log.Println("[main] starting mq-gateway")

	// ------------------------------------------------------------------
	// 1. Connect to IBM MQ (once)
	// ------------------------------------------------------------------
	gateway, err := mqcore.NewGateway()
	if err != nil {
		log.Fatalf("[main] failed to connect to MQ: %v", err)
	}
	defer gateway.Close()
	log.Println("[main] connected to MQ")

	// ------------------------------------------------------------------
	// 2. REST server
	// ------------------------------------------------------------------
	restPort := getenv("REST_PORT", ":8080")

	restHandler := &rest.Handler{
		GW: gateway,
	}

	restServer := &http.Server{
		Addr:         restPort,
		Handler:      restHandler.Routes(),
		ReadTimeout:  15 * time.Second,
		WriteTimeout: 15 * time.Second,
	}

	go func() {
		log.Printf("[REST] listening on %s\n", restPort)
		if err := restServer.ListenAndServe(); err != nil && err != http.ErrServerClosed {
			log.Fatalf("[REST] server error: %v", err)
		}
	}()

	// ------------------------------------------------------------------
	// 3. gRPC server
	// ------------------------------------------------------------------
	grpcPort := getenv("GRPC_PORT", ":9090")

	lis, err := net.Listen("tcp", grpcPort)
	if err != nil {
		log.Fatalf("[gRPC] failed to listen: %v", err)
	}

	grpcServer := grpc.NewServer()
	mqpb.RegisterMQServer(grpcServer, &grpcsrv.Server{
		GW: gateway,
	})

	go func() {
		log.Printf("[gRPC] listening on %s\n", grpcPort)
		if err := grpcServer.Serve(lis); err != nil {
			log.Fatalf("[gRPC] server error: %v", err)
		}
	}()

	// ------------------------------------------------------------------
	// 4. Graceful shutdown
	// ------------------------------------------------------------------
	sigCh := make(chan os.Signal, 1)
	signal.Notify(sigCh, syscall.SIGINT, syscall.SIGTERM)

	sig := <-sigCh
	log.Printf("[main] received signal %s, shutting down", sig)

	// Stop gRPC
	grpcServer.GracefulStop()

	// Stop REST
	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()
	if err := restServer.Shutdown(ctx); err != nil {
		log.Printf("[REST] shutdown error: %v", err)
	}

	log.Println("[main] mq-gateway stopped cleanly")
}
