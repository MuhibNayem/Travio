package main

import (
	"database/sql"
	"encoding/json"
	"fmt"
	"log"
	"os"
	"strconv"
	"strings"
	"time"

	"github.com/google/uuid"
	_ "github.com/jackc/pgx/v5/stdlib"
)

type Metadata struct {
	TotalDistricts int    `json:"total_districts"`
	TotalUpazilas  int    `json:"total_upazilas"`
	DataSource     string `json:"data_source"`
	LastUpdated    string `json:"last_updated"`
}

type District struct {
	ID         string `json:"id"`
	DivisionID string `json:"division_id"`
	Name       string `json:"name"`
	BnName     string `json:"bn_name"`
	Lat        string `json:"lat"`
	Lon        string `json:"lon"`
	URL        string `json:"url"`
}

type Upazila struct {
	ID         string `json:"id"`
	DistrictID string `json:"district_id"`
	Name       string `json:"name"`
	BnName     string `json:"bn_name"`
	URL        string `json:"url"`
}

type GeoData struct {
	Metadata              Metadata               `json:"metadata"`
	DistrictsWithUpazilas []DistrictWithUpazilas `json:"districts_with_upazilas"`
}

type DistrictWithUpazilas struct {
	District District  `json:"district"`
	Upazilas []Upazila `json:"upazilas"`
}

func main() {
	log.Println("Starting station seeding...")

	// 1. Connect to Database
	dbUser := getEnv("DB_USER", "postgres")
	dbPass := getEnv("DB_PASSWORD", "postgres")
	dbHost := getEnv("DB_HOST", "localhost")
	dbPort := getEnv("DB_PORT", "5432")
	dbName := getEnv("DB_NAME", "travio_catalog")
	sslMode := getEnv("DB_SSLMODE", "disable")

	dsn := fmt.Sprintf("postgres://%s:%s@%s:%s/%s?sslmode=%s", dbUser, dbPass, dbHost, dbPort, dbName, sslMode)
	db, err := sql.Open("pgx", dsn)
	if err != nil {
		log.Fatalf("Failed to open DB connection: %v", err)
	}
	defer db.Close()

	if err := db.Ping(); err != nil {
		log.Fatalf("Failed to ping DB: %v", err)
	}
	log.Println("Connected to Database.")

	// 2. Read JSON Data
	paths := []string{
		"seed_data/bangladesh_administrative_divisions.json",       // From catalog root
		"../../seed_data/bangladesh_administrative_divisions.json", // From cmd/seed
		"./bangladesh_administrative_divisions.json",               // From current dir
	}

	var jsonData []byte
	var loadedPath string
	for _, p := range paths {
		b, err := os.ReadFile(p)
		if err == nil {
			jsonData = b
			loadedPath = p
			break
		}
	}

	if jsonData == nil {
		log.Fatalf("Could not find bangladesh_administrative_divisions.json in tried paths")
	}
	log.Printf("Loaded JSON from %s", loadedPath)

	var geoData GeoData
	if err := json.Unmarshal(jsonData, &geoData); err != nil {
		log.Fatalf("Failed to parse JSON: %v", err)
	}

	totalDistricts := 0
	totalUpazilas := 0

	for _, item := range geoData.DistrictsWithUpazilas {
		d := item.District
		totalDistricts++

		// Seed District
		if err := upsertStation(db, d.Name, d.BnName, d.Name, "District Headquarters", d.Lat, d.Lon); err != nil {
			log.Printf("Error seeding district %s: %v", d.Name, err)
		}

		// Seed Upazilas
		for _, u := range item.Upazilas {
			totalUpazilas++
			// Use parent district lat/lon as upazila often lacks it in this dataset, or we use 0.0
			// The new JSON structure for Upazila doesn't show lat/lon, so we rely on District's.
			lat := d.Lat
			lon := d.Lon

			// City for Upazila: Use District Name to avoid collisions between Upazilas with same name in different districts
			city := d.Name

			if err := upsertStation(db, u.Name, u.BnName, city, "Upazila", lat, lon); err != nil {
				log.Printf("Error seeding upazila %s: %v", u.Name, err)
			}
		}
	}

	log.Printf("Seeding completed. Processed %d Districts and %d Upazilas.", totalDistricts, totalUpazilas)
}

func getEnv(key, fallback string) string {
	if v := os.Getenv(key); v != "" {
		return v
	}
	return fallback
}

func upsertStation(db *sql.DB, name, bnName, city, category, latStr, lonStr string) error {
	// Idempotency Logic:
	// 1. Districts (category="District Headquarters") should have Code suffix "001".
	// 2. Upazilas should have Code containing "-".
	// We want to allow a District and Upazila to share the same Name (e.g. "Feni" District vs "Feni" Upazila).
	// But we don't want to insert the same District twice, or same Upazila twice.

	rows, err := db.Query("SELECT id, code FROM stations WHERE name = $1 AND city = $2 AND organization_id IS NULL", name, city)
	if err != nil {
		return fmt.Errorf("lookup error: %w", err)
	}
	defer rows.Close()

	for rows.Next() {
		var id, existingCode string
		if err := rows.Scan(&id, &existingCode); err != nil {
			continue
		}

		if category == "District Headquarters" {
			// If we find an existing record that looks like a District (ends in 001), skip.
			if strings.HasSuffix(existingCode, "-001") {
				return nil
			}
		} else { // Upazila
			// If it's a District, ignore it (we are looking for an existing Upazila).
			if strings.HasSuffix(existingCode, "-001") {
				continue
			}
			// If we find an existing record that looks like an Upazila (contains "-" but not ending in "001"), skip.
			if strings.Contains(existingCode, "-") {
				return nil
			}
		}
	}

	// New Insert
	id := uuid.New().String()
	code := generateCode(name, category)

	lat, _ := strconv.ParseFloat(latStr, 64)
	lon, _ := strconv.ParseFloat(lonStr, 64)

	amenities := []string{"Waiting Room", "Ticket Counter"}
	amenitiesJSON, _ := json.Marshal(amenities)

	now := time.Now()

	query := `INSERT INTO stations (id, organization_id, code, name, city, state, country, 
			  latitude, longitude, timezone, address, amenities, status, created_at, updated_at) 
			  VALUES ($1, $2, $3, $4, $5, $6, $7, $8, $9, $10, $11, $12, $13, $14, $15)`

	_, err = db.Exec(query,
		id, nil, code, name, city, "Bangladesh", "Bangladesh",
		lat, lon, "Asia/Dhaka", city+", Bangladesh",
		amenitiesJSON, "active", now, now,
	)

	// If duplicate code error, retry with different code?
	// For now let it error and we see log.
	return err
}

func generateCode(name, category string) string {
	slug := strings.ToUpper(strings.ReplaceAll(name, " ", ""))
	if len(slug) > 3 {
		slug = slug[:3]
	}
	if category == "District Headquarters" {
		return fmt.Sprintf("%s-001", slug)
	}
	if category == "Upazila" {
		// Use nano time to ensure uniqueness roughly
		return fmt.Sprintf("%s-%d", slug, time.Now().UnixNano()%10000)
	}
	return fmt.Sprintf("%s-%d", slug, time.Now().UnixNano()%100)
}
