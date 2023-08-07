package main

import (
	"log"
	"runtime"

	"github.com/nats-io/nats.go"
	"github.com/nats-io/nats.go/micro"
)

type ServiceRequest struct {
}

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

	schema, err := Schema()
	if err != nil {
		return err
	}

	micro.AddService(nc, micro.Config{
		Name:        "service1",
		Description: "service1",
		Version:     "0.0.1",
		Endpoint: &micro.EndpointConfig{
			Subject: "service1",
			Schema:  schema,
			Handler: micro.HandlerFunc(func(req micro.Request) {
				var r ServiceRequest
				json.Unmasrshal(r.Data(), &r)

			}),
		},
	})

	return nil
}
