package db

import (
	"bytes"
	"check_republic/models"
	"context"
	"encoding/json"
	"fmt"
	"strings"

	"github.com/elastic/go-elasticsearch/v8"
	"github.com/gofiber/fiber/v2/log"
)

type ElasticSearchDB struct {
	es *elasticsearch.Client
}

func InitElasticSearch() {
	db, err := createElasticSearch()
	if err != nil {
		log.Fatal(err)
	}

	indexMapping := `{
		"settings": {
			"number_of_shards": 1,
			"number_of_replicas": 0,
			"refresh_interval": "1s"
		},
		"mappings": {
			"properties": {
				"ID": { "type": "keyword" },
				"mostSpecificRegionID": { "type": "integer" },
				"startDate": { "type": "date", "format": "epoch_millis" },
				"endDate": { "type": "date", "format": "epoch_millis" },
				"price": { "type": "integer" },
				"numberSeats": { "type": "integer" },
				"carType": { "type": "keyword" },
				"hasVollkasko": { "type": "boolean" },
				"freeKilometers": { "type": "integer" }
			}
		}
	}`

	res, err := db.Indices.Create("offers", db.Indices.Create.WithBody(bytes.NewReader([]byte(indexMapping))))
	if err != nil {
		log.Errorf("error creating index: %w", err)
	}
	defer res.Body.Close()

	if res.IsError() {
		log.Errorf("error response from Elasticsearch: %s", res.String())
	}

	log.Info("Index created successfully!")

	DB = ElasticSearchDB{es: db}
	log.Info("Database created")
}

func createElasticSearch() (*elasticsearch.Client, error) {
	es, err := elasticsearch.NewDefaultClient()
	if err != nil {
		log.Fatalf("Error creating Elasticsearch client: %s", err)
	}
	return es, err
}

func (e ElasticSearchDB) refreshElastic() {
	res, err := e.es.Indices.Refresh()
	if err != nil {
		log.Errorf("Error refreshing index: %s", err)
	}
	defer res.Body.Close()

	if res.IsError() {
		log.Errorf("Error response from Elasticsearch: %s", res.String())
	}
}

func (e ElasticSearchDB) CreateOffers(ctx context.Context, offers ...models.Offer) error {
	if len(offers) == 0 {
		log.Warn("No offers provided for bulk creation")
		return nil
	}

	const maxBatchSize = 5000 // Define a batch size for bulk requests
	var buf bytes.Buffer

	for i, offer := range offers {
		// Create metadata for the bulk request
		meta := map[string]interface{}{
			"index": map[string]interface{}{
				"_index": "offers",
				"_id":    offer.ID.String(),
			},
		}

		metaData, err := json.Marshal(meta)
		if err != nil {
			log.Errorf("Error marshalling bulk metadata for offer ID %s: %s", offer.ID, err)
			return err
		}

		buf.Write(metaData)
		buf.WriteByte('\n')

		// Create the document data for the bulk request
		offerMap := map[string]interface{}{
			"ID":                   offer.ID,
			"data":                 offer.Data,
			"mostSpecificRegionID": offer.MostSpecificRegionID,
			"startDate":            offer.StartDate,
			"endDate":              offer.EndDate,
			"price":                offer.Price,
			"numberSeats":          offer.NumberSeats,
			"carType":              offer.CarType,
			"hasVollkasko":         offer.HasVollkasko,
			"freeKilometers":       offer.FreeKilometers,
		}

		offerData, err := json.Marshal(offerMap)
		if err != nil {
			log.Errorf("Error marshalling offer data for ID %s: %s", offer.ID, err)
			return err
		}

		buf.Write(offerData)
		buf.WriteByte('\n')

		// Execute the bulk request in batches
		if (i+1)%maxBatchSize == 0 || i == len(offers)-1 {
			if err := e.executeBulkRequest(ctx, buf); err != nil {
				return err
			}
			buf.Reset()
		}
	}

	log.Info("All offers have been successfully created")
	e.refreshElastic()

	return nil
}

