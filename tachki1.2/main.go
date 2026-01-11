package main

import (
	_ "embed"

	"fyne.io/fyne/v2"
	"fyne.io/fyne/v2/app"
	"fyne.io/fyne/v2/theme"
)

//go:embed icon.png
var iconData []byte

// iconResource is a global variable for the app icon
var iconResource = &fyne.StaticResource{
	StaticName:    "icon.png",
	StaticContent: iconData,
}

func main() {
	// Create app with dark theme
	myApp := app.NewWithID("car.database.manager")
	myApp.Settings().SetTheme(theme.DarkTheme())

	window := myApp.NewWindow("Управление базой данных автомобилей")

	// Set app icon
	if iconResource != nil && len(iconResource.StaticContent) > 0 {
		window.SetIcon(iconResource)
	}

	window.Resize(fyne.NewSize(1200, 800))
	window.CenterOnScreen()

	dbApp := &DatabaseApp{
		app:    myApp,
		window: window,
	}

	// Show Login Screen instead of connecting immediately
	dbApp.showLoginScreen()

	// Run app
	window.ShowAndRun()

	// Clean up connection after window closes
	if dbApp.db != nil {
		dbApp.db.Close()
	}
}
