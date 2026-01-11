package main

import (
	"fmt"

	"strconv"

	"fyne.io/fyne/v2"
	"fyne.io/fyne/v2/canvas"
	"fyne.io/fyne/v2/container"
	"fyne.io/fyne/v2/theme"
	"fyne.io/fyne/v2/widget"
)

func (d *DatabaseApp) createViewTab() *container.Scroll {
	// Title
	titleLabel := widget.NewLabelWithStyle("Просмотр таблиц базы данных", fyne.TextAlignCenter, fyne.TextStyle{Bold: true})

	// Table Selector
	tableSelect := widget.NewSelect([]string{
		"Категории прав",
		"Владельцы",
		"Марки автомобилей",
		"Автомобили",
	}, nil)
	tableSelect.SetSelected("Владельцы")
	tableSelect.PlaceHolder = "Выберите таблицу"

	tableMap := map[string]string{
		"Категории прав":    "driver_categories",
		"Владельцы":         "owners",
		"Марки автомобилей": "car_brands",
		"Автомобили":        "cars",
	}

	// Data Table Area
	dataTable := widget.NewTable(
		func() (int, int) { return 0, 0 },
		func() fyne.CanvasObject {
			label := widget.NewLabel("")
			label.Wrapping = fyne.TextWrapWord
			return label
		},
		func(i widget.TableCellID, o fyne.CanvasObject) {},
	)
	dataTable.SetColumnWidth(0, 80)
	dataTable.SetColumnWidth(1, 150)

	// Refresh Button
	refreshBtn := widget.NewButtonWithIcon("Обновить данные", theme.ViewRefreshIcon(), func() {
		if tableName, ok := tableMap[tableSelect.Selected]; ok {
			d.loadTableData(tableName, dataTable)
		}
	})
	refreshBtn.Importance = widget.MediumImportance

	// Initial Load
	if tableName, ok := tableMap[tableSelect.Selected]; ok {
		d.loadTableData(tableName, dataTable)
	}

	// Change Handler
	tableSelect.OnChanged = func(table string) {
		if tableName, ok := tableMap[table]; ok {
			d.loadTableData(tableName, dataTable)
		}
	}

	// Control Panel
	controlPanel := container.NewVBox(
		titleLabel,
		widget.NewSeparator(),
		widget.NewLabel("Выберите таблицу для просмотра:"),
		tableSelect,
		container.NewGridWithColumns(1, refreshBtn),
		widget.NewSeparator(),
	)

	paddedControlPanel := container.NewPadded(controlPanel)
	split := container.NewVSplit(paddedControlPanel, container.NewPadded(dataTable))
	split.Offset = 0.15
	paddedContent := container.NewPadded(split)

	return container.NewScroll(paddedContent)
}

func (d *DatabaseApp) refreshViewTab(split *container.Split) {
	if table, ok := split.Trailing.(*widget.Table); ok {
		selectWidget := split.Leading.(*fyne.Container).Objects[1].(*widget.Select)

		tableMap := map[string]string{
			"Категории прав":    "driver_categories",
			"Владельцы":         "owners",
			"Марки автомобилей": "car_brands",
			"Автомобили":        "cars",
		}

		if tableName, ok := tableMap[selectWidget.Selected]; ok {
			d.loadTableData(tableName, table)
		}
	}
}

