package main

import (
	"bufio"
	"bytes"
	"encoding/json"
	"fmt"
	"io/ioutil"
	"log"
	"net/http"
	"os"
	"os/signal"
	"syscall"
)

type Request struct {
	Bearer      string `json:"bearer"`
	ContentType string `json:"content-type"`
	Endpoint    string `json:"endpoint"`
	Method      string `json:"method"`
}

type Message struct {
	Content string `json:"content"`
	Role    string `json:"role"`
}

type Data struct {
	Messages []Message `json:"messages"`
	Model    string    `json:"model"`
}

type ChatGPTResponse struct {
	Id      string `json:"id"`
	Object  string `json:"object"`
	Created int    `json:"created"`
	Model   string `json:"model"`
	Usage   struct {
		PromptTokens     int `json:"prompt_tokens"`
		CompletionTokens int `json:"completion_tokens"`
		TotalTokens      int `json:"total_tokens"`
	} `json:"usage"`
	Choices []struct {
		Message struct {
			Role    string `json:"role"`
			Content string `json:"content"`
		} `json:"message"`
		FinishReason string `json:"finish_reason"`
		Index        int    `json:"index"`
	} `json:"choices"`
}

const (
	MODEL = "gpt-3.5-turbo"
)

var (
	myRequest = Request{
		Bearer:      "Bearer " + os.Getenv("OPENAI_API_KEY"),
		ContentType: "application/json",
		Endpoint:    "https://api.openai.com/v1/chat/completions",
		Method:      "POST",
	}
	myData = Data{
		Model:    MODEL,
		Messages: []Message{},
	}
)

func main() {
	fmt.Println("Started new chat session. Press Ctrl+C to stop.")
	cancelChan := make(chan os.Signal, 1)
	signal.Notify(cancelChan, syscall.SIGTERM, syscall.SIGINT)
	go func() {
		for {
			question := Prompt()
			SendPrompt(question)
		}
	}()
	<-cancelChan
	fmt.Println("\nChat has ended.")
}

// Ask the user for input
func Prompt() string {
	fmt.Print("YOU: ")
	reader := bufio.NewReader(os.Stdin)
	text, _ := reader.ReadString('\n')
	return text
}

// Send the prompt to the OpenAI API
func SendPrompt(prompt string) {
	newMessage := Message{
		Role:    "user",
		Content: prompt,
	}
	myData.Messages = append(myData.Messages, newMessage)
	myDataJSON, err := json.Marshal(myData)
	req, err := http.NewRequest(myRequest.Method, myRequest.Endpoint, bytes.NewBuffer(myDataJSON))
	req.Header.Set("Authorization", myRequest.Bearer)
	req.Header.Set("Content-Type", myRequest.ContentType)
	client := &http.Client{}
	resp, err := client.Do(req)
	if err != nil {
		log.Fatal(err)
	}
	defer resp.Body.Close()

	body, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		log.Fatal(err)
	}
	PrintResponse(string([]byte(body)))
}

// Print the response from the OpenAI API
func PrintResponse(response string) {
	myResponseMessage := ChatGPTResponse{}
	json.Unmarshal([]byte(response), &myResponseMessage)
	for _, choice := range myResponseMessage.Choices {
		fmt.Println("AI:", choice.Message.Content, "\n")
	}
}
