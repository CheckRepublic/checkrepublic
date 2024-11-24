package main

import (
	"check_republic/db"
	"check_republic/models"
	"fmt"
	"io"
	"log"
	"log/slog"
	"net/http"
	"os"
	"strconv"
	"time"

	"github.com/gin-gonic/gin"

	_ "go.uber.org/automaxprocs"
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

	r.GET("/api/offers", getHandler)
	r.GET("/api/offers/all", getAllHandler)
	r.POST("/api/offers", postHandler)
	r.DELETE("/api/offers", deleteHandler)

	log.Panic(r.Run(":3000"))
}

// Slow, ugly but a bare necessity - kill with fire as soon as possible
func debugHelper(c *gin.Context) {
	// Query Type
	queryMethod := c.Request.Method

	switch queryMethod {
	case "GET":
		// Append request params to a file
		writeGet(c.Request.URL.Query())
	case "POST":
		// Append request body to a file
		// Read request body and append to a file
		body, err := io.ReadAll(c.Request.Body)
		if err != nil {
			slog.Error("Error reading request body", "error", err)
			return
		}
		writePost(body)
	case "DELETE":
		// New test case - create a new file
		filename = time.Now().String()
	}
}
// writeGet appends the query parameters to a file
func writeGet(queryParams map[string][]string) {
	// Create the content to be written to the file
	content := fmt.Sprintf("Query Params: %v\n", queryParams)

	// Append the content to the file
	f, err := os.OpenFile("result/GET_"+filename, os.O_APPEND|os.O_CREATE|os.O_WRONLY, 0644)
	if err != nil {
		log.Error("Error opening file")
		return
	}
	defer f.Close()

	if _, err := f.WriteString(content); err != nil {
		log.Error("Error writing to file")
	}
}

func writePost(body []byte) {
	// Append the content to the file
	f, err := os.OpenFile("result/POST_"+filename, os.O_APPEND|os.O_CREATE|os.O_WRONLY, 0644)
	if err != nil {
		log.Error("Error opening file")
		return
	}
	defer f.Close()

	if _, err := f.Write(body); err != nil {
		log.Error("Error writing to file")
	}
}

func getAllHandler(c *gin.Context) {
	c.JSON(http.StatusOK, db.DB.GetAllOffers(c.Request.Context()))
	return
}


func postHandler(c *gin.Context) {
	if LogToFile {
		debugHelper(c)
	}
	var offer models.Offers

	// Parse the request body
	if err := c.ShouldBindJSON(&offer); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	db.DB.CreateOffers(c.Request.Context(), offer.Offers...)
	c.String(http.StatusCreated, "Offer created")
}

func getHandler(c *gin.Context) {
	if LogToFile {
		debugHelper(c)
	}
	regionIDParam := c.Query("regionID")
	if regionIDParam == "" {
		c.JSON(http.StatusBadRequest, gin.H{"error": "regionID is required"})
		return
	}
	regionID, _ := strconv.ParseUint(regionIDParam, 10, 64)

	timeRangeStartParam := c.Query("timeRangeStart")
	if timeRangeStartParam == "" {
		c.JSON(http.StatusBadRequest, gin.H{"error": "timeRangeStart is required"})
		return
	}
	timeRangeStart, _ := strconv.ParseUint(timeRangeStartParam, 10, 64)

	timeRangeEndParam := c.Query("timeRangeEnd")
	if timeRangeEndParam == "" {
		c.JSON(http.StatusBadRequest, gin.H{"error": "timeRangeEnd is required"})
		return
	}
	timeRangeEnd, _ := strconv.ParseUint(timeRangeEndParam, 10, 64)

	numberDaysParam := c.Query("numberDays")
	if numberDaysParam == "" {
		c.JSON(http.StatusBadRequest, gin.H{"error": "numberDays is required"})
		return
	}
	numberDays, _ := strconv.ParseUint(numberDaysParam, 10, 64)

	sortOrderParam := c.Query("sortOrder")
	if sortOrderParam == "" {
		c.JSON(http.StatusBadRequest, gin.H{"error": "sortOrder is required"})
		return
	}

	pageParam := c.Query("page")
	if pageParam == "" {
		c.JSON(http.StatusBadRequest, gin.H{"error": "page is required"})
		return
	}
	page, _ := strconv.ParseUint(pageParam, 10, 64)

	pageSizeParam := c.Query("pageSize")
	if pageSizeParam == "" {
		c.JSON(http.StatusBadRequest, gin.H{"error": "pageSize is required"})
		return
	}
	pageSize, _ := strconv.ParseUint(pageSizeParam, 10, 64)

	priceRangeWidthParam := c.Query("priceRangeWidth")
	if priceRangeWidthParam == "" {
		c.JSON(http.StatusBadRequest, gin.H{"error": "priceRangeWidth is required"})
		return
	}
	priceRangeWidth, _ := strconv.ParseUint(priceRangeWidthParam, 10, 32)

	minFreeKilometerWidthParam := c.Query("minFreeKilometerWidth")
	if minFreeKilometerWidthParam == "" {
		c.JSON(http.StatusBadRequest, gin.H{"error": "minFreeKilometerWidth is required"})
		return
	}
	minFreeKilometerWidth, _ := strconv.ParseUint(minFreeKilometerWidthParam, 10, 32)

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

	offers := db.DB.GetFilteredOffers(c.Request.Context(),
		regionID,
		timeRangeStart,
		timeRangeEnd,
		numberDays,
		sortOrderParam,
		page,
		pageSize,
		uint32(priceRangeWidth),
		uint32(minFreeKilometerWidth),
		minNumberSeats,
		minPrice,
		maxPrice,
		carType,
		onlyVollkasko,
		minFreeKilometer)

	c.JSON(http.StatusOK, offers)
}

func deleteHandler(c *gin.Context) {
	if LogToFile {
		debugHelper(c)
	}
	db.DB.DeleteAllOffers(c.Request.Context())
	c.String(http.StatusOK, "All offers deleted")
}
