package datastore

import (
	"errors"
	"gopkg.in/confluentinc/confluent-kafka-go.v1/kafka"
	"log"
)

var (
	topicName = "filecoinBlocks"
)

type Datastore struct {
	producer *kafka.Producer
}

func Init(host string) (*Datastore, error) {
	var (
		ds  Datastore
		err error
	)
	ds.producer, err = kafka.NewProducer(&kafka.ConfigMap{
		"client.id":         "filecoin-analytics",
		"bootstrap.servers": host,
	})
	if err != nil {
		log.Println("[Datastore][Error][Init]", err)
		return nil, errors.New("cannot init config")
	}

	defer ds.producer.Close()

	// Delivery report handler for produced messages
	go func() {
		for e := range ds.producer.Events() {
			switch ev := e.(type) {
			case *kafka.Message:
				if ev.TopicPartition.Error != nil {
					log.Println("[Datastore][Debug]", "Delivery failed:", ev.TopicPartition)
				} else {
					log.Println("[Datastore][Debug]", "Delivered message to", ev.TopicPartition)
				}
			}
		}
	}()

	return &ds, nil
}

func (ds *Datastore) Push(m string) error {
	return ds.producer.Produce(&kafka.Message{
		TopicPartition: kafka.TopicPartition{Topic: &topicName, Partition: kafka.PartitionAny},
		Value:          []byte(m),
	}, nil)
}
