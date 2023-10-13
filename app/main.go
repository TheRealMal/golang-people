package main

import "fmt"

func main() {
	inTopic, outTopic := "FIO", "FIO_FAILED"
	brokers := []string{"0.0.0.0:9092"}
	go kafkaHandler(inTopic, outTopic, brokers)
	fmt.Scanln()
}
