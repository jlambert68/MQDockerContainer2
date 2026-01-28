package main

import (
	"context"
	"fmt"
	"github.com/jlambert68/MQDockerContainer2/mq-gateway/internal/logging"
	"github.com/jlambert68/MQDockerContainer2/mq-gateway/internal/mqcore"

	"log/slog"
	"net"
	"net/http"
	"os"
	"os/signal"
	"syscall"
	"time"

	"google.golang.org/grpc"

	"github.com/jlambert68/MQDockerContainer2/mq-gateway/api/gprcsrv"
	"github.com/jlambert68/MQDockerContainer2/mq-gateway/api/proto/mq_grpc_api"
	// 🔁 CHANGE THIS MODULE PREFIX to match your go.mod
	"github.com/jlambert68/MQDockerContainer2/mq-gateway/api/rest"
)

func getenv(key, def string) string {
	if v := os.Getenv(key); v != "" {
		return v
	}
	return def
}

func main() {

	logging.Init("mq-gateway")

	slog.Info("Sleeping for 10 seconds to allow MQ to start up")
	time.Sleep(10 * time.Second)

	slog.Info("[main] starting github.com/jlambert68/MQDockerContainer2/mq-gateway",
		"id", "b21007ef-2195-45f2-bc4e-6f53b8dc7017")

	slog.Info("[main] Exiting github.com/jlambert68/MQDockerContainer2/mq-gateway",
		"id", "37500e51-aced-4cd1-9fd1-26f19c297982")

	//log.Println("[main] starting github.com/jlambert68/MQDockerContainer2/mq-gateway")

	// ------------------------------------------------------------------
	// 1. Connect to IBM MQ (once)
	// ------------------------------------------------------------------
	gateway, err := mqcore.NewGateway()
	if err != nil {
		slog.Error("[main] failed to connect to MQ",
			"error", err,
			"id", "74a80c22-5b70-43f9-b651-a9334d22d29d")
		os.Exit(1)
	}
	defer gateway.Close()

	slog.Info("[main] connected to MQ",
		"id", "d0a80fb4-71f5-4214-9b31-605a38ea5c97")

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
		slog.Info(fmt.Sprintf("[REST] listening on %s", restPort))
		if err := restServer.ListenAndServe(); err != nil && err != http.ErrServerClosed {
			slog.Error("[REST] server error",
				"error", err,
				"id", "954b394f-07c4-47bd-afc6-790b60e66a8a")
			os.Exit(1)
		}

	}()

	// ------------------------------------------------------------------
	// 3. gRPC server
	// ------------------------------------------------------------------
	grpcPort := getenv("GRPC_PORT", ":9090")

	lis, err := net.Listen("tcp", grpcPort)
	if err != nil {
		slog.Error("[gRPC] failed to listen",
			"error", err,
			"id", "a6e161a3-fc09-4b97-96ae-ae0b0bcfce3b")

		os.Exit(1)
	}
	slog.Info("[gRPC] listening",
		"addr", lis.Addr().String(),
		"id", "e9089512-a789-41fe-a4c6-fcc239f94347",
	)

	grpcServer := grpc.NewServer()
	mq_grpc_api.RegisterMqGrpcServicesServer(grpcServer, &grpcsrv.Server{
		GW: gateway,
	})

	go func() {
		slog.Info("[gRPC] listening",
			"addr", grpcPort,
			"id", "262a76a1-536d-4775-bac8-f3dc457919df",
		)
		if err := grpcServer.Serve(lis); err != nil {
			slog.Error("[gRPC] server error",
				"error", err,
				"id", "0b7852fd-e166-436c-95e4-29d17082294f")
			os.Exit(1)
		}
	}()

	// ------------------------------------------------------------------
	// 4. Graceful shutdown
	// ------------------------------------------------------------------
	sigCh := make(chan os.Signal, 1)
	signal.Notify(sigCh, syscall.SIGINT, syscall.SIGTERM)

	sig := <-sigCh
	slog.Info(fmt.Sprintf("[main] received signal '%s', shutting down", sig))

	// Stop gRPC
	grpcServer.GracefulStop()

	// Stop REST
	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()
	if err := restServer.Shutdown(ctx); err != nil {
		slog.Error("[REST] shutdown error",
			"error", err,
			"id", "040462a5-fff7-4b5a-b0f0-a9d740337349")
		os.Exit(1)
	}

	slog.Info("[main] github.com/jlambert68/MQDockerContainer2/mq-gateway stopped cleanly",
		"id", "4356edef-60e2-42f0-afab-924ffe08008f")
}
