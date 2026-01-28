package main

import (
	"context"
	"fmt"
	"log"
	"os"
	"time"

	"github.com/jlambert68/MQDockerContainer2/mq-gateway/api/proto/mq_grpc_api"
	"github.com/jlambert68/MQDockerContainer2/mq-gateway/pkg/mqgrpc"
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

	// Connect using the wrapper client (plaintext for local dev).
	client, err := mqgrpc.NewClient(mqgrpc.Config{Address: addr})
	if err != nil {
		log.Fatal("dial:", err)
	}
	defer client.Close()

	// Shared request context with timeout.
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	// PUT 1 - seed the queue with a message.
	putResp, err := client.Put(ctx, &mq_grpc_api.PutRequest{
		Queue:   "DEV.QUEUE.1",
		Message: "Hello via gRPC!",
	})
	if err != nil {
		log.Fatal("Put:", err)
	}
	fmt.Printf("PUT resp: %+v\n", putResp)

	// PUT 2 - add another message for browsing.
	putResp2, err := client.Put(ctx, &mq_grpc_api.PutRequest{
		Queue:   "DEV.QUEUE.1",
		Message: "Hello 2 via gRPC!",
	})
	if err != nil {
		log.Fatal("Put:", err)
	}
	fmt.Printf("PUT resp: %+v\n", putResp2)

	// BROWSE FIRST - start a browse cursor (non-destructive).
	browseFirstResp, err := client.BrowseFirst(ctx, &mq_grpc_api.BrowseFirstRequest{
		Queue:       "DEV.QUEUE.1",
		WaitMs:      1000,
		MaxMsgBytes: 65536,
	})
	if err != nil {
		log.Fatal("BrowseFirst:", err)
	}
	fmt.Printf("BROWSE FIRST resp: %+v\n", browseFirstResp)

	// BROWSE NEXT - continue the browse cursor.
	if browseFirstResp.GetBrowseId() != "" {
		browseNextResp, err := client.BrowseNext(ctx, &mq_grpc_api.BrowseNextRequest{
			BrowseId:    browseFirstResp.GetBrowseId(),
			WaitMs:      1000,
			MaxMsgBytes: 65536,
		})
		if err != nil {
			log.Fatal("BrowseNext:", err)
		}
		fmt.Printf("BROWSE NEXT resp: %+v\n", browseNextResp)
	}

	// INQUIRE QUEUE - fetch queue attributes before destructive reads.
	inquireResp, err := client.InquireQueue(ctx, &mq_grpc_api.InquireQueueRequest{
		Queue: "DEV.QUEUE.1",
	})
	if err != nil {
		log.Fatal("InquireQueue:", err)
	}
	fmt.Printf("INQUIRE QUEUE resp: %+v\n", inquireResp)

	// GET - destructive read.
	getResp, err := client.Get(ctx, &mq_grpc_api.GetRequest{
		Queue:       "DEV.QUEUE.1",
		WaitMs:      5000,
		MaxMsgBytes: 65536,
	})
	if err != nil {
		log.Fatal("Get:", err)
	}
	fmt.Printf("GET resp: %+v\n", getResp)

	// INQUIRE QUEUE - check attributes again after one GET.
	inquireResp2, err := client.InquireQueue(ctx, &mq_grpc_api.InquireQueueRequest{
		Queue: "DEV.QUEUE.1",
	})
	if err != nil {
		log.Fatal("InquireQueue:", err)
	}
	fmt.Printf("INQUIRE QUEUE resp: %+v\n", inquireResp2)

	// GET - destructive read of the next message.
	getResp2, err := client.Get(ctx, &mq_grpc_api.GetRequest{
		Queue:       "DEV.QUEUE.1",
		WaitMs:      5000,
		MaxMsgBytes: 65536,
	})
	if err != nil {
		log.Fatal("Get:", err)
	}
	fmt.Printf("GET resp: %+v\n", getResp2)

	// INQUIRE QUEUE - final attribute check after second GET.
	inquireResp3, err := client.InquireQueue(ctx, &mq_grpc_api.InquireQueueRequest{
		Queue: "DEV.QUEUE.1",
	})
	if err != nil {
		log.Fatal("InquireQueue:", err)
	}
	fmt.Printf("INQUIRE QUEUE resp: %+v\n", inquireResp3)

}
