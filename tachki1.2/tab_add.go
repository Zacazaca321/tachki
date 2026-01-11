package main

import (
	"fmt"
	"log"
	"strconv"
	"time"

	"fyne.io/fyne/v2"
	"fyne.io/fyne/v2/container"
	"fyne.io/fyne/v2/layout"
	"fyne.io/fyne/v2/theme"
	"fyne.io/fyne/v2/widget"
)

// --- ADD OWNER TAB ---

func (d *DatabaseApp) createAddOwnerTab() *container.Scroll {
	titleLabel := widget.NewLabelWithStyle("Добавление нового владельца", fyne.TextAlignCenter, fyne.TextStyle{Bold: true})

	firstNameEntry := widget.NewEntry()
	firstNameEntry.SetPlaceHolder("Введите имя")
	firstNameEntry.Validator = func(s string) error {
		if len(s) < 2 {
			return fmt.Errorf("имя должно быть не менее 2 символов")
		}
		return nil
	}

	lastNameEntry := widget.NewEntry()
	lastNameEntry.SetPlaceHolder("Введите фамилию")
	lastNameEntry.Validator = func(s string) error {
		if len(s) < 2 {
			return fmt.Errorf("фамилия должна быть не менее 2 символов")
		}
		return nil
	}

	phoneEntry := widget.NewEntry()
	phoneEntry.SetPlaceHolder("Введите телефон")
	phoneEntry.Validator = func(s string) error {
		if s == "" {
			return nil
		}
		if len(s) < 5 {
			return fmt.Errorf("телефон должен быть не менее 5 символов")
		}
		return nil
	}

	emailEntry := widget.NewEntry()
	emailEntry.SetPlaceHolder("Введите email")
	emailEntry.Validator = func(s string) error {
		if s == "" {
			return nil
		}
		if !contains(s, "@") {
			return fmt.Errorf("email должен содержать @")
		}
		return nil
	}

	categorySelect := widget.NewSelect([]string{}, nil)
	categorySelect.PlaceHolder = "Выберите категорию прав"

	addBtn := widget.NewButtonWithIcon("Добавить владельца", theme.ContentAddIcon(), func() {
		d.addOwnerHandler(firstNameEntry, lastNameEntry, phoneEntry, emailEntry, categorySelect)
	})
	addBtn.Importance = widget.HighImportance

	form := &widget.Form{
		Items: []*widget.FormItem{
			widget.NewFormItem("Имя:", firstNameEntry),
			widget.NewFormItem("Фамилия:", lastNameEntry),
			widget.NewFormItem("Телефон:", phoneEntry),
			widget.NewFormItem("Email:", emailEntry),
			widget.NewFormItem("Категория прав:", categorySelect),
		},
		OnSubmit: func() {
			d.addOwnerHandler(firstNameEntry, lastNameEntry, phoneEntry, emailEntry, categorySelect)
		},
		SubmitText: "Добавить владельца",
	}

	content := container.NewVBox(
		titleLabel,
		widget.NewSeparator(),
		form,
		layout.NewSpacer(),
		addBtn,
	)

	updateCategories := func() {
		categories, err := d.getDriverCategories()
		if err != nil {
			log.Printf("Ошибка получения категорий: %v", err)
			return
		}

		categoryOptions := make([]string, 0, len(categories))
		for _, cat := range categories {
			option := fmt.Sprintf("%s - %s", cat.Code, cat.Name)
			categoryOptions = append(categoryOptions, option)
		}

		categorySelect.Options = categoryOptions
		if len(categoryOptions) > 0 {
			categorySelect.SetSelected(categoryOptions[0])
		} else {
			categorySelect.SetSelected("")
		}
		categorySelect.Refresh()
	}
	updateCategories()

	refreshBtn := widget.NewButtonWithIcon("Обновить категории", theme.ViewRefreshIcon(), func() {
		updateCategories()
		d.showMessage("Обновлено", "Список категорий обновлен")
	})
	refreshBtn.Importance = widget.MediumImportance
	content.Add(refreshBtn)

	return container.NewScroll(container.NewPadded(content))
}

func (d *DatabaseApp) addOwnerHandler(firstNameEntry, lastNameEntry, phoneEntry, emailEntry *widget.Entry, categorySelect *widget.Select) {
	if firstNameEntry.Text == "" || lastNameEntry.Text == "" {
		d.showMessage("Ошибка", "Имя и фамилия обязательны для заполнения")
		return
	}

	if categorySelect.Selected == "" {
		d.showMessage("Ошибка", "Выберите категорию прав")
		return
	}

	categories, err := d.getDriverCategories()
	if err != nil {
		d.showMessage("Ошибка", fmt.Sprintf("Ошибка получения категорий: %v", err))
		return
	}

	var categoryID int
	for _, cat := range categories {
		option := fmt.Sprintf("%s - %s", cat.Code, cat.Name)
		if option == categorySelect.Selected {
			categoryID = cat.ID
			break
		}
	}

	err = d.addOwner(firstNameEntry.Text, lastNameEntry.Text, phoneEntry.Text, emailEntry.Text, categoryID)
	if err != nil {
		d.showMessage("Ошибка", fmt.Sprintf("Не удалось добавить владельца: %v", err))
	} else {
		d.showMessage("Успех", "Владелец успешно добавлен")
		firstNameEntry.SetText("")
		lastNameEntry.SetText("")
		phoneEntry.SetText("")
		emailEntry.SetText("")
	}
}

