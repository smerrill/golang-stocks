package main

import (
	"encoding/csv"
	"errors"
	"fmt"
	// "io/ioutil"
	"net/http"
	"strconv"
	"time"
)

type StockPrice struct {
	symbol string
	price  float64
}

func main() {
	symbols, year := [...]string{"AAPL", "GOOG", "IBM", "MSFT"}, 2008
	numSymbols := len(symbols)
	closingPrices, timeouts := make([]chan StockPrice, numSymbols), make([]<-chan time.Time, numSymbols)

	results := make([]StockPrice, numSymbols)
	// Synchronous version.
	for _, i := range symbols {
		closingPrice, err := getYearEndClosing(i, year)
		if err == nil {
			results = append(results, closingPrice)
		}
	}
	getTopStock(results, year)

	results = make([]StockPrice, numSymbols)
	// Asynchronous version.
	for j, k := range symbols {
		closingPrices[j] = make(chan StockPrice, 1)
		go getYearEndClosingAsync(k, year, closingPrices[j])
		timeouts[j] = time.After(time.Duration(10) * time.Second)
	}

	for j := range symbols {
		select {
		case <-timeouts[j]:
			fmt.Println("Timeout.")
		case closingPrice := <-closingPrices[j]:
			results = append(results, closingPrice)
		}
	}
	getTopStock(results, year)
}

func getTopStock(prices []StockPrice, year int) {
	// @TODO: Error case if this is empty.
	lastHigh, topResult := 0.0, -1
	for i, j := range prices {
		if j.price > lastHigh {
			topResult = i
			lastHigh = j.price
		}
	}

	fmt.Printf("Top stock of %d is %s closing at price %f.\n", year, prices[topResult].symbol, prices[topResult].price)
}

func getYearEndClosingAsync(symbol string, year int, c chan StockPrice) {
	closingPrice, err := getYearEndClosing(symbol, year)
	if err != nil {
		fmt.Println(err.Error())
	}
	c <- closingPrice
}

func getYearEndClosing(symbol string, year int) (StockPrice, error) {
	url := fmt.Sprintf("http://ichart.finance.yahoo.com/table.csv?s=%s&a=11&b=01&c=%d&d=11&e=31&f=%d&g=m",
		symbol, year, year)
	resp, err := http.Get(url)
	if err != nil {
		return StockPrice{}, errors.New(fmt.Sprintf("Error making an HTTP request for stock %s.", symbol))
	}
	defer resp.Body.Close()
	csvReader := csv.NewReader(resp.Body)
	records, err2 := csvReader.ReadAll()
	if err2 != nil {
		return StockPrice{}, errors.New(fmt.Sprintf("Error parsing CSV values for stock %s.", symbol))
	}
	closingPrice, err3 := strconv.ParseFloat(records[1][4], 64)
	if err3 != nil {
		return StockPrice{}, err3
	}
	return StockPrice{symbol, closingPrice}, nil
}
