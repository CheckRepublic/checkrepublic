package models

type CarType int

const (
	Small CarType = iota
	Sports
	Luxury
	Family
)

var carTypeName = map[CarType]string{
    Small:      "small",
    Sports: "sports",
    Luxury:     "luxury",
    Family:  "family",
}

func (carType CarType) String() string {
    return carTypeName[carType]
}
