package main

import (
	"fmt"
	"strconv"
	"time"

	"fyne.io/fyne/v2"
	"fyne.io/fyne/v2/container"
	"fyne.io/fyne/v2/layout"
	"fyne.io/fyne/v2/theme"
	"fyne.io/fyne/v2/widget"
)

func (d *DatabaseApp) createEditTab() *container.Scroll {
	titleLabel := widget.NewLabelWithStyle("Редактирование записей", fyne.TextAlignCenter, fyne.TextStyle{Bold: true})

	tableSelect := widget.NewSelect([]string{"Владельцы", "Автомобили"}, nil)
	tableSelect.SetSelected("Владельцы")
	tableSelect.PlaceHolder = "Выберите таблицу"

	idEntry := widget.NewEntry()
	idEntry.SetPlaceHolder("Введите ID записи")
	idEntry.Validator = func(s string) error {
		if s == "" {
			return nil
		}
		_, err := strconv.Atoi(s)
		if err != nil {
			return fmt.Errorf("ID должен быть числом")
		}
		return nil
	}

	searchBtn := widget.NewButtonWithIcon("Найти запись", theme.SearchIcon(), nil)
	searchBtn.Importance = widget.MediumImportance

	resultContainer := container.NewVBox()
	editContainer := container.NewVBox()

	contentWrapper := container.NewVBox(
		titleLabel,
		widget.NewSeparator(),
		widget.NewLabel("Таблица:"),
		tableSelect,
		widget.NewLabel("ID записи:"),
		idEntry,
		searchBtn,
		widget.NewSeparator(),
		widget.NewLabelWithStyle("Результаты поиска:", fyne.TextAlignLeading, fyne.TextStyle{Bold: true}),
		resultContainer,
		widget.NewSeparator(),
		widget.NewLabelWithStyle("Редактирование:", fyne.TextAlignLeading, fyne.TextStyle{Bold: true}),
		editContainer,
	)

	searchBtn.OnTapped = func() {
		d.searchRecordHandlerWithContainers(tableSelect, idEntry, resultContainer, editContainer)
	}

	return container.NewScroll(container.NewPadded(contentWrapper))
}

func (d *DatabaseApp) searchRecordHandlerWithContainers(tableSelect *widget.Select, idEntry *widget.Entry, resultContainer, editContainer *fyne.Container) {
	if idEntry.Text == "" {
		d.showMessage("Ошибка", "Введите ID записи")
		return
	}

	id, err := strconv.Atoi(idEntry.Text)
	if err != nil {
		d.showMessage("Ошибка", "ID должен быть числом")
		return
	}

	tableMap := map[string]string{
		"Владельцы":  "owners",
		"Автомобили": "cars",
	}

	table := tableSelect.Selected
	tableName, ok := tableMap[table]
	if !ok {
		d.showMessage("Ошибка", "Неизвестная таблица")
		return
	}

	resultContainer.Objects = nil
	editContainer.Objects = nil

	if tableName == "owners" {
		d.handleOwnerEdit(id, resultContainer, editContainer)
	} else if tableName == "cars" {
		d.handleCarEdit(id, resultContainer, editContainer)
	}

	resultContainer.Refresh()
	editContainer.Refresh()
}

