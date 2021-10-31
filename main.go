package main

import (
	"fmt"
	"os"
)

func main() {
	if len(os.Args) < 2 {
		os.Exit(0)
	}
	param := os.Args[1]
	if param == "load" {
		rows := GetCsv("contracheque.csv")
		RunRediSearch(rows)
		RunElasticSearch(rows)
	}
	if param == "search" {
		if len(os.Args) != 3 {
			os.Exit(0)
		}
		text := os.Args[2] //"Azambuja"
		redisResult := SearchRedis(text)
		fmt.Println(redisResult)
		elasticResult := SearchElastic(text)
		fmt.Println(elasticResult)
	}
}
