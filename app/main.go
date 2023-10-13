package main

import (
	"context"
	"fmt"

	"github.com/jackc/pgx/v5"
)

func initApp() {
	// os.Getenv("DATABASE_URL")
	databaseURL := "postgresql://postgres:123456@localhost:5432/EnrichedPeople?sslmode=disable"
	db, err := pgx.Connect(context.Background(), databaseURL)
	if err != nil {
		fmt.Printf("Failed to connect to db: %v\n", err)
		return
	}
	defer db.Close(context.Background())

	inTopic, outTopic := "FIO", "FIO_FAILED"
	brokers := []string{"0.0.0.0:9092"}

	dataChannel := make(chan BodyData, 1)
	dbChannel := make(chan EnrichedData, 1)
	errorsChannel := make(chan []byte, 1)

	go databaseListener(dbChannel, db)
	go enrichListener(dataChannel, dbChannel, errorsChannel)
	go kafkaErrorsHandler(brokers, outTopic, errorsChannel)
	go kafkaHandler(brokers, inTopic, dataChannel, errorsChannel)
	go serverInit(db)
	fmt.Scanln()
}

func main() {
	initApp()
}
