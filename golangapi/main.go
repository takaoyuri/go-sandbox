package main

import (
	"database/sql"
	"fmt"
	"io"
	"log"
	"net/http"
	"os"
	"path/filepath"

	"golang.org/x/text/encoding/japanese"
	"golang.org/x/text/transform"

	myaddress "github.com/takaoyuri/go-sandbox/golangapi/address"
	"github.com/takaoyuri/go-sandbox/golangapi/util"

	"github.com/Jeffail/gabs/v2"
	"github.com/inouet/ken-all/address"

	"github.com/labstack/echo/v4"
	"github.com/labstack/echo/v4/middleware"
	_ "github.com/mattn/go-sqlite3"
)

func getAbsPath(relPath string) string {
	absPath, err := filepath.Abs(relPath)
	if err != nil {
		log.Fatal(err)
	}
	return absPath
}

func initDatabase() (*sql.DB, error) {
	db, err := sql.Open("sqlite3", "./addresses.db")
	if err != nil {
		return nil, err
	}

	createTableSQL := `
	CREATE TABLE IF NOT EXISTS addresses (
		zip_code TEXT PRIMARY KEY,
		prefecture TEXT NOT NULL,
		city TEXT NOT NULL,
		town TEXT NOT NULL,
		created_at DATETIME DEFAULT CURRENT_TIMESTAMP
	);
	`

	_, err = db.Exec(createTableSQL)
	if err != nil {
		return nil, err
	}

	return db, nil
}

func migrateFromCSV(db *sql.DB) error {
	var count int
	err := db.QueryRow("SELECT COUNT(*) FROM addresses").Scan(&count)
	if err != nil {
		return err
	}

	if count > 0 {
		fmt.Println("Database already populated, skipping migration")
		return nil
	}

	fmt.Println("Migrating data from CSV to database...")

	kenAllCsv := "./KEN_ALL.CSV"
	ioReader, err := os.Open(getAbsPath(kenAllCsv))
	if err != nil {
		return err
	}
	defer ioReader.Close()

	reader := address.NewReader(transform.NewReader(ioReader, japanese.ShiftJIS.NewDecoder()))

	tx, err := db.Begin()
	if err != nil {
		return err
	}
	defer tx.Rollback()

	stmt, err := tx.Prepare("INSERT OR REPLACE INTO addresses (zip_code, prefecture, city, town) VALUES (?, ?, ?, ?)")
	if err != nil {
		return err
	}
	defer stmt.Close()

	for {
		cols, err := reader.Read()
		if err == io.EOF {
			break
		}
		if err != nil {
			return err
		}

		rows := address.NewRows(cols)
		for _, row := range rows {
			addr := myaddress.NewAddress(row)
			_, err = stmt.Exec(addr.Zip, addr.Pref, addr.City, addr.Town)
			if err != nil {
				return err
			}
		}
	}

	fmt.Println("Migration completed successfully")
	return tx.Commit()
}

func getAddressFromDB(db *sql.DB, zipCode string) (*myaddress.Address, error) {
	var addr myaddress.Address
	err := db.QueryRow("SELECT zip_code, prefecture, city, town FROM addresses WHERE zip_code = ?", zipCode).Scan(
		&addr.Zip, &addr.Pref, &addr.City, &addr.Town,
	)
	if err != nil {
		return nil, err
	}
	return &addr, nil
}

func main() {
	db, err := initDatabase()
	if err != nil {
		log.Fatal("Failed to initialize database:", err)
	}
	defer db.Close()

	err = migrateFromCSV(db)
	if err != nil {
		log.Fatal("Failed to migrate data:", err)
	}

	fmt.Println("start server")

	e := echo.New()

	e.Use(middleware.Logger())
	e.Use(middleware.Recover())
	e.Use(middleware.CORS())

	e.GET("/zip_code/:zipcode", func(c echo.Context) error {
		zip := util.ParseZipCode(c.Param("zipcode"))
		addr, err := getAddressFromDB(db, zip)
		if err != nil {
			if err == sql.ErrNoRows {
				return echo.NewHTTPError(http.StatusNotFound, "Not Found")
			}
			return echo.NewHTTPError(http.StatusInternalServerError, "Database error")
		}

		jsonObj := gabs.New()
		jsonObj.Set(addr.Town, "town")
		jsonObj.Set(addr.City, "city")
		jsonObj.Set(addr.Pref, "pref")

		cb := c.QueryParam("cb")
		if len(cb) > 0 {
			return c.JSONP(http.StatusOK, cb, jsonObj.Data())
		} else {
			return c.JSON(http.StatusOK, jsonObj.Data())
		}
	})

	e.Logger.Fatal(e.Start(":8080"))
}