func (d *DatabaseApp) handleOwnerEdit(id int, resultContainer, editContainer *fyne.Container) {
	owner, err := d.searchOwnerByID(id)
	if err != nil {
		d.showMessage("Ошибка", fmt.Sprintf("Не удалось найти владельца: %v", err))
		return
	}

	// Show Search Results
	resultContainer.Add(widget.NewLabel(fmt.Sprintf("Найден владелец: %s %s", owner.FirstName, owner.LastName)))
	resultContainer.Add(widget.NewLabel(fmt.Sprintf("Телефон: %s", owner.Phone)))
	resultContainer.Add(widget.NewLabel(fmt.Sprintf("Email: %s", owner.Email)))
	resultContainer.Add(widget.NewLabel(fmt.Sprintf("Категория прав: %s", owner.Category)))
	resultContainer.Add(widget.NewLabel(fmt.Sprintf("Версия записи: загружена %s", time.Now().Format("15:04:05"))))

	// Edit Fields
	firstNameEdit := widget.NewEntry()
	firstNameEdit.SetText(owner.FirstName)
	lastNameEdit := widget.NewEntry()
	lastNameEdit.SetText(owner.LastName)
	phoneEdit := widget.NewEntry()
	phoneEdit.SetText(owner.Phone)
	emailEdit := widget.NewEntry()
	emailEdit.SetText(owner.Email)

	categories, err := d.getDriverCategories()
	if err != nil {
		d.showMessage("Ошибка", fmt.Sprintf("Ошибка получения категорий: %v", err))
		return
	}

	categoryOptions := make([]string, 0, len(categories))
	for _, cat := range categories {
		option := fmt.Sprintf("%s - %s", cat.Code, cat.Name)
		categoryOptions = append(categoryOptions, option)
	}

	categorySelect := widget.NewSelect(categoryOptions, nil)
	currentCategoryFound := false
	for _, cat := range categories {
		if cat.Code == owner.Category {
			categorySelect.SetSelected(fmt.Sprintf("%s - %s", cat.Code, cat.Name))
			currentCategoryFound = true
			break
		}
	}
	if !currentCategoryFound && len(categoryOptions) > 0 {
		categorySelect.SetSelected(categoryOptions[0])
	}

	versionLabel := widget.NewLabel(fmt.Sprintf("Версия: %x", owner.RowVersion))
	versionLabel.Hidden = true

	updateBtn := widget.NewButton("Обновить владельца", func() {
		var categoryID int
		for _, cat := range categories {
			option := fmt.Sprintf("%s - %s", cat.Code, cat.Name)
			if option == categorySelect.Selected {
				categoryID = cat.ID
				break
			}
		}

		err := d.updateOwner(id, firstNameEdit.Text, lastNameEdit.Text, phoneEdit.Text, emailEdit.Text, categoryID, owner.RowVersion)
		if err != nil {
			if err.Error() == "запись была изменена другим пользователем" {
				d.showMessage("Конфликт редактирования", "Запись была изменена другим пользователем. Пожалуйста, обновите данные и попробуйте снова.")
			} else {
				d.showMessage("Ошибка", fmt.Sprintf("Не удалось обновить владельца: %v", err))
			}
		} else {
			d.showMessage("Успех", "Владелец успешно обновлен")
			updatedOwner, err := d.searchOwnerByID(id)
			if err == nil {
				owner.RowVersion = updatedOwner.RowVersion
				versionLabel.SetText(fmt.Sprintf("Версия: %x", owner.RowVersion))
				resultContainer.Objects[4].(*widget.Label).SetText(fmt.Sprintf("Версия записи: обновлена %s", time.Now().Format("15:04:05")))
				resultContainer.Refresh()
			}
		}
	})

	refreshBtn := widget.NewButton("Обновить данные", func() {
		updatedOwner, err := d.searchOwnerByID(id)
		if err != nil {
			d.showMessage("Ошибка", fmt.Sprintf("Не удалось обновить данные: %v", err))
			return
		}
		firstNameEdit.SetText(updatedOwner.FirstName)
		lastNameEdit.SetText(updatedOwner.LastName)
		phoneEdit.SetText(updatedOwner.Phone)
		emailEdit.SetText(updatedOwner.Email)
		for _, cat := range categories {
			if cat.Code == updatedOwner.Category {
				categorySelect.SetSelected(fmt.Sprintf("%s - %s", cat.Code, cat.Name))
				break
			}
		}
		owner.RowVersion = updatedOwner.RowVersion
		versionLabel.SetText(fmt.Sprintf("Версия: %x", owner.RowVersion))

		resultContainer.Objects[0].(*widget.Label).SetText(fmt.Sprintf("Найден владелец: %s %s", updatedOwner.FirstName, updatedOwner.LastName))
		resultContainer.Objects[1].(*widget.Label).SetText(fmt.Sprintf("Телефон: %s", updatedOwner.Phone))
		resultContainer.Objects[2].(*widget.Label).SetText(fmt.Sprintf("Email: %s", updatedOwner.Email))
		resultContainer.Objects[3].(*widget.Label).SetText(fmt.Sprintf("Категория прав: %s", updatedOwner.Category))
		resultContainer.Objects[4].(*widget.Label).SetText(fmt.Sprintf("Версия записи: обновлена %s", time.Now().Format("15:04:05")))
		resultContainer.Refresh()
		d.showMessage("Успех", "Данные успешно обновлены")
	})

	editContainer.Add(widget.NewLabel("Редактирование владельца:"))
	editContainer.Add(widget.NewLabel("Имя:"))
	editContainer.Add(firstNameEdit)
	editContainer.Add(widget.NewLabel("Фамилия:"))
	editContainer.Add(lastNameEdit)
	editContainer.Add(widget.NewLabel("Телефон:"))
	editContainer.Add(phoneEdit)
	editContainer.Add(widget.NewLabel("Email:"))
	editContainer.Add(emailEdit)
	editContainer.Add(widget.NewLabel("Категория прав:"))
	editContainer.Add(categorySelect)
	editContainer.Add(versionLabel)
	editContainer.Add(container.NewHBox(updateBtn, refreshBtn))
}

