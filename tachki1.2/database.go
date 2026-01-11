package main

import (
	"database/sql"
	"fmt"

	_ "github.com/microsoft/go-mssqldb"
)

func (d *DatabaseApp) connectDB(connStr string) error {
	var err error
	// Close existing connection if we are reconnecting
	if d.db != nil {
		d.db.Close()
	}

	d.db, err = sql.Open("sqlserver", connStr)
	if err != nil {
		return err
	}

	return d.db.Ping()
}

// --- Reads ---

func (d *DatabaseApp) getDriverCategories() ([]DriverCategory, error) {
	query := "SELECT category_id, category_code, category_name FROM driver_categories"
	rows, err := d.db.Query(query)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var categories []DriverCategory
	for rows.Next() {
		var cat DriverCategory
		err := rows.Scan(&cat.ID, &cat.Code, &cat.Name)
		if err != nil {
			return nil, err
		}
		categories = append(categories, cat)
	}
	return categories, nil
}

func (d *DatabaseApp) getOwners() ([]Owner, error) {
	query := "SELECT owner_id, first_name, last_name FROM owners"
	rows, err := d.db.Query(query)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var owners []Owner
	for rows.Next() {
		var owner Owner
		err := rows.Scan(&owner.ID, &owner.FirstName, &owner.LastName)
		if err != nil {
			return nil, err
		}
		owners = append(owners, owner)
	}
	return owners, nil
}

func (d *DatabaseApp) getCarBrands() ([]CarBrand, error) {
	query := "SELECT brand_id, brand_name, image_data FROM car_brands"
	rows, err := d.db.Query(query)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var brands []CarBrand
	for rows.Next() {
		var brand CarBrand
		var imageData []byte
		err := rows.Scan(&brand.ID, &brand.Name, &imageData)
		if err != nil {
			return nil, err
		}
		brand.ImageData = imageData
		brands = append(brands, brand)
	}
	return brands, nil
}

func (d *DatabaseApp) searchOwnerByID(id int) (*Owner, error) {
	query := `SELECT o.owner_id, o.first_name, o.last_name, o.phone, o.email, dc.category_code, 
			  o.row_version
			  FROM owners o 
			  JOIN driver_categories dc ON o.license_category_id = dc.category_id 
			  WHERE o.owner_id = @p1`

	row := d.db.QueryRow(query, id)

	var owner Owner
	err := row.Scan(&owner.ID, &owner.FirstName, &owner.LastName, &owner.Phone, &owner.Email, &owner.Category, &owner.RowVersion)
	if err != nil {
		return nil, err
	}
	return &owner, nil
}

func (d *DatabaseApp) searchCarByID(id int) (*Car, error) {
	// Добавили вызов dbo.fn_GetCarDepreciatedValue(price, year) в запрос
	query := `SELECT car_id, owner_id, brand_id, model, year, color, vin_code, price, 
			  dbo.fn_GetCarDepreciatedValue(price, year), 
			  row_version 
			  FROM cars WHERE car_id = @p1`

	row := d.db.QueryRow(query, id)

	var car Car
	// Добавили &car.CurrentPrice в scan
	err := row.Scan(&car.ID, &car.OwnerID, &car.BrandID, &car.Model, &car.Year, &car.Color,
		&car.VIN, &car.Price, &car.CurrentPrice, &car.RowVersion)

	if err != nil {
		return nil, err
	}
	return &car, nil
}

// --- Writes (Create/Update/Delete) ---

func (d *DatabaseApp) addOwner(firstName, lastName, phone, email string, categoryID int) error {
	query := `INSERT INTO owners (first_name, last_name, phone, email, license_category_id) 
              VALUES (@p1, @p2, @p3, @p4, @p5)`
	_, err := d.db.Exec(query, firstName, lastName, phone, email, categoryID)
	return err
}

func (d *DatabaseApp) addCar(ownerID, brandID int, model string, year int, color, vin string, price float64) error {
	query := `INSERT INTO cars (owner_id, brand_id, model, year, color, vin_code, price) 
              VALUES (@p1, @p2, @p3, @p4, @p5, @p6, @p7)`
	_, err := d.db.Exec(query, ownerID, brandID, model, year, color, vin, price)
	return err
}

func (d *DatabaseApp) updateOwner(id int, firstName, lastName, phone, email string, categoryID int, rowVersion []byte) error {
	query := `UPDATE owners 
			  SET first_name = @p1, last_name = @p2, phone = @p3, email = @p4, 
			      license_category_id = @p5, row_version = NEWID()
			  WHERE owner_id = @p6 AND row_version = @p7`

	result, err := d.db.Exec(query, firstName, lastName, phone, email, categoryID, id, rowVersion)
	if err != nil {
		return err
	}

	rowsAffected, err := result.RowsAffected()
	if err != nil {
		return err
	}

	if rowsAffected == 0 {
		return fmt.Errorf("запись была изменена другим пользователем")
	}
	return nil
}

func (d *DatabaseApp) updateCar(id int, ownerID, brandID int, model string, year int, color, vin string, price float64, rowVersion []byte) error {
	query := `UPDATE cars 
			  SET owner_id = @p1, brand_id = @p2, model = @p3, year = @p4, color = @p5, 
			      vin_code = @p6, price = @p7, row_version = NEWID()
			  WHERE car_id = @p8 AND row_version = @p9`

	result, err := d.db.Exec(query, ownerID, brandID, model, year, color, vin, price, id, rowVersion)
	if err != nil {
		return err
	}

	rowsAffected, err := result.RowsAffected()
	if err != nil {
		return err
	}

	if rowsAffected == 0 {
		return fmt.Errorf("запись была изменена другим пользователем")
	}
	return nil
}

func (d *DatabaseApp) deleteRecord(table string, id int) error {
	query := fmt.Sprintf("DELETE FROM %s WHERE %s_id = @p1", table, table[:len(table)-1])
	_, err := d.db.Exec(query, id)
	return err
}
func (d *DatabaseApp) massPriceUpdate(brandID int, percentage float64) error {
	query := "EXEC sp_MassPriceUpdate @BrandID = @p1, @Percentage = @p2"

	result, err := d.db.Exec(query, brandID, percentage)
	if err != nil {
		return err
	}
	// Можно даже проверить, сколько записей затронуто
	rows, _ := result.RowsAffected()
	if rows == 0 {
		return fmt.Errorf("автомобили данного бренда не найдены")
	}

	return nil
}
func (d *DatabaseApp) updateBrandImage(brandID int, imageData []byte) error {
	query := "UPDATE car_brands SET image_data = @p1 WHERE brand_id = @p2"
	_, err := d.db.Exec(query, imageData, brandID)
	return err
}
func (d *DatabaseApp) getBrandImage(brandID int) ([]byte, error) {
	query := "SELECT image_data FROM car_brands WHERE brand_id = @p1"
	row := d.db.QueryRow(query, brandID)

	var imageData []byte
	// Если в базе NULL, Scan запишет nil в imageData, ошибки не будет
	err := row.Scan(&imageData)
	if err != nil {
		return nil, err
	}

	return imageData, nil
}
