package db

import (
	"check_republic/models"

	"github.com/gofiber/fiber/v2/log"
	"github.com/hashicorp/go-memdb"
)

var DB OfferDatabase

type OfferDatabase interface {
	CreateOffers(o ...models.Offer)
	GetAllOffers() models.Offers
	DeleteAllOffers()
}

type MemDB struct {
	Db *memdb.MemDB
}

func Init() {
	db, err := createDB()
	if err != nil {
		log.Fatal(err)
	}
	DB = MemDB{Db: db}
	log.Info("Database created")
}

func createDB() (*memdb.MemDB, error) {
	// Create the DB schema
	schema := &memdb.DBSchema{
		Tables: map[string]*memdb.TableSchema{
			"offer": {
				Name: "offer",
				Indexes: map[string]*memdb.IndexSchema{
					"id": {
						Name:    "id",
						Unique:  true,
						Indexer: &memdb.StringFieldIndex{Field: "ID"}, // TODO change this
					},
				},
			},
		},
	}

	// Create a new data base
	return memdb.NewMemDB(schema)
}

func (m MemDB) CreateOffers(o ...models.Offer) {
	// Start a new transaction for writing
	txn := m.Db.Txn(true)
	for _, offer := range o {
		log.Info("Inserting offer: ", offer)
		err := txn.Insert("offer", offer)
		if err != nil {
			log.Error(err)
		}
	}
	txn.Commit()
}

func (m MemDB) GetAllOffers() models.Offers {
	txn := m.Db.Txn(false)
	defer txn.Abort()

	it, err := txn.Get("offer", "id")
	if err != nil {
		log.Error(err)
	}

	var offers []models.Offer
	for obj := it.Next(); obj != nil; obj = it.Next() {
		log.Info("Found offer: ", obj)
		p := obj.(models.Offer)
		offers = append(offers, p)
	}
	return models.Offers{Offers: offers}
}

func (m MemDB) DeleteAllOffers() {
	txn := m.Db.Txn(true)
	defer txn.Abort()

	_, err := txn.DeleteAll("offer", "id")
	if err != nil {
		log.Error(err)
	}

	txn.Commit()
}
