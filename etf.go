package main

import (
	"encoding/csv"
	"fmt"
	"io"
	"log"
	"math"
	"net/http"
	"strconv"
)

const source = "https://raw.githubusercontent.com/datasets/s-and-p-companies-financials/master/data/constituents-financials.csv"

func main() {
	// 1. Download CSV
	resp, e := http.Get(source)
	if e != nil {
		log.Fatal(e)
		return
	}
	body := resp.Body

	// 2. Parse CSV
	data, largestMarketCap, e := readCSV(csv.NewReader(body))
	defer body.Close()
	if e != nil {
		log.Fatal(e)
		return
	}

	// 3. Determine equity weightings
	weighted := computeRelative(data, largestMarketCap)

	// 4. Output result
	/*
	for stock, multiple := range weighted {
		fmt.Printf("%s %f\n", stock.Symbol, multiple)
	}

	println()
	fmt.Printf("$%f\n", minBudget(weighted))
	println()
	*/

	for stock, numShares := range shares(2000000, minBudget(weighted), weighted) {
		fmt.Printf("%s %d\n", stock.Symbol, numShares)
	}
}

func readCSV(reader *csv.Reader) ([]Stock, float64, error) {
	first, err := reader.Read()
	if err != nil {
		return nil, 0, err
	}

	// Dynamically get field indexes from table header
	var symMap, nameMap, sectMap, pMap, capMap = 0, 0, 0, 0, 0
	for index, value := range first {
		switch value {
		case "Symbol":
			symMap = index
		case "Name":
			nameMap = index
		case "Sector":
			sectMap = index
		case "Price":
			pMap = index
		case "Market Cap":
			capMap = index
		}
	}

	var stocks []Stock
	var largest float64 // keep track of largest market cap for later computational use (without another iteration)

	for {
		record, e := reader.Read()
		if e != nil {
			if e != io.EOF {
				return nil, 0, e
			}

			break
		}

		price, e2 := strconv.ParseFloat(record[pMap], 32)
		if e2 != nil {
			log.Fatal(e2)
			continue
		}

		mktCap, e3 := strconv.ParseFloat(record[capMap], 64)
		if e3 != nil {
			log.Fatal(e3)
			continue
		}
		largest = math.Max(largest, mktCap)

		stocks = append(stocks, Stock{
			Symbol:    record[symMap],
			Name:      record[nameMap],
			Sector:    record[sectMap],
			Price:     float32(price),
			MarketCap: mktCap,
		})
	}

	return stocks, largest, nil
}

func computeRelative(data []Stock, largest float64) (map[Stock]float64) {
	m := make(map[Stock]float64)

	for _, stock := range data {
		m[stock] = largest / stock.MarketCap // equal market weight
	}

	return m
}

func minBudget(weights map[Stock]float64) (float64) {
	var cost float64 = 0

	for stock, weight := range weights {
		cost += float64(stock.Price) * weight
	}

	return cost
}

func shares(portfolio float64, minBudget float64, weights map[Stock]float64) (map[Stock]int) {
	multiple := portfolio / minBudget
	m := make(map[Stock]int)

	for stock, weight := range weights {
		m[stock] = int(math.Round(weight * multiple))
	}

	return m
}

type Stock struct {
	Symbol    string
	Name      string
	Sector    string
	Price     float32
	MarketCap float64
}
