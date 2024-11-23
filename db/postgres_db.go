package db

import (
	"check_republic/models"
	"context"
	"fmt"

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

	// Create car types enum
	_, err = db.Exec(context.Background(), `
	DO $$ BEGIN
CREATE TYPE car_type_enum AS ENUM ('small', 'sports', 'luxury', 'family');
EXCEPTION
    WHEN duplicate_object THEN null;
END $$;
`)
	if err != nil {
		log.Fatal(err)
	}

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
    car_type        text not null,
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

func buildBaseQuery() string {
	return `
	SELECT *
	FROM offers
	WHERE region_id = ANY($1)  -- Match the region or any of its subregions
	AND start_date >= $2 
	AND end_date <= $3
	AND (end_date - start_date) >= ($4 * 24 * 60 * 60 * 1000)  -- The number of full days (24h) the car is available within the rangeStart and rangeEnd
	`
}

func countPriceRange(baseQuery string) string {
	return fmt.Sprintf(`WITH offers AS (
		%s),
	price_ranges AS (
		SELECT * from generate_series(0, (SELECT max(price) FROM offers), $5) AS range_start
	)

    SELECT
        pr.range_start,
		pr.range_start + $5 AS range_end,
        COUNT(*) AS item_count
    FROM
        price_ranges pr
    INNER JOIN
        offers t ON t.price >= pr.range_start AND t.price < pr.range_start + $5
    GROUP BY
        pr.range_start`, baseQuery)
}

func countCarType(baseQuery string) string {
	return fmt.Sprintf(`WITH offers AS (
		%s)
			SELECT
				car_type,
				count(*) as count
			FROM offers
			GROUP BY car_type`, baseQuery)
}

func countSeats(baseQuery string) string {
	return fmt.Sprintf(`
	WITH offers AS (
		%s)
			SELECT
				number_seats,
				count(*) as count
			FROM offers
			GROUP BY number_seats
		`, baseQuery)
}

func countFreeKilometer(baseQuery string) string {
	return fmt.Sprintf(`
	WITH offers AS (
		%s),
	free_kilometer_ranges AS (
    SELECT generate_series(0, (SELECT max(free_kilometers) FROM offers), $5) AS range_start
	)

    SELECT
        fkr.range_start,
		fkr.range_start + $5 AS range_end,
        COUNT(*) AS item_count
    FROM
        free_kilometer_ranges fkr
    INNER JOIN
        offers t ON t.free_kilometers >= fkr.range_start AND t.free_kilometers < fkr.range_start + $5
    GROUP BY
        fkr.range_start
`, baseQuery)
}

func countVollkasko(baseQuery string) string {
	return fmt.Sprintf(`
	WITH offers AS (
		%s)
			SELECT
				has_vollkasko,
				count(*) as count
			FROM offers
			GROUP BY has_vollkasko`, baseQuery)
}

func orderedAndPaginated(baseQuery string, sortOrder string) string {
	return fmt.Sprintf(`
	WITH offers AS (
		%s)
			SELECT *
			FROM offers
			WHERE ($5::int IS NULL OR number_seats >= $5)
		AND ($6::int IS NULL OR price >= $6)
		AND ($7::int IS NULL OR price < $7)
		AND ($8::text IS NULL OR car_type = $8)
		AND (NOT $9 OR has_vollkasko = true)
		AND ($10::int IS NULL OR free_kilometers >= $10)
		ORDER BY price %s
		LIMIT $11 OFFSET $12 `, baseQuery, sortOrder)

}

func (p PostgresDB) GetFilteredOffers(ctx context.Context, regionID uint64, timeRangeStart uint64, timeRangeEnd uint64, numberDays uint64, sortOrder string, page uint64, pageSize uint64, priceRangeWidth uint64, minFreeKilometerWidth uint64, minNumberSeats *uint64, minPrice *uint64, maxPrice *uint64, carType *string, onlyVollkasko *bool, minFreeKilometer *uint64) (dto models.DTO) { // Pagination offset
	baseQuery := buildBaseQuery()
	validRegionIds := models.RegionIdToMostSpecificRegionId[int32(regionID)]
	log.Infof("Valid region ids: %v", validRegionIds)

	offset := page * pageSize

	// Query execution
	rows, err := p.Db.Query(ctx, orderedAndPaginated(baseQuery, parseSortOrder(sortOrder)),
		validRegionIds,
		timeRangeStart,
		timeRangeEnd,
		numberDays,
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
		return dto
	}
	defer rows.Close()

	// Parse offers
	offers, err := pgx.CollectRows(rows, pgx.RowToStructByName[models.Offer])
	if err != nil {
		log.Error("Unable to collect offers: %v\n", err)
	}

	// collect price range
	rows, err = p.Db.Query(ctx, countPriceRange(baseQuery),
		validRegionIds,
		timeRangeStart,
		timeRangeEnd,
		numberDays,
		priceRangeWidth)
	if err != nil {
		log.Error("Unable to fetch price range", err)
		return dto
	}
	defer rows.Close()

	var priceRanges []models.PriceRange
	for rows.Next() {
		var priceRange models.PriceRange
		err = rows.Scan(&priceRange.Start, &priceRange.End, &priceRange.Count)
		if err != nil {
			log.Error("Unable to scan price range", err)
			return dto
		}
		priceRanges = append(priceRanges, priceRange)
	}

	// collect car type count
	rows, err = p.Db.Query(ctx, countCarType(baseQuery),
		validRegionIds,
		timeRangeStart,
		timeRangeEnd,
		numberDays,
	)
	if err != nil {
		log.Error("Unable to fetch car type count", err)
		return dto
	}
	defer rows.Close()

	carTypeCountMap := make(map[string]uint64)
	for rows.Next() {
		var carType string
		var count uint64
		err = rows.Scan(&carType, &count)
		if err != nil {
			log.Error("Unable to scan car type count", err)
			return dto
		}
		carTypeCountMap[carType] = count
	}

	var carTypeCount models.CarTypeCount
	carTypeCount.Small = carTypeCountMap["small"]
	carTypeCount.Sports = carTypeCountMap["sports"]
	carTypeCount.Luxury = carTypeCountMap["luxury"]
	carTypeCount.Family = carTypeCountMap["family"]

	// collect seats count
	rows, err = p.Db.Query(ctx, countSeats(baseQuery),
		validRegionIds,
		timeRangeStart,
		timeRangeEnd,
		numberDays,
	)
	if err != nil {
		log.Error("Unable to fetch seats count", err)
		return dto
	}
	defer rows.Close()

	var seatsCountList []models.SeatsCount
	for rows.Next() {
		var seatsCount models.SeatsCount
		err = rows.Scan(&seatsCount.NumberSeats, &seatsCount.Count)
		if err != nil {
			log.Error("Unable to scan seats count", err)
			return dto
		}
		seatsCountList = append(seatsCountList, seatsCount)
	}

	// collect free kilometer range
	rows, err = p.Db.Query(ctx, countFreeKilometer(baseQuery),
		validRegionIds,
		timeRangeStart,
		timeRangeEnd,
		numberDays,
		minFreeKilometerWidth)
	if err != nil {
		log.Error("Unable to fetch free kilometer range", err)
		return dto
	}
	defer rows.Close()

	var freeKilometerRanges []models.FreeKilometerRange
	for rows.Next() {
		var freeKilometerRange models.FreeKilometerRange
		err = rows.Scan(&freeKilometerRange.Start, &freeKilometerRange.End, &freeKilometerRange.Count)
		if err != nil {
			log.Error("Unable to scan free kilometer range", err)
			return dto
		}
		freeKilometerRanges = append(freeKilometerRanges, freeKilometerRange)
	}

	// collect vollkasko count
	rows, err = p.Db.Query(ctx, countVollkasko(baseQuery),
		validRegionIds,
		timeRangeStart,
		timeRangeEnd,
		numberDays,
	)
	if err != nil {
		log.Error("Unable to fetch vollkasko count", err)
		return dto
	}
	defer rows.Close()

	var vollkaskoCount models.VollkaskoCount
	for rows.Next() {
		var hasVollkasko bool
		var count uint64
		err = rows.Scan(&hasVollkasko, &count)
		if err != nil {
			log.Error("Unable to scan vollkasko count", err)
			return dto
		}
		if hasVollkasko {
			vollkaskoCount.TrueCount = count
		} else {
			vollkaskoCount.FalseCount = count
		}
	}

	return models.DTO{Offers: offers, PriceRanges: priceRanges, CarTypeCounts: carTypeCount, SeatsCount: seatsCountList, FreeKilometerRange: freeKilometerRanges, VollkaskoCount: vollkaskoCount}
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
