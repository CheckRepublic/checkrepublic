package db

import (
	"check_republic/models"

	"github.com/gofiber/fiber/v2/log"
)

func parseHistogramBuckets(agg interface{}, interval uint64) []models.HistogramRange {
	if agg == nil {
		return nil
	}

	buckets, ok := agg.(map[string]interface{})["buckets"].([]interface{})
	if !ok {
		return nil
	}

	var ranges []models.HistogramRange
	for _, bucket := range buckets {
		bucketMap := bucket.(map[string]interface{})
		log.Info(bucketMap)
		count := uint64(bucketMap["doc_count"].(float64))
		if count == 0 {
			continue
		}
		ranges = append(ranges, models.HistogramRange{
			Start: uint64(bucketMap["key"].(float64)),
			End:   uint64(bucketMap["key"].(float64)) + interval,
			Count: count,
		})
	}

	return ranges
}

func parseCarTypeCounts(agg interface{}) models.CarTypeCount {
	if agg == nil {
		return models.CarTypeCount{}
	}

	buckets, ok := agg.(map[string]interface{})["buckets"].([]interface{})
	if !ok {
		return models.CarTypeCount{}
	}
	log.Debug("CarTypeCounts", buckets)

	carTypeCount := models.CarTypeCount{}
	for _, bucket := range buckets {
		bucketMap := bucket.(map[string]interface{})
		key := bucketMap["key"].(string)
		count := uint64(bucketMap["doc_count"].(float64))
		switch key {
		case "small":
			carTypeCount.Small = count
		case "sports":
			carTypeCount.Sports = count
		case "luxury":
			carTypeCount.Luxury = count
		case "family":
			carTypeCount.Family = count
		}
	}

	return carTypeCount
}

func parseSeatsCount(agg interface{}) []models.SeatsCount {
	if agg == nil {
		return nil
	}

	buckets, ok := agg.(map[string]interface{})["buckets"].([]interface{})
	if !ok {
		return nil
	}

	var counts []models.SeatsCount
	for _, bucket := range buckets {
		bucketMap := bucket.(map[string]interface{})
		counts = append(counts, models.SeatsCount{
			NumberSeats: uint64(bucketMap["key"].(float64)),
			Count:       uint64(bucketMap["doc_count"].(float64)),
		})
	}

	return counts
}

func parseVollkaskoCount(agg interface{}) models.VollkaskoCount {
	if agg == nil {
		return models.VollkaskoCount{}
	}

	buckets, ok := agg.(map[string]interface{})["buckets"].([]interface{})
	if !ok {
		return models.VollkaskoCount{}
	}

	log.Debug(buckets)
	vollkaskoCount := models.VollkaskoCount{}
	for _, bucket := range buckets {
		bucketMap := bucket.(map[string]interface{})
		key := uint64(bucketMap["key"].(float64))
		count := uint64(bucketMap["doc_count"].(float64))
		if key == 1 {
			vollkaskoCount.TrueCount = count
		} else {
			vollkaskoCount.FalseCount = count
		}
	}

	return vollkaskoCount
}
