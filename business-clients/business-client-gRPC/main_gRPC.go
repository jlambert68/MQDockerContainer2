package main

import (
	"context"
	"fmt"
	"log"
	"os"
	"time"

	"github.com/jlambert68/MQDockerContainer2/mq-gateway/api/proto/mq_grpc_api"

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

	// PUT 1
	putResp, err := client.Put(ctx, &mq_grpc_api.PutRequest{
		Queue:   "DEV.QUEUE.1",
		Message: "Hello via gRPC!",
	})
	if err != nil {
		log.Fatal("Put:", err)
	}
	fmt.Printf("PUT resp: %+v\n", putResp)

	// PUT 2
	putResp2, err := client.Put(ctx, &mq_grpc_api.PutRequest{
		Queue:   "DEV.QUEUE.1",
		Message: "Hello 2 via gRPC!",
	})
	if err != nil {
		log.Fatal("Put:", err)
	}
	fmt.Printf("PUT resp: %+v\n", putResp2)

	// BROWSE FIRST
	browseFirstResp, err := client.BrowseFirst(ctx, &mq_grpc_api.BrowseFirstRequest{
		Queue:       "DEV.QUEUE.1",
		WaitMs:      1000,
		MaxMsgBytes: 65536,
	})
	if err != nil {
		log.Fatal("BrowseFirst:", err)
	}
	fmt.Printf("BROWSE FIRST resp: %+v\n", browseFirstResp)

	// BROWSE NEXT
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

	// INQUIRE QUEUE
	inquireResp, err := client.InquireQueue(ctx, &mq_grpc_api.InquireQueueRequest{
		Queue: "DEV.QUEUE.1",
	})
	if err != nil {
		log.Fatal("InquireQueue:", err)
	}
	fmt.Printf("INQUIRE QUEUE resp: %+v\n", inquireResp)

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

	// INQUIRE QUEUE
	inquireResp2, err := client.InquireQueue(ctx, &mq_grpc_api.InquireQueueRequest{
		Queue: "DEV.QUEUE.1",
	})
	if err != nil {
		log.Fatal("InquireQueue:", err)
	}
	fmt.Printf("INQUIRE QUEUE resp: %+v\n", inquireResp2)

	// GET
	getResp2, err := client.Get(ctx, &mq_grpc_api.GetRequest{
		Queue:       "DEV.QUEUE.1",
		WaitMs:      5000,
		MaxMsgBytes: 65536,
	})
	if err != nil {
		log.Fatal("Get:", err)
	}
	fmt.Printf("GET resp: %+v\n", getResp2)

	// INQUIRE QUEUE
	inquireResp3, err := client.InquireQueue(ctx, &mq_grpc_api.InquireQueueRequest{
		Queue: "DEV.QUEUE.1",
	})
	if err != nil {
		log.Fatal("InquireQueue:", err)
	}
	fmt.Printf("INQUIRE QUEUE resp: %+v\n", inquireResp3)

}
