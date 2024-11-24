package models

import "sort"

// PriceRange represents the price range details.
type HistogramRange struct {
	Start uint64 `json:"start"`
	End   uint64 `json:"end"`
	Count uint64 `json:"count"`
}

type Bucket HistogramRange

func BucketizeOffersByPrice(offers []*Offer, width uint32) []Bucket {
	if width == 0 {
		panic("Width must be greater than 0")
	}

	bucketMap := make(map[uint32]*Bucket)

	// Distribute offers into buckets
	for _, offer := range offers {
		start := uint32(offer.Price) / width * width
		end := start + width

		if _, exists := bucketMap[start]; !exists {
			bucketMap[start] = &Bucket{Start: uint64(start), End: uint64(end)}
		}
		bucketMap[start].Count++
	}

	// Create a slice from the map and filter out empty buckets
	var buckets []Bucket
	for _, bucket := range bucketMap {
		if bucket.Count > 0 {
			buckets = append(buckets, *bucket)
		}
	}

	// Sort buckets by start ascending
	sort.Slice(buckets, func(i, j int) bool {
		return buckets[i].Start < buckets[j].Start
	})

	return buckets
}

func BucketizeOffersByKilometer(offers []*Offer, width uint32) []Bucket {
	if width == 0 {
		panic("Width must be greater than 0")
	}

	bucketMap := make(map[uint32]*Bucket)

	// Distribute offers into buckets
	for _, offer := range offers {
		start := uint32(offer.FreeKilometers) / width * width
		end := start + width

		if _, exists := bucketMap[start]; !exists {
			bucketMap[start] = &Bucket{Start: uint64(start), End: uint64(end)}
		}
		bucketMap[start].Count++
	}

	// Create a slice from the map and filter out empty buckets
	var buckets []Bucket
	for _, bucket := range bucketMap {
		if bucket.Count > 0 {
			buckets = append(buckets, *bucket)
		}
	}

	// Sort buckets by start ascending
	sort.Slice(buckets, func(i, j int) bool {
		return buckets[i].Start < buckets[j].Start
	})

	return buckets
}

// CarTypeCount represents the count of offers by car type.
type CarTypeCount struct {
	Small  uint64 `json:"small"`
	Sports uint64 `json:"sports"`
	Luxury uint64 `json:"luxury"`
	Family uint64 `json:"family"`
}

func (c *CarTypeCount) Add(carType string) {
	if carType == "small" {
		c.Small++
	} else if carType == "sports" {
		c.Sports++
	} else if carType == "luxury" {
		c.Luxury++
	} else if carType == "family" {
		c.Family++
	}
}

// VollkaskoCount represents the count of offers with and without Vollkasko.
type VollkaskoCount struct {
	TrueCount  uint64 `json:"trueCount"`
	FalseCount uint64 `json:"falseCount"`
}

func (v *VollkaskoCount) Add(hasVollkasko bool) {
	if hasVollkasko {
		v.TrueCount++
	} else {
		v.FalseCount++
	}
}

// SeatsCount represents the count of offers by number of seats.
type KVSeatsCount struct {
	NumberSeats uint64 `json:"numberSeats"`
	Count       uint64 `json:"count"`
}

type SeatsSummary map[uint64]*KVSeatsCount

func (seats *SeatsSummary) Add(numberSeats uint64) {
	if s, ok := (*seats)[numberSeats]; ok {
		s.Count++
	} else {
		(*seats)[numberSeats] = &KVSeatsCount{NumberSeats: numberSeats, Count: 1}
	}
}
