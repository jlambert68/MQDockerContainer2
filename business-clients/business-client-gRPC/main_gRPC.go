package main

import (
	"context"
	"fmt"
	"github.com/jlambert68/MQDockerContainer2/mq-gateway/api/proto/mq_grpc_api"
	"log"
	"os"
	"time"

	"google.golang.org/grpc"
	//"github.com/jlambert68/MQDockerContainer2/mq-gateway/api/proto/mq_grpc_api"
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

	client := mq_grpc_api.NewMqGrpcServicesClient(conn)

	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	// PUT
	putResp, err := client.Put(ctx, &mq_grpc_api.PutRequest{
		Queue:   "DEV.QUEUE.1",
		Message: "Hello via gRPC!",
	})
	if err != nil {
		log.Fatal("Put:", err)
	}
	fmt.Printf("PUT resp: %+v\n", putResp)

	// GET
	getResp, err := client.Get(ctx, &mq_grpc_api.GetRequest{
		Queue:       "DEV.QUEUE.1",
		WaitMs:      5000,
		MaxMsgBytes: 65536,
	})
	if err != nil {
		log.Fatal("Get:", err)
	}
	fmt.Printf("GET resp: %+v\n", getResp)
}
