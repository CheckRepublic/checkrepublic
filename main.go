package main

import (
	"check_republic/db"
	"check_republic/models"
	"log"
	"log/slog"
	"net/http"
	"os"
	"time"

	"github.com/gin-contrib/gzip"
	"github.com/gin-gonic/gin"
)

var filename = time.Now().String()
var LogToFile = os.Getenv("LOG") == "true"

func main() {
	if os.Getenv("DEBUG") == "true" {
		slog.SetLogLoggerLevel(slog.LevelDebug)
	}

	models.InitRegions()

	// db.InitPostgres()
	db.InitMemoryDB()

	gin.SetMode(gin.ReleaseMode)
	r := gin.New()
	r.Use(gin.ErrorLogger())
	r.Use(gin.Recovery())

	// Increase allowed body size
	r.Use(func(c *gin.Context) {
		c.Request.Body = http.MaxBytesReader(c.Writer, c.Request.Body, 3*1024*1024*1024) // 3GB
		c.Next()
	})

	// Gzip compression
	r.Use(gzip.Gzip(gzip.BestSpeed))

	r.GET("/api/offers", getHandler)
	r.POST("/api/offers", postHandler)
	r.DELETE("/api/offers", deleteHandler)

	log.Panic(r.Run(":3000"))
}

func postHandler(c *gin.Context) {
	var offer models.Offers

	// Parse the request body
	if err := c.ShouldBindJSON(&offer); err != nil {
		slog.Error("Error parsing request body", "error", err)
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	err := db.DB.CreateOffers(c.Request.Context(), offer.Offers...)
	if err != nil {
		slog.Error("Error creating offers", "error", err)
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	c.String(http.StatusOK, "Offer created")
}

func getHandler(c *gin.Context) {
	// regionIDParam := c.Query("regionID")
	// regionID, _ := strconv.ParseUint(regionIDParam, 10, 64)

	// timeRangeStartParam := c.Query("timeRangeStart")
	// timeRangeStart, _ := strconv.ParseUint(timeRangeStartParam, 10, 64)
	// timeRangeEndParam := c.Query("timeRangeEnd")
	// timeRangeEnd, _ := strconv.ParseUint(timeRangeEndParam, 10, 64)
	// numberDaysParam := c.Query("numberDays")
	// numberDays, _ := strconv.ParseUint(numberDaysParam, 10, 64)
	// sortOrderParam := c.Query("sortOrder")
	// pageParam := c.Query("page")
	// page, _ := strconv.ParseUint(pageParam, 10, 64)
	// pageSizeParam := c.Query("pageSize")
	// pageSize, _ := strconv.ParseUint(pageSizeParam, 10, 64)
	// priceRangeWidthParam := c.Query("priceRangeWidth")
	// priceRangeWidth, _ := strconv.ParseUint(priceRangeWidthParam, 10, 32)
	// minFreeKilometerWidthParam := c.Query("minFreeKilometerWidth")
	// minFreeKilometerWidth, _ := strconv.ParseUint(minFreeKilometerWidthParam, 10, 32)
	// minNumberSeatsParam := c.Query("minNumberSeats")
	// var minNumberSeats *uint64
	// if minNumberSeatsParam == "" {
	// 	minNumberSeats = nil
	// } else {
	// 	parsedValue, _ := strconv.ParseUint(minNumberSeatsParam, 10, 64)
	// 	minNumberSeats = &parsedValue // Initialize the pointer with a valid address
	// }

	// var minPrice *uint64
	// minPriceParam := c.Query("minPrice")
	// if minPriceParam == "" {
	// 	minPrice = nil
	// } else {
	// 	parsed, _ := strconv.ParseUint(minPriceParam, 10, 64)
	// 	minPrice = &parsed
	// }

	// var maxPrice *uint64
	// maxPriceParam := c.Query("maxPrice")
	// if maxPriceParam == "" {
	// 	maxPrice = nil
	// } else {
	// 	parsed, _ := strconv.ParseUint(maxPriceParam, 10, 64)
	// 	maxPrice = &parsed
	// }

	// var carType *string
	// carTypeParam := c.Query("carType")
	// if carTypeParam == "" {
	// 	carType = nil
	// } else {
	// 	carType = &carTypeParam
	// }

	// var onlyVollkasko *bool
	// onlyVollkaskoParam := c.Query("onlyVollkasko")
	// if onlyVollkaskoParam == "" {
	// 	onlyVollkasko = nil
	// } else {
	// 	parsed, _ := strconv.ParseBool(onlyVollkaskoParam)
	// 	onlyVollkasko = &parsed
	// }

	// var minFreeKilometer *uint64
	// minFreeKilometerParam := c.Query("minFreeKilometer")
	// if minFreeKilometerParam == "" {
	// 	minFreeKilometer = nil
	// } else {
	// 	parsed, _ := strconv.ParseUint(minFreeKilometerParam, 10, 64)
	// 	minFreeKilometer = &parsed
	// }

	// offers := db.DB.GetFilteredOffers(c.Request.Context(),
	// 	regionID,
	// 	timeRangeStart,
	// 	timeRangeEnd,
	// 	numberDays,
	// 	sortOrderParam,
	// 	page,
	// 	pageSize,
	// 	uint32(priceRangeWidth),
	// 	uint32(minFreeKilometerWidth),
	// 	minNumberSeats,
	// 	minPrice,
	// 	maxPrice,
	// 	carType,
	// 	onlyVollkasko,
	// 	minFreeKilometer)
	offers := make([]models.Offers, 0)
	c.JSON(http.StatusOK, offers)
}

func deleteHandler(c *gin.Context) {
	db.DB.DeleteAllOffers(c.Request.Context())
	c.String(http.StatusOK, "All offers deleted")
}
