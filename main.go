package main

import (
	"fmt"
	"io"
	"net/http"
	"sync"
	"time"
)

var maxInFlightRequests = 2
var tokens = make(chan struct{}, maxInFlightRequests)

func hello(w http.ResponseWriter, req *http.Request) {

	// requests limiting logic
	tokens <- struct{}{}
	defer func() {
		<-tokens
	}()

	// takes 5 seconds
	process()

	// write response
	fmt.Fprintf(w, "%s", time.Now())
}

func main() {
	wg := startServer("/hello", hello, ":8080")
	// keep the server running until Ctrl+C is pressed
	defer wg.Wait()

	// 4 is the number of concurrent requests
	sendNRequests("http://localhost:8080/hello", 4)

}

func process() {
	// some processing
	time.Sleep(5 * time.Second)
}

func startServer(url string, handler func(w http.ResponseWriter, req *http.Request), port string) *sync.WaitGroup {
	http.HandleFunc(url, handler)

	var wg sync.WaitGroup

	wg.Add(1)
	go func() {
		http.ListenAndServe(port, nil)
		wg.Done()
	}()
	return &wg
}

func sendNRequests(url string, count int) {
	for i := 0; i < count; i++ {

		go func() {
			r, err := http.Get(url)
			if err != nil {
				panic(err)
			}
			body, err := io.ReadAll(r.Body)
			r.Body.Close()
			if err != nil {
				panic(err)
			}

			fmt.Println(string(body))
		}()
	}
}
