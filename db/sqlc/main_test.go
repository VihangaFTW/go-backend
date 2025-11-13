package db

import (
	"database/sql"
	"log"
	"os"
	"testing"

	"github.com/VihangaFTW/Go-Backend/util"
	_ "github.com/lib/pq"
)

// ? package level variable can be accessed by any file in package db
// memory address pointer definition

var testQueries *Queries
var testDB *sql.DB

func TestMain(m *testing.M) {

	var err error

	config, error := util.LoadConfig("../../")
	if error != nil {
		log.Fatal("cannot load config:", error)
	}

	//? connect to the database
	testDB, err = sql.Open(config.DBDriver, config.DBSource)
	if err != nil {
		log.Fatal("Cannot connect to the database!", err)
	}

	testQueries = New(testDB)
	
	os.Exit((m.Run()))

}
