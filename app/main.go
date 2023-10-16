package main

import (
	"context"
	"log"
	"os"
	"os/signal"
	"sync"

	"github.com/jackc/pgx/v5"
	"github.com/joho/godotenv"
	"github.com/redis/go-redis/v9"
)

const (
	workersNumber = 5
)

var l = log.New(os.Stdout, "[APP] ", 2)

func initApp() {
	godotenv.Load(".env")
	databaseURL := os.Getenv("DATABASE_URL")
	db, err := pgx.Connect(context.Background(), databaseURL)
	if err != nil {
		l.Fatalf("Failed to connect to db: %v\n", err)
	}
	defer db.Close(context.Background())

	rdb := redis.NewClient(&redis.Options{
		Addr:     os.Getenv("REDIS"),
		Password: "",
	})

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
	defer func() {
		<-terminationChannel
		l.Println("Received termination signal; Shutting down...")
		cancelCtx()
		wg.Wait()
	}()

	go func() {
		defer wg.Done()
		l.Println("Starting database goroutine")
		databaseListener(ctx, dbChannel, db, rdb)
	}()
	go func() {
		defer wg.Done()
		l.Println("Starting enricher goroutine")
		enrichListener(ctx, dataChannel, dbChannel, errorsChannel)
	}()
	go func() {
		defer wg.Done()
		l.Println("Starting kafka errors goroutine")
		kafkaErrorsHandler(ctx, brokers, outTopic, errorsChannel)
	}()
	go func() {
		defer wg.Done()
		l.Println("Starting kafka goroutine")
		kafkaHandler(ctx, brokers, inTopic, dataChannel, errorsChannel)
	}()
	go func() {
		defer wg.Done()
		l.Println("Starting server goroutine")
		serverInit(ctx, db, rdb, dbChannel)
	}()
}

func main() {
	initApp()
}
