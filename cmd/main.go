package main

import (
	"encoding/json"
	"fmt"
	"os"
)

// Define the structures for parsing the log entries
type LogEntry struct {
	RequestType string `json:"requestType"`
	Timestamp   string `json:"timestamp"`
	Log         Log    `json:"log"`
}

type Log struct {
	ID             string       `json:"id"`
	StartTime      string       `json:"start_time"`
	Duration       float64      `json:"duration"`
	WriteConfig    *WriteConfig `json:"write_config,omitempty"`
	ExpectedResult *Result      `json:"expected_result,omitempty"`
	ActualResult   *Result      `json:"actual_result,omitempty"`
	SearchConfig   *SearchConfig `json:"search_config,omitempty"`
}

type WriteConfig struct {
	ID     string  `json:"ID"`
	Offers []Offer `json:"Offers"`
}

type Offer struct {
	OfferID        string  `json:"OfferID"`
	RegionID       int     `json:"RegionID"`
	CarType        string  `json:"CarType"`
	NumberDays     int     `json:"NumberDays"`
	NumberSeats    int     `json:"NumberSeats"`
	StartTimestamp string  `json:"StartTimestamp"`
	EndTimestamp   string  `json:"EndTimestamp"`
	Price          int     `json:"Price"`
	HasVollkasko   bool    `json:"HasVollkasko"`
	FreeKilometers int     `json:"FreeKilometers"`
}

type Result struct {
	Offers            []OfferResult      `json:"Offers"`
	CarTypeCounts     map[string]int     `json:"CarTypeCounts"`
	FreeKilometerRanges []RangeCount     `json:"FreeKilometerRanges"`
	PriceRanges       []RangeCount       `json:"PriceRanges"`
	SeatsCounts       map[string]int     `json:"SeatsCounts"`
	VollkaskoCount    map[string]int     `json:"VollkaskoCount"`
}

type OfferResult struct {
	OfferID      string `json:"OfferID"`
	IsDataCorrect bool  `json:"IsDataCorrect"`
}

type RangeCount struct {
	Start int `json:"Start"`
	End   int `json:"End"`
	Count int `json:"Count"`
}

type SearchConfig struct {
	ID              string  `json:"ID"`
	RegionID        int     `json:"RegionID"`
	StartRange      string  `json:"StartRange"`
	EndRange        string  `json:"EndRange"`
	NumberDays      int     `json:"NumberDays"`
	CarType         *string `json:"CarType,omitempty"`
	OnlyVollkasko   *bool   `json:"OnlyVollkasko,omitempty"`
	MinFreeKilometer *int   `json:"MinFreeKilometer,omitempty"`
	MinNumberSeats  *int    `json:"MinNumberSeats,omitempty"`
	MinPrice        *int    `json:"MinPrice,omitempty"`
	MaxPrice        *int    `json:"MaxPrice,omitempty"`
	Pagination      Pagination `json:"Pagination"`
	Order           string  `json:"Order"`
	PriceBucketWidth int    `json:"PriceBucketWidth"`
	FreeKmBucketWidth int   `json:"FreeKmBucketWidth"`
}

type Pagination struct {
	Page     int `json:"Page"`
	PageSize int `json:"PageSize"`
}

func main() {
	if len(os.Args) < 2 {
		fmt.Println("Usage: main <logfile>")
		return
	}
	logFile := os.Args[1]
	file, err := os.Open(logFile)
	if err != nil {
		fmt.Printf("Error opening file: %v\n", err)
		return
	}
	defer file.Close()

	decoder := json.NewDecoder(file)
	for decoder.More() {
		var entry LogEntry
		err := decoder.Decode(&entry)
		if err != nil {
			fmt.Printf("Error decoding log entry: %v\n", err)
			continue
		}

		// check if the expected and actual results are equal and Print the expected and actual results if they are not equal
		if entry.Log.ExpectedResult != nil && entry.Log.ActualResult != nil {
			if !compareResults(entry.Log.ExpectedResult, entry.Log.ActualResult) {
			expectedJSON, _ := json.MarshalIndent(entry.Log.ExpectedResult, "", "  ")
			actualJSON, _ := json.MarshalIndent(entry.Log.ActualResult, "", "  ")
			fmt.Printf("\033[32mExpected: %s\033[0m\n", expectedJSON)
			fmt.Printf("\033[31mActual: %s\033[0m\n", actualJSON)
			}
		}
		fmt.Println("+++++++++++++++++++++++++++++++++++++++++++++++++++++++++++++++++++++++++++")
	}
}

// compareResults compares the expected and actual results and returns true if they are equal, false otherwise
func compareResults(expected, actual *Result) bool {
	if len(expected.Offers) != len(actual.Offers) {
		return false
	}

	for i := range expected.Offers {
		if expected.Offers[i].OfferID != actual.Offers[i].OfferID || expected.Offers[i].IsDataCorrect != actual.Offers[i].IsDataCorrect {
			return false
		}
	}

	if len(expected.CarTypeCounts) != len(actual.CarTypeCounts) {
		return false
	}
	for k, v := range expected.CarTypeCounts {
		if actual.CarTypeCounts[k] != v {
			return false
		}
	}

	if len(expected.FreeKilometerRanges) != len(actual.FreeKilometerRanges) {
		return false
	}
	for i := range expected.FreeKilometerRanges {
		if expected.FreeKilometerRanges[i].Start != actual.FreeKilometerRanges[i].Start || expected.FreeKilometerRanges[i].End != actual.FreeKilometerRanges[i].End || expected.FreeKilometerRanges[i].Count != actual.FreeKilometerRanges[i].Count {
			return false
		}
	}

	if len(expected.PriceRanges) != len(actual.PriceRanges) {
		return false
	}
	for i := range expected.PriceRanges {
		if expected.PriceRanges[i].Start != actual.PriceRanges[i].Start || expected.PriceRanges[i].End != actual.PriceRanges[i].End || expected.PriceRanges[i].Count != actual.PriceRanges[i].Count {
			return false
		}
	}

	if len(expected.SeatsCounts) != len(actual.SeatsCounts) {
		return false
	}
	for k, v := range expected.SeatsCounts {
		if actual.SeatsCounts[k] != v {
			return false
		}
	}

	if len(expected.VollkaskoCount) != len(actual.VollkaskoCount) {
		return false
	}
	for k, v := range expected.VollkaskoCount {
		if actual.VollkaskoCount[k] != v {
			return false
		}
	}

	return true
}
