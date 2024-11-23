package models

// import "github.com/google/uuid"

type Offers struct {
	Offers []Offer `json:"offers"`
}

// Offer represents the offer details.
type Offer struct {
	//ID                   uuid.UUID `json:"ID"`
	ID                   string	   `json:"ID"`
	Data                 string    `json:"data"`
	MostSpecificRegionID uint64     `json:"mostSpecificRegionID"`
	StartDate            uint64     `json:"startDate"`
	EndDate              uint64     `json:"endDate"`
	NumberSeats          uint64     `json:"numberSeats"`
	Price                uint64    `json:"price"`
	CarType              string   `json:"carType"`
	HasVollkasko         bool      `json:"hasVollkasko"`
	FreeKilometers       uint64    `json:"freeKilometers"`
}

func (offers *Offers) FilterByRegion(regionId uint64) (ret *Offers) {
	if offers == nil {
		return offers
	}
	ret = &Offers{}
	validRegions := RegionTree.GetRegionById(regionId).GetLeafs()

	for _, offer := range offers.Offers {
		for _, validRegion := range validRegions {
			if offer.MostSpecificRegionID == validRegion.Id {
				ret.Offers = append(ret.Offers, offer)
			}
		}
	}

	return ret
}

func (offers *Offers) FilterByTimeRange(start uint64, end uint64) (ret *Offers) {
	if offers == nil {
		return offers
	}
	ret = &Offers{}
	for _, offer := range offers.Offers {
		if offer.StartDate >= start && offer.EndDate <= end {
			ret.Offers = append(ret.Offers, offer)
		}
	}

	return ret
}

func (offers *Offers) FilterByNumberDays(num uint64) (ret *Offers) {
	// The number of full days (24h) the car is available within the rangeStart and rangeEnd
	if offers == nil {
		return offers
	}
	ret = &Offers{}
    for _, offer := range offers.Offers {
		if offer.EndDate - offer.StartDate >= num * 24 * 60 * 60 {
			ret.Offers = append(ret.Offers, offer)
		}
	}

	return ret
}

func (offers *Offers) FilterByMinSeats(numSeats *uint64) (ret *Offers) {
	if offers == nil {
		return offers
	}
	ret = &Offers{}
	if numSeats == nil {
		return offers
	}
	
	for _, offer := range offers.Offers {
		if offer.NumberSeats >=  *numSeats {
			ret.Offers = append(ret.Offers, offer)
		}
	}

	return offers
}

func (offers *Offers) FilterByPrice(minPrice *uint64, maxPrice *uint64) (ret *Offers) {
	if offers == nil {
		return offers
	}
	ret = &Offers{}
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
	if offers == nil {
		return offers
	}
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
	if offers == nil {
		return offers
	}
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
	if offers == nil {
		return offers
	}
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

func (offers *Offers) CountPriceRanges(priceRangeWidth uint64) (priceRanges []PriceRange) {
	if offers == nil {
		return priceRanges
	}
	priceRangeMap := make(map[uint64]*PriceRange)
	for _, offer := range offers.Offers {
		priceRange := offer.Price / priceRangeWidth
		if _, ok := priceRangeMap[priceRange]; ok {
			priceRangeMap[priceRange].Count = priceRangeMap[priceRange].Count + 1
		} else {
			priceRangeMap[priceRange] = &PriceRange{priceRange * priceRangeWidth, (priceRange + 1) * priceRangeWidth, 1}
		}
	}

	for _, priceRange := range priceRangeMap {
		priceRanges = append(priceRanges, *priceRange)
	}

	return priceRanges
}

func (offers *Offers) CountCarType () (carTypeCounts CarTypeCount) {
	if offers == nil {
		return carTypeCounts
	}
	for _, offer := range offers.Offers {
		switch offer.CarType {
		case "small":
			carTypeCounts.Small = carTypeCounts.Small + 1
		case "sports":
			carTypeCounts.Sports = carTypeCounts.Sports + 1
		case "luxury":
			carTypeCounts.Luxury = carTypeCounts.Luxury + 1
		case "family":
			carTypeCounts.Family = carTypeCounts.Family + 1
		}
	}

	return carTypeCounts
}

func (offers *Offers) CountNumberSeats() (seatCounts []SeatsCount) {
	if offers == nil {
		return seatCounts
	}
	seatCountMap := make(map[uint64]*SeatsCount)
	for _, offer := range offers.Offers {
		if _, ok := seatCountMap[offer.NumberSeats]; ok {
			seatCountMap[offer.NumberSeats].Count = seatCountMap[offer.NumberSeats].Count + 1
		} else {
			seatCountMap[offer.NumberSeats] = &SeatsCount{offer.NumberSeats, 1}
		}
	}

	for _, seatCount := range seatCountMap {
		seatCounts = append(seatCounts, *seatCount)
	}

	return seatCounts
}

func (offers *Offers) CountFreeKilometerRange(freeKilometerWidth uint64) (freeKilometerRanges []FreeKilometerRange) {
	if offers == nil {
		return freeKilometerRanges
	}
	freeKilometerMap := make(map[uint64]*FreeKilometerRange)
	for _, offer := range offers.Offers {
		freeKilometerRange := offer.FreeKilometers / freeKilometerWidth
		if _, ok := freeKilometerMap[freeKilometerRange]; ok {
			freeKilometerMap[freeKilometerRange].Count = freeKilometerMap[freeKilometerRange].Count + 1
		} else {
			freeKilometerMap[freeKilometerRange] = &FreeKilometerRange{freeKilometerRange * freeKilometerWidth, (freeKilometerRange + 1) * freeKilometerWidth, 1}
		}
	}

	for _, freeKilometerRange := range freeKilometerMap {
		freeKilometerRanges = append(freeKilometerRanges, *freeKilometerRange)
	}

	return freeKilometerRanges
}

func (offers *Offers) CountVollkasko() (vollkaskoCount VollkaskoCount) {
	if offers == nil {
		return vollkaskoCount
	}
	for _, offer := range offers.Offers {
		if offer.HasVollkasko {
			vollkaskoCount.TrueCount = vollkaskoCount.TrueCount + 1
		} else {
			vollkaskoCount.FalseCount = vollkaskoCount.FalseCount + 1
		}
	}

	return vollkaskoCount
}
