package main

import (
	"encoding/json"
	"log"
	"net/http"

	"github.com/formikejo/htmlpdf"
)

type j map[string]interface{}

func swagger(w http.ResponseWriter, r *http.Request) {
	j, err := json.MarshalIndent(j{
		"swagger": "2.0",
		"info": j{
			"version":     "1.0.0",
			"title":       "htmlpdfd",
			"description": "A microservice for generating PDFs from HTML",
		},
		"host":     r.Host,
		"basePath": "/",
		"schemes":  []string{"http"},
		"paths": j{
			"/pdf": j{
				"get": j{
					"description": "Generate a PDF from HTML and return it in the response",
					"produces":    []string{"application/pdf"},
					"responses": j{
						"200": j{
							"description": "The PDF was successfully rendered",
						},
					},
					"parameters": []j{
						j{
							"name":        "url",
							"in":          "query",
							"description": "URL of the HTML to render",
							"required":    true,
							"type":        "string",
						},
					},
				},
			},
		},
	}, "", "  ")
	if err != nil {
		log.Println(err)
		return
	}

	w.Header().Set("Access-Control-Allow-Origin", "*")
	w.Write(j)
}

func main() {
	creator, err := htmlpdf.NewWkhtmltopdf()
	if err != nil {
		log.Fatal(err)
	}

	mux := http.NewServeMux()
	mux.Handle("/pdf", &htmlpdf.PDFHandler{
		Creator: creator,
	})
	mux.HandleFunc("/swagger.json", swagger)
	log.Println("Starting server")
	http.ListenAndServe(":8080", mux)
}
