package main

import (
	"bufio"
	"fmt"
	"log"
	"os"
	"time"

	"github.com/mfmayer/gosk"
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
	// create chat skill
	chatSkill, err := chat.New()
	if err != nil {
		log.Fatal(err)
	}

	// create semantic kernel and add chat skill
	kernel := gosk.NewKernel()
	kernel.AddSkills(chatSkill)

	// start chat
	inputString := waitForInput()
	input := llm.NewContent(inputString).
		SetRole(llm.RoleUser).
		With("date", time.Now().String()).
		With("botName", "Ida").
		With("firstName", "John").
		With("language", "german")
	response, err := kernel.Call("chat", "chatgpt", input)
	if err != nil {
		log.Fatal(err)
	}
	for {
		printOutput(response.String())
		inputString := waitForInput()
		input := llm.NewContent(inputString).
			SetRole(llm.RoleUser).
			WithPredecessor(response)
		response, err = kernel.Call("chat", "chatgpt", input)
		if err != nil {
			log.Fatal(err)
		}
	}
}
