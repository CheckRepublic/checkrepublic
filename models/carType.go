package models

type CarType int

const (
	Small CarType = iota
	Sports
	Luxury
	Family
)

var CarTypeName = map[CarType]string{
    Small:      "small",
    Sports: "sports",
    Luxury:     "luxury",
    Family:  "family",
}

var CarTypeValue = map[string]CarType{
    "small":      Small,
    "sports": Sports,
    "luxury":     Luxury,
    "family":  Family,
}
