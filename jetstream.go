package main

import (
	"log"

	"github.com/nats-io/nats.go"
)

func JetStreamInit(nc *nats.Conn) (nats.JetStreamContext, error) {
	// Create JetStream Context
	js, err := nc.JetStream(nats.PublishAsyncMaxPending(256))
	if err != nil {
		return nil, err
	}

	// Create a stream if it does not exist
	err = CreateStream(js)
	if err != nil {
		return nil, err
	}
	return js, nil
}

func CreateStream(jetStream nats.JetStreamContext) error {
	stream, err := jetStream.StreamInfo(NTPStreamName)

	// stream not found, create it
	if stream == nil {
		log.Printf("Creating stream: %s\n", NTPStreamName)

		_, err = jetStream.AddStream(&nats.StreamConfig{
			Name:     NTPStreamName,
			Subjects: []string{NTPStreamSubjects},
		})
		if err != nil {
			return err
		}
	}
	return nil
}
