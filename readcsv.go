package main

import (
	"encoding/csv"
	"fmt"
	"os"
	"sync"
)

type CsvLine struct {
	Column1 string
	Column2 string
	//Column3 string
}

func GetCsv(name string) []*CsvLine {

	lines, err := readCsv(name)
	if err != nil {
		panic(err)
	}

	var wg sync.WaitGroup
	rows := make([]*CsvLine, len(lines))

	// Loop through lines & turn into object
	for i, line := range lines {
		wg.Add(1)
		go func(i int, line []string) {
			defer wg.Done()
			if line[2] != "" {
				data := CsvLine{
					Column1: line[1],
					Column2: line[2],
					//Column3: line[2],
				}
				//fmt.Println(data.Column1 + " " + data.Column2)
				rows[i] = &data
			}
		}(i, line)
	}
	wg.Wait()
	fmt.Println(len(rows))
	return rows
}

// ReadCsv accepts a file and returns its content as a multi-dimentional type
// with lines and each column. Only parses to string type.
func readCsv(filename string) ([][]string, error) {

	// Open CSV file
	f, err := os.Open(filename)
	if err != nil {
		return [][]string{}, err
	}
	defer f.Close()

	// Read File into a Variable
	lines, err := csv.NewReader(f).ReadAll()
	if err != nil {
		return [][]string{}, err
	}

	return lines, nil
}