func (d *DatabaseApp) loadTableData(tableName string, table *widget.Table) {
	var query string
	var columnNames []string

	switch tableName {
	case "driver_categories":
		query = "SELECT category_id, category_code, category_name, description FROM driver_categories"
		columnNames = []string{"ID", "Код категории", "Название", "Описание"}
	case "owners":
		// Мы используем созданное представление v_owner_details
		// Добавляем experience_years в выборку
		query = `SELECT owner_id, first_name, last_name, phone, email, 
                        category_code, registration_date, car_count, experience_years
                 FROM v_owner_details`
		columnNames = []string{"ID", "Имя", "Фамилия", "Телефон", "Email",
			"Категория", "Дата получения прав", "Кол-во авто", "Стаж (лет)"}
	case "car_brands":
		query = "SELECT brand_id, brand_name, country_origin, founded_year, image_data FROM car_brands"
		columnNames = []string{"ID", "Марка", "Страна", "Год основания", "Логотип"}
	case "cars":
		query = `SELECT c.car_id, 
                        o.first_name + ' ' + o.last_name, 
                        b.brand_name, 
                        c.model, 
                        c.year, 
                        c.color, 
                        c.vin_code, 
                        c.price, -- Цена покупки
                        dbo.fn_GetCarDepreciatedValue(c.price, c.year), -- Текущая цена (функция)
                        c.purchase_date
                 FROM cars c 
                 JOIN owners o ON c.owner_id = o.owner_id 
                 JOIN car_brands b ON c.brand_id = b.brand_id`

		// Добавляем заголовок "Тек. цена"
		columnNames = []string{"ID", "Владелец", "Марка", "Модель", "Год",
			"Цвет", "VIN", "Цена покупки", "Тек. цена (~)", "Дата покупки"}
	default:
		return
	}

	rows, err := d.db.Query(query)
	if err != nil {
		d.showMessage("Ошибка", fmt.Sprintf("Ошибка загрузки данных: %v", err))
		return
	}
	defer rows.Close()

	columnCount := len(columnNames)
	var data [][]interface{}

	for rows.Next() {
		values := make([]interface{}, columnCount)
		valuePtrs := make([]interface{}, columnCount)
		for i := range values {
			valuePtrs[i] = &values[i]
		}

		err := rows.Scan(valuePtrs...)
		if err != nil {
			d.showMessage("Ошибка", fmt.Sprintf("Ошибка чтения данных: %v", err))
			return
		}
		data = append(data, values)
	}

	imageCache := make(map[string]*canvas.Image)

	// Set column widths
	for col := 0; col < columnCount; col++ {
		width := 150
		if tableName == "car_brands" && col == 4 {
			width = 120
		}
		table.SetColumnWidth(col, float32(width))
	}

	table.Length = func() (int, int) {
		return len(data) + 1, columnCount
	}

	table.CreateCell = func() fyne.CanvasObject {
		label := widget.NewLabel("Init")
		label.Wrapping = fyne.TextWrapOff
		label.Truncation = fyne.TextTruncateEllipsis
		return container.NewMax(label)
	}

	table.UpdateCell = func(id widget.TableCellID, cell fyne.CanvasObject) {
		containerObj := cell.(*fyne.Container)
		var label *widget.Label

		if len(containerObj.Objects) > 0 {
			if l, ok := containerObj.Objects[0].(*widget.Label); ok {
				label = l
			}
		}

		if label == nil {
			label = widget.NewLabel("")
			label.Wrapping = fyne.TextWrapOff
			label.Truncation = fyne.TextTruncateEllipsis
		}

		containerObj.Objects = []fyne.CanvasObject{label}

		label.TextStyle = fyne.TextStyle{}
		label.Alignment = fyne.TextAlignLeading

		if id.Row == 0 {
			// Headers
			label.TextStyle = fyne.TextStyle{Bold: true}
			label.Alignment = fyne.TextAlignCenter
			if id.Col < len(columnNames) {
				label.SetText(columnNames[id.Col])
			}
		} else {
			// Data
			rowIndex := id.Row - 1
			if rowIndex < len(data) && id.Col < len(data[rowIndex]) {
				value := data[rowIndex][id.Col]

				if tableName == "car_brands" && id.Col == 4 {
					if imageData, ok := value.([]byte); ok && len(imageData) > 0 {
						cacheKey := fmt.Sprintf("%d_%d", id.Row, id.Col)
						if cachedImg, exists := imageCache[cacheKey]; exists {
							containerObj.Objects = []fyne.CanvasObject{cachedImg}
						} else {
							img, err := createThumbnailFromBytes(imageData)
							if err == nil && img != nil {
								imageCache[cacheKey] = img
								containerObj.Objects = []fyne.CanvasObject{img}
							} else {
								label.SetText("Error")
							}
						}
					} else {
						label.SetText("-")
						label.Alignment = fyne.TextAlignCenter
					}
				} else {
					// Обычный текст
					if value != nil {
						// ПРОВЕРЯЕМ ТИП ДАННЫХ
						switch v := value.(type) {
						case []byte:
							// 1. Превращаем байты в строку (получаем "50000.0000")
							rawString := string(v)

							// 2. Пробуем превратить строку в число (float64)
							if floatVal, err := strconv.ParseFloat(rawString, 64); err == nil {
								// 3. Если успешно — форматируем без дробной части (%.0f)
								// %.0f означает "0 знаков после запятой"
								// %.2f означало бы "2 знака" (например, 50000.00)
								label.SetText(fmt.Sprintf("%.0f", floatVal))
							} else {
								// Если не получилось (вдруг там не число), выводим как есть
								label.SetText(rawString)
							}

						case float64:
							// Для обычных float тоже уберем хвосты, если нужно
							label.SetText(fmt.Sprintf("%.0f", v))

						case int64, int, int32:
							label.SetText(fmt.Sprintf("%d", v))

						default:
							label.SetText(fmt.Sprintf("%v", v))
						}
					}
				}
			}
		}
		containerObj.Refresh()
	}

	// Refined column widths
	for col := 0; col < columnCount; col++ {
		width := 150.0
		if tableName == "driver_categories" && col == 3 {
			width = 400.0
		}
		if tableName == "driver_categories" && col == 0 {
			width = 50.0
		}
		if tableName == "car_brands" && col == 4 {
			width = 100.0
		}
		table.SetColumnWidth(col, float32(width))
	}
	table.Refresh()
}
