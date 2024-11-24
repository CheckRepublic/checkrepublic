package db

import (
	"check_republic/models"
	"context"
)

var DB MemoryDB

type OfferDatabase interface {
	CreateOffers(ctx context.Context, o ...*models.Offer) error
	GetAllOffers(ctx context.Context) models.Offers
	DeleteAllOffers(ctx context.Context) error
	GetFilteredOffers(ctx context.Context, regionID uint64, timeRangeStart uint64, timeRangeEnd uint64, numberDays uint64, sortOrder string, page uint64, pageSize uint64, priceRangeWidth uint64, minFreeKilometerWidth uint64, minNumberSeats *uint64, minPrice *uint64, maxPrice *uint64, carType *string, onlyVollkasko *bool, minFreeKilometer *uint64) models.DTO
}
