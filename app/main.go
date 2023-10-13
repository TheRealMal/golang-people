package main

import "fmt"

func initApp() {
	inTopic, outTopic := "FIO", "FIO_FAILED"
	brokers := []string{"0.0.0.0:9092"}

	dataChannel := make(chan BodyData, 1)
	dbChannel := make(chan EnrichedData, 1)
	errorsChannel := make(chan []byte, 1)

	go enrichListener(dataChannel, dbChannel, errorsChannel)
	go kafkaErrorsHandler(brokers, outTopic, errorsChannel)
	go kafkaHandler(brokers, inTopic, dataChannel, errorsChannel)

	fmt.Scanln()
}

func main() {
	initApp()
}
