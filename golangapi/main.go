package main

import (
	"bufio"
	"github.com/takaoyuri/go-sandbox/golangapi/util"
	"log"
	"net/http"
	"os"
	"path/filepath"
	"strings"

	"github.com/Jeffail/gabs/v2"
	"github.com/ant0ine/go-json-rest/rest"
)

func main() {
	kenAllFileName := "./ken-all.json"
	kenAllPath, err := filepath.Abs(kenAllFileName)
	if err != nil {
		log.Fatal(err)
	}

	file, err := os.Open(kenAllPath)
	if err != nil {
		log.Fatal(err)
	}

	var scanedText []string

	scanner := bufio.NewScanner(file)
	scanner.Scan()

	for scanner.Scan() {
		scanedText = append(scanedText, scanner.Text())
	}

	joinedText := strings.Join(scanedText, ",")
	joinedText = "{\"data\":[" + joinedText + "]}"

	jsonParsed, err := gabs.ParseJSON([]byte(joinedText))
	if err != nil {
		panic(err)
	}

	// addressMap map[zipcode]addressData{}
	var addressList map[string][]*gabs.Container
	addressList = map[string][]*gabs.Container{}
	for _, child := range jsonParsed.Path("data").Children() {
		value, ok := child.Search("zip").Data().(string)
		if _, ok2 := addressList[value]; ok && !ok2 {
			jsonObj := gabs.New()
			jsonObj.Set(child.Path("town").Data().(string), "town")
			jsonObj.Set(child.Path("city").Data().(string), "city")
			jsonObj.Set(child.Path("pref").Data().(string), "pref")

			addressList[value] = append(addressList[value], jsonObj)
		}
	}
	file.Close()
	jsonParsed = nil

	api := rest.NewApi()
	// api.Use(rest.DefaultDevStack...)

	api.Use(
		[]rest.Middleware{
			&rest.AccessLogApacheMiddleware{},
			&rest.TimerMiddleware{},
			&rest.RecorderMiddleware{},
			&rest.RecoverMiddleware{
				EnableResponseStackTrace: true,
			},
			&rest.JsonIndentMiddleware{},
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
			// &rest.GzipMiddleware{},
		}...,
	)

	router, err := rest.MakeRouter(

		rest.Get("/#zipcode", func(w rest.ResponseWriter, req *rest.Request) {
			zip := util.ParseZipCode(req.PathParam("zipcode"))

			if child, ok := addressList[zip]; ok {
				if len(child) == 1 {
					w.WriteJson(child[0].Data())
				} else {
					// todo

				}
			}
		}),
	)

	api.SetApp(router)
	log.Fatal(http.ListenAndServe(":8080", api.MakeHandler()))
}
