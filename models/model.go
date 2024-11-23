package models

import "github.com/google/uuid"

// import "github.com/google/uuid"

type Offers struct {
	Offers []*Offer `json:"offers"`
}

type ByPrice []*Offer

func (a ByPrice) Len() int           { return len(a) }
func (a ByPrice) Swap(i, j int)      { a[i], a[j] = a[j], a[i] }
func (a ByPrice) Less(i, j int) bool { return a[i].Price < a[j].Price }

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

func (offers *Offers) FilterByRegion(regionId uint64) (ret *Offers) {
	ret = &Offers{}
	validRegions := RegionIdToMostSpecificRegionId[int32(regionId)]

	for _, offer := range offers.Offers {
		for _, validRegion := range validRegions {
			if offer.MostSpecificRegionID == uint64(validRegion) {
				ret.Offers = append(ret.Offers, offer)
			}
		}
	}

	return ret
}

func (offers *Offers) FilterByTimeRange(start uint64, end uint64) (ret *Offers) {
	ret = &Offers{}

	for _, offer := range offers.Offers {
		if offer.StartDate >= start && offer.EndDate <= end {
			ret.Offers = append(ret.Offers, offer)
		}
	}

	return ret
}

func (offers *Offers) FilterByNumberDays(num uint64) (ret *Offers) {
	ret = &Offers{}

	// The number of full days (24h) the car is available within the rangeStart and rangeEnd
	for _, offer := range offers.Offers {
		if offer.EndDate-offer.StartDate >= num*24*60*60 {
			ret.Offers = append(ret.Offers, offer)
		}
	}

	return ret
}

func (offers *Offers) FilterByMinSeats(numSeats *uint64) (ret *Offers) {
	ret = &Offers{}

	if numSeats == nil {
		return offers
	}

	for _, offer := range offers.Offers {
		if offer.NumberSeats >= *numSeats {
			ret.Offers = append(ret.Offers, offer)
		}
	}

	return offers
}

func (offers *Offers) FilterByPrice(minPrice *uint64, maxPrice *uint64) (ret *Offers) {
	ret = &Offers{}

	if minPrice == nil && maxPrice == nil {
		return offers
	}

	if minPrice != nil {
		for _, offer := range offers.Offers {
			if offer.Price >= *minPrice {
				ret.Offers = append(ret.Offers, offer)
			}
		}
	}

	if maxPrice != nil {
		for _, offer := range offers.Offers {
			if offer.Price <= *maxPrice {
				ret.Offers = append(ret.Offers, offer)
			}
		}
	}

	return ret
}

func (offers *Offers) FilterByCarType(carType *string) (ret *Offers) {
	ret = &Offers{}

	if carType == nil {
		return offers
	}

	for _, offer := range offers.Offers {
		if offer.CarType == *carType {
			ret.Offers = append(ret.Offers, offer)
		}
	}

	return ret
}

func (offers *Offers) FilterByVollkasko(vollKasko *bool) (ret *Offers) {
	ret = &Offers{}

	if vollKasko == nil {
		return offers
	}

	for _, offer := range offers.Offers {
		if offer.HasVollkasko == *vollKasko {
			ret.Offers = append(ret.Offers, offer)
		}
	}

	return ret
}

func (offers *Offers) FilterByMinFreeKm(km *uint64) (ret *Offers) {
	ret = &Offers{}

	if km == nil {
		return offers
	}

	for _, offer := range offers.Offers {
		if offer.FreeKilometers >= *km {
			ret.Offers = append(ret.Offers, offer)
		}
	}

	return ret
}
