package main

import (
	"context"
	"fmt"
	"net/http"
	"strconv"
	"strings"

	"github.com/gin-gonic/gin"
	"github.com/jackc/pgx/v5"
)

func serverInit(ctx context.Context, db *pgx.Conn) {
	ginApp := gin.Default()
	ginApp.GET("/enriched-data/", getEnrichedData(db))
	server := &http.Server{
		Addr:    ":8080",
		Handler: ginApp,
	}
	go server.ListenAndServe()
	<-ctx.Done()
	server.Shutdown(context.Background())
	fmt.Printf("Server stopped.")
}

func getEnrichedData(db *pgx.Conn) gin.HandlerFunc {
	fn := func(c *gin.Context) {
		age := c.DefaultQuery("age", "")
		gender := c.DefaultQuery("gender", "")
		nationality := c.DefaultQuery("nationality", "")
		page := c.DefaultQuery("page", "1")

		pageInt, err := strconv.Atoi(page)
		if err != nil {
			c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid page parameter"})
			return
		}

		limit := 10
		offset := (pageInt - 1) * limit

		query := strings.Builder{}
		query.WriteString("SELECT * FROM enriched_data WHERE 1=1")
		args := []interface{}{}
		if age != "" {
			args = append(args, age)
			query.WriteString(" AND age = $")
			query.WriteString(strconv.Itoa(len(args)))
		}
		if gender != "" {
			args = append(args, gender)
			query.WriteString(" AND gender = $")
			query.WriteString(strconv.Itoa(len(args)))
		}
		if nationality != "" {
			args = append(args, nationality)
			query.WriteString(" AND nationality = $")
			query.WriteString(strconv.Itoa(len(args)))
		}
		args = append(args, limit)
		query.WriteString(" LIMIT $")
		query.WriteString(strconv.Itoa(len(args)))
		args = append(args, offset)
		query.WriteString(" OFFSET $")
		query.WriteString(strconv.Itoa(len(args)))
		fmt.Println(query.String())
		rows, err := db.Query(context.Background(), query.String(), args...)
		if err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to query database"})
			fmt.Printf("%v\n", err)
			return
		}
		defer rows.Close()

		results := []EnrichedData{}
		for rows.Next() {
			var data EnrichedData
			var id int
			err := rows.Scan(&id, &data.Name, &data.Surname, &data.Patronymic, &data.Age, &data.Gender, &data.Nationality)
			if err != nil {
				c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to scan rows"})
				return
			}
			results = append(results, data)
		}
		c.JSON(http.StatusOK, results)
	}
	return gin.HandlerFunc(fn)
}
