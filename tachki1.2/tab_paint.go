package main

import (
	"bytes"
	"fmt"
	"image"
	"image/color"
	"image/draw"
	"image/png"
	"io"

	_ "image/jpeg"
	_ "image/png"
	"math"

	"fyne.io/fyne/v2"
	"fyne.io/fyne/v2/canvas"
	"fyne.io/fyne/v2/container"
	"fyne.io/fyne/v2/dialog"
	"fyne.io/fyne/v2/storage"
	"fyne.io/fyne/v2/theme"
	"fyne.io/fyne/v2/widget"
)

// Стандартный размер холста для редактирования
const (
	CanvasWidth  = 500
	CanvasHeight = 500
)

// --- КАСТОМНЫЙ ВИДЖЕТ ДЛЯ РИСОВАНИЯ ---

type DrawingCanvas struct {
	widget.BaseWidget

	img          *image.RGBA   // Буфер пикселей
	canvasImg    *canvas.Image // Объект Fyne для отображения
	brushColor   color.Color   // Текущий цвет кисти
	brushSize    float64       // Размер кисти
	lastX, lastY float32       // Последние координаты мыши
}

func NewDrawingCanvas() *DrawingCanvas {
	// Создаем пустое белое изображение 500x500
	rect := image.Rect(0, 0, CanvasWidth, CanvasHeight)
	img := image.NewRGBA(rect)
	// Заливаем белым
	draw.Draw(img, rect, &image.Uniform{color.White}, image.Point{}, draw.Src)

	dc := &DrawingCanvas{
		img:        img,
		canvasImg:  canvas.NewImageFromImage(img),
		brushColor: color.Black, // По умолчанию черный
		brushSize:  3.0,
	}

	dc.canvasImg.ScaleMode = canvas.ImageScalePixels
	dc.ExtendBaseWidget(dc)
	return dc
}

// Переопределяем MinSize, чтобы виджет занимал нужное место
func (dc *DrawingCanvas) MinSize() fyne.Size {
	return fyne.NewSize(float32(CanvasWidth), float32(CanvasHeight))
}

// Загрузка изображения и размещение его по центру холста
func (dc *DrawingCanvas) LoadImage(data []byte) error {
	// 1. Очищаем холст (заливаем белым)
	rect := dc.img.Bounds()
	draw.Draw(dc.img, rect, &image.Uniform{color.White}, image.Point{}, draw.Src)

	if len(data) > 0 {
		// 2. Декодируем загруженную картинку
		src, _, err := image.Decode(bytes.NewReader(data))
		if err != nil {
			return err
		}

		// 3. Вычисляем позицию для центрирования
		srcBounds := src.Bounds()
		srcW := srcBounds.Dx()
		srcH := srcBounds.Dy()

		// Если картинка огромная - не масштабируем (просто обрежется),
		// Если маленькая - встанет по центру.
		x := (CanvasWidth - srcW) / 2
		y := (CanvasHeight - srcH) / 2

		targetRect := image.Rect(x, y, x+srcW, y+srcH)

		// 4. Рисуем загруженную картинку поверх белого фона
		draw.Draw(dc.img, targetRect, src, srcBounds.Min, draw.Over)
	}

	// Обновляем Fyne
	dc.canvasImg.Refresh()
	dc.Refresh()
	return nil
}

func (dc *DrawingCanvas) CreateRenderer() fyne.WidgetRenderer {
	return widget.NewSimpleRenderer(dc.canvasImg)
}

func (dc *DrawingCanvas) Tapped(ev *fyne.PointEvent) {
	dc.lastX = ev.Position.X
	dc.lastY = ev.Position.Y
	dc.drawPoint(ev.Position.X, ev.Position.Y)
}

func (dc *DrawingCanvas) Dragged(ev *fyne.DragEvent) {
	dc.drawLine(dc.lastX, dc.lastY, ev.Position.X, ev.Position.Y)
	dc.lastX = ev.Position.X
	dc.lastY = ev.Position.Y
}

func (dc *DrawingCanvas) DragEnd() {}

func (dc *DrawingCanvas) drawPoint(x, y float32) {
	// Проверка границ, чтобы не выйти за пределы массива
	ix := int(x)
	iy := int(y)
	if ix < 0 || iy < 0 || ix >= CanvasWidth || iy >= CanvasHeight {
		return
	}

	r := int(dc.brushSize)
	// Рисуем квадрат/круг кисти
	for dy := -r; dy <= r; dy++ {
		for dx := -r; dx <= r; dx++ {
			nx := ix + dx
			ny := iy + dy
			// Проверка границ для каждого пикселя кисти
			if nx >= 0 && nx < CanvasWidth && ny >= 0 && ny < CanvasHeight {
				if dx*dx+dy*dy <= r*r { // Круглая форма
					dc.img.Set(nx, ny, dc.brushColor)
				}
			}
		}
	}
	dc.canvasImg.Refresh()
}

func (dc *DrawingCanvas) drawLine(x0, y0, x1, y1 float32) {
	dist := math.Hypot(float64(x1-x0), float64(y1-y0))
	if dist == 0 {
		return
	}
	step := 1.0 / dist
	for t := 0.0; t <= 1.0; t += step {
		x := x0 + float32(t)*(x1-x0)
		y := y0 + float32(t)*(y1-y0)
		dc.drawPoint(x, y)
	}
}

func (dc *DrawingCanvas) GetBytes() ([]byte, error) {
	var buf bytes.Buffer
	err := png.Encode(&buf, dc.img)
	if err != nil {
		return nil, err
	}
	return buf.Bytes(), nil
}

// --- ИНТЕРФЕЙС ВКЛАДКИ ---

