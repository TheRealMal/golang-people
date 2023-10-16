package main

import (
	"context"
	"encoding/json"
	"os"
	"os/signal"

	"github.com/IBM/sarama"
)

type BodyData struct {
	Name       *string `json:"name,omitempty"`
	Surname    *string `json:"surname,omitempty"`
	Patronymic *string `json:"patronymic,omitempty"`
}

type ErrorData struct {
	Body     *string `json:"body,omitempty"`
	ErrorMsg *string `json:"error,omitempty"`
}

// Creates Kafka consumer for FIO topic and producer for FIO_FAILED
// If consumer receives message with wrong format producer sends
// error message to FIO_FAILED topic
func kafkaHandler(ctx context.Context, brokers []string, topic string, dataChannel chan<- BodyData, errorsChannel chan<- []byte) {
	// Create new Kafka consumer
	consumer, err := sarama.NewConsumer(brokers, nil)
	if err != nil {
		l.Fatalf("Error creating Kafka consumer: %v", err)
	}
	defer consumer.Close()
	// Create new Kafka topic consumer
	partitionConsumer, err := consumer.ConsumePartition(topic, 0, sarama.OffsetNewest)
	if err != nil {
		l.Fatalf("Error creating Kafka partition consumer: %v", err)
	}
	defer partitionConsumer.Close()

	// Channel for termination signal
	sigterm := make(chan os.Signal, 1)
	signal.Notify(sigterm, os.Interrupt)
	for {
		select {
		case msg := <-partitionConsumer.Messages():
			l.Printf("Received message: %s\n", msg.Value)
			var bodyFIO BodyData
			if err := json.Unmarshal(msg.Value, &bodyFIO); err != nil {
				errorsChannel <- prepareErrorBytes[BodyData]("Failed while parsing json", &bodyFIO)
				continue
			}
			if bodyFIO.Name == nil || bodyFIO.Surname == nil {
				errorsChannel <- prepareErrorBytes[BodyData]("Required fields not found", &bodyFIO)
				continue
			}
			if bodyFIO.Patronymic == nil {
				tmp := ""
				bodyFIO.Patronymic = &tmp
			}
			dataChannel <- bodyFIO
		case <-ctx.Done():
			l.Println("Kafka queue listener stopped.")
			return
		}
	}
}

// Returns bytes array for given message and body
// that led to error
func prepareErrorBytes[T any](message string, data *T) []byte {
	bodyBytes, _ := json.Marshal(*data)
	body := string(bodyBytes[:])
	jsonErrorMsg, _ := json.Marshal(ErrorData{
		ErrorMsg: &message,
		Body:     &body,
	})
	return jsonErrorMsg
}

// Sends error message to Kafka queue
func sendErrorToQueue(producer *sarama.SyncProducer, topic string, data []byte) {
	failedMsg := &sarama.ProducerMessage{
		Topic: topic,
		Value: sarama.StringEncoder(string(data)),
	}
	if _, _, err := (*producer).SendMessage(failedMsg); err != nil {
		l.Printf("Failed to send message to %s queue: %v\n", topic, err)
	}
	l.Printf("Successfully sent error to %s\n queue", topic)
}

// Listens error channel
// Produces message to provided Kafka topic
func kafkaErrorsHandler(ctx context.Context, brokers []string, topic string, errorsChannel <-chan []byte) {
	producer, err := sarama.NewSyncProducer(brokers, nil)
	if err != nil {
		l.Fatalf("Error creating Kafka producer: %v", err)
	}
	defer producer.Close()
	for {
		select {
		case msg := <-errorsChannel:
			sendErrorToQueue(&producer, topic, msg)
		case <-ctx.Done():
			l.Println("Kafka errors listener stopped.")
			return
		}
	}
}
