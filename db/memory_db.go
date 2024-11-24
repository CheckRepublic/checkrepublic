package db

import (
	"check_republic/models"
	"context"
	"log/slog"
	"sort"
	"sync"

	"github.com/google/uuid"
)

type MemoryDB struct {
	db        models.Offers
	dataStore map[uuid.UUID]string
	rwlock    *sync.RWMutex
}

func InitMemoryDB() {
	DB = MemoryDB{
		db:        models.Offers{},
		rwlock:    &sync.RWMutex{},
		dataStore: map[uuid.UUID]string{},
	}
	slog.Info("Database created")
}

func (m *MemoryDB) CreateOffers(ctx context.Context, offers ...models.PostOfferDTO) error {
	m.rwlock.Lock()
	defer m.rwlock.Unlock()

	for _, offer := range offers {
		in_memory_offer := models.OfferInMemory{
			ID:                   offer.ID,
			MostSpecificRegionID: offer.MostSpecificRegionID,
			StartDate:            offer.StartDate,
			EndDate:              offer.EndDate,
			NumberSeats:          offer.NumberSeats,
			Price:                offer.Price,
			CarType:              offer.CarType,
			HasVollkasko:         offer.HasVollkasko,
			FreeKilometers:       offer.FreeKilometers,
		}

		m.db = append(m.db, &in_memory_offer)
		m.dataStore[offer.ID] = offer.Data
	}

	return nil
}

func (m *MemoryDB) GetFilteredOffers(ctx context.Context, regionID uint64, timeRangeStart uint64, timeRangeEnd uint64, numberDays uint64, sortOrder string, page uint64, pageSize uint64, priceRangeWidth uint32, minFreeKilometerWidth uint32, minNumberSeats *uint64, minPrice *uint64, maxPrice *uint64, carType *string, onlyVollkasko *bool, minFreeKilometer *uint64) models.DTO {
	m.rwlock.RLock()
	defer m.rwlock.RUnlock()

	required_ofs := m.db.FilterMandatory(regionID, timeRangeStart, timeRangeEnd, numberDays)

	// Optional filters
	aggs := required_ofs.FilterAggregations(minNumberSeats, minPrice, maxPrice, carType, onlyVollkasko, minFreeKilometer)
	optional_ofs := aggs.OptionalAgg

	pricesRange := models.BucketizeOffersByPrice(aggs.PricesAgg, priceRangeWidth)
	freeKilometerRange := models.BucketizeOffersByKilometer(aggs.FreeKmAgg, minFreeKilometerWidth)

	// Sorting
	if sortOrder == "price-asc" {
		sort.Sort(models.ByPrice{Offers: optional_ofs, Asc: true})
	} else if sortOrder == "price-desc" {
		sort.Sort(models.ByPrice{Offers: optional_ofs, Asc: false})
	}

	// Calculate the starting and ending indices for pagination
	startIndex := page * pageSize
	endIndex := startIndex + pageSize

	// Ensure indices are within bounds
	if startIndex > uint64(len(*optional_ofs)) {
		startIndex = uint64(len(*optional_ofs))
	}
	if endIndex > uint64(len(*optional_ofs)) {
		endIndex = uint64(len(*optional_ofs))
	}

	// Slice the offers list for pagination
	paginatedOffers := (*optional_ofs)[startIndex:endIndex]

	seatsCountSlice := []*models.KVSeatsCount{}
	// Transform the data correctly
	for _, v := range aggs.SeatsCount {
		seatsCountSlice = append(seatsCountSlice, v)
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

	var dto_offers []*models.OfferDTO
	for _, offer := range paginatedOffers {
		dto_offers = append(dto_offers, &models.OfferDTO{
			ID:   offer.ID.String(),
			Data: m.dataStore[offer.ID],
		})
	}

	return models.DTO{
		Offers:             dto_offers,
		CarTypeCounts:      aggs.CarTypeCount,
		VollkaskoCount:     aggs.VollKaskoCount,
		SeatsCount:         seatsCountSlice,
		PriceRanges:        transformedPricesRange,
		FreeKilometerRange: transformedKmRange,
	}
}

func regionIsLeaf(regionID uint64, check uint64) bool {
	regions := models.RegionIdToMostSpecificRegionId[int32(regionID)]
	for _, region := range regions {
		if region == int32(check) {
			return true
		}
	}

	return false
}

func (m *MemoryDB) DeleteAllOffers(ctx context.Context) error {
	m.rwlock.Lock()
	defer m.rwlock.Unlock()

	m.db = []*models.OfferInMemory{}
	m.dataStore = map[uuid.UUID]string{}

	return nil
}
