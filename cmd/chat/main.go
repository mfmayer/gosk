package main

import (
	"bufio"
	"fmt"
	"log"
	"os"
	"time"

	"github.com/mfmayer/gosk"
	"github.com/mfmayer/gosk/pkg/gpt"
	"github.com/mfmayer/gosk/pkg/llm"
	"github.com/mfmayer/gosk/pkg/skills/chat"
)

func waitForInput() string {
	fmt.Print("\nYou: ")
	reader := bufio.NewReader(os.Stdin)
	text, _ := reader.ReadString('\n')
	return text
}

func printOutput(output string) {
	fmt.Printf("Bot: %s", output)
}

func main() {
	// create semantic kernel and add chat skill
	kernel := gosk.NewKernel()
	kernel.RegisterGenerators(gpt.Register)
	kernel.RegisterSkills(chat.Register)

	chatFunction, err := kernel.FindFunction("chat", "chatgpt")
	if err != nil {
		log.Fatal(err)
	}

	// start chat
	inputString := waitForInput()
	input := llm.NewContent(inputString).
		SetRole(llm.RoleUser).With("date", time.Now().String()).
		With("botName", "Ida").With("firstName", "John").With("language", "german")
	// response, err := kernel.Call(input, chatFunction)
	// if err != nil {
	// 	log.Fatal(err)
	// }
	var response llm.Content
	for {
		response, err = kernel.Call(input, chatFunction)
		if err != nil {
			log.Fatal(err)
		}
		printOutput(response.String())
		inputString := waitForInput()
		input = llm.NewContent(inputString).
			SetRole(llm.RoleUser).
			WithPredecessor(response)
		// response, err = kernel.CallWithName(input, "chat", "chatgpt")
		// if err != nil {
		// 	log.Fatal(err)
		// }
	}
}
