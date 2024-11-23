package db

import (
	"check_republic/models"

	"github.com/gofiber/fiber/v2/log"
	"github.com/hashicorp/go-memdb"
)

var Db *memdb.MemDB

func Init() {
	db, err := createDB()
	if err != nil {
		log.Fatal(err)
	}
	Db = db
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

func CreateOffers(o ...models.Offer) {
	// Start a new transaction for writing
	txn := Db.Txn(true)
	for _, offer := range o {
		log.Info("Inserting offer: ", offer)
		err := txn.Insert("offer", offer)
		if err != nil {
			log.Error(err)
		}
	}
	txn.Commit()
}

func GetAllOffers() models.Offers {
	txn := Db.Txn(false)
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
