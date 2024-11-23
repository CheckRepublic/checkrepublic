package main

import (
	"check_republic/models"
	"log"

	"github.com/gofiber/fiber/v2"
)

func main() {
	// Create the database
	var err error
	db, err = CreateDB()
	if err != nil {
		log.Fatal(err)
	}
	log.Println("Database created")

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

	CreateOffers(offer.Offers...)
	return c.SendString("Offer created")
}

func getHandler(c *fiber.Ctx) error {
    //log the request
    log.Println(c.Queries())

	regionID := c.Query("regionID")
    if regionID == "" {
        return c.Status(fiber.StatusBadRequest).SendString("regionID is required")
    }

    timeRangeStart := c.Query("timeRangeStart")
    if timeRangeStart == "" {
        return c.Status(fiber.StatusBadRequest).SendString("timeRangeStart is required")
    }

    timeRangeEnd := c.Query("timeRangeEnd")
    if timeRangeEnd == "" {
        return c.Status(fiber.StatusBadRequest).SendString("timeRangeEnd is required")
    }

    numberDays := c.Query("numberDays")
    if numberDays == "" {
        return c.Status(fiber.StatusBadRequest).SendString("numberDays is required")
    }

    sortOrder := c.Query("sortOrder")
    if sortOrder == "" {
        return c.Status(fiber.StatusBadRequest).SendString("sortOrder is required")
    }

    page := c.Query("page")
    if page == "" {
        return c.Status(fiber.StatusBadRequest).SendString("page is required")
    }

    pageSize := c.Query("pageSize")
    if pageSize == "" {
        return c.Status(fiber.StatusBadRequest).SendString("pageSize is required")
    }

    priceRangeWidth := c.Query("priceRangeWidth")
    if priceRangeWidth == "" {
        return c.Status(fiber.StatusBadRequest).SendString("priceRangeWidth is required")
    }

    minFreeKilometerWidth := c.Query("minFreeKilometerWidth")
    if minFreeKilometerWidth == "" {
        return c.Status(fiber.StatusBadRequest).SendString("minFreeKilometerWidth is required")
    }

    minNumberSeats := c.Query("minNumberSeats")
    minNumberSeats = minNumberSeats
    minPrice := c.Query("minPrice")
    minPrice = minPrice
    maxPrice := c.Query("maxPrice")
    maxPrice = maxPrice
    carType := c.Query("carType")
    carType = carType
    onlyVollkasko := c.Query("onlyVollkasko")
    onlyVollkasko = onlyVollkasko
    minFreeKilometer := c.Query("minFreeKilometer")
    minFreeKilometer = minFreeKilometer

	return c.JSON(GetAllOffers())
}
