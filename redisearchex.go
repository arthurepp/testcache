package main

import (
	"fmt"
	"log"
	"strconv"
	"strings"
	"sync"
	"time"

	"github.com/RediSearch/redisearch-go/redisearch"
	"github.com/gomodule/redigo/redis"
)

var (
	indexRedis = "redistest"
)

func RunRediSearch(rows []*CsvLine) {
	pool := &redis.Pool{
		MaxIdle:     100,
		MaxActive:   100000,
		IdleTimeout: 200 * time.Second,
		Dial: func() (redis.Conn, error) {
			return redis.Dial("tcp", ":6382", redis.DialKeepAlive(30*time.Second), redis.DialReadTimeout(10*time.Second))
		}}

	c := redisearch.NewClientFromPool(pool, indexRedis)

	// Create a schema
	sc := redisearch.NewSchema(redisearch.DefaultOptions).
		AddField(redisearch.NewTextField("cargo")).
		AddField(redisearch.NewTextFieldOptions("nome", redisearch.TextFieldOptions{Weight: 5.0, Sortable: true}))

	c.Drop()

	if err := c.CreateIndex(sc); err != nil {
		log.Fatal(err)
	}

	var wg sync.WaitGroup
	var total = len(rows) - 68
	for i := 0; i < total; i += 100 {
		wg.Add(1)
		go func(i int, rows []*CsvLine) {
			defer wg.Done()
			for j := i; j < (i + 100); j++ {
				if rows[j] != nil {
					docId := strconv.Itoa(j)
					doc := redisearch.NewDocument(docId, 1.0)
					doc.Set("nome", rows[j].Name).
						Set("cargo", rows[j].Position)
					if err := c.IndexOptions(redisearch.DefaultIndexingOptions, doc); err != nil {
						fmt.Println(err)
					}
				}
			}
		}(i, rows)
	}
	wg.Wait()
}

func SearchRedis(text string) string {
	c := redisearch.NewClient("localhost:6382", indexRedis)
	start := time.Now()
	docs, _, _ := c.Search(redisearch.NewQuery(text).AddFilter(
		redisearch.Filter{
			Field: "Name",
		},
	))
	duration := time.Since(start)
	fmt.Println(duration)
	var sb strings.Builder
	for _, doc := range docs {
		sb.WriteString(fmt.Sprintln(doc))
	}
	return sb.String()
}
