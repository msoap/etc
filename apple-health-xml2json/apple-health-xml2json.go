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
	"fmt"
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

	data, err := loadXMLFile(os.Args[1])
	if err != nil {
		log.Fatalf("Load XML file: %s", err)
	}

	encoder := json.NewEncoder(os.Stdout)
	encoder.SetIndent("", " ")
	if err := encoder.Encode(data.HealthData); err != nil {
		log.Fatalf("JSON marshal: %s", err)
	}
}

func loadXMLFile(zipFile string) (*healthData, error) {
	zp, err := zip.OpenReader(zipFile)
	if err != nil {
		return nil, fmt.Errorf("Open zip file: %s", err)
	}
	defer zp.Close()

	fl, err := zp.Open("apple_health_export/export.xml")
	if err != nil {
		return nil, fmt.Errorf("Open file from zip: %s", err)
	}
	defer fl.Close()

	var data healthData
	if err := xml.NewDecoder(fl).Decode(&data); err != nil {
		return nil, fmt.Errorf("XML decode error: %s", err)
	}

	return &data, nil
}
