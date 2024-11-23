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

	log.Fatal(app.Listen(":3000"))
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
	return c.JSON(GetAllOffers())
}
