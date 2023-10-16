package main

import (
	"context"
	"fmt"

	"github.com/jackc/pgx/v5"
	"github.com/redis/go-redis/v9"
)

func insertEnrichedData(db *pgx.Conn, data *EnrichedData) error {
	query := "INSERT INTO enriched_data (name, surname, patronymic, age, gender, nationality) VALUES ($1, $2, $3, $4, $5, $6) RETURNING id"
	row := db.QueryRow(context.Background(), query, data.Name, data.Surname, data.Patronymic, data.Age, data.Gender, data.Nationality)

	var id int
	if err := row.Scan(&id); err != nil {
		return err
	}
	return nil
}

func databaseListener(ctx context.Context, dbChannel <-chan EnrichedData, db *pgx.Conn, rdb *redis.Client) {
	for {
		select {
		case data := <-dbChannel:
			if err := insertEnrichedData(db, &data); err != nil {
				fmt.Printf("Failed to insert row: %v\n", err)
				break
			}
			rdb.FlushDB(context.Background())
			l.Println("Successfully inserted data to db.")
		case <-ctx.Done():
			l.Printf("Database listener stopped.\n")
			return
		}
	}
}
