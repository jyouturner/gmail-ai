package integration

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io/ioutil"
	"net/http"
	"strings"
)

// ChatGPTClient is a client for the ChatGPT API
type ChatGPTClient struct {
	ChatGPTAPIURL string
	APIKey        string
	client        *http.Client
}

// NewChatGPTClient creates a new ChatGPT client
func NewChatGPTClient(url string, apiKey string, options ...Option) *ChatGPTClient {
	// Create an HTTP client and set the API key in the header
	client := Wrap(http.DefaultClient, options...)
	return &ChatGPTClient{
		ChatGPTAPIURL: url,
		APIKey:        apiKey,
		client:        client,
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

// IsRejectionEmail returns true if the given email content is a rejection email
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

	req, err := http.NewRequest("POST", c.ChatGPTAPIURL, bytes.NewReader(requestBodyBytes))
	if err != nil {
		return false, err
	}
	req.Header.Add("Authorization", fmt.Sprintf("Bearer %s", c.APIKey))
	req.Header.Add("Content-Type", "application/json")

	// Send the API request
	resp, err := c.client.Do(req)
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
