package models

import (
	"github.com/google/uuid"
)

const msFactor = 24 * 60 * 60 * 1000

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
	NumberSeats          uint64    `json:"numberSeats"`
	Price                uint64    `json:"price"`
	CarType              string    `json:"carType"`
	HasVollkasko         bool      `json:"hasVollkasko"`
	FreeKilometers       uint64    `json:"freeKilometers"`
}

func (offers *Offers) FilterMandatory(regionId uint64, start uint64, end uint64, num uint64) (ret *Offers) {
	ret = &Offers{}
	validRegions := RegionIdToMostSpecificRegionId[int32(regionId)]

	for _, offer := range offers.Offers {
		for _, validRegion := range validRegions {
			// Check regions
			if offer.MostSpecificRegionID == uint64(validRegion) {
				// Check start and end date
				if offer.StartDate >= start && offer.EndDate <= end {
					// Check number of days
					if offer.EndDate-offer.StartDate == num*msFactor {
						ret.Offers = append(ret.Offers, offer)
					}
				}
			}
		}
	}

	return ret
}

type Aggregations struct {
	PricesAgg    *Offers
	FreeKmAgg    *Offers
	CarTypeAgg   *Offers
	VollKaskoAgg *Offers
	SeatsAgg     *Offers
	OptionalAgg  *Offers
}

func (offers *Offers) FilterAggregations(numSeats *uint64, minPrice *uint64, maxPrice *uint64, carType *string, onlyVollkasko *bool, minFreeKilometer *uint64) (ret *Aggregations) {
	ret = &Aggregations{
		PricesAgg:    &Offers{},
		FreeKmAgg:    &Offers{},
		CarTypeAgg:   &Offers{},
		VollKaskoAgg: &Offers{},
		SeatsAgg:     &Offers{},
		OptionalAgg:  &Offers{},
	}

	for _, offer := range offers.Offers {
		// For prices aggregation
		if (numSeats == nil || offer.NumberSeats >= *numSeats) &&
			(carType == nil || offer.CarType == *carType) &&
			(onlyVollkasko == nil || offer.HasVollkasko == *onlyVollkasko) &&
			(minFreeKilometer == nil || offer.FreeKilometers >= *minFreeKilometer) {
			ret.PricesAgg.Offers = append(ret.PricesAgg.Offers, offer)
		}

		// For free km aggregation
		if (numSeats == nil || offer.NumberSeats >= *numSeats) &&
			(carType == nil || offer.CarType == *carType) &&
			(onlyVollkasko == nil || offer.HasVollkasko == *onlyVollkasko) &&
			((minPrice == nil && maxPrice == nil) || (minPrice == nil || offer.Price >= *minPrice) && (maxPrice == nil || offer.Price < *maxPrice)) {
			ret.FreeKmAgg.Offers = append(ret.FreeKmAgg.Offers, offer)
		}

		// For car type aggregation
		if (numSeats == nil || offer.NumberSeats >= *numSeats) &&
			(minFreeKilometer == nil || offer.FreeKilometers >= *minFreeKilometer) &&
			(onlyVollkasko == nil || offer.HasVollkasko == *onlyVollkasko) &&
			((minPrice == nil && maxPrice == nil) || (minPrice == nil || offer.Price >= *minPrice) && (maxPrice == nil || offer.Price < *maxPrice)) {
			ret.CarTypeAgg.Offers = append(ret.CarTypeAgg.Offers, offer)
		}

		// For vollkasko aggregation
		if (numSeats == nil || offer.NumberSeats >= *numSeats) &&
			(minFreeKilometer == nil || offer.FreeKilometers >= *minFreeKilometer) &&
			(carType == nil || offer.CarType == *carType) &&
			((minPrice == nil && maxPrice == nil) || (minPrice == nil || offer.Price >= *minPrice) && (maxPrice == nil || offer.Price < *maxPrice)) {
			ret.VollKaskoAgg.Offers = append(ret.VollKaskoAgg.Offers, offer)
		}

		// For seats aggregation
		if (onlyVollkasko == nil || offer.HasVollkasko == *onlyVollkasko) &&
			(minFreeKilometer == nil || offer.FreeKilometers >= *minFreeKilometer) &&
			(carType == nil || offer.CarType == *carType) &&
			((minPrice == nil && maxPrice == nil) || (minPrice == nil || offer.Price >= *minPrice) && (maxPrice == nil || offer.Price < *maxPrice)) {
			ret.SeatsAgg.Offers = append(ret.SeatsAgg.Offers, offer)
		}

		// For optional aggregation
		if (numSeats == nil || offer.NumberSeats >= *numSeats) &&
			(onlyVollkasko == nil || offer.HasVollkasko == *onlyVollkasko) &&
			(minFreeKilometer == nil || offer.FreeKilometers >= *minFreeKilometer) &&
			(carType == nil || offer.CarType == *carType) &&
			((minPrice == nil && maxPrice == nil) || (minPrice == nil || offer.Price >= *minPrice) && (maxPrice == nil || offer.Price < *maxPrice)) {
			ret.OptionalAgg.Offers = append(ret.OptionalAgg.Offers, offer)
		}

	}

	return ret
}
