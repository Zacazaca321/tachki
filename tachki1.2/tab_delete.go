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

func (d *DatabaseApp) createDeleteTab() *container.Scroll {
	titleLabel := widget.NewLabelWithStyle("Удаление записей", fyne.TextAlignCenter, fyne.TextStyle{Bold: true})

	tableSelect := widget.NewSelect([]string{"Автомобили", "Владельцы"}, nil)
	tableSelect.SetSelected("Автомобили")
	tableSelect.PlaceHolder = "Выберите таблицу"

	tableMap := map[string]string{
		"Автомобили": "cars",
		"Владельцы":  "owners",
	}

	idEntry := widget.NewEntry()
	idEntry.SetPlaceHolder("Введите ID записи для удаления")
	idEntry.Validator = func(s string) error {
		if s == "" {
			return fmt.Errorf("введите ID записи")
		}
		_, err := strconv.Atoi(s)
		if err != nil {
			return fmt.Errorf("ID должен быть числом")
		}
		return nil
	}

	deleteBtn := widget.NewButtonWithIcon("Удалить запись", theme.DeleteIcon(), func() {
		if tableName, ok := tableMap[tableSelect.Selected]; ok {
			d.deleteRecordHandler(tableSelect, idEntry, tableName)
		}
	})
	deleteBtn.Importance = widget.DangerImportance

	warningLabel := widget.NewLabel("⚠️ Внимание:\nПри удалении владельца также удаляются все его автомобили (CASCADE).\nОперация необратима!")
	warningLabel.Wrapping = fyne.TextWrapWord

	form := &widget.Form{
		Items: []*widget.FormItem{
			widget.NewFormItem("Таблица:", tableSelect),
			widget.NewFormItem("ID записи:", idEntry),
		},
		OnSubmit: func() {
			if tableName, ok := tableMap[tableSelect.Selected]; ok {
				d.deleteRecordHandler(tableSelect, idEntry, tableName)
			}
		},
		SubmitText: "Удалить запись",
	}

	content := container.NewVBox(
		titleLabel,
		widget.NewSeparator(),
		form,
		widget.NewSeparator(),
		warningLabel,
		layout.NewSpacer(),
		deleteBtn,
	)

	return container.NewScroll(container.NewPadded(content))
}

func (d *DatabaseApp) deleteRecordHandler(tableSelect *widget.Select, idEntry *widget.Entry, tableName string) {
	if idEntry.Text == "" {
		d.showMessage("Ошибка", "Введите ID записи")
		return
	}

	id, err := strconv.Atoi(idEntry.Text)
	if err != nil {
		d.showMessage("Ошибка", "ID должен быть числом")
		return
	}

	err = d.deleteRecord(tableName, id)
	if err != nil {
		d.showMessage("Ошибка", fmt.Sprintf("Не удалось удалить запись: %v", err))
	} else {
		d.showMessage("Успех", fmt.Sprintf("Запись с ID %d успешно удалена из таблицы %s", id, tableSelect.Selected))
		idEntry.SetText("")
	}
}
