package main

import (
	"encoding/csv"
	"fmt"
	"os"
	"sync"
)

type CsvLine struct {
	Name     string
	Position string
}

func GetCsv(name string) []*CsvLine {

	lines, err := readCsv(name)
	if err != nil {
		panic(err)
	}

	var wg sync.WaitGroup
	rows := make([]*CsvLine, len(lines))

	for i, line := range lines {
		wg.Add(1)
		go func(i int, line []string) {
			defer wg.Done()
			if line[2] != "" {
				data := CsvLine{
					Name:     line[1],
					Position: line[2],
				}
				//fmt.Println(data.Name + " " + data.Position)
				rows[i] = &data
			}
		}(i, line)
	}
	wg.Wait()
	fmt.Println(len(rows))
	return rows
}

func readCsv(filename string) ([][]string, error) {

	f, err := os.Open(filename)
	if err != nil {
		return [][]string{}, err
	}
	defer f.Close()

	lines, err := csv.NewReader(f).ReadAll()
	if err != nil {
		return [][]string{}, err
	}

	return lines, nil
}
