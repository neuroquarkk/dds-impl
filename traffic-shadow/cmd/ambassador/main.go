package main

import (
	"bytes"
	"flag"
	"io"
	"log"
	"net/http"
	"strconv"
	"sync/atomic"
	"time"
)

var (
	// net/http servers each request on its own goroutine
	// so this file has many of this three groups running concurrently
	// plain increment would rance. atmoic keeps every read/write safe
	totalReq       int32 = 0
	v1Error        int32 = 0
	v2Error        int32 = 0
	v1TotalLatency int64 = 0
	v2TotalLatency int64 = 0
	client               = &http.Client{
		Timeout: 2 * time.Second,
	}
)

func main() {
	var port, v1Port, v2Port int
	flag.IntVar(&port, "port", 8081, "port for the ambassador")
	flag.IntVar(&v1Port, "v1", 8082, "port of the v1 primary server (default 8082)")
	flag.IntVar(&v2Port, "v2", 8083, "port of the v2 shadow server (default 8083)")
	flag.Parse()

	url1 := "http://localhost:" + strconv.Itoa(v1Port)
	url2 := "http://localhost:" + strconv.Itoa(v2Port)

	mux := http.NewServeMux()

	mux.HandleFunc("POST /", func(w http.ResponseWriter, r *http.Request) {
		atomic.AddInt32(&totalReq, 1)

		// read the body once and clone it into two separate buffers
		// reusing r.Body across two requets would race since only can consume
		bodyBytes, _ := io.ReadAll(r.Body)
		r.Body.Close()

		reqA, _ := http.NewRequest("POST", url1, bytes.NewBuffer(bodyBytes))
		reqB, _ := http.NewRequest("POST", url2, bytes.NewBuffer(bodyBytes))

		reqA.Header = r.Header.Clone()
		reqB.Header = r.Header.Clone()

		// shadow request
		go func() {
			start := time.Now()
			respB, err := client.Do(reqB)
			elapsed := time.Since(start)
			atomic.AddInt64(&v2TotalLatency, elapsed.Milliseconds())

			if err != nil {
				return
			}
			defer respB.Body.Close()

			if respB.StatusCode >= 400 {
				atomic.AddInt32(&v2Error, 1)
			}
		}()

		// primary request
		start := time.Now()
		respA, err := client.Do(reqA)
		elapsed := time.Since(start)
		atomic.AddInt64(&v1TotalLatency, elapsed.Milliseconds())

		if err != nil {
			http.Error(w, "Primary server unavailable", http.StatusBadGateway)
			return
		}
		defer respA.Body.Close()
		if respA.StatusCode >= 400 {
			atomic.AddInt32(&v1Error, 1)
		}

		w.WriteHeader(respA.StatusCode)
		io.Copy(w, respA.Body)
	})

	server := http.Server{
		Addr:    ":" + strconv.Itoa(port),
		Handler: mux,
	}

	// preiodic aggregate metrics
	go func() {
		ticker := time.NewTicker(5 * time.Second)
		for range ticker.C {
			total := atomic.LoadInt32(&totalReq)
			e1 := atomic.LoadInt32(&v1Error)
			e2 := atomic.LoadInt32(&v2Error)

			l1 := atomic.LoadInt64(&v1TotalLatency)
			l2 := atomic.LoadInt64(&v2TotalLatency)

			if total > 0 {
				v1Rate := float64(e1) / float64(total) * 100
				v2Rate := float64(e2) / float64(total) * 100

				v1Avg := float64(l1) / float64(total)
				v2Avg := float64(l2) / float64(total)

				log.Printf("[METRICS] v1 (primary) errors: %d (%.1f%%) | avg latency: %.1fms\n", e1, v1Rate, v1Avg)
				log.Printf("[METRICS] v2 (shadow)  errors: %d (%.1f%%) | avg latency: %.1fms\n", e2, v2Rate, v2Avg)
			}
		}
	}()

	log.Println("starting server on port", port)
	server.ListenAndServe()
}
