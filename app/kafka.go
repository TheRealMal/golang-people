package main

import (
	"encoding/json"
	"fmt"
	"log"
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
func kafkaHandler(topic string, brokers []string, dataChannel chan<- BodyData, errorsChannel chan<- []byte) {
	// Create new Kafka consumer
	consumer, err := sarama.NewConsumer(brokers, nil)
	if err != nil {
		log.Fatalf("Error creating Kafka consumer: %v", err)
	}
	defer consumer.Close()
	// Create new Kafka topic consumer
	partitionConsumer, err := consumer.ConsumePartition(topic, 0, sarama.OffsetNewest)
	if err != nil {
		log.Fatalf("Error creating Kafka partition consumer: %v", err)
	}
	defer partitionConsumer.Close()

	// Channel for termination signal
	sigterm := make(chan os.Signal, 1)
	signal.Notify(sigterm, os.Interrupt)
	fmt.Printf("Waiting for Kafka messages...")
	for {
		select {
		case msg := <-partitionConsumer.Messages():
			fmt.Printf("Received message: %s\n", msg.Value)
			var bodyFIO BodyData
			if err := json.Unmarshal(msg.Value, &bodyFIO); err != nil {
				errorMsg := "Failed while parsing json"
				body := string(msg.Value[:])
				log.Printf("%s: %v", errorMsg, err)
				jsonErrorMsg, _ := json.Marshal(ErrorData{
					ErrorMsg: &errorMsg,
					Body:     &body,
				})
				errorsChannel <- jsonErrorMsg
				// sendErrorToQueue(&producer, outTopic, jsonErrorMsg)
				continue
			}
			if bodyFIO.Name == nil || bodyFIO.Surname == nil {
				errorMsg := "Required fields not found"
				bodyBytes, _ := json.Marshal(bodyFIO)
				body := string(bodyBytes[:])
				log.Printf("%s", errorMsg)
				jsonErrorMsg, _ := json.Marshal(ErrorData{
					ErrorMsg: &errorMsg,
					Body:     &body,
				})
				errorsChannel <- jsonErrorMsg
				// sendErrorToQueue(&producer, outTopic, jsonErrorMsg)
				continue
			}

		case <-sigterm:
			fmt.Println("Received SIGINT. Shutting down.")
			return
		}
	}
}

func kafkaErrorsHandler(brokers []string, topic string, errorsChannel <-chan []byte) {
	// Create new Kafka producer
	producer, err := sarama.NewSyncProducer(brokers, nil)
	if err != nil {
		log.Fatalf("Error creating Kafka producer: %v", err)
	}
	defer producer.Close()
	for {
		select {
		case msg := <-errorsChannel:
			sendErrorToQueue(&producer, topic, msg)
		}
	}
}

// Sends error message to Kafka queue
func sendErrorToQueue(producer *sarama.SyncProducer, topic string, data []byte) {
	failedMsg := &sarama.ProducerMessage{
		Topic: topic,
		Value: sarama.StringEncoder(string(data)),
	}
	if _, _, err := (*producer).SendMessage(failedMsg); err != nil {
		log.Printf("Failed to send message to %s queue: %v", topic, err)
	}
}
