package db

import (
	"check_republic/models"
	"context"

	"github.com/gofiber/fiber/v2/log"
	"github.com/jackc/pgx/v5"
)

type PostgresDB struct {
	Db *pgx.Conn
}

func InitPostgres() {
	db, err := createPostgres()
	if err != nil {
		log.Fatal(err)
	}
	DB = PostgresDB{Db: db}
	log.Info("Database created")

	// Create table
	_, err = db.Exec(context.Background(), `
		CREATE TABLE IF NOT EXISTS offers
(
    id              uuid     not null
        primary key,
    data            text     not null,
    region_id       integer  not null,
    start_date      bigint   not null,
    end_date        bigint   not null,
    number_seats    smallint not null,
    price           integer  not null,
    car_type        text     not null,
    has_vollkasko   boolean  not null,
    free_kilometers smallint not null
);`)
	if err != nil {
		log.Fatal(err)
	}

}

func createPostgres() (*pgx.Conn, error) {
	ctx := context.Background()
	conn, err := pgx.Connect(ctx, "postgres://postgres:postgres@localhost/postgres")
	if err != nil {
		log.Fatalf("Unable to connect to database: %v\n", err)
	}
	return conn, err
}

func (p PostgresDB) CreateOffers(ctx context.Context, o ...models.Offer) error {
	tx, err := p.Db.Begin(ctx)
	if err != nil {
		return err
	}
	defer tx.Rollback(ctx)

	insertQuery := `
        INSERT INTO public.offers (
            id, data, region_id, start_date, end_date,
            number_seats, price, car_type, has_vollkasko, free_kilometers
        )
        VALUES ($1, $2, $3, $4, $5, $6, $7, $8, $9, $10)
    `

	for _, offer := range o {
		_, err := tx.Exec(ctx, insertQuery,
			offer.ID,
			offer.Data,
			offer.MostSpecificRegionID,
			offer.StartDate,
			offer.EndDate,
			offer.NumberSeats,
			offer.Price,
			offer.CarType,
			offer.HasVollkasko,
			offer.FreeKilometers,
		)
		if err != nil {
			log.Error("Unable to create offer", err)
			return err
		}
		log.Debug("Offer created")
	}

	// Commit transaction
	if err := tx.Commit(ctx); err != nil {
		log.Error("Unable to commit transaction")
		return err
	}

	return nil
}

func (p PostgresDB) GetAllOffers(ctx context.Context) models.Offers {
	query := `SELECT * FROM offers`
	rows, err := p.Db.Query(ctx, query)
	if err != nil {
		log.Fatalf("Unable to fetch offers: %v\n", err)
	}

	offers, err := pgx.CollectRows(rows, pgx.RowToStructByName[models.Offer])
	if err != nil {
		log.Fatalf("Unable to collect offers: %v\n", err)
	}

	return models.Offers{
		Offers: offers,
	}
}

func (p PostgresDB) GetFilteredOffers(ctx context.Context, regionID uint64, timeRangeStart uint64, timeRangeEnd uint64, numberDays uint64, sortOrder string, page uint64, pageSize uint64, priceRangeWidth uint64, minFreeKilometerWidth uint64, minNumberSeats *uint64, minPrice *uint64, maxPrice *uint64, carType *string, onlyVollkasko *bool, minFreeKilometer *uint64) models.Offers {
	// Query for offers
	query := `
		SELECT *
		FROM offers
		WHERE region_id = any($1)  -- Match the region or any of its subregions
		AND start_date >= $2 AND end_date <= $3
		AND ($4::int IS NULL OR number_seats >= $4)
		AND ($5::int IS NULL OR price >= $5)
		AND ($6::int IS NULL OR price < $6)
		AND ($7::text IS NULL OR car_type = $7)
		AND (NOT $8 OR has_vollkasko = true)
		AND ($9::int IS NULL OR free_kilometers >= $9)
		ORDER BY price ` + parseSortOrder(sortOrder) + `
		LIMIT $10 OFFSET $11
    `

	// Pagination offset
	offset := (page - 1) * pageSize

	// Query execution
	rows, err := p.Db.Query(ctx, query,
		regionIdToMostSpecificRegionId[regionID],
		timeRangeStart,
		timeRangeEnd,
		minNumberSeats,
		minPrice,
		maxPrice,
		carType,
		onlyVollkasko,
		minFreeKilometer,
		pageSize,
		offset,
	)
	if err != nil {
		log.Error("Unable to fetch offers", err)
		return models.Offers{}
	}
	defer rows.Close()

	// Parse offers
	offers, err := pgx.CollectRows(rows, pgx.RowToStructByName[models.Offer])
	if err != nil {
		log.Error("Unable to collect offers: %v\n", err)
	}

	return models.Offers{Offers: offers}
}

