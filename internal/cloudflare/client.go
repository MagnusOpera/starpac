package cloudflare

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"net/url"
	"regexp"
	"strings"
	"time"
)

const defaultAPIBaseURL = "https://api.cloudflare.com/client/v4"

var databaseIDPattern = regexp.MustCompile(`^[0-9a-fA-F]{8}-[0-9a-fA-F]{4}-[0-9a-fA-F]{4}-[0-9a-fA-F]{4}-[0-9a-fA-F]{12}$`)

type Client struct {
	accountID  string
	database   string
	resolvedID string
	apiToken   string
	baseURL    string
	httpClient *http.Client
}

type Config struct {
	AccountID  string
	Database   string
	APIToken   string
	BaseURL    string
	HTTPClient *http.Client
}

type apiResponse struct {
	Success  bool            `json:"success"`
	Errors   []responseInfo  `json:"errors"`
	Messages []responseInfo  `json:"messages"`
	Result   json.RawMessage `json:"result"`
}

type responseInfo struct {
	Code    int    `json:"code"`
	Message string `json:"message"`
}

type queryResult struct {
	Success bool             `json:"success"`
	Results []map[string]any `json:"results"`
}

type databaseRecord struct {
	UUID string `json:"uuid"`
	Name string `json:"name"`
}

func New(config Config) (*Client, error) {
	if strings.TrimSpace(config.AccountID) == "" {
		return nil, fmt.Errorf("Cloudflare account id is required")
	}
	if strings.TrimSpace(config.Database) == "" {
		return nil, fmt.Errorf("D1 database name or id is required")
	}
	if strings.TrimSpace(config.APIToken) == "" {
		return nil, fmt.Errorf("Cloudflare API token is required")
	}
	baseURL := strings.TrimRight(config.BaseURL, "/")
	if baseURL == "" {
		baseURL = defaultAPIBaseURL
	}
	httpClient := config.HTTPClient
	if httpClient == nil {
		httpClient = &http.Client{Timeout: 45 * time.Second}
	}
	client := &Client{
		accountID:  strings.TrimSpace(config.AccountID),
		database:   strings.TrimSpace(config.Database),
		apiToken:   strings.TrimSpace(config.APIToken),
		baseURL:    baseURL,
		httpClient: httpClient,
	}
	if databaseIDPattern.MatchString(client.database) {
		client.resolvedID = client.database
	}
	return client, nil
}

func (client *Client) Query(ctx context.Context, sql string) ([]map[string]any, error) {
	databaseID, err := client.databaseID(ctx)
	if err != nil {
		return nil, err
	}
	body := map[string]any{"sql": sql}
	var results []queryResult
	if err := client.request(
		ctx,
		http.MethodPost,
		"/accounts/"+url.PathEscape(client.accountID)+"/d1/database/"+url.PathEscape(databaseID)+"/query",
		body,
		&results,
	); err != nil {
		return nil, err
	}
	if len(results) == 0 {
		return nil, fmt.Errorf("Cloudflare D1 query returned no result")
	}
	for _, result := range results {
		if !result.Success {
			return nil, fmt.Errorf("Cloudflare D1 query failed")
		}
	}
	return results[len(results)-1].Results, nil
}

func (client *Client) Batch(ctx context.Context, statements []string) error {
	if len(statements) == 0 {
		return nil
	}
	databaseID, err := client.databaseID(ctx)
	if err != nil {
		return err
	}
	batch := make([]map[string]string, 0, len(statements))
	for _, statement := range statements {
		if strings.TrimSpace(statement) == "" {
			continue
		}
		batch = append(batch, map[string]string{"sql": statement})
	}
	var results []queryResult
	if err := client.request(
		ctx,
		http.MethodPost,
		"/accounts/"+url.PathEscape(client.accountID)+"/d1/database/"+url.PathEscape(databaseID)+"/query",
		map[string]any{"batch": batch},
		&results,
	); err != nil {
		return err
	}
	if len(results) == 0 {
		return fmt.Errorf("Cloudflare D1 batch returned no results")
	}
	for index, result := range results {
		if !result.Success {
			return fmt.Errorf("Cloudflare D1 batch statement %d failed", index+1)
		}
	}
	return nil
}

func (client *Client) databaseID(ctx context.Context) (string, error) {
	if client.resolvedID != "" {
		return client.resolvedID, nil
	}
	query := url.Values{}
	query.Set("name", client.database)
	var databases []databaseRecord
	if err := client.request(
		ctx,
		http.MethodGet,
		"/accounts/"+url.PathEscape(client.accountID)+"/d1/database?"+query.Encode(),
		nil,
		&databases,
	); err != nil {
		return "", err
	}
	for _, database := range databases {
		if database.Name == client.database {
			client.resolvedID = database.UUID
			return client.resolvedID, nil
		}
	}
	return "", fmt.Errorf("D1 database %q was not found in account %s", client.database, client.accountID)
}

func (client *Client) request(ctx context.Context, method, path string, body any, result any) error {
	var requestBody io.Reader
	if body != nil {
		encoded, err := json.Marshal(body)
		if err != nil {
			return err
		}
		requestBody = bytes.NewReader(encoded)
	}
	request, err := http.NewRequestWithContext(ctx, method, client.baseURL+path, requestBody)
	if err != nil {
		return err
	}
	request.Header.Set("Authorization", "Bearer "+client.apiToken)
	request.Header.Set("Content-Type", "application/json")
	response, err := client.httpClient.Do(request)
	if err != nil {
		return err
	}
	defer response.Body.Close()
	responseBody, err := io.ReadAll(io.LimitReader(response.Body, 8<<20))
	if err != nil {
		return err
	}
	var envelope apiResponse
	if err := json.Unmarshal(responseBody, &envelope); err != nil {
		return fmt.Errorf("Cloudflare API returned HTTP %d with invalid JSON: %w", response.StatusCode, err)
	}
	if response.StatusCode < 200 || response.StatusCode >= 300 || !envelope.Success {
		return apiError(response.StatusCode, envelope.Errors)
	}
	if result == nil {
		return nil
	}
	if err := json.Unmarshal(envelope.Result, result); err != nil {
		return fmt.Errorf("decode Cloudflare API result: %w", err)
	}
	return nil
}

func apiError(status int, errors []responseInfo) error {
	messages := make([]string, 0, len(errors))
	for _, item := range errors {
		if item.Code != 0 {
			messages = append(messages, fmt.Sprintf("%d: %s", item.Code, item.Message))
			continue
		}
		messages = append(messages, item.Message)
	}
	if len(messages) == 0 {
		return fmt.Errorf("Cloudflare API request failed with HTTP %d", status)
	}
	return fmt.Errorf("Cloudflare API request failed with HTTP %d: %s", status, strings.Join(messages, "; "))
}
