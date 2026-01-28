package main

import (
	"context"
	"fmt"
	"log"
	"os"
	"time"

	"github.com/jlambert68/MQDockerContainer2/mq-gateway/api/proto/mq_grpc_api"
	"github.com/jlambert68/MQDockerContainer2/mq-gateway/pkg/mqgrpc"
)

func getenv(k, d string) string {
	if v := os.Getenv(k); v != "" {
		return v
	}
	return d
}

func main() {
	addr := getenv("MQ_GRPC_ADDR", "localhost:9090")

	client, err := mqgrpc.NewClient(mqgrpc.Config{Address: addr})
	if err != nil {
		log.Fatal("dial:", err)
	}
	defer client.Close()

	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	putResp, err := client.Put(ctx, &mq_grpc_api.PutRequest{
		Queue:   "DEV.QUEUE.1",
		Message: "Hello from OtherProject!",
	})
	if err != nil {
		log.Fatal("Put:", err)
	}
	fmt.Printf("PUT resp: %+v\n", putResp)

	browseFirstResp, err := client.BrowseFirst(ctx, &mq_grpc_api.BrowseFirstRequest{
		Queue:       "DEV.QUEUE.1",
		WaitMs:      1000,
		MaxMsgBytes: 65536,
	})
	if err != nil {
		log.Fatal("BrowseFirst:", err)
	}
	fmt.Printf("BROWSE FIRST resp: %+v\n", browseFirstResp)

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

	inquireResp, err := client.InquireQueue(ctx, &mq_grpc_api.InquireQueueRequest{
		Queue: "DEV.QUEUE.1",
	})
	if err != nil {
		log.Fatal("InquireQueue:", err)
	}
	fmt.Printf("INQUIRE QUEUE resp: %+v\n", inquireResp)

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
