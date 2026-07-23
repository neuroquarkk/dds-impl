package main

import (
	"bytes"
	"encoding/json"
	"flag"
	"log"
	"math/rand/v2"
	"net/http"
	"strconv"
	"time"
)

func main() {
	var port int
	flag.IntVar(&port, "port", 8081, "port of the ambassador to send traffic to")
	flag.Parse()

	url := "http://localhost:" + strconv.Itoa(port)
	client := &http.Client{}

	payload := map[string]string{"foo": "bar"}
	data, _ := json.Marshal(payload)

	log.Println("starting client...")
	for {
		req, _ := http.NewRequest("POST", url, bytes.NewBuffer(data))
		resp, err := client.Do(req)
		if err != nil {
			log.Printf("request failed: %v\n", err)
		} else {
			log.Printf("response: %d\n", resp.StatusCode)
			resp.Body.Close()
		}

		t := time.Duration(rand.IntN(700-500) + 500)
		time.Sleep(t * time.Millisecond)
	}
}
