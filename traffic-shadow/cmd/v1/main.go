package main

import (
	"flag"
	"log"
	"math/rand/v2"
	"net/http"
	"strconv"
	"time"
)

func main() {
	var port int
	flag.IntVar(&port, "port", 8082, "port for the v1 primary server to listen on")
	flag.Parse()

	mux := http.NewServeMux()

	mux.HandleFunc("POST /", func(w http.ResponseWriter, r *http.Request) {
		// fast (20-40ms) and reliable (10% simulated error)
		t := time.Duration(rand.IntN(40-20) + 20)
		time.Sleep(t * time.Millisecond)

		e := rand.Float32()
		if e < 0.1 {
			http.Error(w, "simulated error", http.StatusInternalServerError)
			return
		}

		w.WriteHeader(http.StatusOK)
		w.Write([]byte("OK"))
	})

	server := http.Server{
		Addr:    ":" + strconv.Itoa(port),
		Handler: mux,
	}

	log.Println("starting server on port", port)
	if err := server.ListenAndServe(); err != nil {
		log.Fatalf("failed to start server: %v\n", err)
	}
}
