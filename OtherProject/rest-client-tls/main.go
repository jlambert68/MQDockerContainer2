package main

import (
	"context"
	"crypto/tls"
	"fmt"
	"log"
	"os"
	"time"

	"github.com/jlambert68/MQDockerContainer2/mq-gateway/pkg/mqrest"
	"github.com/jlambert68/MQDockerContainer2/mq-gateway/pkg/mqtls"
)

func getenv(k, d string) string {
	if v := os.Getenv(k); v != "" {
		return v
	}
	return d
}

func main() {
	baseURL := getenv("MQ_REST_URL", "https://localhost:8080")

	// TLS input (PEM by default).
	certFile := getenv("MQ_TLS_CERT", "")
	keyFile := getenv("MQ_TLS_KEY", "")
	caFile := getenv("MQ_TLS_CA", "")
	keyPassEnv := getenv("MQ_TLS_KEY_PASSWORD_ENV", "")

	// Optional PKCS#12 (.p12/.pfx).
	p12File := getenv("MQ_TLS_P12", "")
	p12PassEnv := getenv("MQ_TLS_P12_PASSWORD_ENV", "")

	var tlsCfg *tls.Config
	var err error
	if p12File != "" {
		tlsCfg, err = mqtls.LoadClientTLSConfigFromP12Env(p12File, caFile, p12PassEnv)
	} else {
		tlsCfg, err = mqtls.LoadClientTLSConfigFromEnv(certFile, keyFile, caFile, keyPassEnv)
	}
	if err != nil {
		log.Fatal("TLS config:", err)
	}

	client := mqrest.NewClient(mqrest.Config{
		BaseURL:   baseURL,
		Timeout:   10 * time.Second,
		TLSConfig: tlsCfg,
	})

	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	putResp, err := client.Put(ctx, mqrest.PutRequest{
		Queue:   "DEV.QUEUE.1",
		Message: "Hello via REST + TLS!",
	})
	if err != nil {
		log.Fatal("Put:", err)
	}
	fmt.Printf("PUT resp: %+v\n", putResp)
}
