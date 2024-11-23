package main

import (
	"check_republic/db"
	"check_republic/logic"
	"check_republic/models"
	"log"
	"strconv"

	"github.com/gofiber/fiber/v2"
)

var offDB db.OfferDatabase

func main() {
	offDB = db.Init()

	app := fiber.New()

	app.Get("/api/offers", getHandler)
	app.Post("/api/offers", postHandler)
	app.Delete("/api/offers", helloHandler)

	log.Fatal(app.Listen(":80"))
}

func helloHandler(c *fiber.Ctx) error {
	return c.SendString("Hello, World!")
}

func postHandler(c *fiber.Ctx) error {
	var offer models.Offers

	// Parse the request body
	if err := c.BodyParser(&offer); err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
			"error": err.Error(),
		})
	}

	offDB.CreateOffers(offer.Offers...)
	return c.SendString("Offer created")
}

func getHandler(c *fiber.Ctx) error {
	regionIDParam := c.Query("regionID")
	if regionIDParam == "" {
		return c.Status(fiber.StatusBadRequest).SendString("regionID is required")
	}
	regionID := strconv.ParseUint(regionIDParam, 10, 64)

	timeRangeStartParam := c.Query("timeRangeStart")
	if timeRangeStartParam == "" {
		return c.Status(fiber.StatusBadRequest).SendString("timeRangeStart is required")
	}
	timeRangeStart := strconv.ParseUint(timeRangeStartParam, 10, 64)

	timeRangeEndParam := c.Query("timeRangeEnd")
	if timeRangeEndParam == "" {
		return c.Status(fiber.StatusBadRequest).SendString("timeRangeEnd is required")
	}
	timeRangeEnd := strconv.ParseUint(timeRangeEndParam, 10, 64)

	numberDaysParam := c.Query("numberDays")
	if numberDaysParam == "" {
		return c.Status(fiber.StatusBadRequest).SendString("numberDays is required")
	}
	numberDays := strconv.ParseUint(numberDaysParam, 10, 64)

	sortOrderParam := c.Query("sortOrder")
	if sortOrderParam == "" {
		return c.Status(fiber.StatusBadRequest).SendString("sortOrder is required")
	}

	pageParam := c.Query("page")
	if pageParam == "" {
		return c.Status(fiber.StatusBadRequest).SendString("page is required")
	}
	page := strconv.ParseUint(pageParam, 10, 64)

	pageSizeParam := c.Query("pageSize")
	if pageSizeParam == "" {
		return c.Status(fiber.StatusBadRequest).SendString("pageSize is required")
	}
	pageSize := strconv.ParseUint(pageSizeParam, 10, 64)

	priceRangeWidthParam := c.Query("priceRangeWidth")
	if priceRangeWidthParam == "" {
		return c.Status(fiber.StatusBadRequest).SendString("priceRangeWidth is required")
	}
	priceRangeWidth := strconv.ParseUint(priceRangeWidthParam, 10, 64)

	minFreeKilometerWidthParam := c.Query("minFreeKilometerWidth")
	if minFreeKilometerWidthParam == "" {
		return c.Status(fiber.StatusBadRequest).SendString("minFreeKilometerWidth is required")
	}
	minFreeKilometerWidth := strconv.ParseUint(minFreeKilometerWidthParam, 10, 64)

	minNumberSeatsParam := c.Query("minNumberSeats")
	minNumberSeats := strconv.ParseUint(minNumberSeatsParam, 10, 64)

	minPriceParam := c.Query("minPrice")
	minPrice = strconv.ParseUint(minPriceParam, 10, 64)

	maxPriceParam := c.Query("maxPrice")
	maxPrice = strconv.ParseUint(maxPriceParam, 10, 64)

	carTypeParam := c.Query("carType")

	onlyVollkaskoParam := c.Query("onlyVollkasko")
	onlyVollkasko := strconv.ParseBool(onlyVollkaskoParam)

	minFreeKilometerParam := c.Query("minFreeKilometer")
	minFreeKilometer := strconv.ParseUint(minFreeKilometerParam, 10, 64)

	offers := logic.Filter(regionID,
		timeRangeStart,
		timeRangeEnd,
		numberDays,
		sortOrder,
		page,
		pageSize,
		priceRangeWidth,
		minFreeKilometerWidth,
		minNumberSeats,
		minPrice,
		maxPrice,
		carType,
		onlyVollkasko,
		minFreeKilometer)

	return c.JSON(offers)
}
