package models

// PriceRange represents the price range details.
type PriceRange struct {
	Start uint16 `json:"start"`
	End   uint16 `json:"end"`
	Count uint32 `json:"count"`
}

// CarTypeCount represents the count of offers by car type.
type CarTypeCount struct {
	Small  uint32 `json:"small"`
	Sports uint32 `json:"sports"`
	Luxury uint32 `json:"luxury"`
	Family uint32 `json:"family"`
}

// VollkaskoCount represents the count of offers with and without Vollkasko.
type VollkaskoCount struct {
	TrueCount  uint32 `json:"trueCount"`
	FalseCount uint32 `json:"falseCount"`
}

// SeatsCount represents the count of offers by number of seats.
type SeatsCount struct {
	NumberSeats uint8  `json:"numberSeats"`
	Count       uint32 `json:"count"`
}

// FreeKilometerRange represents the range of free kilometers and the count of offers in this range.
type FreeKilometerRange struct {
	Start uint16 `json:"start"`
	End   uint16 `json:"end"`
	Count uint32 `json:"count"`
}