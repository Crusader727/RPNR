package main

import (
	"log"

	"./server"
)

func main() {
	log.Fatal(server.RunHTTPServer())
}
