package db

import (
	"check_republic/models"
	"context"
	"sort"

	"github.com/gofiber/fiber/v2/log"
)

type MemoryDB struct {
	db     []*models.Offer
	// takes a inner node region and returns all leaf offers in leaf regions
	regionIdToOffers map[int32][]*models.Offer
}

func InitMemoryDB() {
	DB = MemoryDB{
		db:     []*models.Offer{},
		regionIdToOffers: make(map[int32][]*models.Offer),
	}
	log.Info("Database created")
}

func (m *MemoryDB) CreateOffers(ctx context.Context, offers ...*models.Offer) error {
	for _, offer := range offers {
		m.db = append(m.db, offer)
		for _, anchecstor := range models.SpecificRegionToAnchestor[int32(offer.MostSpecificRegionID)]{
			m.regionIdToOffers[anchecstor] = append(m.regionIdToOffers[anchecstor], offer)
		}
	}

	log.Infof("map %v", m.regionIdToOffers)
	return nil
}

func (m *MemoryDB) GetAllOffers(ctx context.Context) models.Offers {
	return models.Offers{Offers: m.db}
}

func (m *MemoryDB) GetFilteredOffers(ctx context.Context, regionID uint64, timeRangeStart uint64, timeRangeEnd uint64, numberDays uint64, sortOrder string, page uint64, pageSize uint64, priceRangeWidth uint32, minFreeKilometerWidth uint32, minNumberSeats *uint64, minPrice *uint64, maxPrice *uint64, carType *string, onlyVollkasko *bool, minFreeKilometer *uint64) models.DTO {
	ofs := &models.Offers{Offers: m.regionIdToOffers[int32(regionID)]}
	log.Infof("offers %v", ofs)
	required_ofs := ofs.
		FilterByTimeRange(timeRangeStart, timeRangeEnd).
		FilterByNumberDays(numberDays)

	// Optional filters
	optional_ofs := required_ofs.
		FilterByMinSeats(minNumberSeats).
		FilterByPrice(minPrice, maxPrice).
		FilterByCarType(carType).
		FilterByVollkasko(onlyVollkasko).
		FilterByMinFreeKm(minFreeKilometer)

	carTypeCount := models.CarTypeCount{}
	onlyVollkaskoCount := models.VollkaskoCount{}
	seatsCount := models.SeatsCount{}

	pricesRange := models.BucketizeOffersByPrice(required_ofs.
		FilterByMinSeats(minNumberSeats).
		FilterByCarType(carType).
		FilterByVollkasko(onlyVollkasko).
		FilterByMinFreeKm(minFreeKilometer).Offers, priceRangeWidth)
	freeKilometerRange := models.BucketizeOffersByKilometer(required_ofs.
		FilterByMinSeats(minNumberSeats).
		FilterByPrice(minPrice, maxPrice).
		FilterByCarType(carType).
		FilterByVollkasko(onlyVollkasko).Offers, minFreeKilometerWidth)

	for _, offer := range required_ofs.
		FilterByMinSeats(minNumberSeats).
		FilterByPrice(minPrice, maxPrice).
		FilterByVollkasko(onlyVollkasko).
		FilterByMinFreeKm(minFreeKilometer).Offers {
		carTypeCount.Add(offer.CarType)
	}

	for _, offer := range required_ofs.
		FilterByMinSeats(minNumberSeats).
		FilterByPrice(minPrice, maxPrice).
		FilterByCarType(carType).
		FilterByMinFreeKm(minFreeKilometer).Offers {
		onlyVollkaskoCount.Add(offer.HasVollkasko)
	}

	for _, offer := range required_ofs.
		FilterByPrice(minPrice, maxPrice).
		FilterByCarType(carType).
		FilterByVollkasko(onlyVollkasko).
		FilterByMinFreeKm(minFreeKilometer).Offers {
		seatsCount.Add(offer.NumberSeats)
	}

	// Sorting
	if sortOrder == "price-asc" {
		sort.Sort(models.ByPrice{Offers: optional_ofs.Offers, Asc: true})
	} else if sortOrder == "price-desc" {
		sort.Sort(models.ByPrice{Offers: optional_ofs.Offers, Asc: false})
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
	m.db = []*models.Offer{}

	return nil
}