func (d *DatabaseApp) refreshAddOwnerTab(content *fyne.Container) {
	var categorySelect *widget.Select

	for _, obj := range content.Objects {
		if form, ok := obj.(*widget.Form); ok {
			for _, item := range form.Items {
				if selectWidget, ok := item.Widget.(*widget.Select); ok {
					categorySelect = selectWidget
					break
				}
			}
		}
	}

	if categorySelect == nil {
		log.Printf("Не удалось найти виджет выбора категории")
		return
	}

	categories, err := d.getDriverCategories()
	if err != nil {
		log.Printf("Ошибка получения категорий: %v", err)
		return
	}

	categoryOptions := make([]string, 0, len(categories))
	for _, cat := range categories {
		option := fmt.Sprintf("%s - %s", cat.Code, cat.Name)
		categoryOptions = append(categoryOptions, option)
	}

	categorySelect.Options = categoryOptions
	if len(categoryOptions) > 0 {
		categorySelect.SetSelected(categoryOptions[0])
	} else {
		categorySelect.SetSelected("")
	}
	categorySelect.Refresh()
}

// --- ADD CAR TAB ---

func (d *DatabaseApp) createAddCarTab() *container.Scroll {
	titleLabel := widget.NewLabelWithStyle("Добавление нового автомобиля", fyne.TextAlignCenter, fyne.TextStyle{Bold: true})

	ownerSelect := widget.NewSelect([]string{}, nil)
	ownerSelect.PlaceHolder = "Выберите владельца"

	brandSelect := widget.NewSelect([]string{}, nil)
	brandSelect.PlaceHolder = "Выберите марку"

	modelEntry := widget.NewEntry()
	modelEntry.SetPlaceHolder("Введите модель")
	modelEntry.Validator = func(s string) error {
		if len(s) < 1 {
			return fmt.Errorf("модель не может быть пустой")
		}
		return nil
	}

	yearEntry := widget.NewEntry()
	yearEntry.SetPlaceHolder("Введите год выпуска")
	yearEntry.Validator = func(s string) error {
		if s == "" {
			return nil
		}
		year, err := strconv.Atoi(s)
		if err != nil {
			return fmt.Errorf("год должен быть числом")
		}
		currentYear := time.Now().Year()
		if year < 1886 || year > currentYear+1 {
			return fmt.Errorf("год должен быть между 1886 и %d", currentYear+1)
		}
		return nil
	}

	colorEntry := widget.NewEntry()
	colorEntry.SetPlaceHolder("Введите цвет")

	vinEntry := widget.NewEntry()
	vinEntry.SetPlaceHolder("Введите VIN код")
	vinEntry.Validator = func(s string) error {
		if s == "" {
			return nil
		}
		if len(s) < 10 {
			return fmt.Errorf("VIN должен быть не менее 10 символов")
		}
		return nil
	}

	priceEntry := widget.NewEntry()
	priceEntry.SetPlaceHolder("Введите цену")
	priceEntry.Validator = func(s string) error {
		if s == "" {
			return nil
		}
		_, err := strconv.ParseFloat(s, 64)
		if err != nil {
			return fmt.Errorf("цена должна быть числом")
		}
		return nil
	}

	addBtn := widget.NewButtonWithIcon("Добавить автомобиль", theme.ContentAddIcon(), func() {
		d.addCarHandler(ownerSelect, brandSelect, modelEntry, yearEntry, colorEntry, vinEntry, priceEntry)
	})
	addBtn.Importance = widget.HighImportance

	form := &widget.Form{
		Items: []*widget.FormItem{
			widget.NewFormItem("Владелец:", ownerSelect),
			widget.NewFormItem("Марка:", brandSelect),
			widget.NewFormItem("Модель:", modelEntry),
			widget.NewFormItem("Год выпуска:", yearEntry),
			widget.NewFormItem("Цвет:", colorEntry),
			widget.NewFormItem("VIN код:", vinEntry),
			widget.NewFormItem("Цена:", priceEntry),
		},
		OnSubmit: func() {
			d.addCarHandler(ownerSelect, brandSelect, modelEntry, yearEntry, colorEntry, vinEntry, priceEntry)
		},
		SubmitText: "Добавить автомобиль",
	}

	content := container.NewVBox(
		titleLabel,
		widget.NewSeparator(),
		form,
		layout.NewSpacer(),
		addBtn,
	)

	updateLists := func() {
		owners, err := d.getOwners()
		if err != nil {
			log.Printf("Ошибка получения владельцев: %v", err)
		} else {
			ownerOptions := make([]string, 0, len(owners))
			for _, owner := range owners {
				option := fmt.Sprintf("%d: %s %s", owner.ID, owner.FirstName, owner.LastName)
				ownerOptions = append(ownerOptions, option)
			}
			ownerSelect.Options = ownerOptions
			if len(ownerOptions) > 0 {
				ownerSelect.SetSelected(ownerOptions[0])
			}
			ownerSelect.Refresh()
		}

		brands, err := d.getCarBrands()
		if err != nil {
			log.Printf("Ошибка получения брендов: %v", err)
		} else {
			brandOptions := make([]string, 0, len(brands))
			for _, brand := range brands {
				brandOptions = append(brandOptions, brand.Name)
			}
			brandSelect.Options = brandOptions
			if len(brandOptions) > 0 {
				brandSelect.SetSelected(brandOptions[0])
			}
			brandSelect.Refresh()
		}
	}
	updateLists()

	refreshBtn := widget.NewButtonWithIcon("Обновить списки", theme.ViewRefreshIcon(), func() {
		updateLists()
		d.showMessage("Обновлено", "Списки владельцев и марок обновлены")
	})
	refreshBtn.Importance = widget.MediumImportance
	content.Add(refreshBtn)

	return container.NewScroll(container.NewPadded(content))
}

