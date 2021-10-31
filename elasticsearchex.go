package main

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"log"
	"strconv"
	"strings"
	"sync"
	"sync/atomic"
	"time"

	"github.com/elastic/go-elasticsearch/v8"
	"github.com/elastic/go-elasticsearch/v8/esutil"
)

var (
	indexElastic = "elastictest"
)

func RunElasticSearch(rows []*CsvLine) {
	ctx := context.Background()

	esclient, err := elasticsearch.NewDefaultClient()
	if err != nil {
		log.Fatalf("Error creating the client: %s", err)
	}

	bi, err := esutil.NewBulkIndexer(esutil.BulkIndexerConfig{
		Index:  indexElastic,
		Client: esclient,
	})
	if err != nil {
		fmt.Println(err)
	}

	var wg sync.WaitGroup
	var total = len(rows) - 68
	for i := 0; i < total; i += 100 {
		wg.Add(1)
		go func(i int, rows []*CsvLine, bi esutil.BulkIndexer) {
			defer wg.Done()
			for j := i; j < (i + 100); j++ {
				if rows[j] != nil {

					dataJSON, err := json.Marshal(rows[j])
					if err != nil {
						fmt.Println(err)
					}
					var countSuccessful uint64
					err = bi.Add(
						context.Background(),
						esutil.BulkIndexerItem{
							Action:     "index",
							DocumentID: strconv.Itoa(j),
							Body:       bytes.NewReader(dataJSON),
							OnSuccess: func(ctx context.Context, item esutil.BulkIndexerItem, res esutil.BulkIndexerResponseItem) {
								atomic.AddUint64(&countSuccessful, 1)
							},
							OnFailure: func(ctx context.Context, item esutil.BulkIndexerItem, res esutil.BulkIndexerResponseItem, err error) {
								if err != nil {
									log.Printf("ERROR: %s", err)
								} else {
									log.Printf("ERROR: %s: %s", res.Error.Type, res.Error.Reason)
								}
							},
						},
					)
					if err != nil {
						fmt.Println(err)
					}
					fmt.Println(rows[j].Name)
				}
			}
		}(i, rows, bi)
	}
	wg.Wait()
	bi.Close(ctx)
}

func SearchElastic(text string) string {
	es, err := elasticsearch.NewDefaultClient()
	if err != nil {
		log.Fatalf("Error creating the client: %s", err)
	}
	key := "Name"
	value := text
	var buf bytes.Buffer
	query := map[string]interface{}{
		"query": map[string]interface{}{
			"match": map[string]interface{}{
				key: value,
			},
		},
	}
	if err := json.NewEncoder(&buf).Encode(query); err != nil {
		log.Fatalf("Error encoding query: %s", err)
	}
	res, err := es.Search(
		es.Search.WithContext(context.Background()),
		es.Search.WithIndex(indexElastic),
		es.Search.WithBody(&buf),
		es.Search.WithTimeout(90*time.Second),
	)
	if err != nil {
		log.Fatalf("Error getting response: %s", err)
	}
	defer res.Body.Close()

	if res.IsError() {
		var e map[string]interface{}
		if err := json.NewDecoder(res.Body).Decode(&e); err != nil {
			log.Fatalf("Error parsing the response body: %s", err)
		} else {
			log.Fatalf("[%s] %s: %s",
				res.Status(),
				e["error"].(map[string]interface{})["type"],
				e["error"].(map[string]interface{})["reason"],
			)
		}
	}
	var r map[string]interface{}
	if err := json.NewDecoder(res.Body).Decode(&r); err != nil {
		log.Fatalf("Error parsing the response body: %s", err)
	}

	log.Printf(
		"[%s] %d hits; took: %dms",
		res.Status(),
		int(r["hits"].(map[string]interface{})["total"].(map[string]interface{})["value"].(float64)),
		int(r["took"].(float64)),
	)
	var sb strings.Builder
	for _, hit := range r["hits"].(map[string]interface{})["hits"].([]interface{}) {
		sb.WriteString(fmt.Sprintf(" * ID=%s, %s\n", hit.(map[string]interface{})["_id"], hit.(map[string]interface{})["_source"]))
	}
	return sb.String()
}
