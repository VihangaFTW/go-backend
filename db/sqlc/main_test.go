package db

import (
	"database/sql"
	"log"
	"os"
	"testing"

	_ "github.com/lib/pq"
)

const (
	dbDriver = "postgres"
	dbSource = "postgres://root:12345@localhost:5432/simple_bank?sslmode=disable"
)

// ? package level variable can be accessed by any file in package db
// memory address pointer definition
var testQueries *Queries
var testDB *sql.DB

func TestMain(m *testing.M) {

	var err error

	//? connect to the database
	testDB, err = sql.Open(dbDriver, dbSource)
	if err != nil {
		log.Fatal("Cannot connect to the database!", err)
	}

	testQueries = New(testDB)

	os.Exit((m.Run()))

}
