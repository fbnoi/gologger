package main

import (
	"log"
	"time"
)

func main() {
	timeout := time.NewTimer(1000 * time.Millisecond)
	select {
	case <-timeout.C:
		log.Println("timeout")
	default:
		time.Sleep(2000 * time.Millisecond)
		log.Println("done")
	}
}
