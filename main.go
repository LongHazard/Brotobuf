package main

import (
	"log"
	"sync"
)

func main() {
	log.Println("Starting...")
	// Let's assume that publisher and consumer are services running on different servers.
	// So run publisher and consumer asynchronously to see how it works
	var wg sync.WaitGroup

	wg.Add(1)
	go func() {
		defer wg.Done()
		KeepAliveSubscribe("event")
	}()

	wg.Wait()

	log.Println("Exit...")

}
