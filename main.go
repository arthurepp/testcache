package main

import (
	"fmt"
	"log"
	"strconv"
	"sync"
	"time"

	"github.com/RediSearch/redisearch-go/redisearch"
	"github.com/gomodule/redigo/redis"
)

func main() {
	rows := GetCsv("contracheque.csv")
	fmt.Println("carregou o arquivo")

	pool := &redis.Pool{
		MaxIdle:     100,
		MaxActive:   100000,
		IdleTimeout: 200 * time.Second,
		Dial: func() (redis.Conn, error) {
			return redis.Dial("tcp", ":6382", redis.DialKeepAlive(30*time.Second), redis.DialReadTimeout(10*time.Second))
		}}

	// Create a client. By default a client is schemaless
	// unless a schema is provided when creating the index
	//c := redisearch.NewClient("localhost:6382", "myIndex2")
	c := redisearch.NewClientFromPool(pool, "myIndex2")

	// Create a schema
	sc := redisearch.NewSchema(redisearch.DefaultOptions).
		//AddField(redisearch.NewNumericField("date")).
		AddField(redisearch.NewTextField("cargo")).
		AddField(redisearch.NewTextFieldOptions("nome", redisearch.TextFieldOptions{Weight: 5.0, Sortable: true}))

	// Drop an existing index. If the index does not exist an error is returned
	c.Drop()

	// Create the index with the given schema
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
					doc.Set("nome", rows[j].Column1).
						Set("cargo", rows[j].Column2)
					// if err := c.Index(doc); err != nil {
					// 	log.Fatal(err)
					// }
					if err := c.IndexOptions(redisearch.DefaultIndexingOptions, doc); err != nil {
						//log.Fatal(err)
						fmt.Println(err)
					}
				}
				//fmt.Println(rows[j].Column1)
			}
		}(i, rows)
		// docId := strconv.Itoa(i)
		// doc := redisearch.NewDocument(docId, 1.0)
		// doc.Set("nome", rows[i].Column1).
		// 	Set("cargo", rows[i].Column2)
		// 	//Set("city", rows[i].Column2)
		// // if err := c.Index(doc); err != nil {
		// // 	log.Fatal(err)
		// // }
		// if err := c.IndexOptions(redisearch.DefaultIndexingOptions, doc); err != nil {
		// 	//log.Fatal(err)
		// 	fmt.Println(err)
		// 	fmt.Println(rows[i].Column1)
		// }
	}
	wg.Wait()
	fmt.Println(rows[1].Column1)
	fmt.Println("fim")

	docs, _, _ := c.Search(redisearch.NewQuery("Aldir"))
	fmt.Println(docs)

	// Searching with limit and sorting
	//docs, total, err := c.Search(redisearch.NewQuery(""))
	//fmt.Println(len(docs), total, err)
	//fmt.Println(docs[0].Id, docs[0].Properties["city"], total, err)
	// Output: doc1 Hello world 1 <nil>
}
