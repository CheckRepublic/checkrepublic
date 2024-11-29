package db

import (
	"check_republic/models"
	"context"
	"log/slog"
	"sync"
)

type MemoryDB struct {
	// takes a inner node region and returns all leaf offers in leaf regions
	offers []*models.Offer
	tail   uint
	rwlock *sync.RWMutex

	vollcascoBitmask     models.BitMask
	numberOfDaysBitmask  map[int]*models.BitMask
	numberOfSeatsBitmask map[int]*models.BitMask
	carTypeBitmask       map[string]*models.BitMask

	regionIdToOffers map[int32]*models.BitMask
}

// MARK: InitMemoryDB
func InitMemoryDB() {
	DB = MemoryDB{
		offers: make([]*models.Offer, 0, 10_000),
		tail:   0,
		rwlock: &sync.RWMutex{},

		vollcascoBitmask:     models.BitMask{},
		numberOfDaysBitmask:  make(map[int]*models.BitMask),
		numberOfSeatsBitmask: make(map[int]*models.BitMask),
		carTypeBitmask:       make(map[string]*models.BitMask),

		regionIdToOffers: make(map[int32]*models.BitMask),
	}

	for _, carType := range []string{"small", "sports", "luxury", "family"} {
		DB.carTypeBitmask[carType] = &models.BitMask{}
	}

	for _, leaf := range models.RegionTree.GetLeafIds() {
		DB.regionIdToOffers[int32(leaf)] = &models.BitMask{}
	}

	slog.Info("Database created")
}

// MARK: CreateOffers
func (m *MemoryDB) CreateOffers(ctx context.Context, offers ...*models.Offer) error {
	m.rwlock.Lock()
	defer m.rwlock.Unlock()

	for _, offer := range offers {
		// TODO fill bitmaps with data
		if offer.HasVollkasko {
			m.vollcascoBitmask.Set(m.tail, true)
		} else {
			m.vollcascoBitmask.Set(m.tail, false)
		}

		// loop over the regionIdToOffers map
		for regionId, mask := range m.regionIdToOffers {
			if offer.MostSpecificRegionID == uint64(regionId) {
				mask.Set(m.tail, true)
			} else {
				mask.Set(m.tail, false)
			}
		}

		// Ensure the bitmasks are initialized
		for carType, mask := range m.carTypeBitmask {
			if carType == offer.CarType {
				mask.Set(m.tail, true)
			} else {
				mask.Set(m.tail, false)
			}
		}

		found := false
		for numSeats, mask := range m.numberOfSeatsBitmask {
			if int(offer.NumberSeats) == numSeats {
				mask.Set(m.tail, true)
				found = true
			} else {
				mask.Set(m.tail, false)
			}
		}
		if !found {
			mask := models.BitMask{}
			mask.Set(m.tail, true)
			m.numberOfSeatsBitmask[int(offer.NumberSeats)] = &mask
		}

		found = false
		for numDays, mask := range m.numberOfDaysBitmask {
			if int(offer.NumberDays) == numDays {
				mask.Set(m.tail, true)
				found = true
			} else {
				mask.Set(m.tail, false)
			}
		}
		if !found {
			mask := models.BitMask{}
			mask.Set(m.tail, true)
			m.numberOfDaysBitmask[int(offer.NumberDays)] = &mask
		}

		m.offers = append(m.offers, offer)
		m.tail++
	}

	slog.Info("Vollkasko bitmask")
	m.vollcascoBitmask.Print()
	for k, v := range m.numberOfDaysBitmask {
		slog.Info("Number of days", "days", k)
		v.Print()
	}
	for k, v := range m.carTypeBitmask {
		slog.Info("Car type", "type", k)
		v.Print()
	}
	for region, mask := range m.regionIdToOffers {
		slog.Info("Region", "region", region)
		mask.Print()
	}

	return nil
}

