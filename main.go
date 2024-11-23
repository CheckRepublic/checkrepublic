package main

import (
	"check_republic/db"
	"check_republic/models"
	"fmt"
	"io/ioutil"
	"os"
	"strconv"
	"time"

	"github.com/gofiber/fiber/v2"
	"github.com/gofiber/fiber/v2/log"
)

func main() {
	if os.Getenv("DEBUG") == "true" {
		log.SetLevel(log.LevelDebug)
	}

	models.InitRegions()

	// db.InitPostgres()
	db.InitElasticSearch()

	app := fiber.New()
	log.Info("Server started")

	app.Get("/api/offers", printHandler)
	//app.Get("/api/offers/all", getAllHandler)
	app.Post("/api/offers", printHandler)
	app.Delete("/api/offers", deleteHandler)

	log.Fatal(app.Listen(":3000"))
}

func getAllHandler(c *fiber.Ctx) error {
	return c.JSON(db.DB.GetAllOffers(c.Context()))
}

func postHandler(c *fiber.Ctx) error {
	var offer models.Offers

	// Parse the request body
	if err := c.BodyParser(&offer); err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
			"error": err.Error(),
		})
	}

	db.DB.CreateOffers(c.Context(), offer.Offers...)
	return c.SendString("Offer created my ass")
}

func printHandler(c *fiber.Ctx) error {
	log.Info("Request received")
	// Get all query parameters
	queryParams := c.Queries()

	// Query Type 
		queryMethod := c.Method()

	// Get request body
	body := c.Body()

	// Create a timestamped filename
	timestamp := time.Now()
	filename := fmt.Sprintf("request_%s_%s.txt", queryMethod, timestamp)

	// Create the content to be written to the file
	content := fmt.Sprintf("Query Params: %v\nBody: %s", queryParams, body)

	// Write the content to the file
	err := ioutil.WriteFile("result/" + filename, []byte(content), 0644)
	if err != nil {
		return c.Status(fiber.StatusInternalServerError).SendString("Failed to write to file")
	}

	return c.SendString("Request data saved successfully")
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
		parsedValue, _ := strconv.ParseUint(minNumberSeatsParam, 10, 64)
		minNumberSeats = &parsedValue // Initialize the pointer with a valid address
	}

	var minPrice *uint64
	minPriceParam := c.Query("minPrice")
	if minPriceParam == "" {
		minPrice = nil
	} else {
		parsed, _ := strconv.ParseUint(minPriceParam, 10, 64)
		minPrice = &parsed
	}

	var maxPrice *uint64
	maxPriceParam := c.Query("maxPrice")
	if maxPriceParam == "" {
		maxPrice = nil
	} else {
		parsed, _ := strconv.ParseUint(maxPriceParam, 10, 64)
		maxPrice = &parsed
	}

	var carType *string
	carTypeParam := c.Query("carType")
	if carTypeParam == "" {
		carType = nil
	} else {
		carType = &carTypeParam
	}

	var onlyVollkasko *bool
	onlyVollkaskoParam := c.Query("onlyVollkasko")
	if onlyVollkaskoParam == "" {
		onlyVollkasko = nil
	} else {
		parsed, _ := strconv.ParseBool(onlyVollkaskoParam)
		onlyVollkasko = &parsed
	}

	var minFreeKilometer *uint64
	minFreeKilometerParam := c.Query("minFreeKilometer")
	if minFreeKilometerParam == "" {
		minFreeKilometer = nil
	} else {
		parsed, _ := strconv.ParseUint(minFreeKilometerParam, 10, 64)
		minFreeKilometer = &parsed
	}

	offers := db.DB.GetFilteredOffers(c.Context(),
		regionID,
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

func deleteHandler(c *fiber.Ctx) error {
	db.DB.DeleteAllOffers(c.Context())
	return c.SendString("All offers deleted")
}
