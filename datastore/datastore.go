package datastore

import (
	"context"
	"encoding/json"
	"errors"
	"github.com/p2p-org/mbelt-filecoin-streamer/config"
	"github.com/segmentio/kafka-go"
	"github.com/segmentio/kafka-go/lz4"
	"log"
	"strings"
	"time"
)

const (
	kafkaPartition         = 0
	TopicBlocks            = "blocks_stream"
	TopicMessages          = "messages_stream"
	TopicMessageReceipts   = "message_receipts_stream"
	TopicTipSets           = "tipsets_stream"
	TopicTipsetsToRevert   = "tipsets_to_revert_stream"
	TopicActorStates       = "actor_states_stream"
	TopicMinerInfos        = "miner_infos_stream"
	TopicMinerSectors      = "miner_sectors_stream"
	TopicRewardActorStates = "reward_actor_states_stream"
)

type KafkaDatastore struct {
	config       *config.Config
	kafkaWriters map[string]*kafka.Writer
	// ack   chan kafka.Event
}

type kafkaMessage struct {
	topic   string
	message map[string]interface{}
}

func Init(config *config.Config) (*KafkaDatastore, error) {
	ds := &KafkaDatastore{
		config:       config,
		kafkaWriters: make(map[string]*kafka.Writer),
	}

	for _, topic := range []string{TopicBlocks, TopicTipsetsToRevert, TopicMessages, TopicMessageReceipts, TopicTipSets,
		TopicActorStates, TopicMinerInfos, TopicMinerSectors, TopicRewardActorStates} {
		topicWithPrefix := strings.ToUpper(config.KafkaPrefix + "_" + topic)
		var writer *kafka.Writer
		if config.KafkaAsyncWrite {
			writer = kafka.NewWriter(kafka.WriterConfig{
				Brokers:          []string{ds.config.KafkaHosts},
				Topic:            topicWithPrefix,
				CompressionCodec: &lz4.CompressionCodec{},
				Async:            true,
				MaxAttempts:      5,
				Balancer:         &kafka.RoundRobin{},
				QueueCapacity:    10000,
				BatchSize:        10000,
				BatchTimeout:     10 * time.Second,
				WriteTimeout:     15 * time.Second,
				RequiredAcks:     1,
			})
		} else {
			writer = kafka.NewWriter(kafka.WriterConfig{
				Brokers:          []string{ds.config.KafkaHosts},
				Topic:            topicWithPrefix,
				CompressionCodec: &lz4.CompressionCodec{},
				Async:            false,
				MaxAttempts:      5,
				Balancer:         &kafka.LeastBytes{},
			})
		}
		if writer == nil {
			return nil, errors.New("cannot create kafka writer")
		}
		ds.kafkaWriters[topic] = writer
	}

	return ds, nil
}

func (ds *KafkaDatastore) Push(topic string, m map[string]interface{}, ctx context.Context) {
	var kMsgs []kafka.Message

	for key, value := range m {
		data, err := json.Marshal(value)
		if err != nil {
			log.Println("[KafkaDatastore][Error][runPusher]", "Cannot marshal push data", err)
		}
		kMsgs = append(kMsgs, kafka.Message{
			Key:   []byte(key),
			Value: data,
		})
	}

	if len(kMsgs) == 0 {
		return
	}

	err := ds.kafkaWriters[topic].WriteMessages(ctx, kMsgs...)

	if err != nil {
		log.Println("[KafkaDatastore][Error][runPusher]", "Cannot produce data", err)
	}
}
