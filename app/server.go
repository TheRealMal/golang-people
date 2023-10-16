package main

import (
	"app/graph"
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"strconv"
	"strings"
	"time"

	"github.com/99designs/gqlgen/graphql/handler"
	"github.com/99designs/gqlgen/graphql/playground"
	"github.com/gin-gonic/gin"
	"github.com/jackc/pgx/v5"
	"github.com/redis/go-redis/v9"
)

type EnrichedDataWithID struct {
	EnrichedData
	ID int
}

func serverInit(ctx context.Context, db *pgx.Conn, rdb *redis.Client, dbChannel chan<- EnrichedData) {
	router := gin.Default()

	router.GET("/enriched-data/", checkCache(rdb), getEnrichedData(db, rdb))
	router.POST("/enriched-data/", addEnrichedData(dbChannel, rdb))
	router.DELETE("/enriched-data/:id", delEnrichedData(db, rdb))
	router.PUT("/enriched-data/:id", updateEnrichedData(db, rdb))

	router.POST("/query", graphqlHandler())
	router.GET("/", playgroundHandler())

	server := &http.Server{
		Addr:    ":8080",
		Handler: router,
	}
	go server.ListenAndServe()

	<-ctx.Done()
	server.Shutdown(context.Background())
	l.Println("Server stopped.")
}

func checkCache(rdb *redis.Client) gin.HandlerFunc {
	fn := func(c *gin.Context) {
		cacheKey := strings.Builder{}
		cacheKey.WriteString(c.DefaultQuery("age", ""))
		cacheKey.WriteString(c.DefaultQuery("gender", ""))
		cacheKey.WriteString(c.DefaultQuery("nationality", ""))
		cacheKey.WriteString(c.DefaultQuery("page", "1"))
		val, err := rdb.Get(context.Background(), cacheKey.String()).Bytes()
		if err != nil {
			return
		}
		response := []EnrichedDataWithID{}
		json.Unmarshal(val, &response)
		c.JSON(http.StatusOK, response)
		c.Abort()
	}
	return gin.HandlerFunc(fn)
}

func cacheResponse(a *string, g *string, n *string, p *string, response *[]EnrichedDataWithID, rdb *redis.Client) error {
	cacheKey := strings.Builder{}
	cacheKey.WriteString(*a)
	cacheKey.WriteString(*g)
	cacheKey.WriteString(*n)
	cacheKey.WriteString(*p)
	resultsJSON, err := json.Marshal(*response)
	if err != nil {
		l.Printf("Failed to marshal data: %v\n", err)
		return err
	}
	cacheErr := rdb.Set(context.Background(), cacheKey.String(), resultsJSON, 360*time.Second).Err()
	if cacheErr != nil {
		l.Printf("Something went wrong while caching: %v\n", cacheErr)
		return cacheErr
	}
	return nil
}

func graphqlHandler() gin.HandlerFunc {
	h := handler.NewDefaultServer(
		graph.NewExecutableSchema(
			graph.Config{
				Resolvers: &graph.Resolver{},
			},
		),
	)
	return func(c *gin.Context) {
		h.ServeHTTP(c.Writer, c.Request)
	}
}

func playgroundHandler() gin.HandlerFunc {
	h := playground.Handler("GraphQL Playground", "/query")
	return func(c *gin.Context) {
		h.ServeHTTP(c.Writer, c.Request)
	}
}

func parseUpdateRequestParams(c *gin.Context) ([]string, []interface{}) {
	argsToCheck := []string{"Name", "Surname", "Patronymic", "Age", "Gender", "Nationality"}
	argsString := []string{}
	args := []interface{}{}
	for _, key := range argsToCheck {
		value := c.DefaultQuery(key, "")
		if value != "" {
			queryBuilder := strings.Builder{}
			queryBuilder.WriteString(key)
			queryBuilder.WriteString(" = $")
			args = append(args, value)
			queryBuilder.WriteString(strconv.Itoa(len(args)))
			argsString = append(argsString, queryBuilder.String())
		}
	}
	return argsString, args
}

func updateEnrichedData(db *pgx.Conn, rdb *redis.Client) gin.HandlerFunc {
	fn := func(c *gin.Context) {
		id := c.Param("id")
		if _, err := strconv.Atoi(id); err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": "Provided ID type is not int"})
			fmt.Printf("%v\n", err)
			return
		}

		argsString, args := parseUpdateRequestParams(c)
		if len(argsString) == 0 {
			c.JSON(http.StatusBadRequest, gin.H{"error": "No parameters passed"})
			return
		}

		args = append(args, id)
		query := "UPDATE enriched_data SET " + strings.Join(argsString, ", ") + " WHERE id = $" + strconv.Itoa(len(args))
		res, err := db.Exec(context.Background(), query, args...)
		if err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to update row", "message": err.Error()})
			fmt.Printf("%v\n", err)
			return
		}
		if res.RowsAffected() == 0 {
			c.JSON(http.StatusNotFound, gin.H{"error": "Provided ID not found"})
			return
		}
		rdb.FlushDB(context.Background())
		c.JSON(http.StatusOK, gin.H{"status": "ok"})
	}
	return gin.HandlerFunc(fn)
}

func delEnrichedData(db *pgx.Conn, rdb *redis.Client) gin.HandlerFunc {
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
			return
		}
		rdb.FlushDB(context.Background())
		c.JSON(http.StatusOK, gin.H{"status": "ok"})
	}
	return gin.HandlerFunc(fn)
}

func addEnrichedData(dbChannel chan<- EnrichedData, rdb *redis.Client) gin.HandlerFunc {
	fn := func(c *gin.Context) {
		requestBody := new(EnrichedData)
		if err := c.Bind(requestBody); err != nil {
			c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid request body", "message": err.Error()})
			return
		}
		rdb.FlushDB(context.Background())
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

func getEnrichedData(db *pgx.Conn, rdb *redis.Client) gin.HandlerFunc {
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

		cacheResponse(&age, &gender, &nationality, &page, &results, rdb)
		c.JSON(http.StatusOK, results)
	}
	return gin.HandlerFunc(fn)
}
