package main

import (
	"log"
	"runtime"

	"github.com/nats-io/nats.go"
	"github.com/nats-io/nats.go/micro"
)

func main() {
	err := Connect()
	if err != nil {
		log.Fatal(err)
	}

	runtime.Goexit()
}

func Connect() error {
	nc, err := nats.Connect(nats.DefaultURL)
	if err != nil {
		return err
	}

	_, err = micro.AddService(nc, micro.Config{
		Name:        "service1",
		Description: "service1",
		Version:     "0.0.1",
		Endpoint: &micro.EndpointConfig{
			Subject: "service1",
			Handler: micro.HandlerFunc(func(req micro.Request) {
				//var r ServiceRequest
				//json.Unmarshal(r.Data(), &r)

			}),
			Metadata: nil,
		},
	})
	if err != nil {
		return err
	}

	return nil
}