func (d *DatabaseApp) addCarHandler(ownerSelect, brandSelect *widget.Select, modelEntry, yearEntry, colorEntry, vinEntry, priceEntry *widget.Entry) {
	if modelEntry.Text == "" {
		d.showMessage("Ошибка", "Модель обязательна для заполнения")
		return
	}

	if ownerSelect.Selected == "" || brandSelect.Selected == "" {
		d.showMessage("Ошибка", "Выберите владельца и марку автомобиля")
		return
	}

	owners, err := d.getOwners()
	if err != nil {
		d.showMessage("Ошибка", fmt.Sprintf("Ошибка получения владельцев: %v", err))
		return
	}

	var ownerID int
	for _, owner := range owners {
		option := fmt.Sprintf("%d: %s %s", owner.ID, owner.FirstName, owner.LastName)
		if option == ownerSelect.Selected {
			ownerID = owner.ID
			break
		}
	}

	brands, err := d.getCarBrands()
	if err != nil {
		d.showMessage("Ошибка", fmt.Sprintf("Ошибка получения брендов: %v", err))
		return
	}

	var brandID int
	for _, brand := range brands {
		if brand.Name == brandSelect.Selected {
			brandID = brand.ID
			break
		}
	}

	year, _ := strconv.Atoi(yearEntry.Text)
	price, _ := strconv.ParseFloat(priceEntry.Text, 64)

	err = d.addCar(ownerID, brandID, modelEntry.Text, year, colorEntry.Text, vinEntry.Text, price)
	if err != nil {
		d.showMessage("Ошибка", fmt.Sprintf("Не удалось добавить автомобиль: %v", err))
	} else {
		d.showMessage("Успех", "Автомобиль успешно добавлен")
		modelEntry.SetText("")
		yearEntry.SetText("")
		colorEntry.SetText("")
		vinEntry.SetText("")
		priceEntry.SetText("")
	}
}

func (d *DatabaseApp) refreshAddCarTab(content *fyne.Container) {
	var ownerSelect, brandSelect *widget.Select

	for _, obj := range content.Objects {
		if form, ok := obj.(*widget.Form); ok {
			for _, item := range form.Items {
				if selectWidget, ok := item.Widget.(*widget.Select); ok {
					if ownerSelect == nil {
						ownerSelect = selectWidget
					} else if brandSelect == nil {
						brandSelect = selectWidget
					}
				}
			}
		}
	}

	if ownerSelect == nil || brandSelect == nil {
		log.Printf("Не удалось найти виджеты выбора владельца или марки")
		return
	}

	owners, err := d.getOwners()
	if err != nil {
		log.Printf("Ошибка получения владельцев: %v", err)
	} else {
		ownerOptions := make([]string, 0, len(owners))
		for _, owner := range owners {
			option := fmt.Sprintf("%d: %s %s", owner.ID, owner.FirstName, owner.LastName)
			ownerOptions = append(ownerOptions, option)
		}
		ownerSelect.Options = ownerOptions
		if len(ownerOptions) > 0 {
			ownerSelect.SetSelected(ownerOptions[0])
		}
		ownerSelect.Refresh()
	}

	brands, err := d.getCarBrands()
	if err != nil {
		log.Printf("Ошибка получения брендов: %v", err)
	} else {
		brandOptions := make([]string, 0, len(brands))
		for _, brand := range brands {
			brandOptions = append(brandOptions, brand.Name)
		}
		brandSelect.Options = brandOptions
		if len(brandOptions) > 0 {
			brandSelect.SetSelected(brandOptions[0])
		}
		brandSelect.Refresh()
	}
}