func (d *DatabaseApp) handleCarEdit(id int, resultContainer, editContainer *fyne.Container) {
	// 1. Получаем данные автомобиля
	car, err := d.searchCarByID(id)
	if err != nil {
		d.showMessage("Ошибка", fmt.Sprintf("Не удалось найти автомобиль: %v", err))
		return
	}

	// 2. Очищаем контейнер результатов (по просьбе не выводим характеристики отдельно)
	resultContainer.Objects = nil
	// Можно оставить только заголовок, чтобы подтвердить, что поиск прошел успешно
	resultContainer.Add(widget.NewLabelWithStyle(fmt.Sprintf("Редактирование записи #%d", car.ID), fyne.TextAlignCenter, fyne.TextStyle{Bold: true}))
	resultContainer.Refresh()

	// 3. Получаем списки для выпадающих меню
	owners, err := d.getOwners()
	if err != nil {
		d.showMessage("Ошибка", fmt.Sprintf("Ошибка получения владельцев: %v", err))
		return
	}

	brands, err := d.getCarBrands()
	if err != nil {
		d.showMessage("Ошибка", fmt.Sprintf("Ошибка получения брендов: %v", err))
		return
	}

	// 4. Подготавливаем виджеты редактирования
	ownerOptions := make([]string, 0, len(owners))
	for _, owner := range owners {
		option := fmt.Sprintf("%d: %s %s", owner.ID, owner.FirstName, owner.LastName)
		ownerOptions = append(ownerOptions, option)
	}

	brandOptions := make([]string, 0, len(brands))
	for _, brand := range brands {
		brandOptions = append(brandOptions, brand.Name)
	}

	ownerSelect := widget.NewSelect(ownerOptions, nil)
	brandSelect := widget.NewSelect(brandOptions, nil)

	// Установка текущего владельца
	currentOwnerFound := false
	for _, owner := range owners {
		if owner.ID == car.OwnerID {
			ownerSelect.SetSelected(fmt.Sprintf("%d: %s %s", owner.ID, owner.FirstName, owner.LastName))
			currentOwnerFound = true
			break
		}
	}
	if !currentOwnerFound && len(ownerOptions) > 0 {
		ownerSelect.SetSelected(ownerOptions[0])
	}

	// Установка текущего бренда
	currentBrandFound := false
	for _, brand := range brands {
		if brand.ID == car.BrandID {
			brandSelect.SetSelected(brand.Name)
			currentBrandFound = true
			break
		}
	}
	if !currentBrandFound && len(brandOptions) > 0 {
		brandSelect.SetSelected(brandOptions[0])
	}

	// Текстовые поля
	modelEdit := widget.NewEntry()
	modelEdit.SetText(car.Model)

	yearEdit := widget.NewEntry()
	yearEdit.SetText(strconv.Itoa(car.Year))

	colorEdit := widget.NewEntry()
	colorEdit.SetText(car.Color)

	vinEdit := widget.NewEntry()
	vinEdit.SetText(car.VIN)

	// Поле ввода цены покупки
	priceEdit := widget.NewEntry()
	priceEdit.SetText(fmt.Sprintf("%.0f", car.Price))

	// Метка для отображения Текущей цены (расчетной)
	// Она будет обновляться при изменении данных
	currentPriceLabel := widget.NewLabel(fmt.Sprintf("Текущая рыночная цена (~): %.0f", car.CurrentPrice))
	currentPriceLabel.TextStyle = fyne.TextStyle{Italic: true} // Курсив, чтобы выделить, что это справочная инфо

	// Скрытое поле версии
	versionLabel := widget.NewLabel(fmt.Sprintf("Версия: %x", car.RowVersion))
	versionLabel.Hidden = true

	// 5. Логика кнопки "Обновить автомобиль"
	updateBtn := widget.NewButton("Сохранить изменения", func() {
		// Ищем ID владельца
		var ownerID int
		for _, owner := range owners {
			option := fmt.Sprintf("%d: %s %s", owner.ID, owner.FirstName, owner.LastName)
			if option == ownerSelect.Selected {
				ownerID = owner.ID
				break
			}
		}

		// Ищем ID бренда
		var brandID int
		for _, brand := range brands {
			if brand.Name == brandSelect.Selected {
				brandID = brand.ID
				break
			}
		}

		year, _ := strconv.Atoi(yearEdit.Text)
		price, _ := strconv.ParseFloat(priceEdit.Text, 64)

		// Отправляем запрос в БД
		err := d.updateCar(id, ownerID, brandID, modelEdit.Text, year, colorEdit.Text, vinEdit.Text, price, car.RowVersion)
		if err != nil {
			if err.Error() == "запись была изменена другим пользователем" {
				d.showMessage("Конфликт редактирования", "Запись была изменена другим пользователем. Пожалуйста, обновите данные.")
			} else {
				d.showMessage("Ошибка", fmt.Sprintf("Не удалось обновить автомобиль: %v", err))
			}
		} else {
			d.showMessage("Успех", "Автомобиль успешно обновлен")

			// Получаем обновленные данные (включая пересчитанную CurrentPrice)
			updatedCar, err := d.searchCarByID(id)
			if err == nil {
				car = updatedCar // Обновляем локальную переменную
				versionLabel.SetText(fmt.Sprintf("Версия: %x", car.RowVersion))

				// Обновляем метку с текущей ценой
				currentPriceLabel.SetText(fmt.Sprintf("Текущая рыночная цена (~): %.0f", car.CurrentPrice))
				currentPriceLabel.Refresh()
			}
		}
	})

	// 6. Логика кнопки "Сбросить / Обновить"
	refreshBtn := widget.NewButton("Сбросить", func() {
		updatedCar, err := d.searchCarByID(id)
		if err != nil {
			d.showMessage("Ошибка", fmt.Sprintf("Не удалось обновить данные: %v", err))
			return
		}

		// Обновляем поля ввода
		modelEdit.SetText(updatedCar.Model)
		yearEdit.SetText(strconv.Itoa(updatedCar.Year))
		colorEdit.SetText(updatedCar.Color)
		vinEdit.SetText(updatedCar.VIN)
		priceEdit.SetText(fmt.Sprintf("%.0f", updatedCar.Price))

		// Обновляем селекторы
		for _, owner := range owners {
			if owner.ID == updatedCar.OwnerID {
				ownerSelect.SetSelected(fmt.Sprintf("%d: %s %s", owner.ID, owner.FirstName, owner.LastName))
				break
			}
		}
		for _, brand := range brands {
			if brand.ID == updatedCar.BrandID {
				brandSelect.SetSelected(brand.Name)
				break
			}
		}

		car = updatedCar // Обновляем ссылку
		versionLabel.SetText(fmt.Sprintf("Версия: %x", car.RowVersion))

		// Обновляем метку с текущей ценой
		currentPriceLabel.SetText(fmt.Sprintf("Текущая рыночная цена (~): %.0f", updatedCar.CurrentPrice))
		d.showMessage("Успех", "Данные формы сброшены к значениям из БД")
	})

	// 7. Сборка формы
	// Используем FormLayout или просто добавляем метки и поля последовательно
	editContainer.Add(widget.NewLabel("Владелец:"))
	editContainer.Add(ownerSelect)

	editContainer.Add(widget.NewLabel("Марка:"))
	editContainer.Add(brandSelect)

	editContainer.Add(widget.NewLabel("Модель:"))
	editContainer.Add(modelEdit)

	editContainer.Add(widget.NewLabel("Год выпуска:"))
	editContainer.Add(yearEdit)

	editContainer.Add(widget.NewLabel("Цвет:"))
	editContainer.Add(colorEdit)

	editContainer.Add(widget.NewLabel("ВИН код:"))
	editContainer.Add(vinEdit)

	editContainer.Add(widget.NewLabel("Цена при покупке:"))
	editContainer.Add(priceEdit)

	// ДОБАВЛЯЕМ ТЕКУЩУЮ ЦЕНУ СРАЗУ ПОД ЦЕНОЙ ПОКУПКИ
	editContainer.Add(currentPriceLabel)

	editContainer.Add(versionLabel)

	editContainer.Add(layout.NewSpacer())
	editContainer.Add(container.NewHBox(updateBtn, refreshBtn))
}
