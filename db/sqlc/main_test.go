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
	dbSource = "postgresql://root:secret@localhost:5432/simple_bank?sslmode=disable"
)

var testQueries *Queries
var testDB *sql.DB

// fonksiyonu, Go test paketinin sunduğu özel bir fonsiyondur. Testler başlamadan önce çalışır; bu nedenle genellikle:
// Test ortamının kurulması (örneğin veritabanı bağlantısı açma, gerekli yapılandırmaların ayarlanması),
// Kaynakların paylaşılması,
// Testler bitiminde temizleme işlemlerinin yapılması (örneğin veritabanı bağlantısının kapatılması),
func TestMain(m *testing.M) {
	var err error

	testDB, err = sql.Open(dbDriver, dbSource)
	if err != nil {
		log.Fatal("cannot connect to db:", err)
	}
	testQueries = New(testDB)
	os.Exit(m.Run())

}
