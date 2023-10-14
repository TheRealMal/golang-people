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
	ginApp.POST("/enriched-data/", addEnrichedData(db))
	server := &http.Server{
		Addr:    ":8080",
		Handler: ginApp,
	}
	go server.ListenAndServe()
	<-ctx.Done()
	server.Shutdown(context.Background())
	fmt.Printf("Server stopped.")
}

func addEnrichedData(db *pgx.Conn) gin.HandlerFunc {
	fn := func(c *gin.Context) {

	}
	return gin.HandlerFunc(fn)
}

func sqlCheckAndWriteArg(arg string, argName string, args *[]interface{}, queryBuilder *strings.Builder) {
	if arg == "" {
		return
	}
	*args = append(*args, arg)
	queryBuilder.WriteString(" AND ")
	queryBuilder.WriteString(argName)
	queryBuilder.WriteString(" = $")
	queryBuilder.WriteString(strconv.Itoa(len(*args)))
}

func sqlWriteArg(arg int, argName string, args *[]interface{}, queryBuilder *strings.Builder) {
	*args = append(*args, arg)
	queryBuilder.WriteString(" ")
	queryBuilder.WriteString(argName)
	queryBuilder.WriteString(" $")
	queryBuilder.WriteString(strconv.Itoa(len(*args)))
}

func generateGetQuery(age string, gender string, nationality string, pageInt int) (string, []interface{}) {
	limit := 10
	offset := (pageInt - 1) * limit

	queryBuilder := strings.Builder{}
	queryBuilder.WriteString("SELECT * FROM enriched_data WHERE 1=1")
	args := []interface{}{}

	sqlCheckAndWriteArg(age, "age", &args, &queryBuilder)
	sqlCheckAndWriteArg(gender, "gender", &args, &queryBuilder)
	sqlCheckAndWriteArg(nationality, "nationality", &args, &queryBuilder)
	sqlWriteArg(limit, "LIMIT", &args, &queryBuilder)
	sqlWriteArg(offset, "OFFSET", &args, &queryBuilder)

	return queryBuilder.String(), args
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

		query, queryArgs := generateGetQuery(age, gender, nationality, pageInt)
		rows, err := db.Query(context.Background(), query, queryArgs...)
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
