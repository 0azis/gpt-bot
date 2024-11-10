package api

import (
	"bytes"
	"encoding/json"
	"fmt"
	"net/http"
)

const baseUrl = "https://api.runware.ai/v1"

type runwareInterface interface {
	SendMessage(prompt string) error
}

type runwareClient struct {
	token string
}

type bodyRequest struct {
	TaskType       string `json:"taskType"`
	TaskUUID       string `json:"taskUUID"`
	PositivePrompt string `json:"positivePrompt"`
	Width          int    `json:"width"`
	Height         int    `json:"height"`
	ModelID        string `json:"modelId"`
}

type bodyResult struct {
	Data interface{} `json:"data"`
	// TaskType  string `json:"taskType"`
	// TaskUUID  string `json:"taskUUID"`
	// ImageUUID string `json:"imageUUID"`
	// ImageURL  string `json:"imageURL"`
}

func newBody(prompt string) bodyRequest {
	return bodyRequest{
		PositivePrompt: prompt,
	}
}

func newRunware(token string) runwareInterface {
	return runwareClient{token}
}

func (rc runwareClient) SendMessage(prompt string) error {
	body := bodyRequest{
		TaskType:       "imageInference",
		TaskUUID:       "39d7207a-87ef-4c93-8082-1431f9c1dc97",
		PositivePrompt: prompt,
		Width:          512,
		Height:         512,
		ModelID:        "civitai:102438@133677",
	}
	b, err := json.Marshal(body)
	if err != nil {
		return err
	}

	req, err := http.NewRequest("POST", baseUrl, bytes.NewBuffer(b))
	if err != nil {
		return err
	}
	req.Header = http.Header{
		"Content-Type":  {"application/json"},
		"Authorization": {fmt.Sprintf("Bearer %s", rc.token)},
	}
	fmt.Println(*req)
	// req.Header.Set("Authorization", rc.token)
	// req.Header.Set("Content-Type", "application/json")

	client := &http.Client{}
	resp, err := client.Do(req)
	fmt.Println(*resp)
	if err != nil {
		return err
	}
	defer resp.Body.Close()

	var respBody bodyResult
	err = json.NewDecoder(resp.Body).Decode(&respBody)
	if err != nil {
		return err
	}
	fmt.Println(respBody)

	return nil
}
