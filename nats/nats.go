package main

import (
	"encoding/json"
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

type ServiceRequest struct {
	Text string `json:"text"`
}

type ServiceResponse struct {
	Text string `json:"text"`
}

func Connect() error {
	nc, err := nats.Connect(nats.DefaultURL)
	if err != nil {
		return err
	}

	micro.AddService(nc, micro.Config{
		Name:        "service1",
		Description: "service1",
		Version:     "0.0.1",
		Endpoint: &micro.EndpointConfig{
			Subject: "service1",
			Handler: micro.HandlerFunc(func(req micro.Request) {
				var r ServiceRequest
				err := json.Unmarshal(req.Data(), &r)
				if err != nil {
					req.Error("400", err.Error(), nil)
					return
				}
				req.RespondJSON(&ServiceResponse{""})
			}),
			Metadata: nil,
		},
	})

	log.Println("Listening in 'service1'", nc.ConnectedAddr())

	return nil
}