// func (m *MemoryDB) GetFilteredOffers(ctx context.Context, regionID uint64, timeRangeStart uint64, timeRangeEnd uint64, numberDays uint64, sortOrder string, page uint64, pageSize uint64, priceRangeWidth uint32, minFreeKilometerWidth uint32, minNumberSeats *uint64, minPrice *uint64, maxPrice *uint64, carType *string, onlyVollkasko *bool, minFreeKilometer *uint64) models.DTO {
// 	// m.rwlock.RLock()
// 	ofs := &models.Offers{Offers: m.regionIdToOffers[int32(regionID)]}
// 	// m.rwlock.RUnlock()
// 	required_ofs := ofs.FilterMandatory(timeRangeStart, timeRangeEnd, numberDays)

// 	// Optional filters
// 	aggs := required_ofs.FilterAggregations(minNumberSeats, minPrice, maxPrice, carType, onlyVollkasko, minFreeKilometer)

// 	optional_ofs := aggs.OptionalAgg

// 	pricesRange := models.BucketizeOffersByPrice(aggs.PricesAgg.Offers, priceRangeWidth)
// 	freeKilometerRange := models.BucketizeOffersByKilometer(aggs.FreeKmAgg.Offers, minFreeKilometerWidth)

// 	// Sorting
// 	if sortOrder == "price-asc" {
// 		sort.Sort(models.ByPrice{Offers: optional_ofs.Offers, Asc: true})
// 	} else if sortOrder == "price-desc" {
// 		sort.Sort(models.ByPrice{Offers: optional_ofs.Offers, Asc: false})
// 	}

// 	// Calculate the starting and ending indices for pagination
// 	startIndex := page * pageSize
// 	endIndex := startIndex + pageSize

// 	// Ensure indices are within bounds
// 	if startIndex > uint64(len(optional_ofs.Offers)) {
// 		startIndex = uint64(len(optional_ofs.Offers))
// 	}
// 	if endIndex > uint64(len(optional_ofs.Offers)) {
// 		endIndex = uint64(len(optional_ofs.Offers))
// 	}

// 	// Slice the offers list for pagination
// 	paginatedOffers := optional_ofs.Offers[startIndex:endIndex]

// 	var dto_offers = make([]*models.OfferDTO, 0, len(paginatedOffers))
// 	for _, offer := range paginatedOffers {
// 		dto_offers = append(dto_offers, &models.OfferDTO{
// 			ID:   offer.ID.String(),
// 			Data: offer.Data,
// 		})
// 	}

// 	seatsCountSlice := []*models.KVSeatsCount{}
// 	// Transform the data correctly
// 	for _, v := range aggs.SeatsCount {
// 		seatsCountSlice = append(seatsCountSlice, v)
// 	}
// 	sort.Slice(seatsCountSlice, func(i, j int) bool {
// 		return seatsCountSlice[i].NumberSeats < seatsCountSlice[j].NumberSeats
// 	})

// 	transformedPricesRange := make([]models.HistogramRange, 0, len(pricesRange))
// 	for _, offer := range pricesRange {
// 		transformedPricesRange = append(transformedPricesRange, models.HistogramRange{
// 			Start: offer.Start,
// 			End:   offer.End,
// 			Count: offer.Count,
// 		})
// 	}

// 	transformedKmRange := make([]models.HistogramRange, 0, len(freeKilometerRange))
// 	for _, offer := range freeKilometerRange {
// 		transformedKmRange = append(transformedKmRange, models.HistogramRange{
// 			Start: offer.Start,
// 			End:   offer.End,
// 			Count: offer.Count,
// 		})
// 	}

// 	return models.DTO{
// 		Offers:             dto_offers,
// 		CarTypeCounts:      aggs.CarTypeCount,
// 		VollkaskoCount:     aggs.VollKaskoCount,
// 		SeatsCount:         seatsCountSlice,
// 		PriceRanges:        transformedPricesRange,
// 		FreeKilometerRange: transformedKmRange,
// 	}
// }

// MARK: DeleteAllOffers
func (m *MemoryDB) DeleteAllOffers(ctx context.Context) error {
	m.offers = make([]*models.Offer, 0, 10_000)
	m.tail = 0

	m.vollcascoBitmask = models.BitMask{}
	m.numberOfDaysBitmask = make(map[int]*models.BitMask)
	m.numberOfSeatsBitmask = make(map[int]*models.BitMask)
	m.carTypeBitmask = make(map[string]*models.BitMask)

	return nil
}
