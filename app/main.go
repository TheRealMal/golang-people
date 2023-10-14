package main

import (
	"context"
	"fmt"
	"log"
	"os"
	"os/signal"
	"sync"

	"github.com/jackc/pgx/v5"
	"github.com/joho/godotenv"
)

const workersNumber = 5

func initApp() {
	godotenv.Load(".env")
	databaseURL := os.Getenv("DATABASE_URL")
	db, err := pgx.Connect(context.Background(), databaseURL)
	if err != nil {
		log.Fatalf("Failed to connect to db: %v\n", err)
	}
	defer db.Close(context.Background())

	inTopic, outTopic := os.Getenv("INPUT_TOPIC"), os.Getenv("OUTPUT_TOPIC")
	brokers := []string{os.Getenv("KAFKA_BROKER")}

	dataChannel := make(chan BodyData, 1)
	dbChannel := make(chan EnrichedData, 1)
	errorsChannel := make(chan []byte, 1)

	// Termination part
	terminationChannel := make(chan os.Signal)
	signal.Notify(terminationChannel, os.Interrupt)
	ctx, cancelCtx := context.WithCancel(context.Background())
	wg := &sync.WaitGroup{}
	wg.Add(5)

	go func() {
		defer wg.Done()
		databaseListener(ctx, dbChannel, db)
	}()
	go func() {
		defer wg.Done()
		enrichListener(ctx, dataChannel, dbChannel, errorsChannel)
	}()
	go func() {
		defer wg.Done()
		kafkaErrorsHandler(ctx, brokers, outTopic, errorsChannel)
	}()
	go func() {
		defer wg.Done()
		kafkaHandler(ctx, brokers, inTopic, dataChannel, errorsChannel)
	}()
	go func() {
		defer wg.Done()
		serverInit(ctx, db, dbChannel)
	}()
	<-terminationChannel
	fmt.Printf("Received termination signal; Shutting down...\n")
	cancelCtx()
	wg.Wait()
}

func main() {
	initApp()
}
