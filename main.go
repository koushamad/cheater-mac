package main

import (
	"fyne.io/fyne/v2"
	"fyne.io/fyne/v2/app"
	"fyne.io/fyne/v2/container"
	"fyne.io/fyne/v2/dialog"
	"fyne.io/fyne/v2/widget"
	"github.com/atotto/clipboard"
)

var (
	req chan Message
	res chan Message

	cheatContent = make(chan Message, 1)
	contentText  = ""

	myApp            = app.New()
	myWindow         = myApp.NewWindow("Smart Cheater AI")
	label            = widget.NewLabelWithStyle(contentText, fyne.TextAlignLeading, fyne.TextStyle{Monospace: true})
	body             = container.NewScroll(label)
	loadingIndicator = widget.NewProgressBarInfinite()
	askButton        = widget.NewButton("Answer", onAnswerTapped)
	copyButton       = widget.NewButton("Copy", copyText)
	promptInput      = widget.NewEntry()
	buttons          = container.NewHBox(askButton, copyButton)
	footer           = container.NewHSplit(promptInput, buttons)
	content          = container.NewVSplit(body, footer)
)

func init() {
	req = make(chan Message)
	res = make(chan Message)

	myWindow.Resize(fyne.NewSize(800, 600))
	label.Wrapping = fyne.TextWrapWord
	label.Resize(fyne.NewSize(300, label.MinSize().Height))
	promptInput.SetPlaceHolder("Enter Your Prompt")
	promptInput.SetText("Solve this challenge")
	footer.SetOffset(0.99)
	content.SetOffset(0.99)
	myWindow.SetContent(content)
	buttons.Add(loadingIndicator)
	loadingIndicator.Hide()
}

func main() {
	go runServer()
	go listenServer()

	myWindow.ShowAndRun()
}

func listenServer() {
	for {
		select {
		case msg := <-cheatContent:
			contentText = msg.Content + "\n\n" + "--------------------------------------------------------------------------------------------------\n\n"
			label.SetText(label.Text + contentText)
		}
	}
}

func runServer() {
	go Listen(req, res)

	for {
		select {
		case m := <-req:
			cheatContent <- m
		}
	}
}

func showErrorDialog(err error) {
	dialog.ShowError(err, myWindow)
}

func copyText() {
	err := clipboard.WriteAll(contentText)
	if err != nil {
		showErrorDialog(err)
		return
	}
}

func onAnswerTapped() {
	loadingIndicator.Show()

	prompt := promptInput.Text + "\n\n" + contentText

	res <- Message{
		APIKey:  apiKey,
		Client:  "ios",
		Content: prompt,
	}

	loadingIndicator.Hide()
}
