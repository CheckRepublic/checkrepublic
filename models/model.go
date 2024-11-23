package models

import "github.com/google/uuid"

// SearchResultOffer represents the search result offer.
type SearchResultOffer struct {
	ID   uuid.UUID `json:"ID"`
	Data string    `json:"data"`
}

type Offers struct {
	Offers []Offer `json:"offers"`
}

// Offer represents the offer details.
type Offer struct {
	//ID                   uuid.UUID `json:"ID"`
	ID                   string	   `json:"ID"`
	Data                 string    `json:"data"`
	MostSpecificRegionID int32     `json:"mostSpecificRegionID"`
	StartDate            int64     `json:"startDate"`
	EndDate              int64     `json:"endDate"`
	NumberSeats          uint8     `json:"numberSeats"`
	Price                uint16    `json:"price"`
	CarType              string    `json:"carType"`
	HasVollkasko         bool      `json:"hasVollkasko"`
	FreeKilometers       uint16    `json:"freeKilometers"`
}


