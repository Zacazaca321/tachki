package main

import (
	"fmt"
	"strconv"

	"fyne.io/fyne/v2"
	"fyne.io/fyne/v2/container"
	"fyne.io/fyne/v2/layout"
	"fyne.io/fyne/v2/theme"
	"fyne.io/fyne/v2/widget"
)

func (d *DatabaseApp) createOperationsTab() *container.Scroll {
	titleLabel := widget.NewLabelWithStyle("Массовые операции (Stored Procedures)", fyne.TextAlignCenter, fyne.TextStyle{Bold: true})

	// --- Секция 1: Индексация цен ---

	lblHeader := widget.NewLabel("Массовое изменение цен по марке авто:")

	// Выбор бренда
	brandSelect := widget.NewSelect([]string{}, nil)
	brandSelect.PlaceHolder = "Выберите марку (например, BMW)"

	// Ввод процента
	percentEntry := widget.NewEntry()
	percentEntry.PlaceHolder = "Процент (например: 10 или -5)"

	// Кнопка выполнения
	execBtn := widget.NewButtonWithIcon("Применить индексацию", theme.MediaPlayIcon(), func() {
		if brandSelect.Selected == "" {
			d.showMessage("Ошибка", "Выберите марку автомобиля")
			return
		}
		if percentEntry.Text == "" {
			d.showMessage("Ошибка", "Введите процент")
			return
		}

		// Парсинг процента
		percent, err := strconv.ParseFloat(percentEntry.Text, 64)
		if err != nil {
			d.showMessage("Ошибка", "Процент должен быть числом")
			return
		}

		// Получаем ID бренда (нужно снова найти ID по имени)
		brands, _ := d.getCarBrands() // В реальном коде лучше кэшировать или хранить map
		var brandID int
		for _, b := range brands {
			if b.Name == brandSelect.Selected {
				brandID = b.ID
				break
			}
		}

		// Вызов процедуры
		err = d.massPriceUpdate(brandID, percent)
		if err != nil {
			d.showMessage("Ошибка", fmt.Sprintf("Сбой процедуры: %v", err))
		} else {
			d.showMessage("Успех", fmt.Sprintf("Цены для %s успешно изменены на %.1f%%!", brandSelect.Selected, percent))
		}
	})
	execBtn.Importance = widget.WarningImportance // Оранжевая кнопка (опасно!)

	// Функция обновления списка брендов
	updateBrands := func() {
		brands, err := d.getCarBrands()
		if err != nil {
			return
		}
		options := make([]string, 0, len(brands))
		for _, b := range brands {
			options = append(options, b.Name)
		}
		brandSelect.Options = options
		brandSelect.Refresh()
	}

	// Кнопка обновления списка
	refreshBtn := widget.NewButtonWithIcon("", theme.ViewRefreshIcon(), updateBrands)

	// Начальная загрузка
	updateBrands()

	// Компоновка формы
	formContainer := container.NewVBox(
		lblHeader,
		container.NewBorder(nil, nil, nil, refreshBtn, brandSelect), // Селект с кнопкой обновления
		widget.NewForm(widget.NewFormItem("Процент изменения:", percentEntry)),
		layout.NewSpacer(),
		execBtn,
	)

	// Обертка в карточку
	card := widget.NewCard("Индексация цен", "Изменение стоимости всех авто бренда", container.NewPadded(formContainer))

	content := container.NewVBox(
		titleLabel,
		widget.NewSeparator(),
		card,
	)

	return container.NewScroll(container.NewPadded(content))
}

// Вспомогательный метод для обновления вкладки при переключении
func (d *DatabaseApp) refreshOperationsTab(content *fyne.Container) {
	// Здесь можно реализовать логику обновления, если нужно
	// В данном простом примере мы обновляем бренды внутри самой кнопки refreshBtn
}
