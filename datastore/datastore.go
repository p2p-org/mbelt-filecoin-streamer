package datastore

import (
	"context"
	"encoding/json"
	"errors"
	"github.com/p2p-org/mbelt-filecoin-streamer/config"
	"github.com/segmentio/kafka-go"
	"log"
	"time"
)

const (
	kafkaPartition         = 0
	kafkaWorkers           = 3
	kafkaPushChanBuffer    = 20000
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
	pushChan chan kafkaMessage
}

type kafkaMessage struct {
	topic   string
	message map[string]interface{}
}

func Init(config *config.Config, ctx context.Context) (*KafkaDatastore, error) {
	ds := &KafkaDatastore{
		config:       config,
		kafkaWriters: make(map[string]*kafka.Writer),
		pushChan:     make(chan kafkaMessage, kafkaPushChanBuffer),
	}

	for _, topic := range []string{TopicBlocks, TopicTipsetsToRevert, TopicMessages, TopicMessageReceipts, TopicTipSets,
		TopicActorStates, TopicMinerInfos, TopicMinerSectors, TopicRewardActorStates} {
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

	for workers := 0; workers < kafkaWorkers; workers++ {
		go ds.runPusher(ctx)
	}

	return ds, nil
}

func (ds *KafkaDatastore) runPusher(ctx context.Context) {
	timer := time.Tick(1 * time.Minute)
	for {
		select {
		case <-timer:
			log.Println("[KafkaDatastore][Debug][runPusher]", "kafka pusher chan size:", len(ds.pushChan))

		case <-ctx.Done():
			for m := range ds.pushChan {
				ds.push(m.topic, m.message)
			}
			return

		case m := <-ds.pushChan:
			ds.push(m.topic, m.message)
		}
	}
}

func (ds *KafkaDatastore) push(topic string, m map[string]interface{}) {
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

	err := ds.kafkaWriters[topic].WriteMessages(context.Background(), kMsgs...)

	if err != nil {
		log.Println("[KafkaDatastore][Error][runPusher]", "Cannot produce data", err)
	}
}

func (ds *KafkaDatastore) Push(topic string, m map[string]interface{}) (err error) {
	if ds.kafkaWriters == nil {
		log.Println("[KafkaDatastore][Error][push]", "Kafka writers not initialized")
		return errors.New("couldn't push messages to kafka, seems like writers not initialized")
	}

	if _, ok := ds.kafkaWriters[topic]; !ok {
		log.Println("[KafkaDatastore][Error][push]", "Kafka writer not initialized for topic", topic)
		return errors.New("couldn't push messages to kafka, seems like there is no writer for topic " + topic)
	}

	ds.pushChan <- kafkaMessage{topic: topic, message: m}
	return err
}
