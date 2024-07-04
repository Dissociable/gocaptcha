package gocaptcha

import (
	"bytes"
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"net/http"
	"strconv"

	"github.com/justhyped/gocaptcha/internal"
)

type CapSolver struct {
	baseUrl string
	apiKey  string
}

func NewCapSolver(apiKey string) *CapSolver {
	return &CapSolver{
		apiKey:  apiKey,
		baseUrl: "https://api.capsolver.com",
	}
}

func (a *CapSolver) SolveImageCaptcha(
	ctx context.Context, settings *Settings, payload *ImageCaptchaPayload,
) (ICaptchaResponse, error) {
	task := map[string]any{
		"type": "ImageToTextTask",
		"body": payload.Base64String,
		"case": payload.CaseSensitive,
	}
	if payload.Score >= 0.0 {
		task["score"] = payload.Score
	}
	if payload.Module != "" {
		task["module"] = payload.Module
	}

	result, err := a.solveTask(ctx, settings, task, nil, nil)
	if err != nil {
		return nil, err
	}

	result.reportBad = a.report("/feedbackTask", result, settings, false)
	result.reportGood = a.report("/feedbackTask", result, settings, true)
	return result, nil
}

func (a *CapSolver) SolveRecaptchaV2(
	ctx context.Context, settings *Settings, payload *RecaptchaV2Payload, cookies Cookies,
) (ICaptchaResponse, error) {
	task := map[string]any{
		"type":        "ReCaptchaV2TaskProxyLess",
		"websiteURL":  payload.EndpointUrl,
		"websiteKey":  payload.EndpointKey,
		"isInvisible": payload.IsInvisibleCaptcha,
	}

	result, err := a.solveTask(ctx, settings, task, nil, cookies)
	if err != nil {
		return nil, err
	}

	result.reportBad = a.report("/feedbackTask", result, settings, false)
	result.reportGood = a.report("/feedbackTask", result, settings, true)
	return result, nil
}

func (a *CapSolver) SolveRecaptchaV2Proxy(
	ctx context.Context, settings *Settings, payload *RecaptchaV2Payload, proxy *Proxy, cookies Cookies,
) (ICaptchaResponse, error) {
	if proxy.IsEmpty() {
		return nil, errors.New("proxy is empty")
	}

	task := map[string]any{
		"type":        "ReCaptchaV2Task",
		"websiteURL":  payload.EndpointUrl,
		"websiteKey":  payload.EndpointKey,
		"isInvisible": payload.IsInvisibleCaptcha,
	}

	result, err := a.solveTask(ctx, settings, task, proxy, cookies)
	if err != nil {
		return nil, err
	}

	result.reportBad = a.report("/feedbackTask", result, settings, false)
	result.reportGood = a.report("/feedbackTask", result, settings, true)
	return result, nil
}

func (a *CapSolver) SolveRecaptchaV3(
	ctx context.Context, settings *Settings, payload *RecaptchaV3Payload, cookies Cookies,
) (ICaptchaResponse, error) {
	task := map[string]any{
		"type":       "RecaptchaV3TaskProxyless",
		"websiteURL": payload.EndpointUrl,
		"websiteKey": payload.EndpointKey,
		"minScore":   payload.MinScore,
		"pageAction": payload.Action,
	}

	result, err := a.solveTask(ctx, settings, task, nil, cookies)
	if err != nil {
		return nil, err
	}

	result.reportBad = a.report("/feedbackTask", result, settings, false)
	result.reportGood = a.report("/feedbackTask", result, settings, true)
	return result, nil
}

func (a *CapSolver) SolveHCaptcha(ctx context.Context, settings *Settings, payload *HCaptchaPayload) (
	ICaptchaResponse, error,
) {
	task := map[string]any{
		"type":       "HCaptchaTaskProxyless",
		"websiteURL": payload.EndpointUrl,
		"websiteKey": payload.EndpointKey,
	}

	result, err := a.solveTask(ctx, settings, task, nil, nil)
	if err != nil {
		return nil, err
	}

	return result, nil
}

func (a *CapSolver) SolveTurnstile(
	ctx context.Context, settings *Settings, payload *TurnstilePayload,
) (ICaptchaResponse, error) {
	task := map[string]any{
		"type":       "TurnstileTaskProxyless",
		"websiteURL": payload.EndpointUrl,
		"websiteKey": payload.EndpointKey,
	}

	result, err := a.solveTask(ctx, settings, task, nil, nil)
	if err != nil {
		return nil, err
	}

	return result, nil
}

func (a *CapSolver) solveTask(
	ctx context.Context, settings *Settings, task map[string]any, proxy *Proxy, cookies Cookies,
) (*CaptchaResponse, error) {
	taskId, syncAnswer, err := a.createTask(ctx, settings, task, proxy, cookies)
	if err != nil {
		return nil, err
	}

	if syncAnswer != nil {
		return &CaptchaResponse{solution: *syncAnswer, taskId: taskId}, nil
	}

	if err := internal.SleepWithContext(ctx, settings.initialWaitTime); err != nil {
		return nil, err
	}

	for i := 0; i < settings.maxRetries; i++ {
		answer, err := a.getTaskResult(ctx, settings, taskId)
		if err != nil {
			return nil, err
		}

		if answer != "" {
			return &CaptchaResponse{solution: answer, taskId: taskId}, nil
		}

		if err := internal.SleepWithContext(ctx, settings.pollInterval); err != nil {
			return nil, err
		}
	}

	return nil, errors.New("max tries exceeded")
}

