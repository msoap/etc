/*
Apple Health export data converter from XML to JSON or sqlite3 DB

Usage:

	go run apple-health-import.go export.zip -json export.json -db export.db

Install:

	go get github.com/msoap/etc/apple-health-import@latest
*/
package main

import (
	"archive/zip"
	"encoding/json"
	"encoding/xml"
	"flag"
	"fmt"
	"log"
	"os"
	"strconv"
	"strings"
	"time"

	"github.com/jmoiron/sqlx"
	_ "github.com/mattn/go-sqlite3"
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

type dbRecord struct {
	CreationDate time.Time `db:"creation_date"`
	StartDate    time.Time `db:"start_date"`
	EndDate      time.Time `db:"end_date"`
	Type         string    `db:"type"`
	Unit         string    `db:"unit"`
	Value        float64   `db:"value"`
	SourceName   string    `db:"source_name"`
}

func main() {
	jsonFile := flag.String("json", "", "output JSON file")
	dbFile := flag.String("db", "", "output sqlite3 DB file")
	flag.Parse()

	zipFile := ""
	if args := flag.Args(); len(args) != 1 {
		log.Fatal("need zip file as first argument")
	} else {
		zipFile = args[0]
	}

	log.Printf("Load XML file: %s, out JSON: '%s', out to DB: '%s'", zipFile, *jsonFile, *dbFile)

	data, err := loadXMLFile(zipFile)
	if err != nil {
		log.Fatalf("Load XML file: %s", err)
	}
	log.Printf("Loaded %d records from XML", len(data.HealthData))

	if *jsonFile != "" {
		saveJSON(*jsonFile, data)
	}

	if *dbFile != "" {
		saveDB(*dbFile, data)
	}
}

func loadXMLFile(zipFile string) (*healthData, error) {
	zp, err := zip.OpenReader(zipFile)
	if err != nil {
		return nil, fmt.Errorf("Open zip file %s: %s", zipFile, err)
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

func saveJSON(jsonFile string, data *healthData) {
	fileOut, err := os.Create(jsonFile)
	if err != nil {
		log.Fatalf("Create JSON file: %s", err)
	}
	encoder := json.NewEncoder(fileOut)
	encoder.SetIndent("", " ")
	if err := encoder.Encode(data.HealthData); err != nil {
		log.Fatalf("JSON marshal: %s", err)
	}
	if err := fileOut.Close(); err != nil {
		log.Fatalf("Close JSON file: %s", err)
	}
	log.Printf("Saved to JSON file: %s", jsonFile)
}

func saveDB(dbFile string, data *healthData) {
	var dbData []dbRecord

	parseDateFn := func(s string) time.Time {
		t, err := time.Parse("2006-01-02 15:04:05 -0700", s)
		if err != nil {
			log.Fatalf("Parse date: %s", err)
		}
		return t
	}

	for _, rec := range data.HealthData {
		value, err := strconv.ParseFloat(rec.Value, 64)
		if err != nil {
			// support for float values
			continue
		}

		recType := strings.TrimPrefix(
			strings.TrimPrefix(
				strings.TrimPrefix(
					rec.Type,
					"HKDataType"),
				"HKCategoryTypeIdentifier"),
			"HKQuantityTypeIdentifier")

		dbData = append(dbData, dbRecord{
			CreationDate: parseDateFn(rec.CreationDate),
			StartDate:    parseDateFn(rec.StartDate),
			EndDate:      parseDateFn(rec.EndDate),
			Type:         recType,
			Unit:         rec.Unit,
			Value:        value,
			SourceName:   rec.SourceName,
		})
	}

	cnt, err := insertDBData(dbFile, dbData)
	if err != nil {
		log.Fatalf("Save to DB: %s", err)
	}

	log.Printf("Saved to DB file: %s, %d records, saved: %d", dbFile, len(dbData), cnt)
}

func insertDBData(dbName string, data []dbRecord) (int, error) {
	db, err := sqlx.Open("sqlite3", dbName)
	if err != nil {
		return 0, fmt.Errorf("opening DB %s: %s", dbName, err)
	}

	defer func() {
		if err := db.Close(); err != nil {
			log.Fatalf("closing DB: %s", err)
		}
	}()

	// create table
	if _, err := db.Exec(`
	CREATE TABLE IF NOT EXISTS health (
		creation_date DATETIME,
		start_date    DATETIME,
		end_date      DATETIME,
		type          TEXT,
		unit          TEXT,
		value         DECIMAL(10,5),
		source_name   TEXT,

		UNIQUE (creation_date, type)
	)`); err != nil {
		return 0, fmt.Errorf("creating table: %s", err)
	}

	// insert data
	sqlQuery := `
		INSERT INTO health (
			creation_date,
			start_date,
			end_date,
			type,
			unit,
			value,
			source_name
		) VALUES (
			:creation_date,
			:start_date,
			:end_date,
			:type,
			:unit,
			:value,
			:source_name
		)
		ON CONFLICT(creation_date, type) DO NOTHING
	`
	stmt, err := db.PrepareNamed(sqlQuery)
	if err != nil {
		return 0, fmt.Errorf("preparing statement: %s", err)
	}

	cnt := 0
	for _, rec := range data {
		res, err := stmt.Exec(rec)
		if err != nil {
			return 0, fmt.Errorf("inserting record %#v: %s", rec, err)
		}

		n, err := res.RowsAffected()
		if err != nil {
			return 0, fmt.Errorf("getting rows affected: %s", err)
		}

		cnt += int(n)
	}

	return cnt, nil
}
