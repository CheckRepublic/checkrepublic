package models

type OfferDTO struct {
	ID   string `json:"ID"`
	Data string `json:"data"`
}

type DTO struct {
	Offers        []*OfferDTO      `json:"offers"`
	PriceRanges   []HistogramRange `json:"priceRanges"`
	CarTypeCounts CarTypeCount     `json:"carTypeCounts"`
	SeatsCount    []struct {
		NumberSeats uint64 `json:"numberSeats"`
		Count       uint64 `json:"count"`
	} `json:"seatsCount"`
	FreeKilometerRange []HistogramRange `json:"freeKilometerRange"`
	VollkaskoCount     VollkaskoCount   `json:"vollkaskoCount"`
}
