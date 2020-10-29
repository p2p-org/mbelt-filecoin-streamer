package datastore

import (
	"context"
	"encoding/json"
	"errors"
	"github.com/p2p-org/mbelt-filecoin-streamer/config"
	"github.com/segmentio/kafka-go"
	"log"
)

const (
	kafkaPartition       = 0
	TopicBlocks          = "blocks_stream"
	TopicMessages        = "messages_stream"
	TopicTipSets         = "tipsets_stream"
	TopicTipsetsToRevert = "tipsets_to_revert_stream"
	TopicActorStates     = "actor_states_stream"
)

type KafkaDatastore struct {
	config       *config.Config
	kafkaWriters map[string]*kafka.Writer
	// ack   chan kafka.Event
	// pushChan chan interface{}
}

func Init(config *config.Config) (*KafkaDatastore, error) {
	ds := &KafkaDatastore{
		config:       config,
		kafkaWriters: make(map[string]*kafka.Writer),
		// pushChan:     make(chan interface{}),
	}

	for _, topic := range []string{TopicBlocks, TopicTipsetsToRevert, TopicMessages, TopicTipSets, TopicActorStates} {
		writer := kafka.NewWriter(kafka.WriterConfig{
			Brokers:  []string{ds.config.KafkaHosts},
			Topic:    topic,
			Balancer: &kafka.LeastBytes{},
		})
		if writer == nil {
			return nil, errors.New("cannot create kafka writer")
		}
		ds.kafkaWriters[topic] = writer
	}

	return ds, nil
}

func (ds *KafkaDatastore) Push(topic string, m map[string]interface{}) (err error) {
	var (
		kMsgs []kafka.Message
	)
	// log.Println("[KafkaDatastore][push][Debug] Push data to kafka")

	if ds.kafkaWriters == nil {
		log.Println("[KafkaDatastore][Error][push]", "Kafka writers not initialized")
		return errors.New("cannot push")
	}

	if _, ok := ds.kafkaWriters[topic]; !ok {
		log.Println("[KafkaDatastore][Error][push]", "Kafka writer not initialized for topic", topic)
	}

	for key, value := range m {
		data, err := json.Marshal(value)
		if err != nil {
			log.Println("[KafkaDatastore][Error][push]", "Cannot marshal push data", err)
			return errors.New("cannot push")
		}
		kMsgs = append(kMsgs, kafka.Message{
			Key:   []byte(key),
			Value: data,
		})
	}

	if len(kMsgs) == 0 {
		return nil
	}

	err = ds.kafkaWriters[topic].WriteMessages(context.Background(), kMsgs...)

	if err != nil {
		log.Println("[KafkaDatastore][Error][push]", "Cannot produce data", err)
	}
	return err
}
