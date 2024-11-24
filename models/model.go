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

func (offers *Offers) FilterByMinSeats(numSeats *uint64) (ret *Offers) {
	if numSeats == nil {
		return offers
	}
	ret = &Offers{}

	for _, offer := range offers.Offers {
		if offer.NumberSeats >= *numSeats {
			ret.Offers = append(ret.Offers, offer)
		}
	}

	return ret
}

func (offers *Offers) FilterByPrice(minPrice *uint64, maxPrice *uint64) (ret *Offers) {
	if minPrice == nil && maxPrice == nil {
		return offers
	}
	ret = &Offers{}

	for _, offer := range offers.Offers {
		isOkay := true
		isOkay = isOkay && (minPrice == nil || offer.Price >= *minPrice)
		isOkay = isOkay && (maxPrice == nil || offer.Price < *maxPrice)

		if isOkay {
			ret.Offers = append(ret.Offers, offer)
		}
	}

	return ret
}

func (offers *Offers) FilterByCarType(carType *string) (ret *Offers) {
	if carType == nil {
		return offers
	}
	ret = &Offers{}

	for _, offer := range offers.Offers {
		if offer.CarType == *carType {
			ret.Offers = append(ret.Offers, offer)
		}
	}

	return ret
}

func (offers *Offers) FilterByVollkasko(vollKasko *bool) (ret *Offers) {
	if vollKasko == nil || *vollKasko == false {
		return offers
	}
	ret = &Offers{}

	for _, offer := range offers.Offers {
		if offer.HasVollkasko == *vollKasko {
			ret.Offers = append(ret.Offers, offer)
		}
	}

	return ret
}

func (offers *Offers) FilterByMinFreeKm(km *uint64) (ret *Offers) {
	if km == nil {
		return offers
	}
	ret = &Offers{}

	for _, offer := range offers.Offers {
		if offer.FreeKilometers >= *km {
			ret.Offers = append(ret.Offers, offer)
		}
	}

	return ret
}
