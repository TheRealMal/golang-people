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

func kafkaListener(topic string, brokers []string) {
	fmt.Println(topic, brokers)
	// Настройки Kafka
	kafkaConfig := sarama.NewConfig()
	kafkaConfig.Consumer.Offsets.Initial = sarama.OffsetOldest

	// Создаем нового потребителя Kafka
	consumer, err := sarama.NewConsumer(brokers, kafkaConfig)
	if err != nil {
		log.Fatalf("Error creating Kafka consumer: %s", err)
	}
	defer consumer.Close()

	// Создаем потребителя для темы
	partitionConsumer, err := consumer.ConsumePartition(topic, 0, sarama.OffsetNewest)
	if err != nil {
		log.Fatalf("Error creating Kafka partition consumer: %s", err)
	}
	defer partitionConsumer.Close()

	// Канал для ожидания сигнала завершения
	sigterm := make(chan os.Signal, 1)
	signal.Notify(sigterm, os.Interrupt)

	for {
		select {
		case msg := <-partitionConsumer.Messages():
			// Обработка сообщения
			fmt.Printf("Received message: %s\n", msg.Value)

			// Попробуем разобрать JSON сообщения
			var fioData FIOData
			if err := json.Unmarshal(msg.Value, &fioData); err != nil {
				// Если сообщение не является валидным JSON, отправляем его в очередь FIO_FAILED
				log.Printf("Error parsing JSON: %s", err)
				sendToFailedQueue(msg.Value, brokers)
				continue
			}

			// Далее можно продолжить с обогащением и сохранением данных в БД
		case <-sigterm:
			// Завершение приложения при получении сигнала завершения
			fmt.Println("Received SIGINT. Shutting down.")
			return
		}
	}
}

func sendToFailedQueue(data []byte, brokers []string) {
	// Создаем продюсера для отправки сообщений в очередь FIO_FAILED
	failedProducer, err := sarama.NewSyncProducer(brokers, nil)
	if err != nil {
		log.Fatalf("Error creating Kafka producer: %s", err)
	}
	defer failedProducer.Close()

	// Создаем сообщение и отправляем его в очередь FIO_FAILED
	failedMsg := &sarama.ProducerMessage{
		Topic: "FIO_FAILED",
		Value: sarama.StringEncoder(string(data)),
	}

	if _, _, err := failedProducer.SendMessage(failedMsg); err != nil {
		log.Printf("Failed to send message to FIO_FAILED queue: %s", err)
	}
}
