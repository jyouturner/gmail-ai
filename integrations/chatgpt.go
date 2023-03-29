package integration

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io/ioutil"
	"net/http"
	"strings"

	"golang.org/x/time/rate"
)

type ChatGPTClient struct {
	ChatGPTAPIURL string
	APIKey        string
	limiter       *rate.Limiter
}

func NewChatGPTClient(apiKey string) *ChatGPTClient {

	return &ChatGPTClient{
		ChatGPTAPIURL: "https://api.openai.com/v1/engines/text-davinci-003/completions",
		APIKey:        apiKey,
	}
}

type ChatGPTRequest struct {
	Prompt    string `json:"prompt"`
	MaxTokens int    `json:"max_tokens"`
}

type ChatGPTResponse struct {
	Choices []struct {
		Text string `json:"text"`
	} `json:"choices"`
}

func (c *ChatGPTClient) IsRejectionEmail(emailContent string) (bool, error) {
	// Prepare the API request
	requestBody := &ChatGPTRequest{
		Prompt:    fmt.Sprintf("Given the following email, is it a rejection email?  Just reply Yes or No. \"%s\"", emailContent),
		MaxTokens: 10,
	}
	requestBodyBytes, err := json.Marshal(requestBody)
	if err != nil {
		return false, err
	}

	// Create an HTTP client and set the API key in the header
	client := Wrap(http.DefaultClient,
		WithRateLimit())
	req, err := http.NewRequest("POST", c.ChatGPTAPIURL, bytes.NewReader(requestBodyBytes))
	if err != nil {
		return false, err
	}
	req.Header.Add("Authorization", fmt.Sprintf("Bearer %s", c.APIKey))
	req.Header.Add("Content-Type", "application/json")

	// Send the API request
	resp, err := client.Do(req)
	if err != nil {
		return false, err
	}
	defer resp.Body.Close()

	// Parse the API response
	responseBodyBytes, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		return false, err
	}
	var chatGPTResponse ChatGPTResponse
	if err := json.Unmarshal(responseBodyBytes, &chatGPTResponse); err != nil {
		return false, err
	}
	//fmt.Println(chatGPTResponse.Choices)
	// Check if the response indicates a rejection email
	if len(chatGPTResponse.Choices) == 0 {
		// no response from API, but don't fail
		return false, nil
	}
	answer := strings.TrimSpace(chatGPTResponse.Choices[0].Text)
	return strings.EqualFold(answer, "yes"), nil
}
