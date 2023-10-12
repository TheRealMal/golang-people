package main

func main() {
	topic := "FIO"
	brokers := []string{"localhost:9092"}
	kafkaListener(topic, brokers)
}
