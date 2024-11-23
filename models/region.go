package models

import (
	"encoding/json"
	"io/ioutil"
	"log"
	"os"
)

type Region struct {
	Id         uint64   `json:"id"`
	Name       string   `json:"name"`
	SubRegions []Region `json:"subRegions"`
}

var RegionTree Region = readRegions("regions.json")

func (region *Region) GetRegionById(id uint64) *Region {
	if region.Id == id {
		return region
	}

	for _, subRegion := range region.SubRegions {
		if found := subRegion.GetRegionById(id); found != nil {
			return found
		}
	}

	return nil
}

func (region *Region) GetLeafs() []Region {
	if len(region.SubRegions) == 0 {
		return []Region{*region}
	}

	leafs := make([]Region, 0)
	for _, subRegion := range region.SubRegions {
		leafs = append(leafs, subRegion.GetLeafs()...)
	}
	return leafs
}

func readRegions(filePath string) Region {
	file, err := os.Open(filePath)
	if err != nil {
		log.Fatalf("failed to open file: %v", err)
	}
	defer file.Close()

	byteValue, err := ioutil.ReadAll(file)
	if err != nil {
		log.Fatalf("failed to read file: %v", err)
	}

	var region Region
	if err := json.Unmarshal(byteValue, &region); err != nil {
		log.Fatalf("failed to unmarshal JSON: %v", err)
	}

	return region
}