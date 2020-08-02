package datastore

import (
	"encoding/json"
	"errors"
	"github.com/filecoin-project/lotus/chain/types"
	"github.com/p2p-org/mbelt-filecoin-streamer/config"
	"gopkg.in/confluentinc/confluent-kafka-go.v1/kafka"
	"log"
	"time"
)

const (
	topicBlocks   = "blocks_stream"
	topicMessages = "messages_stream"
)

type Datastore struct {
	config   *config.Config
	producer *kafka.Producer
	// ack   chan kafka.Event
	pushChan chan interface{}
}

func Init(config *config.Config) (*Datastore, error) {
	return &Datastore{
		config: config,
	}, nil
}

func (ds *Datastore) SetupProducer() error {
	var err error
	if ds.producer != nil {
		return nil
	}

	// https://github.com/edenhill/librdkafka/blob/master/CONFIGURATION.md
	ds.producer, err = kafka.NewProducer(&kafka.ConfigMap{
		"client.id":         "filecoin-analytics",
		"bootstrap.servers": ds.config.KafkaHosts,
	})
	if err != nil {
		log.Println("[Datastore][Error][SetupProducer]", err)
		return errors.New("cannot setup producer")
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
			case *kafka.Error:
				log.Println("[Datastore][Error][Kafka]", "Kafka error", ev.String())
			default:
				log.Println("[Datastore][Debug][Kafka][Event]", "Kafka event", ev.String())
			}
		}
	}()

	ds.pushChan = make(chan interface{})

	go func() {
		for push := range ds.pushChan {
			for attempts := 5; attempts > 0; attempts-- {
				err := ds.push(push)
				if err == nil {
					break
				}
				log.Println("[Datastore][Debug]", "Kafka push timeout")
				time.Sleep(time.Millisecond * 200)
			}
		}
	}()
	return nil
}

func (ds *Datastore) initConsumer(topic string) (*kafka.Consumer, error) {
	consumer, err := kafka.NewConsumer(&kafka.ConfigMap{
		"bootstrap.servers": ds.config.KafkaHosts,
		"group.id":          "filecoin-analytics-group",
		"auto.offset.reset": "earliest",
	})

	if err != nil {
		log.Println("[Datastore][Error][SetupConsumers]", err)
		return nil, errors.New("cannot setup consumers")
	}

	consumer.Subscribe(topic, nil)

	if err != nil {
		log.Println("[Datastore][Error][SetupConsumers]", err)
		return nil, errors.New("cannot subsribe topic")
	}

	return consumer, nil
}

func (ds *Datastore) InitBlocksTopic() (*kafka.Consumer, error) {
	return ds.initConsumer(topicBlocks)
}

func (ds *Datastore) InitMessageTopic() (*kafka.Consumer, error) {
	return ds.initConsumer(topicMessages)
}

func (ds *Datastore) Push(i interface{}) {
	if ds.pushChan == nil {
		log.Println("[Datastore][Error][Push]", "Push channel not initialized")
	}
	if ds.producer == nil {
		log.Println("[Datastore][Error][Push]", "Kafka producer not initialized")
	}

	ds.pushChan <- i
}

func (ds *Datastore) push(i interface{}) (err error) {
	var topic string
	// log.Println("[Datastore][push][Debug] Push data to kafka")

	if ds.producer == nil {
		log.Println("[Datastore][Error][push]", "Kafka producer not initialized")
		return errors.New("cannot push")
	}

	switch i.(type) {
	case types.BlockHeader:
		topic = topicBlocks
	case types.Message:
		topic = topicMessages
	default:
		log.Println("[Datastore][Error][push]", "Unsupported struct")
		return errors.New("cannot push")
	}

	data, err := json.Marshal(i)
	if err != nil {
		log.Println("[Datastore][Error][push]", "Cannot marshal push data", err)
		return errors.New("cannot push")
	}

	log.Println("[Datastore][Debug][push]", "Push to topic", topic)

	// Check gor working
	defer func() {
		if r := recover(); r != nil {
			err = errors.New("kafka failed")
		}
	}()

	err = ds.producer.Produce(&kafka.Message{
		TopicPartition: kafka.TopicPartition{Topic: &topic, Partition: kafka.PartitionAny},
		Value:          data,
	}, nil)

	if err != nil {
		log.Println("[Datastore][Error][push]", "Cannot produce data", err)
	}
	return err
}
