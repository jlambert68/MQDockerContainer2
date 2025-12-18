package main

import (
	"fmt"
	"net/http"
	"os"
	"time"
)

func main() {
	http.HandleFunc("/hello", func(w http.ResponseWriter, r *http.Request) {
		msg := fmt.Sprintf("hello from Linux at %s pid=%d\n", time.Now().Format(time.RFC3339), os.Getpid())
		w.Write([]byte(msg))
	})

	fmt.Println("listening on :8080")
	// Put a breakpoint on the next line to verify debugging works.
	if err := http.ListenAndServe(":8080", nil); err != nil {
		panic(err)
	}
}