// Helper method to execute a bulk request
func (e ElasticSearchDB) executeBulkRequest(ctx context.Context, buf bytes.Buffer) error {
	res, err := e.es.Bulk(bytes.NewReader(buf.Bytes()), e.es.Bulk.WithContext(ctx))
	if err != nil {
		log.Errorf("Error executing bulk operation: %s", err)
		return err
	}
	defer res.Body.Close()

	if res.IsError() {
		log.Errorf("Bulk operation response error: %s", res.String())
		return fmt.Errorf("bulk operation failed: %s", res.String())
	}

	log.Info("Bulk operation completed successfully")
	return nil
}

func (e ElasticSearchDB) GetAllOffers(ctx context.Context) models.Offers {
	// Prepare the search request
	req := map[string]interface{}{
		"query": map[string]interface{}{
			"match_all": map[string]interface{}{},
		},
	}

	// Serialize the request to JSON
	reqBody, err := json.Marshal(req)
	if err != nil {
		log.Errorf("Error marshalling search request: %s", err)
		return models.Offers{}
	}

	// Execute the search request
	res, err := e.es.Search(
		e.es.Search.WithContext(ctx),
		e.es.Search.WithIndex("offers"),
		e.es.Search.WithBody(bytes.NewReader(reqBody)),
		e.es.Search.WithTrackTotalHits(true),
	)
	if err != nil {
		log.Errorf("Error executing search request: %s", err)
		return models.Offers{}
	}
	defer res.Body.Close()

	// Check for errors in the response
	if res.IsError() {
		log.Errorf("Search response error: %s", res.String())
		return models.Offers{}
	}

	// Parse the response
	var searchResult struct {
		Hits struct {
			Hits []struct {
				Source models.Offer `json:"_source"`
			} `json:"hits"`
		} `json:"hits"`
	}
	if err := json.NewDecoder(res.Body).Decode(&searchResult); err != nil {
		log.Errorf("Error parsing search response: %s", err)
		return models.Offers{}
	}

	// Collect the offers
	var offers []models.Offer
	for _, hit := range searchResult.Hits.Hits {
		offers = append(offers, hit.Source)
	}

	log.Info("Retrieved all offers successfully")
	return models.Offers{Offers: offers}
}

