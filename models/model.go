package models

import "github.com/google/uuid"

// import "github.com/google/uuid"

type Offers struct {
	Offers []Offer `json:"offers"`
}

// Offer represents the offer details.
type Offer struct {
	ID                   uuid.UUID `json:"ID"`
	Data                 string    `json:"data"`
	MostSpecificRegionID uint64    `json:"mostSpecificRegionID" db:"region_id"`
	StartDate            uint64    `json:"startDate"`
	EndDate              uint64    `json:"endDate"`
	NumberSeats          uint64    `json:"numberSeats"`
	Price                uint64    `json:"price"`
	CarType              string    `json:"carType"`
	HasVollkasko         bool      `json:"hasVollkasko"`
	FreeKilometers       uint64    `json:"freeKilometers"`
}