// Helper to parse sortOrder
func parseSortOrder(sortOrder string) string {
	if sortOrder == "price-desc" {
		return "DESC"
	}
	return "ASC"
}

func (p PostgresDB) DeleteAllOffers(ctx context.Context) error {
	query := `DELETE FROM offers`
	_, err := p.Db.Exec(ctx, query)
	return err
}

var regionIdToMostSpecificRegionId = map[uint64][]uint64{
	58:  {58},
	59:  {59},
	21:  {58, 59},
	60:  {60},
	61:  {61},
	22:  {60, 61},
	62:  {62},
	63:  {63},
	23:  {62, 63},
	7:   {58, 59, 60, 61, 62, 63},
	64:  {64},
	65:  {65},
	24:  {64, 65},
	66:  {66},
	67:  {67},
	25:  {66, 67},
	68:  {68},
	69:  {69},
	26:  {68, 69},
	70:  {70},
	71:  {71},
	27:  {70, 71},
	72:  {72},
	73:  {73},
	28:  {72, 73},
	8:   {64, 65, 66, 67, 68, 69, 70, 71, 72, 73},
	74:  {74},
	75:  {75},
	29:  {74, 75},
	76:  {76},
	77:  {77},
	30:  {76, 77},
	9:   {74, 75, 76, 77},
	1:   {58, 59, 60, 61, 62, 63, 64, 65, 66, 67, 68, 69, 70, 71, 72, 73, 74, 75, 76, 77},
	78:  {78},
	79:  {79},
	80:  {80},
	81:  {81},
	31:  {78, 79, 80, 81},
	82:  {82},
	83:  {83},
	32:  {82, 83},
	84:  {84},
	85:  {85},
	33:  {84, 85},
	86:  {86},
	87:  {87},
	34:  {86, 87},
	88:  {88},
	89:  {89},
	35:  {88, 89},
	10:  {78, 79, 80, 81, 82, 83, 84, 85, 86, 87, 88, 89},
	90:  {90},
	91:  {91},
	36:  {90, 91},
	92:  {92},
	93:  {93},
	37:  {92, 93},
	11:  {90, 91, 92, 93},
	2:   {78, 79, 80, 81, 82, 83, 84, 85, 86, 87, 88, 89, 90, 91, 92, 93},
	94:  {94},
	95:  {95},
	38:  {94, 95},
	96:  {96},
	97:  {97},
	39:  {96, 97},
	12:  {94, 95, 96, 97},
	98:  {98},
	99:  {99},
	40:  {98, 99},
	100: {100},
	41:  {100},
	101: {101},
	102: {102},
	42:  {101, 102},
	13:  {98, 99, 100, 101, 102},
	103: {103},
	43:  {103},
	104: {104},
	105: {105},
	44:  {104, 105},
	14:  {103, 104, 105},
	3:   {94, 95, 96, 97, 98, 99, 100, 101, 102, 103, 104, 105},
	106: {106},
	107: {107},
	45:  {106, 107},
	108: {108},
	109: {109},
	46:  {108, 109},
	15:  {106, 107, 108, 109},
	110: {110},
	111: {111},
	47:  {110, 111},
	112: {112},
	113: {113},
	48:  {112, 113},
	16:  {110, 111, 112, 113},
	4:   {106, 107, 108, 109, 110, 111, 112, 113},
	114: {114},
	115: {115},
	49:  {114, 115},
	116: {116},
	117: {117},
	50:  {116, 117},
	17:  {114, 115, 116, 117},
	118: {118},
	51:  {118},
	119: {119},
	120: {120},
	52:  {119, 120},
	18:  {118, 119, 120},
	5:   {114, 115, 116, 117, 118, 119, 120},
	121: {121},
	53:  {121},
	122: {122},
	54:  {122},
	123: {123},
	124: {124},
	55:  {123, 124},
	19:  {121, 122, 123, 124},
	56:  {56},
	57:  {57},
	20:  {56, 57},
	6:   {121, 122, 123, 124, 56, 57},
}
