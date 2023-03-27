package main

import (
	"fyne.io/fyne/v2"
	"fyne.io/fyne/v2/app"
	"fyne.io/fyne/v2/container"
	"fyne.io/fyne/v2/dialog"
	"fyne.io/fyne/v2/widget"
	"github.com/atotto/clipboard"
)

const line = "\n\n----------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------\n\n"
const defaultPrompt = "Solve this challenge and develop a code for it in Golang  and  explain your solution in English language B1 level." + line

var (
	req  chan Message
	res  chan Message
	done chan bool

	cheatContent = make(chan Message, 1)
	contentText  = defaultPrompt

	myApp            = app.New()
	myWindow         = myApp.NewWindow("Smart Cheater AI")
	label            = widget.NewMultiLineEntry()
	body             = container.NewScroll(label)
	loadingIndicator = widget.NewProgressBarInfinite()
	askButton        = widget.NewButton("Answer", onAnswerTapped)
	copyButton       = widget.NewButton("Copy", copyText)
	clearButton      = widget.NewButton("Clear", clearText)
	promptInput      = widget.NewEntry()
	buttons          = container.NewHBox(askButton, copyButton, clearButton)
	footer           = container.NewHSplit(promptInput, buttons)
	content          = container.NewVSplit(body, footer)
)

func init() {
	req = make(chan Message)
	res = make(chan Message)
	done = make(chan bool)

	myWindow.Resize(fyne.NewSize(800, 600))
	label.Wrapping = fyne.TextWrapWord
	label.Resize(fyne.NewSize(300, label.MinSize().Height))
	promptInput.SetPlaceHolder("Enter Your Prompt")
	promptInput.OnSubmitted = onEnterPressed
	promptInput.SetText("")
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
			contentText += msg.Content + line
			label.SetText(contentText)
		}
	}
}

func runServer() {
	go Listen(req, res, done)

	for {
		select {
		case m := <-req:
			cheatContent <- m
		case <-done:
			return
		}
	}
}

func showErrorDialog(err error) {
	dialog.ShowError(err, myWindow)
}

func onEnterPressed(text string) {
	loadingIndicator.Show()

	prompt := promptInput.Text
	promptInput.SetText("")

	res <- Message{
		APIKey:  apiKey,
		Client:  "ios",
		Content: prompt,
	}

	loadingIndicator.Hide()
}

func copyText() {
	err := clipboard.WriteAll(contentText)
	if err != nil {
		showErrorDialog(err)
		return
	}
}

func clearText() {
	label.SetText("defaultPrompt")
	contentText = defaultPrompt
}

func onAnswerTapped() {
	loadingIndicator.Show()

	prompt := contentText

	res <- Message{
		APIKey:  apiKey,
		Client:  "ios",
		Content: prompt,
	}

	loadingIndicator.Hide()
}
