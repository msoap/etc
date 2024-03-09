/*
Apple Health export data converter from XML to JSON

Usage:

	go run apple-health-xml2json.go export.zip > export.json

Install:

	GO111MODULE=off go get -u github.com/msoap/etc/apple-health-xml2json
*/
package main

import (
	"archive/zip"
	"encoding/json"
	"encoding/xml"
	"log"
	"os"
)

type healthData struct {
	HealthData []record `xml:"Record"`
}

type record struct {
	Type          string `xml:"type,attr"          json:"type"`
	SourceName    string `xml:"sourceName,attr"    json:"source_name"`
	SourceVersion string `xml:"sourceVersion,attr" json:"source_version"`
	Unit          string `xml:"unit,attr"          json:"unit"`
	CreationDate  string `xml:"creationDate,attr"  json:"creation_date"`
	StartDate     string `xml:"startDate,attr"     json:"start_date"`
	EndDate       string `xml:"endDate,attr"       json:"end_date"`
	Value         string `xml:"value,attr"         json:"value"`
	Device        string `xml:"device,attr"        json:"device"`
}

func main() {
	if len(os.Args) < 2 {
		log.Fatal("need zip file as first argument")
	}

	zp, err := zip.OpenReader(os.Args[1])
	if err != nil {
		log.Fatalf("Open zip file error: %s", err)
	}
	defer zp.Close()

	fl, err := zp.Open("apple_health_export/export.xml")
	if err != nil {
		log.Fatalf("Open file from zip error: %s", err)
	}
	defer fl.Close()

	var data healthData
	if err := xml.NewDecoder(fl).Decode(&data); err != nil {
		log.Fatalf("XML decode error: %s", err)
	}

	encoder := json.NewEncoder(os.Stdout)
	encoder.SetIndent("", " ")
	if err := encoder.Encode(data.HealthData); err != nil {
		log.Fatalf("JSON marshal error: %s", err)
	}
}
