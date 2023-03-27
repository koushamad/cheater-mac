package main

import (
	"fmt"
	"fyne.io/fyne/v2"
	"fyne.io/fyne/v2/app"
	"fyne.io/fyne/v2/container"
	"fyne.io/fyne/v2/dialog"
	"fyne.io/fyne/v2/widget"
	"github.com/0x9ef/openai-go"
	"github.com/atotto/clipboard"
	"golang.org/x/net/context"
	"log"
	"sync"
)

type textData struct {
	Text string `json:"text"`
}

var (
	lock = sync.Mutex{}

	cheatContent              = make(chan Message, 1)
	aiAns                     = make(chan string, 1)
	aiTranslate               = make(chan string, 1)
	cheatContentText          = ""
	aiAnsText                 = ""
	aiTranslateText           = ""
	programmingLanguages      = []string{"Go", "Python", "JavaScript", "Java", "C++", "C#", "Ruby", "PHP", "Swift", "Rust", "Scala", "Kotlin", "Dart", "Objective-C", "Perl", "Bash Script"}
	languages                 = []string{"Level A1", "Level A2", "Level B1", "Level B2", "Level C1", "Level C2"}
	defaultLanguage           = "Level B1"
	defaultProgramingLanguage = "Go"
	defaultApiToken           = "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"

	myApp                    = app.New()
	myWindow                 = myApp.NewWindow("Smart Cheater AI")
	label1                   = widget.NewLabelWithStyle(cheatContentText, fyne.TextAlignLeading, fyne.TextStyle{Monospace: true})
	label2                   = widget.NewLabelWithStyle(aiAnsText, fyne.TextAlignLeading, fyne.TextStyle{Monospace: true})
	scroll1                  = container.NewScroll(label1)
	scroll2                  = container.NewScroll(label2)
	apiTokenInput            = widget.NewEntry()
	maxTokensInput           = widget.NewEntry()
	maxTokens                = 1000
	loadingIndicator         = widget.NewProgressBarInfinite()
	selectProgramingLanguage = widget.NewSelect(programmingLanguages, func(selected string) { defaultProgramingLanguage = selected })
	selectLanguage           = widget.NewSelect(languages, func(selected string) { defaultLanguage = selected })
	askButton                = widget.NewButton("Answer", onAskAITapped)
	translateButton          = widget.NewButton("Translate", onTranslateAITapped)
	copyButton               = widget.NewButton("Copy", copyText)
	promptInput              = widget.NewEntry()
	configInputs             = container.NewHBox(selectProgramingLanguage, selectLanguage)
	apiInput                 = container.NewHSplit(apiTokenInput, maxTokensInput)
	header                   = container.NewHSplit(configInputs, apiInput)
	split                    = container.NewHSplit(scroll1, scroll2)
	body                     = container.NewVSplit(header, split)
	buttons                  = container.NewHBox(askButton, translateButton, copyButton)
	footer                   = container.NewHSplit(promptInput, buttons)
	content                  = container.NewVSplit(body, footer)
)

func init() {
	myWindow.Resize(fyne.NewSize(800, 600))
	apiTokenInput.SetPlaceHolder("Enter Chat GPT API Token")
	label1.Wrapping = fyne.TextWrapWord
	label2.Wrapping = fyne.TextWrapWord
	label1.Resize(fyne.NewSize(300, label1.MinSize().Height))
	label2.Resize(fyne.NewSize(300, label2.MinSize().Height))
	apiTokenInput.SetText(defaultApiToken)
	maxTokensInput.SetText(fmt.Sprintf("%d", maxTokens))
	promptInput.SetPlaceHolder("Enter Your Prompt")
	promptInput.SetText("Solve this challenge")
	selectProgramingLanguage.SetSelected(defaultProgramingLanguage)
	selectLanguage.SetSelected(defaultLanguage)
	footer.SetOffset(0.99)
	apiInput.SetOffset(0.9)
	header.SetOffset(0.01)
	body.SetOffset(0.01)
	content.SetOffset(0.99)
	myWindow.SetContent(content)
	buttons.Add(loadingIndicator)
	loadingIndicator.Hide()
	lock.Lock()
}

func main() {
	log.SetFlags(log.LstdFlags)

	go runServer()

	go func() {
		for {
			select {
			case msg := <-cheatContent:
				cheatContentText += msg.Content +
					"\n\n" + "-----------------------------------------------------------------------------\n\n"
				label1.SetText(cheatContentText)
				lock.Unlock()
			case text := <-aiAns:
				aiAnsText = text
				label2.SetText(aiAnsText)
			case text := <-aiTranslate:
				aiTranslateText = text
				label1.SetText(aiTranslateText)
			}
		}
	}()

	myWindow.ShowAndRun()
}

func updateAttributes() {
	defaultApiToken = apiTokenInput.Text
	maxTokens, _ = fmt.Sscanf(maxTokensInput.Text, "%d", &maxTokens)
}

func showErrorDialog(err error) {
	dialog.ShowError(err, myWindow)
}

func copyText() {
	prompt := promptInput.Text + " in " + defaultProgramingLanguage + ":\n\n" + cheatContentText
	err := clipboard.WriteAll(prompt)
	if err != nil {
		fmt.Println("Error setting clipboard:", err)
		return
	}
}

func onAskAITapped() {
	updateAttributes()
	go func() {
		if lock.TryLock() {
			loadingIndicator.Show()
			prompt := promptInput.Text + " in " + defaultProgramingLanguage + " programing language and explain solution in the " + defaultLanguage + "\n\n" + cheatContentText

			e := openai.New(defaultApiToken)
			r, err := e.Completion(context.Background(), &openai.CompletionOptions{
				Model:       openai.ModelGPT3TextDavinci003,
				MaxTokens:   maxTokens,
				Prompt:      []string{prompt},
				Temperature: 0.3,
				N:           1,
			})

			if err != nil {
				showErrorDialog(err)
			} else {
				aiAns <- r.Choices[0].Text
			}
			loadingIndicator.Hide()
			lock.Unlock()
		}
	}()
}

func onTranslateAITapped() {
	updateAttributes()
	go func() {
		if lock.TryLock() {
			loadingIndicator.Show()
			prompt := "Translate this text to " + defaultLanguage + "\n\n" + cheatContentText

			e := openai.New(defaultApiToken)
			r, err := e.Completion(context.Background(), &openai.CompletionOptions{
				Model:       openai.ModelGPT3TextDavinci003,
				MaxTokens:   maxTokens,
				Prompt:      []string{prompt},
				Temperature: 0.3,
				N:           1,
			})

			if err != nil {
				showErrorDialog(err)
			} else {
				aiTranslate <- r.Choices[0].Text
			}

			loadingIndicator.Hide()
			lock.Unlock()
		}
	}()
}

func runServer() {

	msg := make(chan Message)
	go ListenWS(msg)

	for {
		select {
		case m := <-msg:
			cheatContent <- m
		}
	}
}
