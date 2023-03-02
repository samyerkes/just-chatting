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

type Response struct {
	Choices []struct {
		Message struct {
			Role    string `json:"role"`
			Content string `json:"content"`
		} `json:"message"`
		FinishReason string `json:"finish_reason"`
		Index        int    `json:"index"`
	} `json:"choices"`
}

var (
	myRequest = Request{
		Bearer:      "Bearer " + os.Getenv("OPENAI_API_KEY"),
		ContentType: "application/json",
		Endpoint:    "https://api.openai.com/v1/chat/completions",
		Method:      "POST",
	}
	myData = Data{
		Model:    "gpt-3.5-turbo",
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
	if err != nil {
		log.Fatal(err)
	}
	headers := map[string]string{
		"Authorization": myRequest.Bearer,
		"Content-Type":  myRequest.ContentType,
	}
	response, err := MakeHttpRequest(myRequest.Method, myRequest.Endpoint, headers, myDataJSON)
	if err != nil {
		log.Fatal(err)
	}
	PrintResponse(response)
}

func MakeHttpRequest(method string, endpoint string, headers map[string]string, data []byte) (string, error) {
	req, err := http.NewRequest(method, endpoint, bytes.NewBuffer(data))
	for k, v := range headers {
		req.Header.Set(k, v)
	}
	client := &http.Client{}
	resp, err := client.Do(req)
	if err != nil {
		return "", err
	}
	defer resp.Body.Close()

	body, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		return "", err
	}
	return string([]byte(body)), nil
}

// Print the response from the OpenAI API
func PrintResponse(response string) {
	myResponseMessage := Response{}
	json.Unmarshal([]byte(response), &myResponseMessage)
	for _, choice := range myResponseMessage.Choices {
		fmt.Println("AI:", choice.Message.Content, "\n")
	}
}
