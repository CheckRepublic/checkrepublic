package db

import (
	"check_republic/models"
	"context"
	"log/slog"
	"sort"
	"sync"
)

type MemoryDB struct {
	// takes a inner node region and returns all leaf offers in leaf regions
	regionIdToPresortedOffers map[int32]*TwoWayPresortedOffers
	rwlock                    *sync.RWMutex
}

type TwoWayPresortedOffers struct {
	PriceAsc  []*models.Offer
	PriceDesc []*models.Offer
}

func InitMemoryDB() {
	DB = MemoryDB{
		regionIdToPresortedOffers: make(map[int32]*TwoWayPresortedOffers),
		rwlock:                    &sync.RWMutex{},
	}
	slog.Info("Database created")
}

func (m *MemoryDB) CreateOffers(ctx context.Context, offers ...*models.Offer) error {
	m.rwlock.Lock()
	defer m.rwlock.Unlock()

	regionIdToOffers := make(map[int32][]*models.Offer)

	for _, offer := range offers {
		for _, anchecstor := range models.SpecificRegionToAnchestor[int32(offer.MostSpecificRegionID)] {
			regionIdToOffers[anchecstor] = append(regionIdToOffers[anchecstor], offer)
		}
	}

	for regionID, offers := range regionIdToOffers {
		m.regionIdToPresortedOffers[regionID] = &TwoWayPresortedOffers{
			PriceAsc:  offers,
			PriceDesc: []*models.Offer{},
		}
		m.regionIdToPresortedOffers[regionID].PriceDesc = append(m.regionIdToPresortedOffers[regionID].PriceDesc, offers...)

		sort.Sort(models.ByPrice{Offers: m.regionIdToPresortedOffers[regionID].PriceAsc, Asc: true})
		sort.Sort(models.ByPrice{Offers: m.regionIdToPresortedOffers[regionID].PriceDesc, Asc: false})
	}

	return nil
}

func (m *MemoryDB) GetFilteredOffers(ctx context.Context, regionID uint64, timeRangeStart uint64, timeRangeEnd uint64, numberDays uint64, sortOrder string, page uint64, pageSize uint64, priceRangeWidth uint32, minFreeKilometerWidth uint32, minNumberSeats *uint64, minPrice *uint64, maxPrice *uint64, carType *string, onlyVollkasko *bool, minFreeKilometer *uint64) models.DTO {
	m.rwlock.RLock()
	defer m.rwlock.RUnlock()

	ofs := &models.Offers{}
	if sortOrder == "price-asc" {
		ofs.Offers = m.regionIdToPresortedOffers[int32(regionID)].PriceAsc
	} else {
		ofs.Offers = m.regionIdToPresortedOffers[int32(regionID)].PriceDesc
	}
	required_ofs := ofs.FilterMandatory(timeRangeStart, timeRangeEnd, numberDays)

	// Optional filters
	aggs := required_ofs.FilterAggregations(minNumberSeats, minPrice, maxPrice, carType, onlyVollkasko, minFreeKilometer)

	optional_ofs := aggs.OptionalAgg

	carTypeCount := models.CarTypeCount{}
	onlyVollkaskoCount := models.VollkaskoCount{}
	seatsCount := models.SeatsCount{}

	pricesRange := models.BucketizeOffersByPrice(aggs.PricesAgg.Offers, priceRangeWidth)
	freeKilometerRange := models.BucketizeOffersByKilometer(aggs.FreeKmAgg.Offers, minFreeKilometerWidth)

	for _, offer := range aggs.CarTypeAgg.Offers {
		carTypeCount.Add(offer.CarType)
	}

	for _, offer := range aggs.VollKaskoAgg.Offers {
		onlyVollkaskoCount.Add(offer.HasVollkasko)
	}

	for _, offer := range aggs.SeatsAgg.Offers {
		seatsCount.Add(offer.NumberSeats)
	}

	// Calculate the starting and ending indices for pagination
	startIndex := page * pageSize
	endIndex := startIndex + pageSize

	// Ensure indices are within bounds
	if startIndex > uint64(len(optional_ofs.Offers)) {
		startIndex = uint64(len(optional_ofs.Offers))
	}
	if endIndex > uint64(len(optional_ofs.Offers)) {
		endIndex = uint64(len(optional_ofs.Offers))
	}

	// Slice the offers list for pagination
	paginatedOffers := optional_ofs.Offers[startIndex:endIndex]

	var dto_offers []*models.OfferDTO
	for _, offer := range paginatedOffers {
		dto_offers = append(dto_offers, &models.OfferDTO{
			ID:   offer.ID.String(),
			Data: offer.Data,
		})
	}

	// Transform the data correctly
	seatsCountSlice := []struct {
		NumberSeats uint64 `json:"numberSeats"`
		Count       uint64 `json:"count"`
	}{}
	for k, v := range seatsCount {
		seatsCountSlice = append(seatsCountSlice, struct {
			NumberSeats uint64 `json:"numberSeats"`
			Count       uint64 `json:"count"`
		}{NumberSeats: k, Count: v})
	}
	sort.Slice(seatsCountSlice, func(i, j int) bool {
		return seatsCountSlice[i].NumberSeats < seatsCountSlice[j].NumberSeats
	})

	transformedPricesRange := []models.HistogramRange{}
	for _, offer := range pricesRange {
		transformedPricesRange = append(transformedPricesRange, models.HistogramRange{
			Start: offer.Start,
			End:   offer.End,
			Count: offer.Count,
		})
	}

	transformedKmRange := []models.HistogramRange{}
	for _, offer := range freeKilometerRange {
		transformedKmRange = append(transformedKmRange, models.HistogramRange{
			Start: offer.Start,
			End:   offer.End,
			Count: offer.Count,
		})
	}

	return models.DTO{
		Offers:             dto_offers,
		CarTypeCounts:      carTypeCount,
		VollkaskoCount:     onlyVollkaskoCount,
		SeatsCount:         seatsCountSlice,
		PriceRanges:        transformedPricesRange,
		FreeKilometerRange: transformedKmRange,
	}
}

func (m *MemoryDB) DeleteAllOffers(ctx context.Context) error {
	m.rwlock.Lock()
	defer m.rwlock.Unlock()

	m.regionIdToPresortedOffers = make(map[int32]*TwoWayPresortedOffers)

	return nil
}