func (d *DatabaseApp) createPaintTab() *container.Scroll {
	// 1. Селектор
	brandSelect := widget.NewSelect([]string{}, nil)
	brandSelect.PlaceHolder = "Выберите логотип"

	// 2. Холст (500x500)
	drawingArea := NewDrawingCanvas()

	// Обертка с рамкой, чтобы видеть границы
	canvasBorder := container.NewPadded(
		container.NewBorder(
			canvas.NewRectangle(color.Black), canvas.NewRectangle(color.Black),
			canvas.NewRectangle(color.Black), canvas.NewRectangle(color.Black),
			drawingArea,
		),
	)

	// 3. Палитра цветов
	// Набор цветов
	colors := []struct {
		Name string
		Col  color.Color
	}{
		{"Черный", color.Black},
		{"Красный", color.RGBA{255, 0, 0, 255}},
		{"Зеленый", color.RGBA{0, 200, 0, 255}},
		{"Синий", color.RGBA{0, 0, 255, 255}},
		{"Желтый", color.RGBA{255, 255, 0, 255}},
		{"Ластик", color.White},
	}

	colorButtons := container.NewGridWithColumns(6)
	for _, c := range colors {
		col := c.Col
		btn := widget.NewButton(c.Name, func() {
			drawingArea.brushColor = col
		})
		if c.Name == "Ластик" {
			btn.Icon = theme.ContentClearIcon()
		}
		colorButtons.Add(btn)
	}

	// 4. Слайдер размера
	sizeSlider := widget.NewSlider(1, 20)
	sizeSlider.Value = 3
	sizeLabel := widget.NewLabel("3 px")
	sizeSlider.OnChanged = func(v float64) {
		drawingArea.brushSize = v
		sizeLabel.SetText(fmt.Sprintf("%.0f px", v))
	}

	toolsPanel := container.NewVBox(
		widget.NewLabelWithStyle("Инструменты рисования:", fyne.TextAlignLeading, fyne.TextStyle{Bold: true}),
		colorButtons,
		container.NewBorder(nil, nil, widget.NewLabel("Размер кисти:"), sizeLabel, sizeSlider),
	)

	// 5. Кнопки действий
	loadBtn := widget.NewButtonWithIcon("Загрузить из БД", theme.DownloadIcon(), func() {
		if brandSelect.Selected == "" {
			d.showMessage("Ошибка", "Выберите бренд из списка")
			return
		}

		brands, _ := d.getCarBrands()
		var brandID int
		for _, b := range brands {
			if b.Name == brandSelect.Selected {
				brandID = b.ID
				break
			}
		}

		imgData, err := d.getBrandImage(brandID)
		if err != nil {
			d.showMessage("Ошибка БД", fmt.Sprintf("%v", err))
			return
		}

		if len(imgData) == 0 {
			d.showMessage("Инфо", "Картинка в базе пустая, создан чистый лист.")
		} else {
			fmt.Printf("Загружено байт: %d\n", len(imgData))
		}

		// Загружаем (даже если пусто, создастся белый фон)
		err = drawingArea.LoadImage(imgData)
		if err != nil {
			d.showMessage("Ошибка", "Не удалось прочитать формат картинки")
		}
	})

	importBtn := widget.NewButtonWithIcon("Импорт (ПК)", theme.FolderOpenIcon(), func() {
		fd := dialog.NewFileOpen(func(reader fyne.URIReadCloser, err error) {
			if err == nil && reader != nil {
				data, _ := io.ReadAll(reader)
				drawingArea.LoadImage(data)
			}
		}, d.window)
		fd.SetFilter(storage.NewExtensionFileFilter([]string{".png", ".jpg", ".jpeg"}))
		fd.Show()
	})

	clearBtn := widget.NewButtonWithIcon("Очистить", theme.ContentClearIcon(), func() {
		drawingArea.LoadImage(nil) // Сброс в белый
	})

	saveBtn := widget.NewButtonWithIcon("Сохранить в БД", theme.DocumentSaveIcon(), func() {
		if brandSelect.Selected == "" {
			d.showMessage("Ошибка", "Выберите бренд!")
			return
		}
		data, _ := drawingArea.GetBytes()

		brands, _ := d.getCarBrands()
		for _, b := range brands {
			if b.Name == brandSelect.Selected {
				d.updateBrandImage(b.ID, data)
				d.showMessage("Успех", "Логотип обновлен")
				break
			}
		}
	})
	saveBtn.Importance = widget.HighImportance

	// Обновление списка
	updateBrands := func() {
		brands, err := d.getCarBrands()
		if err == nil {
			opts := make([]string, 0)
			for _, b := range brands {
				opts = append(opts, b.Name)
			}
			brandSelect.Options = opts
			brandSelect.Refresh()
		}
	}
	refreshListBtn := widget.NewButtonWithIcon("", theme.ViewRefreshIcon(), updateBrands)
	updateBrands()

	// 6. Компоновка
	topControls := container.NewVBox(
		container.NewBorder(nil, nil, nil, refreshListBtn, brandSelect),
		container.NewGridWithColumns(3, loadBtn, importBtn, clearBtn),
		widget.NewSeparator(),
		toolsPanel,
		widget.NewSeparator(),
	)

	// Сам холст кладем в Scroll, чтобы если экран маленький, можно было прокрутить
	// Но ставим холсту фиксированный размер, он не будет сжиматься
	content := container.NewBorder(
		topControls,
		container.NewPadded(saveBtn),
		nil, nil,
		container.NewScroll(container.NewCenter(canvasBorder)), // Центрируем холст
	)

	return container.NewScroll(content)
}

func (d *DatabaseApp) refreshPaintTab(content *fyne.Container) {
	// Можно добавить логику обновления списка
}
