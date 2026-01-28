package main

import (
	"context"
	"fmt"
	"log"
	"os"
	"time"

	"github.com/jlambert68/MQDockerContainer2/mq-gateway/pkg/mqrest"
)

type PutRequest struct {
	// Target queue name.
	Queue string `json:"queue"`
	// Payload to put.
	Message string `json:"message"`
}
type PutResponse struct {
	Status string `json:"status"`
	Error  string `json:"error"`
}
type GetRequest struct {
	// Target queue name.
	Queue string `json:"queue"`
	// Wait interval in milliseconds.
	WaitMs int `json:"wait_ms"`
	// Max message size in bytes.
	MaxMsgBytes int `json:"max_msg_bytes"`
}
type GetResponse struct {
	Status  string `json:"status"`
	Message string `json:"message"`
	Empty   bool   `json:"empty"`
	Error   string `json:"error"`
}
type BrowseFirstRequest struct {
	// Target queue name.
	Queue string `json:"queue"`
	// Wait interval in milliseconds.
	WaitMs int `json:"wait_ms"`
	// Max message size in bytes.
	MaxMsgBytes int `json:"max_msg_bytes"`
}
type BrowseNextRequest struct {
	// Browse session token returned by /browse/first.
	BrowseID string `json:"browse_id"`
	// Wait interval in milliseconds.
	WaitMs int `json:"wait_ms"`
	// Max message size in bytes.
	MaxMsgBytes int `json:"max_msg_bytes"`
}
type BrowseResponse struct {
	Status  string `json:"status"`
	Message string `json:"message"`
	Empty   bool   `json:"empty"`
	// BrowseID is set only for BrowseFirst.
	BrowseID string `json:"browse_id"`
	Error    string `json:"error"`
}
type InquireQueueRequest struct {
	// Target queue name.
	Queue string `json:"queue"`
}
type InquireQueueResponse struct {
	Status string `json:"status"`
	// Queue is the resolved queue name (may be normalized by MQ).
	Queue           string `json:"queue"`
	QueueDesc       string `json:"queue_desc"`
	QueueType       int32  `json:"queue_type"`
	QueueUsage      int32  `json:"queue_usage"`
	DefPersistence  int32  `json:"def_persistence"`
	InhibitGet      int32  `json:"inhibit_get"`
	InhibitPut      int32  `json:"inhibit_put"`
	CurrentQDepth   int32  `json:"current_q_depth"`
	MaxQDepth       int32  `json:"max_q_depth"`
	OpenInputCount  int32  `json:"open_input_count"`
	OpenOutputCount int32  `json:"open_output_count"`
	Error           string `json:"error"`
}

func getenv(k, d string) string {
	if v := os.Getenv(k); v != "" {
		return v
	}
	return d
}

func main() {
	// Base URL for the REST gateway.
	base := getenv("MQ_GATEWAY_URL", "http://localhost:8080")

	restClient := mqrest.NewClient(mqrest.Config{
		BaseURL: base,
		Timeout: 10 * time.Second,
	})

	// PUT 1 - seed the queue with a message.
	p := mqrest.PutRequest{Queue: "DEV.QUEUE.1", Message: "Hello via REST!"}
	pres, err := restClient.Put(context.Background(), p)
	if err != nil {
		log.Fatal(err)
	}
	fmt.Printf("PUT resp: %+v\n", pres)

	// PUT 2 - add another message for browsing.
	p2 := mqrest.PutRequest{Queue: "DEV.QUEUE.1", Message: "Hello 2 via REST!"}
	pres2, err := restClient.Put(context.Background(), p2)
	if err != nil {
		log.Fatal(err)
	}
	fmt.Printf("PUT resp: %+v\n", pres2)

	// BROWSE FIRST - non-destructive peek and start browse cursor.
	bf := mqrest.BrowseFirstRequest{Queue: "DEV.QUEUE.1", WaitMs: 1000, MaxMsgBytes: 65536}
	bres, err := restClient.BrowseFirst(context.Background(), bf)
	if err != nil {
		log.Fatal(err)
	}
	fmt.Printf("BROWSE FIRST resp: %+v\n", bres)

	// BROWSE NEXT - continue the browse cursor.
	if bres.BrowseID != "" {
		bn := mqrest.BrowseNextRequest{BrowseID: bres.BrowseID, WaitMs: 1000, MaxMsgBytes: 65536}
		bnres, err := restClient.BrowseNext(context.Background(), bn)
		if err != nil {
			log.Fatal(err)
		}
		fmt.Printf("BROWSE NEXT resp: %+v\n", bnres)
	}

	// INQUIRE QUEUE - fetch queue attributes before destructive reads.
	iq := mqrest.InquireQueueRequest{Queue: "DEV.QUEUE.1"}
	iqres, err := restClient.InquireQueue(context.Background(), iq)
	if err != nil {
		log.Fatal(err)
	}
	fmt.Printf("INQUIRE QUEUE resp: %+v\n", iqres)

	// GET - destructive read.
	g := mqrest.GetRequest{Queue: "DEV.QUEUE.1", WaitMs: 5000, MaxMsgBytes: 65536}
	gres, err := restClient.Get(context.Background(), g)
	if err != nil {
		log.Fatal(err)
	}
	fmt.Printf("GET resp: %+v\n", gres)

	// INQUIRE QUEUE - check attributes again after one GET.
	iq2 := mqrest.InquireQueueRequest{Queue: "DEV.QUEUE.1"}
	iqres2, err := restClient.InquireQueue(context.Background(), iq2)
	if err != nil {
		log.Fatal(err)
	}
	fmt.Printf("INQUIRE QUEUE resp: %+v\n", iqres2)

	// GET - destructive read of the next message.
	g2 := mqrest.GetRequest{Queue: "DEV.QUEUE.1", WaitMs: 5000, MaxMsgBytes: 65536}
	gres2, err := restClient.Get(context.Background(), g2)
	if err != nil {
		log.Fatal(err)
	}
	fmt.Printf("GET resp: %+v\n", gres2)

	// INQUIRE QUEUE - final attribute check after second GET.
	iq3 := mqrest.InquireQueueRequest{Queue: "DEV.QUEUE.1"}
	iqres3, err := restClient.InquireQueue(context.Background(), iq3)
	if err != nil {
		log.Fatal(err)
	}
	fmt.Printf("INQUIRE QUEUE resp: %+v\n", iqres3)
}
