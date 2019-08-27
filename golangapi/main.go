package main

import (
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
)

//
func getAbsPath(relPath string) string {
	absPath, err := filepath.Abs(relPath)
	if err != nil {
		log.Fatal(err)
	}
	return absPath
}

func main() {

	kenAllCsv := "./KEN_ALL.CSV"
	ioReader, err := os.Open(getAbsPath(kenAllCsv))

	if err != nil {
		log.Fatal(err)
	}

	defer ioReader.Close()

	reader := address.NewReader(transform.NewReader(ioReader, japanese.ShiftJIS.NewDecoder()))

	var myaddresses map[string]myaddress.Address
	myaddresses = map[string]myaddress.Address{}

	for {
		cols, err := reader.Read()

		if err == io.EOF {
			break
		}

		rows := address.NewRows(cols)

		for _, row := range rows {
			address := myaddress.NewAddress(row)
			myaddresses[address.Zip] = address
		}
	}

	fmt.Println("start server")

	e := echo.New()

	e.Use(middleware.Logger())
	e.Use(middleware.Recover())
	e.Use(middleware.CORS())

	e.GET("/zip_code/:zipcode", func(c echo.Context) error {

		zip := util.ParseZipCode(c.Param("zipcode"))
		if child, ok := myaddresses[zip]; ok {

			jsonObj := gabs.New()
			jsonObj.Set(child.Town, "town")
			jsonObj.Set(child.City, "city")
			jsonObj.Set(child.Pref, "pref")

			cb := c.QueryParam("cb")
			if len(cb) > 0 {
				return c.JSONP(http.StatusOK, cb, jsonObj.Data())
			} else {
				return c.JSON(http.StatusOK, jsonObj.Data())
			}

		} else {
			return echo.NewHTTPError(http.StatusNotFound, "Not Found")
		}
	})

	e.Logger.Fatal(e.Start(":8080"))
}
