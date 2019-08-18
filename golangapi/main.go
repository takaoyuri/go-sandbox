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
	"github.com/ant0ine/go-json-rest/rest"
	"github.com/inouet/ken-all/address"
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

	api := rest.NewApi()

	api.Use(
		[]rest.Middleware{
			&rest.AccessLogApacheMiddleware{},
			&rest.TimerMiddleware{},
			&rest.RecorderMiddleware{},
			&rest.RecoverMiddleware{},
			&rest.ContentTypeCheckerMiddleware{},
			&rest.CorsMiddleware{
				RejectNonCorsRequests: false,
				OriginValidator: func(origin string, request *rest.Request) bool {
					return true //origin == "http://my.other.host"
				},
				AllowedMethods: []string{"GET", "POST", "PUT"},
				AllowedHeaders: []string{
					"Accept", "Content-Type", "X-Custom-Header", "Origin"},
				AccessControlAllowCredentials: true,
				AccessControlMaxAge:           3600,
			},
			&rest.JsonpMiddleware{
				CallbackNameKey: "cb",
			},
		}...,
	)

	router, err := rest.MakeRouter(
		rest.Get("/zip_code/#zipcode", func(w rest.ResponseWriter, req *rest.Request) {
			zip := util.ParseZipCode(req.PathParam("zipcode"))
			if child, ok := myaddresses[zip]; ok {
				jsonObj := gabs.New()
				jsonObj.Set(child.Town, "town")
				jsonObj.Set(child.City, "city")
				jsonObj.Set(child.Pref, "pref")
				w.WriteJson(jsonObj.Data())
			} else {
				rest.NotFound(w, req)
			}
		}),
	)

	api.SetApp(router)
	log.Fatal(http.ListenAndServe(":8080", api.MakeHandler()))
}
