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
			"refresh_interval": "30s"
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

func (e ElasticSearchDB) CreateOffers(ctx context.Context, o ...models.Offer) error {
	var buf bytes.Buffer

	for _, offer := range o {
		// Create the metadata line for the bulk request
		meta := map[string]interface{}{
			"index": map[string]interface{}{
				"_index": "offers",
				"_id":    offer.ID.String(),
			},
		}

		metaData, err := json.Marshal(meta)
		if err != nil {
			log.Errorf("Error marshalling bulk metadata: %s", err)
			return err
		}

		// Append metadata to the buffer
		buf.Write(metaData)
		buf.WriteByte('\n')

		// Create the document for the bulk request
		offerMap := map[string]interface{}{
			"ID":                   offer.ID,
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
			log.Errorf("Error marshalling offer: %s", err)
			return err
		}

		// Append document data to the buffer
		buf.Write(offerData)
		buf.WriteByte('\n')
	}

	// Execute the bulk operation
	res, err := e.es.Bulk(bytes.NewReader(buf.Bytes()), e.es.Bulk.WithContext(ctx))
	if err != nil {
		log.Errorf("Error executing bulk operation: %s", err)
		return err
	}
	defer res.Body.Close()

	// Check for errors in the bulk response
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
) models.Offers {
	// Construct the query dynamically
	query := map[string]interface{}{
		"bool": map[string]interface{}{
			"must":   []map[string]interface{}{},
			"filter": []map[string]interface{}{},
		},
	}

	// Add region filter
	query["bool"].(map[string]interface{})["filter"] = append(
		query["bool"].(map[string]interface{})["filter"].([]map[string]interface{}),
		map[string]interface{}{
			"term": map[string]interface{}{
				"mostSpecificRegionID": regionID,
			},
		},
	)

	// Add time range filter
	query["bool"].(map[string]interface{})["filter"] = append(
		query["bool"].(map[string]interface{})["filter"].([]map[string]interface{}),
		map[string]interface{}{
			"range": map[string]interface{}{
				"startDate": map[string]interface{}{
					"gte": timeRangeStart,
					"lte": timeRangeEnd,
				},
			},
		},
	)

	// Add price range filter if provided
	if minPrice != nil || maxPrice != nil {
		priceRange := map[string]interface{}{}
		if minPrice != nil {
			priceRange["gte"] = *minPrice
		}
		if maxPrice != nil {
			priceRange["lte"] = *maxPrice
		}
		query["bool"].(map[string]interface{})["filter"] = append(
			query["bool"].(map[string]interface{})["filter"].([]map[string]interface{}),
			map[string]interface{}{
				"range": map[string]interface{}{
					"price": priceRange,
				},
			},
		)
	}

	// Add minimum free kilometer filter if provided
	if minFreeKilometer != nil {
		query["bool"].(map[string]interface{})["filter"] = append(
			query["bool"].(map[string]interface{})["filter"].([]map[string]interface{}),
			map[string]interface{}{
				"range": map[string]interface{}{
					"freeKilometers": map[string]interface{}{
						"gte": *minFreeKilometer,
					},
				},
			},
		)
	}

	// Add minimum number of seats filter if provided
	if minNumberSeats != nil {
		query["bool"].(map[string]interface{})["filter"] = append(
			query["bool"].(map[string]interface{})["filter"].([]map[string]interface{}),
			map[string]interface{}{
				"range": map[string]interface{}{
					"numberSeats": map[string]interface{}{
						"gte": *minNumberSeats,
					},
				},
			},
		)
	}

	// Add car type filter if provided
	if carType != nil {
		query["bool"].(map[string]interface{})["filter"] = append(
			query["bool"].(map[string]interface{})["filter"].([]map[string]interface{}),
			map[string]interface{}{
				"term": map[string]interface{}{
					"carType": *carType,
				},
			},
		)
	}

	// Add Vollkasko filter if provided
	if onlyVollkasko != nil && *onlyVollkasko {
		query["bool"].(map[string]interface{})["filter"] = append(
			query["bool"].(map[string]interface{})["filter"].([]map[string]interface{}),
			map[string]interface{}{
				"term": map[string]interface{}{
					"hasVollkasko": true,
				},
			},
		)
	}

	// Sorting
	sortField := "price"   // Default field
	sortDirection := "asc" // Default direction

	if sortOrder != "" {
		parts := strings.Split(sortOrder, "-")
		if len(parts) == 2 {
			sortField = parts[0]
			sortDirection = parts[1]
		} else {
			log.Warnf("Invalid sortOrder format, using default: %s", sortOrder)
		}
	}

	sort := []map[string]interface{}{
		{
			sortField: map[string]interface{}{
				"order": sortDirection,
			},
		},
	}

	// Pagination
	from := (page - 1) * pageSize

	// Construct the search request
	reqBody := map[string]interface{}{
		"query": query,
		"from":  from,
		"size":  pageSize,
		"sort":  sort,
	}

	// Serialize the request
	reqJSON, err := json.Marshal(reqBody)
	if err != nil {
		log.Errorf("Error marshalling search request: %s", err)
		return models.Offers{}
	}

	// Execute the search
	res, err := e.es.Search(
		e.es.Search.WithContext(ctx),
		e.es.Search.WithIndex("offers"),
		e.es.Search.WithBody(bytes.NewReader(reqJSON)),
	)
	if err != nil {
		log.Errorf("Error executing search: %s", err)
		return models.Offers{}
	}
	defer res.Body.Close()

	// Check for response errors
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

	// Aggregate results
	var offers []models.Offer
	for _, hit := range searchResult.Hits.Hits {
		offers = append(offers, hit.Source)
	}

	log.Infof("Retrieved %d offers with filtering", len(offers))
	return models.Offers{Offers: offers}
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
	return nil
}