func (e ElasticSearchDB) GetFilteredOffers(
	ctx context.Context,
	regionID uint64,
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
	minFreeKilometer *uint64,
) models.DTO {
	// Build the base query for mandatory filters
	baseQuery := map[string]interface{}{
		"bool": map[string]interface{}{
			"filter": []map[string]interface{}{
				// Mandatory Filters
				{
					"terms": map[string]interface{}{
						"mostSpecificRegionID": models.RegionIdToMostSpecificRegionId[int32(regionID)],
					},
				},
				{
					"range": map[string]interface{}{
						"startDate": map[string]interface{}{
							"gte": timeRangeStart,
						},
					},
				},
				{
					"range": map[string]interface{}{
						"endDate": map[string]interface{}{
							"lte": timeRangeEnd,
						},
					},
				},
			},
		},
	}

	// Add a script to calculate the number of full days the car is available in the range
	// Precompute the required overlap in milliseconds
	requiredOverlapMillis := numberDays * 24 * 60 * 60 * 1000

	baseQuery["bool"].(map[string]interface{})["filter"] = append(
		baseQuery["bool"].(map[string]interface{})["filter"].([]map[string]interface{}),
		map[string]interface{}{
			"script": map[string]interface{}{
				"script": map[string]interface{}{
					"source": `
					long carStart = doc['startDate'].value.toInstant().toEpochMilli();
					long carEnd = doc['endDate'].value.toInstant().toEpochMilli();

					double actualStart = Math.max(carStart, params.rangeStart);
					double actualEnd = Math.min(carEnd, params.rangeEnd);

					double overlapMillis = Math.max(0, actualEnd - actualStart);
					return overlapMillis >= params.requiredOverlapMillis;
				`,
					"params": map[string]interface{}{
						"rangeStart":            timeRangeStart,
						"rangeEnd":              timeRangeEnd,
						"requiredOverlapMillis": requiredOverlapMillis,
					},
				},
			},
		},
	)

	// Optional Filters
	optionalFilters := []map[string]interface{}{}
	if minPrice != nil || maxPrice != nil {
		priceRange := map[string]interface{}{}
		if minPrice != nil {
			priceRange["gte"] = *minPrice
		}
		if maxPrice != nil {
			priceRange["lt"] = *maxPrice
		}
		optionalFilters = append(optionalFilters, map[string]interface{}{
			"range": map[string]interface{}{
				"price": priceRange,
			},
		})
	}

	if minFreeKilometer != nil {
		optionalFilters = append(optionalFilters, map[string]interface{}{
			"range": map[string]interface{}{
				"freeKilometers": map[string]interface{}{
					"gte": *minFreeKilometer,
				},
			},
		})
	}

	if minNumberSeats != nil {
		optionalFilters = append(optionalFilters, map[string]interface{}{
			"range": map[string]interface{}{
				"numberSeats": map[string]interface{}{
					"gte": *minNumberSeats,
				},
			},
		})
	}

	if carType != nil {
		optionalFilters = append(optionalFilters, map[string]interface{}{
			"term": map[string]interface{}{
				"carType": *carType,
			},
		})
	}

	if onlyVollkasko != nil && *onlyVollkasko {
		optionalFilters = append(optionalFilters, map[string]interface{}{
			"term": map[string]interface{}{
				"hasVollkasko": true,
			},
		})
	}

	// Aggregations based on mandatory filters only
	aggregations := map[string]interface{}{
		"price_ranges": map[string]interface{}{
			"histogram": map[string]interface{}{
				"field":    "price",
				"interval": priceRangeWidth,
			},
		},
		"car_type_counts": map[string]interface{}{
			"terms": map[string]interface{}{
				"field": "carType",
			},
		},
		"seats_count": map[string]interface{}{
			"terms": map[string]interface{}{
				"field": "numberSeats",
			},
		},
		"free_kilometer_range": map[string]interface{}{
			"histogram": map[string]interface{}{
				"field":    "freeKilometers",
				"interval": minFreeKilometerWidth,
			},
		},
		"vollkasko_count": map[string]interface{}{
			"terms": map[string]interface{}{
				"field": "hasVollkasko",
			},
		},
	}

	// Construct the query to include mandatory filters and aggregations
	baseRequestBody := map[string]interface{}{
		"query": baseQuery,
		"size":  0, // Aggregations only
		"aggs":  aggregations,
	}

	// Execute base aggregation query
	baseRequestJSON, err := json.Marshal(baseRequestBody)
	if err != nil {
		log.Errorf("Error marshalling base aggregation request: %s", err)
		return models.DTO{}
	}

	baseRes, err := e.es.Search(
		e.es.Search.WithContext(ctx),
		e.es.Search.WithIndex("offers"),
		e.es.Search.WithBody(bytes.NewReader(baseRequestJSON)),
	)
	if err != nil {
		log.Errorf("Error executing base aggregation query: %s", err)
		return models.DTO{}
	}
	defer baseRes.Body.Close()

	// Check for errors in the response
	if baseRes.IsError() {
		log.Errorf("Base aggregation query response error: %s", baseRes.String())
		return models.DTO{}
	}

	// Parse base aggregation results
	var baseAggResult struct {
		Aggregations map[string]interface{} `json:"aggregations"`
	}
	if err := json.NewDecoder(baseRes.Body).Decode(&baseAggResult); err != nil {
		log.Errorf("Error parsing base aggregation response: %s", err)
		return models.DTO{}
	}

	// Add optional filters for final query
	finalQuery := map[string]interface{}{
		"bool": map[string]interface{}{
			"filter": append(baseQuery["bool"].(map[string]interface{})["filter"].([]map[string]interface{}), optionalFilters...),
		},
	}
	log.Infof("Final Query %v", finalQuery)

	// Sorting and pagination
	sortField := "price"
	sortDirection := "asc"
	if sortOrder != "" {
		parts := strings.Split(sortOrder, "-")
		if len(parts) == 2 {
			sortField = parts[0]
			sortDirection = parts[1]
		}
	}
	sort := []map[string]interface{}{
		{
			sortField: map[string]interface{}{
				"order": sortDirection,
			},
		},
	}
	from := (page - 1) * pageSize

	// Final query request body
	finalRequestBody := map[string]interface{}{
		"query": finalQuery,
		"from":  from,
		"size":  pageSize,
		"sort":  sort,
	}

	// Execute final query for results
	finalRequestJSON, err := json.Marshal(finalRequestBody)
	if err != nil {
		log.Errorf("Error marshalling final request: %s", err)
		return models.DTO{}
	}

	finalRes, err := e.es.Search(
		e.es.Search.WithContext(ctx),
		e.es.Search.WithIndex("offers"),
		e.es.Search.WithBody(bytes.NewReader(finalRequestJSON)),
	)
	if err != nil {
		log.Errorf("Error executing final query: %s", err)
		return models.DTO{}
	}
	defer finalRes.Body.Close()

	// Parse final query results
	var finalQueryResult struct {
		Hits struct {
			Hits []struct {
				Source models.Offer `json:"_source"`
			} `json:"hits"`
		} `json:"hits"`
	}
	log.Info("Final Query Result", finalRes.Body)
	if err := json.NewDecoder(finalRes.Body).Decode(&finalQueryResult); err != nil {
		log.Errorf("Error parsing final query response: %s", err)
		return models.DTO{}
	}

	// Parse final hits
	var offers []models.OfferDTO
	log.Debug("Final Hits", finalQueryResult.Hits.Hits)
	for _, hit := range finalQueryResult.Hits.Hits {
		offers = append(offers, models.OfferDTO{ID: hit.Source.ID.String(), Data: hit.Source.Data})
	}

	// Parse aggregations
	priceRanges := parseHistogramBuckets(baseAggResult.Aggregations["price_ranges"], priceRangeWidth)
	carTypeCounts := parseCarTypeCounts(baseAggResult.Aggregations["car_type_counts"])
	seatsCount := parseSeatsCount(baseAggResult.Aggregations["seats_count"])
	freeKilometerRange := parseHistogramBuckets(baseAggResult.Aggregations["free_kilometer_range"], minFreeKilometerWidth)
	vollkaskoCount := parseVollkaskoCount(baseAggResult.Aggregations["vollkasko_count"])

	return models.DTO{
		Offers:             offers,
		PriceRanges:        priceRanges,
		CarTypeCounts:      carTypeCounts,
		SeatsCount:         seatsCount,
		FreeKilometerRange: freeKilometerRange,
		VollkaskoCount:     vollkaskoCount,
	}
}

func (e ElasticSearchDB) DeleteAllOffers(ctx context.Context) error {
	// Prepare the query to match all documents
	query := map[string]interface{}{
		"query": map[string]interface{}{
			"match_all": map[string]interface{}{},
		},
	}

	// Serialize the query to JSON
	queryBody, err := json.Marshal(query)
	if err != nil {
		log.Errorf("Error marshalling delete query: %s", err)
		return err
	}

	// Execute the delete by query request
	res, err := e.es.DeleteByQuery(
		[]string{"offers"},
		bytes.NewReader(queryBody),
		e.es.DeleteByQuery.WithContext(ctx),
	)
	if err != nil {
		log.Errorf("Error executing delete by query: %s", err)
		return err
	}
	defer res.Body.Close()

	// Check for errors in the response
	if res.IsError() {
		log.Errorf("Delete by query response error: %s", res.String())
		return fmt.Errorf("delete by query failed: %s", res.String())
	}

	log.Info("All offers deleted successfully")
	e.refreshElastic()
	return nil
}
