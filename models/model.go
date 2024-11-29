package models

import (
	"github.com/google/uuid"
)

const MsFactor = 24 * 60 * 60 * 1000

type Offers struct {
	Offers []*Offer `json:"offers"`
}

type ByPrice struct {
	Offers []*Offer
	Asc    bool
}

func (a ByPrice) Len() int      { return len(a.Offers) }
func (a ByPrice) Swap(i, j int) { a.Offers[i], a.Offers[j] = a.Offers[j], a.Offers[i] }
func (a ByPrice) Less(i, j int) bool {
	if a.Offers[i].Price == a.Offers[j].Price {
		return a.Offers[i].ID.String() < a.Offers[j].ID.String()
	}

	if !a.Asc {
		return a.Offers[i].Price > a.Offers[j].Price
	}
	return a.Offers[i].Price < a.Offers[j].Price
}

// Offer represents the offer details.
type Offer struct {
	ID                   uuid.UUID `json:"ID"`
	Data                 string    `json:"data"`
	MostSpecificRegionID uint64    `json:"mostSpecificRegionID" db:"region_id"`
	StartDate            uint64    `json:"startDate"`
	EndDate              uint64    `json:"endDate"`
	NumberDays           uint64    `json:"-"`
	NumberSeats          uint64    `json:"numberSeats"`
	Price                uint64    `json:"price"`
	CarType              string    `json:"carType"`
	HasVollkasko         bool      `json:"hasVollkasko"`
	FreeKilometers       uint64    `json:"freeKilometers"`
}

type Aggregations struct {
	PricesAgg   *Offers
	FreeKmAgg   *Offers
	OptionalAgg *Offers
}

func (offers *Offers) FilterAggregations(timeStart uint64, timeEnd uint64, minPrice *uint64, maxPrice *uint64, minFreeKilometer *uint64) (ret *Aggregations) {
	ret = &Aggregations{
		PricesAgg: &Offers{
			Offers: make([]*Offer, 0, len(offers.Offers)/2),
		},
		FreeKmAgg: &Offers{
			Offers: make([]*Offer, 0, len(offers.Offers)/2),
		},
		OptionalAgg: &Offers{
			Offers: make([]*Offer, 0, len(offers.Offers)/2),
		},
	}

	for _, offer := range offers.Offers {
		var boolTime = offer.StartDate >= timeStart && offer.EndDate <= timeEnd
		var boolFreeKm = minFreeKilometer == nil || offer.FreeKilometers >= *minFreeKilometer
		var boolPrice = ((minPrice == nil && maxPrice == nil) || (minPrice == nil || offer.Price >= *minPrice) && (maxPrice == nil || offer.Price < *maxPrice))

		// For prices aggregation
		if boolTime &&
			boolFreeKm {
			ret.PricesAgg.Offers = append(ret.PricesAgg.Offers, offer)
		}

		// For free km aggregation
		if boolTime &&
			boolPrice {
			ret.FreeKmAgg.Offers = append(ret.FreeKmAgg.Offers, offer)
		}

		// For optional aggregation
		if boolTime &&
			boolFreeKm &&
			boolPrice {
			ret.OptionalAgg.Offers = append(ret.OptionalAgg.Offers, offer)
		}

	}

	return ret
}
