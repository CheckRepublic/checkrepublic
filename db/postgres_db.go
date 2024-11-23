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
        WITH RECURSIVE region_tree AS (
			SELECT id
			FROM regions
			WHERE id = $1  -- Starting region
			UNION ALL
			SELECT r.id
			FROM regions r
			INNER JOIN region_tree rt ON r.parent_id = rt.id
		)
		SELECT *
		FROM offers
		WHERE region_id IN (SELECT id FROM region_tree)  -- Match the region or any of its subregions
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
		regionID,
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
