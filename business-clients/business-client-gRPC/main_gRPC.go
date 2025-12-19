package main

import (
	"context"
	"fmt"
	"log"
	"os"
	"time"

	"google.golang.org/grpc"

	"github.com/jlambert68/mq-gateway/api/proto/mqpb"
)

func getenv(k, d string) string {
	if v := os.Getenv(k); v != "" {
		return v
	}
	return d
}

func main() {
	addr := getenv("MQ_GRPC_ADDR", "localhost:9090")

	conn, err := grpc.Dial(addr, grpc.WithInsecure())
	if err != nil {
		log.Fatal("dial:", err)
	}
	defer conn.Close()

	client := mqpb.NewMQClient(conn)

	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	// PUT
	putResp, err := client.Put(ctx, &mqpb.PutRequest{
		Queue:   "DEV.QUEUE.1",
		Message: "Hello via gRPC!",
	})
	if err != nil {
		log.Fatal("Put:", err)
	}
	fmt.Printf("PUT resp: %+v\n", putResp)

	// GET
	getResp, err := client.Get(ctx, &mqpb.GetRequest{
		Queue:       "DEV.QUEUE.1",
		WaitMs:      5000,
		MaxMsgBytes: 65536,
	})
	if err != nil {
		log.Fatal("Get:", err)
	}
	fmt.Printf("GET resp: %+v\n", getResp)
}
