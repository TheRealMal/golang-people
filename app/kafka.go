package main

import (
	"encoding/json"
	"fmt"
	"log"
	"os"
	"os/signal"

	"github.com/IBM/sarama"
)

type FIOData struct {
	Name       string `json:"name"`
	Surname    string `json:"surname"`
	Patronymic string `json:"patronymic,omitempty"`
}

/*
	Creates Kafka consumer for FIO topic and producer for FIO_FAILED
	If consumer receives message with wrong format producer sends
	error message to FIO_FAILED topic
*/
func kafkaHandler(inTopic string, outTopic string, brokers []string) {
	// Create new Kafka producer ()
	producer, err := sarama.NewSyncProducer(brokers, nil)
	if err != nil {
		log.Fatalf("Error creating Kafka producer: %v", err)
	}
	defer producer.Close()
	// Create new Kafka consumer
	consumer, err := sarama.NewConsumer(brokers, kafkaConfig)
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
	go func(){
		for {
			select {
			case msg := <-partitionConsumer.Messages():
				fmt.Printf("Received message: %s\n", msg.Value)
				var fioData FIOData
				if err := json.Unmarshal(msg.Value, &fioData); err != nil {
					log.Printf("Error parsing JSON: %v", err)
					sendToFailedQueue(&producer, outTopic, msg.Value)
					continue
				}
			case <-sigterm:
				fmt.Println("Received SIGINT. Shutting down.")
				return
			}
		}
	}
}

func sendFailToQueue(producer *SyncProducer, topic string, data []byte) {
	// Создаем продюсера для отправки сообщений в очередь FIO_FAILED
	failedProducer, err := sarama.NewSyncProducer(brokers, nil)
	if err != nil {
		log.Fatalf("Error creating Kafka producer: %v", err)
	}
	defer failedProducer.Close()

	// Create message for FIO_FAILED topic
	failedMsg := &sarama.ProducerMessage{
		Topic: topic,
		Value: sarama.StringEncoder(string(data)),
	}

	if _, _, err := failedProducer.SendMessage(failedMsg); err != nil {
		log.Printf("Failed to send message to %s queue: %v", topic, err)
	}
}
