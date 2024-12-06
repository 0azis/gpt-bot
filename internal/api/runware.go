package api

import (
	"encoding/json"
	"fmt"
	"log/slog"

	"github.com/valyala/fasthttp"
)

type runwareInterface interface {
	SendMessage(prompt string) (string, error)
}

type runwareClient struct {
	baseUrl string
	token   string
}

func newRunware(token string) runwareInterface {
	return runwareClient{"https://api.runware.ai/v1", token}
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
	// body := []requestBody{newBody(prompt)}
	body := []byte(fmt.Sprintf(`[{"taskType":"imageInference","taskUUID":"39d7207a-87ef-4c93-8082-1431f9c1dc97","positivePrompt":"%s","width":512,"height":512,"modelId":"civitai:102438@133677"}]`, prompt))

	resp, err := rc.doRequest(body)
	if err != nil {
		slog.Error(err.Error())
		return imageLink, err
	}

	var respBody responseBody
	err = json.Unmarshal(resp, &respBody)
	if err != nil {
		slog.Error(err.Error())
		return imageLink, err
	}

	return respBody.Data[0].ImageURL, nil
}

func (rc runwareClient) doRequest(body []byte) ([]byte, error) {
	req := fasthttp.AcquireRequest()
	resp := fasthttp.AcquireResponse()
	defer fasthttp.ReleaseRequest(req)   // <- do not forget to release
	defer fasthttp.ReleaseResponse(resp) // <- do not forget to release

	req.SetBody(body)
	req.SetRequestURI(rc.baseUrl)
	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("Authorization", fmt.Sprintf("Bearer %s", rc.token))

	err := fasthttp.Do(req, resp)
	if err != nil {
		return []byte{}, err
	}

	bodyBytes := resp.Body()
	return bodyBytes, nil
}
