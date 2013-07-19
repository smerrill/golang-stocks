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

func main() {
	symbols, year := [...]string{"AAPL", "GOOG", "IBM", "MSFT"}, 2008
	numSymbols := len(symbols)
	closingPrices, timeouts := make([]chan float64, numSymbols), make([]<-chan time.Time, numSymbols)

	// Synchronous version.
	for _, i := range symbols {
		closingPrice, err := getYearEndClosing(i, year)
		if err == nil {
			fmt.Printf("%s: %f\n", i, closingPrice)
		}
	}

	// Asynchronous version.
	for j, k := range symbols {
		closingPrices[j] = make(chan float64, 1)
		go getYearEndClosingAsync(k, year, closingPrices[j])
		timeouts[j] = time.After(time.Duration(2000) * time.Millisecond)

		select {
		case closingPrice := <-closingPrices[j]:
			fmt.Printf("%s: %f\n", k, closingPrice)
		case <-timeouts[j]:
			fmt.Println("Timeout.")
		}
	}
}

func getYearEndClosingAsync(symbol string, year int, c chan float64) {
	closingPrice, err := getYearEndClosing(symbol, year)
	if err != nil {
		fmt.Println(err.Error())
	}
	c <- closingPrice
}

func getYearEndClosing(symbol string, year int) (float64, error) {
	url := fmt.Sprintf("http://ichart.finance.yahoo.com/table.csv?s=%s&a=11&b=01&c=%d&d=11&e=31&f=%d&g=m",
		symbol, year, year)
	resp, err := http.Get(url)
	if err != nil {
		return 0, errors.New(fmt.Sprintf("Error making an HTTP request for stock %s.", symbol))
	}
	defer resp.Body.Close()
	csvReader := csv.NewReader(resp.Body)
	records, err2 := csvReader.ReadAll()
	if err2 != nil {
		return 0, errors.New(fmt.Sprintf("Error parsing CSV values for stock %s.", symbol))
	}
	closingPrice, err3 := strconv.ParseFloat(records[1][4], 64)
	if err3 != nil {
		return 0, err3
	}
	return closingPrice, nil
}
