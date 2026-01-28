package main

import (
	"bytes"
	"encoding/json"
	"fmt"
	"log"
	"net/http"
	"os"
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

	// PUT 1 - seed the queue with a message.
	p := PutRequest{Queue: "DEV.QUEUE.1", Message: "Hello via REST!"}
	buf, _ := json.Marshal(p)
	resp, err := http.Post(base+"/put", "application/json", bytes.NewReader(buf))
	if err != nil {
		log.Fatal(err)
	}
	defer resp.Body.Close()
	var pres PutResponse
	_ = json.NewDecoder(resp.Body).Decode(&pres)
	fmt.Printf("PUT resp: %+v\n", pres)

	// PUT 2 - add another message for browsing.
	p2 := PutRequest{Queue: "DEV.QUEUE.1", Message: "Hello 2 via REST!"}
	buf, _ = json.Marshal(p2)
	resp, err = http.Post(base+"/put", "application/json", bytes.NewReader(buf))
	if err != nil {
		log.Fatal(err)
	}
	defer resp.Body.Close()
	var pres2 PutResponse
	_ = json.NewDecoder(resp.Body).Decode(&pres2)
	fmt.Printf("PUT resp: %+v\n", pres2)

	// BROWSE FIRST - non-destructive peek and start browse cursor.
	bf := BrowseFirstRequest{Queue: "DEV.QUEUE.1", WaitMs: 1000, MaxMsgBytes: 65536}
	buf, _ = json.Marshal(bf)
	resp, err = http.Post(base+"/browse/first", "application/json", bytes.NewReader(buf))
	if err != nil {
		log.Fatal(err)
	}
	defer resp.Body.Close()
	var bres BrowseResponse
	_ = json.NewDecoder(resp.Body).Decode(&bres)
	fmt.Printf("BROWSE FIRST resp: %+v\n", bres)

	// BROWSE NEXT - continue the browse cursor.
	if bres.BrowseID != "" {
		bn := BrowseNextRequest{BrowseID: bres.BrowseID, WaitMs: 1000, MaxMsgBytes: 65536}
		buf, _ = json.Marshal(bn)
		resp, err = http.Post(base+"/browse/next", "application/json", bytes.NewReader(buf))
		if err != nil {
			log.Fatal(err)
		}
		defer resp.Body.Close()
		var bnres BrowseResponse
		_ = json.NewDecoder(resp.Body).Decode(&bnres)
		fmt.Printf("BROWSE NEXT resp: %+v\n", bnres)
	}

	// INQUIRE QUEUE - fetch queue attributes before destructive reads.
	iq := InquireQueueRequest{Queue: "DEV.QUEUE.1"}
	buf, _ = json.Marshal(iq)
	resp, err = http.Post(base+"/inquire/queue", "application/json", bytes.NewReader(buf))
	if err != nil {
		log.Fatal(err)
	}
	defer resp.Body.Close()
	var iqres InquireQueueResponse
	_ = json.NewDecoder(resp.Body).Decode(&iqres)
	fmt.Printf("INQUIRE QUEUE resp: %+v\n", iqres)

	// GET - destructive read.
	g := GetRequest{Queue: "DEV.QUEUE.1", WaitMs: 5000, MaxMsgBytes: 65536}
	buf, _ = json.Marshal(g)
	resp, err = http.Post(base+"/get", "application/json", bytes.NewReader(buf))
	if err != nil {
		log.Fatal(err)
	}
	defer resp.Body.Close()
	var gres GetResponse
	_ = json.NewDecoder(resp.Body).Decode(&gres)
	fmt.Printf("GET resp: %+v\n", gres)

	// INQUIRE QUEUE - check attributes again after one GET.
	iq2 := InquireQueueRequest{Queue: "DEV.QUEUE.1"}
	buf, _ = json.Marshal(iq2)
	resp, err = http.Post(base+"/inquire/queue", "application/json", bytes.NewReader(buf))
	if err != nil {
		log.Fatal(err)
	}
	defer resp.Body.Close()
	var iqres2 InquireQueueResponse
	_ = json.NewDecoder(resp.Body).Decode(&iqres2)
	fmt.Printf("INQUIRE QUEUE resp: %+v\n", iqres2)

	// GET - destructive read of the next message.
	g2 := GetRequest{Queue: "DEV.QUEUE.1", WaitMs: 5000, MaxMsgBytes: 65536}
	buf, _ = json.Marshal(g2)
	resp, err = http.Post(base+"/get", "application/json", bytes.NewReader(buf))
	if err != nil {
		log.Fatal(err)
	}
	defer resp.Body.Close()
	var gres2 GetResponse
	_ = json.NewDecoder(resp.Body).Decode(&gres2)
	fmt.Printf("GET resp: %+v\n", gres2)

	// INQUIRE QUEUE - final attribute check after second GET.
	iq3 := InquireQueueRequest{Queue: "DEV.QUEUE.1"}
	buf, _ = json.Marshal(iq3)
	resp, err = http.Post(base+"/inquire/queue", "application/json", bytes.NewReader(buf))
	if err != nil {
		log.Fatal(err)
	}
	defer resp.Body.Close()
	var iqres3 InquireQueueResponse
	_ = json.NewDecoder(resp.Body).Decode(&iqres3)
	fmt.Printf("INQUIRE QUEUE resp: %+v\n", iqres3)
}
