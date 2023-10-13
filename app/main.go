package main

import "fmt"

func initApp() {
	inTopic, outTopic := "FIO", "FIO_FAILED"
	brokers := []string{"0.0.0.0:9092"}
	dataChannel := make(chan BodyData)
	go kafkaHandler(inTopic, outTopic, brokers, dataChannel)
	fmt.Scanln()
}
