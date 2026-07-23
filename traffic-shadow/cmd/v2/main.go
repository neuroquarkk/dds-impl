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
	flag.IntVar(&port, "port", 8083, "port for the v2 primary server to listen on")
	flag.Parse()

	mux := http.NewServeMux()

	mux.HandleFunc("POST /", func(w http.ResponseWriter, r *http.Request) {
		// slow (60-100) and reliable (40% simulated error)
		t := time.Duration(rand.IntN(100-60) + 60)
		time.Sleep(t * time.Millisecond)

		e := rand.Float32()
		if e < 0.4 {
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
