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
	Queue   string `json:"queue"`
	Message string `json:"message"`
}
type PutResponse struct {
	Status string `json:"status"`
	Error  string `json:"error"`
}
type GetRequest struct {
	Queue       string `json:"queue"`
	WaitMs      int    `json:"wait_ms"`
	MaxMsgBytes int    `json:"max_msg_bytes"`
}
type GetResponse struct {
	Status  string `json:"status"`
	Message string `json:"message"`
	Empty   bool   `json:"empty"`
	Error   string `json:"error"`
}
type BrowseFirstRequest struct {
	Queue       string `json:"queue"`
	WaitMs      int    `json:"wait_ms"`
	MaxMsgBytes int    `json:"max_msg_bytes"`
}
type BrowseNextRequest struct {
	BrowseID    string `json:"browse_id"`
	WaitMs      int    `json:"wait_ms"`
	MaxMsgBytes int    `json:"max_msg_bytes"`
}
type BrowseResponse struct {
	Status   string `json:"status"`
	Message  string `json:"message"`
	Empty    bool   `json:"empty"`
	BrowseID string `json:"browse_id"`
	Error    string `json:"error"`
}

func getenv(k, d string) string {
	if v := os.Getenv(k); v != "" {
		return v
	}
	return d
}

func main() {
	base := getenv("MQ_GATEWAY_URL", "http://localhost:8080")

	// PUT
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

	// GET
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

	// BROWSE FIRST
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

	// BROWSE NEXT
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
}
