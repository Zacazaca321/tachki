package main

import (
	"fmt"
	"image/color" // Add this

	"fyne.io/fyne/v2"
	"fyne.io/fyne/v2/canvas" // Add this
	"fyne.io/fyne/v2/container"
	"fyne.io/fyne/v2/dialog"
	"fyne.io/fyne/v2/layout"
	"fyne.io/fyne/v2/theme"
	"fyne.io/fyne/v2/widget"
)

func (d *DatabaseApp) showLoginScreen() {
	// 1. Create Input Fields
	serverEntry := widget.NewEntry()
	serverEntry.SetPlaceHolder("e.g., localhost, 1433")
	serverEntry.SetText("localhost")

	dbEntry := widget.NewEntry()
	dbEntry.SetPlaceHolder("e.g., master")
	dbEntry.SetText("master")

	userEntry := widget.NewEntry()
	userEntry.SetPlaceHolder("User ID")
	userEntry.SetText("sa")

	passEntry := widget.NewPasswordEntry()
	passEntry.SetPlaceHolder("Password")

	// 2. Create the Form
	// using HintText to help the user
	form := &widget.Form{
		Items: []*widget.FormItem{
			{Text: "Server Address", Widget: serverEntry, HintText: "Host name or IP"},
			{Text: "Database Name", Widget: dbEntry, HintText: "Target DB"},
			{Text: "Username", Widget: userEntry},
			{Text: "Password", Widget: passEntry},
		},
	}

	// 3. Create 'Connect' Button with Icon
	connectBtn := widget.NewButtonWithIcon("Connect to Database", theme.LoginIcon(), func() {
		// Construct connection string
		connStr := fmt.Sprintf("server=%s;user id=%s;password=%s;database=%s;",
			serverEntry.Text,
			userEntry.Text,
			passEntry.Text,
			dbEntry.Text,
		)

		// Disable button to prevent double clicks during connection attempt
		// (In a real app you might want to show a loading spinner here)

		err := d.connectDB(connStr)
		if err != nil {
			dialog.ShowError(fmt.Errorf("Connection Failed:\n%v", err), d.window)
			return
		}

		// If successful, load the main UI
		d.createUI()
	})
	connectBtn.Importance = widget.HighImportance

	// 4. Layout & Styling

	// Create an invisible spacer to force the card to be at least 450px wide.
	// This solves the "text boxes too narrow" issue.
	widthSpacer := canvas.NewRectangle(color.Transparent)
	widthSpacer.SetMinSize(fyne.NewSize(450, 0))

	// Group elements vertically
	contentVBox := container.NewVBox(
		widthSpacer,
		widget.NewLabelWithStyle("Please enter your credentials", fyne.TextAlignCenter, fyne.TextStyle{Italic: true}),
		widget.NewSeparator(),
		form,
		layout.NewSpacer(), // Pushes button to bottom if resized (though Card fits content)
		connectBtn,
	)

	// Wrap inside a Card for a nice border and background look
	loginCard := widget.NewCard(
		"SQL Server Connection",
		"Car Database Manager",
		container.NewPadded(contentVBox),
	)

	// Center the card in the middle of the window
	centeredLayout := container.NewCenter(loginCard)

	// Set a background image or gradient could go here,
	// but for now, we set the centered layout as content.
	d.window.SetContent(centeredLayout)
}
func (d *DatabaseApp) createUI() {
	// ...
	tabs := container.NewAppTabs(
		container.NewTabItemWithIcon("📊 Просмотр", theme.VisibilityIcon(), d.createViewTab()),
		container.NewTabItemWithIcon("👤 Добавить владельца", theme.ContentAddIcon(), d.createAddOwnerTab()),
		container.NewTabItemWithIcon("🚗 Добавить автомобиль", theme.ContentAddIcon(), d.createAddCarTab()),
		container.NewTabItemWithIcon("✏️ Редактирование", theme.DocumentCreateIcon(), d.createEditTab()),
		container.NewTabItemWithIcon("⚙️ Операции", theme.SettingsIcon(), d.createOperationsTab()),
		container.NewTabItemWithIcon("🎨 Логотипы", theme.ColorPaletteIcon(), d.createPaintTab()),
		container.NewTabItemWithIcon("🗑️ Удаление", theme.DeleteIcon(), d.createDeleteTab()),
	)

	// Настраиваем стиль вкладок
	tabs.SetTabLocation(container.TabLocationTop)
	tabs.SelectTabIndex(0)

	// Обработчик смены вкладок для обновления данных
	tabs.OnSelected = func(tab *container.TabItem) {
		switch tab.Text {
		case "📊 Просмотр":
			if scroll, ok := tab.Content.(*container.Scroll); ok {
				if split, ok := scroll.Content.(*container.Split); ok {
					d.refreshViewTab(split)
				}
			}
		case "👤 Добавить владельца":
			if scroll, ok := tab.Content.(*container.Scroll); ok {
				if content, ok := scroll.Content.(*fyne.Container); ok {
					d.refreshAddOwnerTab(content)
				}
			}
		case "🚗 Добавить автомобиль":
			if scroll, ok := tab.Content.(*container.Scroll); ok {
				if content, ok := scroll.Content.(*fyne.Container); ok {
					d.refreshAddCarTab(content)
				}
			}
		case "⚙️ Операции":
			// Если нужно обновлять данные при входе на вкладку операций
			if scroll, ok := tab.Content.(*container.Scroll); ok {
				if content, ok := scroll.Content.(*fyne.Container); ok {
					d.refreshOperationsTab(content)
				}
			}
		}
	}

	// Создаем основной контейнер с отступами
	mainContainer := container.NewPadded(container.NewMax(tabs))

	// Добавляем футер с информацией
	footer := widget.NewLabel("База данных автомобилей © 2026 | Подключено к SQL Server")
	footer.Alignment = fyne.TextAlignCenter

	// Собираем окончательный интерфейс
	finalContainer := container.NewBorder(
		nil,    // Верхняя панель
		footer, // Нижняя панель (футер)
		nil,    // Левая панель
		nil,    // Правая панель
		mainContainer,
	)

	d.window.SetContent(finalContainer)
}
