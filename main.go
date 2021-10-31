package main

import (
	"fmt"
	"log"
	"strconv"
	"sync"

	"github.com/RediSearch/redisearch-go/redisearch"
)

func main() {
	rows := GetCsv("caso.csv")
	fmt.Println("carregou o arquivo")
	// Create a client. By default a client is schemaless
	// unless a schema is provided when creating the index
	c := redisearch.NewClient("localhost:6382", "myIndex")

	// Create a schema
	sc := redisearch.NewSchema(redisearch.DefaultOptions).
		AddField(redisearch.NewNumericField("Column1")).
		AddField(redisearch.NewTextField("Column2")).
		AddField(redisearch.NewTextFieldOptions("Column3", redisearch.TextFieldOptions{Sortable: true}))

	// Drop an existing index. If the index does not exist an error is returned
	c.Drop()

	// Create the index with the given schema
	if err := c.CreateIndex(sc); err != nil {
		log.Fatal(err)
	}

	var wg sync.WaitGroup
	var total = 100000 //len(rows)
	for i := 0; i < total; i += 10000 {
		wg.Add(1)
		go func(i int, rows []*CsvLine) {
			defer wg.Done()
			for j := i; j < (i + 1000); j++ {
				docId := strconv.Itoa(j)
				doc := redisearch.NewDocument(docId, 1.0)
				doc.Set("date", rows[j].Column1).
					Set("state", rows[j].Column2).
					Set("city", rows[j].Column3)
				if err := c.Index(doc); err != nil {
					log.Fatal(err)
				}
				//fmt.Println(rows[j].Column3)
			}
		}(i, rows)
	}
	wg.Wait()
	fmt.Println("fim")

	// Searching with limit and sorting
	docs, total, err := c.Search(redisearch.NewQuery("Mar Vermelho").
		Limit(0, 2).
		SetReturnFields("city"))
	fmt.Println(len(docs), total, err)
	//fmt.Println(docs[0].Id, docs[0].Properties["city"], total, err)
	// Output: doc1 Hello world 1 <nil>
}
