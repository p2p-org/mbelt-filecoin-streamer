package datastore

import (
	"context"
	"encoding/json"
	"errors"
	"log"
	"time"

	"github.com/sirupsen/logrus"

	"github.com/p2p-org/mbelt-filecoin-streamer/config"
	"github.com/segmentio/kafka-go"
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
	conn         *kafka.Conn
	// ack   chan kafka.Event
	// pushChan chan interface{}
}

func Init(config *config.Config) (*KafkaDatastore, error) {
	ds := &KafkaDatastore{
		config:       config,
		kafkaWriters: make(map[string]*kafka.Writer),
		// pushChan:     make(chan interface{}),
	}
	conn, err := kafka.Dial("tcp", ds.config.KafkaHosts)
	if err != nil {
		log.Println("[KafkaDatastore][Error][Init]", "kafka.Dial() error: ", err)
		return nil, err
	}
	ds.conn = conn
	for _, topic := range []string{TopicBlocks, TopicTipsetsToRevert, TopicMessages, TopicMessageReceipts, TopicTipSets,
		TopicActorStates, TopicMinerInfos, TopicMinerSectors, TopicRewardActorStates} {
		writer := kafka.NewWriter(kafka.WriterConfig{
			Brokers:      []string{ds.config.KafkaHosts},
			Topic:        topic,
			WriteTimeout: 1 * time.Second,
			ReadTimeout:  1 * time.Second,
			Balancer:     &kafka.LeastBytes{},
		})
		if writer == nil {
			return nil, errors.New("cannot create kafka writer")
		}

		//pp.Println(conn.Brokers())
		ds.kafkaWriters[topic] = writer
	}

	err = ds.testKafkaConnection()
	if err != nil {
		return nil, err
	}

	return ds, nil
}

func (kd *KafkaDatastore) testKafkaConnection() error {
	err := kd.conn.CreateTopics(kafka.TopicConfig{
		Topic:              "test_topic1233r452354afq34563",
		NumPartitions:      1,
		ReplicationFactor:  1,
		ReplicaAssignments: nil,
		ConfigEntries:      nil,
	})
	if err != nil {
		log.Println("[KafkaDatastore][Error][testKafkaConnection]", "CreateTopics error: ", err)
		return err
	}

	//err = kd.conn.DeleteTopics("test_topic123")
	//if err != nil {
	//	log.Println("[KafkaDatastore][Error][testKafkaConnection]", "DeleteTopics error: ", err)
	//	return err
	//}

	logrus.Info("kafka connection tested successfully")
	return nil
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
		log.Println("[KafkaDatastore][Error][push]", "Cannot write data", err)
	}
	return err
}