func (a *CapSolver) createTask(
	ctx context.Context, settings *Settings, task map[string]any, proxy *Proxy, cookies Cookies,
) (string, *string, error) {
	type capSolverSolution struct {
		Text              string `json:"text"`
		RecaptchaResponse string `json:"gRecaptchaResponse"`
		UserAgent         string `json:"userAgent"`
	}

	type capSolverCreateResponse struct {
		ErrorID          int               `json:"errorId"`
		ErrorCode        any               `json:"errorCode"`
		ErrorDescription string            `json:"errorDescription"`
		TaskID           any               `json:"taskId"`
		Status           string            `json:"status"`
		Solution         capSolverSolution `json:"solution"`
	}

	m := map[string]any{"clientKey": a.apiKey, "task": task}

	if proxy != nil && !proxy.IsEmpty() {
		// extend the map with the proxy.Map()
		for k, v := range proxy.Map() {
			m[k] = v
		}
	}

	if cookies != nil {
		m["cookies"] = cookies
	}

	jsonValue, err := json.Marshal(m)
	if err != nil {
		return "", nil, err
	}

	req, err := http.NewRequestWithContext(ctx, http.MethodPost, a.baseUrl+"/createTask", bytes.NewBuffer(jsonValue))
	if err != nil {
		return "", nil, err
	}
	req.Header.Set("content-type", "application/json")

	resp, err := settings.client.Do(req)
	if err != nil {
		return "", nil, err
	}
	defer resp.Body.Close()

	respBody, err := io.ReadAll(resp.Body)
	if err != nil {
		return "", nil, err
	}

	var responseAsJSON capSolverCreateResponse
	if err := json.Unmarshal(respBody, &responseAsJSON); err != nil {
		return "", nil, err
	}

	if responseAsJSON.ErrorID != 0 {
		return "", nil, errors.New(responseAsJSON.ErrorDescription)
	}

	// if the task is solved synchronously, the solution is returned immediately
	var result *string
	if responseAsJSON.Status == "ready" {
		if responseAsJSON.Solution.RecaptchaResponse != "" {
			result = &responseAsJSON.Solution.RecaptchaResponse
		} else {
			result = &responseAsJSON.Solution.Text
		}
	}

	switch responseAsJSON.TaskID.(type) {
	case string:
		// taskId is a string with CapSolver
		return responseAsJSON.TaskID.(string), result, nil
	case float64:
		// taskId is a float64 with CapSolver
		return strconv.FormatFloat(responseAsJSON.TaskID.(float64), 'f', 0, 64), result, nil
	}

	// if you encounter this error with a custom provider, please open an issue
	return "", nil, errors.New("unexpected taskId type, expecting string or float64")
}

func (a *CapSolver) getTaskResult(ctx context.Context, settings *Settings, taskId string) (string, error) {
	type capSolverSolution struct {
		Text              string `json:"text"`
		RecaptchaResponse string `json:"gRecaptchaResponse"`
		UserAgent         string `json:"userAgent"`
	}

	type resultResponse struct {
		Status           string            `json:"status"`
		ErrorID          int               `json:"errorId"`
		ErrorCode        any               `json:"errorCode"`
		ErrorDescription string            `json:"errorDescription"`
		Solution         capSolverSolution `json:"solution"`
	}

	resultData := map[string]string{"clientKey": a.apiKey, "taskId": taskId}
	jsonValue, err := json.Marshal(resultData)
	if err != nil {
		return "", err
	}

	req, err := http.NewRequestWithContext(ctx, http.MethodPost, a.baseUrl+"/getTaskResult", bytes.NewBuffer(jsonValue))
	if err != nil {
		return "", err
	}

	resp, err := settings.client.Do(req)
	if err != nil {
		return "", nil
	}
	defer resp.Body.Close()

	respBody, err := io.ReadAll(resp.Body)
	if err != nil {
		return "", err
	}

	var respJson resultResponse
	if err := json.Unmarshal(respBody, &respJson); err != nil {
		return "", err
	}

	if respJson.ErrorID != 0 {
		return "", errors.New(respJson.ErrorDescription)
	}

	if respJson.Status != "ready" {
		return "", nil
	}

	if respJson.Solution.Text != "" {
		return respJson.Solution.Text, nil
	}

	if respJson.Solution.RecaptchaResponse != "" {
		return respJson.Solution.RecaptchaResponse, nil
	}

	return "", nil
}

func (a *CapSolver) report(
	path string, result *CaptchaResponse, settings *Settings, correct bool,
) func(ctx context.Context) error {
	type response struct {
		ErrorID          int64  `json:"errorId"`
		ErrorCode        string `json:"errorCode"`
		ErrorDescription string `json:"errorDescription"`
	}

	return func(ctx context.Context) error {
		payload := map[string]any{
			"clientKey": a.apiKey,
			"taskId":    result.taskId,
			"result": map[string]any{
				"invalid": !correct,
			},
		}
		rawPayload, err := json.Marshal(payload)
		if err != nil {
			return err
		}

		req, err := http.NewRequestWithContext(ctx, http.MethodPost, a.baseUrl+path, bytes.NewBuffer(rawPayload))
		if err != nil {
			return err
		}

		resp, err := settings.client.Do(req)
		if err != nil {
			return err
		}
		defer resp.Body.Close()

		respBody, err := io.ReadAll(resp.Body)
		if err != nil {
			return err
		}

		var respJson response
		if err := json.Unmarshal(respBody, &respJson); err != nil {
			return err
		}

		if respJson.ErrorID != 0 {
			return fmt.Errorf("%v: %v", respJson.ErrorCode, respJson.ErrorDescription)
		}

		return nil
	}
}

var _ IProvider = (*CapSolver)(nil)
