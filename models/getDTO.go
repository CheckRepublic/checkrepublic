package models

type DTO struct {
	Offers Offers `json:"offers"`
	PriceRanges []PriceRange `json:"priceRanges"`
	CarTypeCounts CarTypeCount `json:"carTypeCounts"`
	SeatsCount []SeatsCount `json:"seatsCount"`
	FreeKilometerRange []FreeKilometerRange `json:"freeKilometerRange"`
	VollkaskoCount VollkaskoCount `json:"vollkaskoCount"`
}