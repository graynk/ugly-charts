package main

import (
	"github.com/gotk3/gotk3/gtk"
	"github.com/graynk/ugly-charts"
	"log"
	"math/rand"
)

func main() {
	// Инициализируем GTK.
	gtk.Init(nil)

	// Создаём билдер
	b, err := gtk.BuilderNew()
	if err != nil {
		log.Fatal("Ошибка:", err)
	}

	// Загружаем в билдер окно из файла Glade
	err = b.AddFromFile("example/example.glade")
	if err != nil {
		log.Fatal("Ошибка:", err)
	}

	// Получаем объект главного окна по ID
	obj, err := b.GetObject("main_window")
	if err != nil {
		log.Fatal("Ошибка:", err)
	}

	// Преобразуем из объекта именно окно типа gtk.Window
	// и соединяем с сигналом "destroy" чтобы можно было закрыть
	// приложение при закрытии окна
	win := obj.(*gtk.ApplicationWindow)
	win.Connect("destroy", func() {
		gtk.MainQuit()
	})

	// Получаем кнопку
	obj, _ = b.GetObject("add_button")
	addButton := obj.(*gtk.Button)
	obj, _ = b.GetObject("drawing_area")
	chartArea := obj.(*gtk.DrawingArea)
	// Data
	chart := uglycharts.NewLineChart(chartArea)
	chart.SetDrawMarker(true)
	chart.SetMarkerSize(4)
	chart.SetLineWidth(2)
	chart.SetMinX(0)
	chart.SetMinY(0)
	chart.SetMaxX(10)
	chart.SetMaxY(100)
	chart.SetAutoRangingX(true)
	chart.SetAutoRangingY(true)
	chart.SetTitle("Stonks")
	series := make([]uglycharts.Series, 4)
	for i := 0; i < 4; i++ {
		s := uglycharts.NewFloatSeries(60)
		chart.AddSeries(s)
		series[i] = s
	}
	count := 0
	var growY float64 = 0
	addButton.Connect("clicked", func() {
		for i := 0; i < 4; i++ {
			y := (rand.Float64() * 20) + 57
			//y := rand.Float64()
			point := uglycharts.FloatPoint{X: float64(count), Y: y + growY}
			series[i].Add(point)
		}
		count++
		growY += 5
	})
	win.ShowAll()
	gtk.Main()
}
