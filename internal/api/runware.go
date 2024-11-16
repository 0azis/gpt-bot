package api

import (
	"bytes"
	"encoding/json"
	"fmt"
	"log/slog"
	"net/http"
)

type runwareInterface interface {
	SendMessage(prompt string) (string, error)
}

type runwareClient struct {
	baseUrl string
	token   string
}

func newRunware(token string) runwareInterface {
	return runwareClient{token, "https://api.runware.ai/v1"}
}

type requestBody struct {
	TaskType       string `json:"taskType"`
	TaskUUID       string `json:"taskUUID"`
	PositivePrompt string `json:"positivePrompt"`
	Width          int    `json:"width"`
	Height         int    `json:"height"`
	ModelID        string `json:"modelId"`
}

type responseBody struct {
	Data []struct {
		ImageURL string `json:"imageURL"`
	} `json:"data"`
}

func newBody(prompt string) requestBody {
	return requestBody{
		TaskType:       "imageInference",
		TaskUUID:       "39d7207a-87ef-4c93-8082-1431f9c1dc97",
		PositivePrompt: prompt,
		Width:          512,
		Height:         512,
		ModelID:        "civitai:102438@133677",
	}
}

func (rc runwareClient) SendMessage(prompt string) (string, error) {
	var imageLink string
	var body []requestBody
	body = append(body, newBody(prompt))

	b, err := json.Marshal(body)
	if err != nil {
		slog.Error(err.Error())
		return imageLink, err
	}

	req, err := http.NewRequest("POST", rc.baseUrl, bytes.NewBuffer(b))
	if err != nil {
		slog.Error(err.Error())
		return imageLink, err
	}
	req.Header = http.Header{
		"Content-Type":  {"application/json"},
		"Authorization": {fmt.Sprintf("Bearer %s", rc.token)},
	}

	client := &http.Client{}
	resp, err := client.Do(req)
	if err != nil {
		slog.Error(err.Error())
		return imageLink, err
	}
	defer resp.Body.Close()

	var respBody responseBody
	err = json.NewDecoder(resp.Body).Decode(&respBody)
	if err != nil {
		slog.Error(err.Error())
		return imageLink, err
	}

	return respBody.Data[0].ImageURL, nil
}
