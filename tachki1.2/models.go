package main

import (
	"database/sql"

	"fyne.io/fyne/v2"
)

// DatabaseApp holds the application state
type DatabaseApp struct {
	db     *sql.DB
	app    fyne.App
	window fyne.Window
}

// DriverCategory represents a row in driver_categories
type DriverCategory struct {
	ID   int
	Code string
	Name string
}

// Owner represents a row in owners
type Owner struct {
	ID         int
	FirstName  string
	LastName   string
	Phone      string
	Email      string
	Category   string
	RowVersion []byte // For optimistic locking
}

// CarBrand represents a row in car_brands
type CarBrand struct {
	ID        int
	Name      string
	ImageData []byte
}

// Car represents a row in cars
type Car struct {
	ID           int
	OwnerID      int
	BrandID      int
	Model        string
	Year         int
	Color        string
	VIN          string
	Price        float64
	CurrentPrice float64
	RowVersion   []byte
}
