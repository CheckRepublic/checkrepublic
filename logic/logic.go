package logic

import (
	"check_republic/db"
	"check_republic/models"
)

func Filter(regionID uint64,
	timeRangeStart uint64,
	timeRangeEnd uint64,
	numberDays uint64,
	sortOrder string,
	page uint64,
	pageSize uint64,
	priceRangeWidth uint64,
	minFreeKilometerWidth uint64,
	minNumberSeats *uint64,
	minPrice *uint64,
	maxPrice *uint64,
	carType *string,
	onlyVollkasko *bool,
	minFreeKilometer *uint64) (offersDTO models.DTO) {
	var offers models.Offers
	offers = db.DB.GetAllOffers()
	offers = *offers.FilterByRegion(regionID).FilterByTimeRange(timeRangeStart, timeRangeEnd).FilterByNumberDays(numberDays).FilterByMinSeats(minNumberSeats).FilterByPrice(minPrice, maxPrice).FilterByCarType(carType).FilterByVollkasko(onlyVollkasko).FilterByMinFreeKm(minFreeKilometer)
	
	offersDTO.Offers = offers.Offers
	offersDTO.PriceRanges = offers.CountPriceRanges(priceRangeWidth)
	offersDTO.CarTypeCounts = offers.CountCarType()
	offersDTO.SeatsCount = offers.CountNumberSeats()
	offersDTO.FreeKilometerRange = offers.CountFreeKilometerRange(minFreeKilometerWidth)
	offersDTO.VollkaskoCount = offers.CountVollkasko()


	return offersDTO
}
