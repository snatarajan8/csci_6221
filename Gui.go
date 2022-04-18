package main

import (
	"fmt"
	"fyne.io/fyne/v2"
	"fyne.io/fyne/v2/app"
	"fyne.io/fyne/v2/container"
	"fyne.io/fyne/v2/dialog"
	"fyne.io/fyne/v2/widget"
	"log"
)

func run() {
	log.Println("run!")
	log.Println("Input Address is", InputAd)
}
func filter(value string) {
	log.Println("Select set to", value)
}

var InputAd string  // Input File's Address
var OutputAd string // output File's Address

func main() {
	myApp := app.New()
	myWindow := myApp.NewWindow("Go Group")
	//var InputAd string
	input := widget.NewButton("Input .mp4 files", func() {
		file_Dialog := dialog.NewFileOpen(
			func(r fyne.URIReadCloser, _ error) {
				InputAd = r.URI().Path()
				log.Println("Input Address is", InputAd)
			}, myWindow)
		file_Dialog.Show()
		// Show file selection dialog.
	})
	output := widget.NewButton("Output Address", func() {
		file_Dialog := dialog.NewFolderOpen(
			func(r fyne.ListableURI, _ error) {
				OutputAd = r.Path()
				fmt.Println(OutputAd)
			}, myWindow)
		file_Dialog.Show()
		// Show file selection dialog.
		log.Println("Output Address is", OutputAd)
	})
	combo := widget.NewSelect([]string{"Black & White", "Option 2"}, filter)
	run := widget.NewButton("Run", run)
	myWindow.SetContent(container.NewVBox(input, combo, output, run))
	myWindow.Resize(fyne.NewSize(400, 400))
	myWindow.ShowAndRun()

}
