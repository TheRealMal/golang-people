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

type EnrichedDataWithID struct {
	EnrichedData
	ID int
}

func serverInit(ctx context.Context, db *pgx.Conn, dbChannel chan<- EnrichedData) {
	ginApp := gin.Default()
	ginApp.GET("/enriched-data/", getEnrichedData(db))
	ginApp.POST("/enriched-data/", addEnrichedData(dbChannel))
	ginApp.DELETE("/enriched-data/:id", delEnrichedData(db))
	server := &http.Server{
		Addr:    ":8080",
		Handler: ginApp,
	}
	go server.ListenAndServe()
	<-ctx.Done()
	server.Shutdown(context.Background())
	fmt.Printf("Server stopped.")
}

func delEnrichedData(db *pgx.Conn) gin.HandlerFunc {
	fn := func(c *gin.Context) {
		id := c.Param("id")
		if _, err := strconv.Atoi(id); err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": "Provided ID type is not int"})
			fmt.Printf("%v\n", err)
			return
		}
		query := "DELETE FROM enriched_data WHERE id = $1"
		res, err := db.Exec(context.Background(), query, id)
		if err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to delete row", "message": err.Error()})
			fmt.Printf("%v\n", err)
			return
		}
		if res.RowsAffected() == 0 {
			c.JSON(http.StatusNotFound, gin.H{"error": "Provided ID not found"})
			fmt.Printf("%v\n", err)
			return
		}
		c.JSON(http.StatusOK, gin.H{"status": "ok"})
	}
	return gin.HandlerFunc(fn)
}

func addEnrichedData(dbChannel chan<- EnrichedData) gin.HandlerFunc {
	fn := func(c *gin.Context) {
		requestBody := new(EnrichedData)
		if err := c.Bind(requestBody); err != nil {
			c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid request body", "message": err.Error()})
			return
		}
		dbChannel <- *requestBody
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

		results := []EnrichedDataWithID{}
		for rows.Next() {
			var data EnrichedDataWithID
			err := rows.Scan(&data.ID, &data.Name, &data.Surname, &data.Patronymic, &data.Age, &data.Gender, &data.Nationality)
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
