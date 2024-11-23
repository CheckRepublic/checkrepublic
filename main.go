package main

import (
	"check_republic/db"
	"check_republic/logic"
	"check_republic/models"
	"log"
	"strconv"

	"github.com/gofiber/fiber/v2"
)

func main() {
	db.Init()

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

	db.DB.CreateOffers(offer.Offers...)
	return c.SendString("Offer created")
}

func getHandler(c *fiber.Ctx) error {
	regionIDParam := c.Query("regionID")
	if regionIDParam == "" {
		return c.Status(fiber.StatusBadRequest).SendString("regionID is required")
	}
	regionID, _ := strconv.ParseUint(regionIDParam, 10, 64)

	timeRangeStartParam := c.Query("timeRangeStart")
	if timeRangeStartParam == "" {
		return c.Status(fiber.StatusBadRequest).SendString("timeRangeStart is required")
	}
	timeRangeStart, _ := strconv.ParseUint(timeRangeStartParam, 10, 64)

	timeRangeEndParam := c.Query("timeRangeEnd")
	if timeRangeEndParam == "" {
		return c.Status(fiber.StatusBadRequest).SendString("timeRangeEnd is required")
	}
	timeRangeEnd, _ := strconv.ParseUint(timeRangeEndParam, 10, 64)

	numberDaysParam := c.Query("numberDays")
	if numberDaysParam == "" {
		return c.Status(fiber.StatusBadRequest).SendString("numberDays is required")
	}
	numberDays, _ := strconv.ParseUint(numberDaysParam, 10, 64)

	sortOrderParam := c.Query("sortOrder")
	if sortOrderParam == "" {
		return c.Status(fiber.StatusBadRequest).SendString("sortOrder is required")
	}

	pageParam := c.Query("page")
	if pageParam == "" {
		return c.Status(fiber.StatusBadRequest).SendString("page is required")
	}
	page, _ := strconv.ParseUint(pageParam, 10, 64)

	pageSizeParam := c.Query("pageSize")
	if pageSizeParam == "" {
		return c.Status(fiber.StatusBadRequest).SendString("pageSize is required")
	}
	pageSize, _ := strconv.ParseUint(pageSizeParam, 10, 64)

	priceRangeWidthParam := c.Query("priceRangeWidth")
	if priceRangeWidthParam == "" {
		return c.Status(fiber.StatusBadRequest).SendString("priceRangeWidth is required")
	}
	priceRangeWidth, _ := strconv.ParseUint(priceRangeWidthParam, 10, 64)

	minFreeKilometerWidthParam := c.Query("minFreeKilometerWidth")
	if minFreeKilometerWidthParam == "" {
		return c.Status(fiber.StatusBadRequest).SendString("minFreeKilometerWidth is required")
	}
	minFreeKilometerWidth, _ := strconv.ParseUint(minFreeKilometerWidthParam, 10, 64)

	minNumberSeatsParam := c.Query("minNumberSeats")
	var minNumberSeats *uint64
	if minNumberSeatsParam == "" {
		minNumberSeats = nil
	} else {
		*minNumberSeats, _ = strconv.ParseUint(minNumberSeatsParam, 10, 64)
	}

    var minPrice *uint64
	minPriceParam := c.Query("minPrice")
    if minPriceParam == "" {
        minPrice = nil
    } else {
        *minPrice, _ = strconv.ParseUint(minPriceParam, 10, 64)
    }

    var maxPrice *uint64
	maxPriceParam := c.Query("maxPrice")
    if maxPriceParam == "" {
        maxPrice = nil
    } else {
        *maxPrice, _ = strconv.ParseUint(maxPriceParam, 10, 64)
    }

    var carType *string
	carTypeParam := c.Query("carType")
    if carTypeParam == "" {
        carType = nil
    } else {
        *carType = carTypeParam
    }

    var onlyVollkasko *bool
	onlyVollkaskoParam := c.Query("onlyVollkasko")
    if onlyVollkaskoParam == "" {
        onlyVollkasko = nil
    } else {
        *onlyVollkasko, _ = strconv.ParseBool(onlyVollkaskoParam)
    }

    var minFreeKilometer *uint64
	minFreeKilometerParam := c.Query("minFreeKilometer")
    if minFreeKilometerParam == "" {
        minFreeKilometer = nil
    } else {
        *minFreeKilometer, _ = strconv.ParseUint(minFreeKilometerParam, 10, 64)
    }

	offers := logic.Filter(regionID,
		timeRangeStart,
		timeRangeEnd,
		numberDays,
		sortOrderParam,
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
