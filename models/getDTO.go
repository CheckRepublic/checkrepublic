package models

type DTO struct {
	Offers             []Offer          `json:"offers"`
	PriceRanges        []HistogramRange `json:"priceRanges"`
	CarTypeCounts      CarTypeCount     `json:"carTypeCounts"`
	SeatsCount         []SeatsCount     `json:"seatsCount"`
	FreeKilometerRange []HistogramRange `json:"freeKilometerRange"`
	VollkaskoCount     VollkaskoCount   `json:"vollkaskoCount"`
}
