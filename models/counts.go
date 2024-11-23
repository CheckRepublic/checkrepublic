package models

// PriceRange represents the price range details.
type HistogramRange struct {
	Start uint64 `json:"start"`
	End   uint64 `json:"end"`
	Count uint64 `json:"count"`
}

// CarTypeCount represents the count of offers by car type.
type CarTypeCount struct {
	Small  uint64 `json:"small"`
	Sports uint64 `json:"sports"`
	Luxury uint64 `json:"luxury"`
	Family uint64 `json:"family"`
}

// VollkaskoCount represents the count of offers with and without Vollkasko.
type VollkaskoCount struct {
	TrueCount  uint64 `json:"trueCount"`
	FalseCount uint64 `json:"falseCount"`
}

// SeatsCount represents the count of offers by number of seats.
type SeatsCount struct {
	NumberSeats uint64 `json:"numberSeats"`
	Count       uint64 `json:"count"`
}
