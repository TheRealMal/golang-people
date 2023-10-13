package main

import "fmt"

func initApp() {
	inTopic, outTopic := "FIO", "FIO_FAILED"
	brokers := []string{"0.0.0.0:9092"}

	dataChannel := make(chan BodyData)
	dbChannel := make(chan EnrichedData)
	errorsChannel := make(chan []byte)

	go kafkaErrorsHandler(brokers, outTopic, errorsChannel)
	go enrichListener(dataChannel, dbChannel, errorsChannel)
	go kafkaHandler(inTopic, brokers, dataChannel, errorsChannel)

	fmt.Scanln()
}
