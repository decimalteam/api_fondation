package main

import (
	"encoding/json"
	"log"
	"runtime"

	"github.com/nats-io/nats.go"
	"github.com/nats-io/nats.go/micro"
)

func main() {
	err := Connect("service1", "service1 description", "0.0.1", "service1subject")
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

func Connect(svcName, svcDescription, svcVersion, svcSubject string) error {
	nc, err := nats.Connect(nats.DefaultURL)
	if err != nil {
		return err
	}

	_, err = micro.AddService(nc, micro.Config{
		Name:        svcName,
		Description: svcDescription,
		Version:     svcVersion,
		Endpoint: &micro.EndpointConfig{
			Subject: svcSubject,
			Handler: micro.HandlerFunc(func(req micro.Request) {
				var r ServiceRequest
				err = json.Unmarshal(req.Data(), &r)
				if err != nil {
					err = req.Error("400", err.Error(), nil)
					if err != nil {
						return
					}
				}
				err = req.RespondJSON(&ServiceResponse{"response"})
				if err != nil {
					return
				}
			}),
			Metadata: nil,
		},
	})
	if err != nil {
		return err
	}

	log.Println("Listening on 'service1'", nc.ConnectedAddr())

	return nil
}
