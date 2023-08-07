package main

import (
	"log"

	"github.com/nats-io/nats.go"
)

func main() {
	err := Connect()
	if err != nil {
		log.Fatal(err)
	}
}

func Connect() error {
	_, err := nats.Connect(nats.DefaultURL)
	if err != nil {
		return err
	}

	return nil
}
