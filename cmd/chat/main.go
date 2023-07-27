package main

import (
	"bufio"
	"fmt"
	"log"
	"os"
	"time"

	"github.com/mfmayer/gosk"
	"github.com/mfmayer/gosk/pkg/gptgenerator"
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

	kernel := gosk.NewKernel()
	generator, err := gptgenerator.NewGPT35Generator()
	if err != nil {
		log.Fatal(err)
	}
	generators := map[string]llm.Generator{
		"gpt35": generator,
	}
	chatSkill, err := chat.NewChatSkill(generators)
	if err != nil {
		log.Fatal(err)
	}
	kernel.AddSkills(chatSkill)

	inputString := waitForInput()
	input := llm.NewContent(inputString).
		WithRoleOption(llm.RoleUser).
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
			WithRoleOption(llm.RoleUser).
			WithPredecessor(response)
		response, err = kernel.Call("chat", "chatgpt", input)
		if err != nil {
			log.Fatal(err)
		}
	}
}
