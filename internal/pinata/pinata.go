package pinata

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"mime/multipart"
	"net/http"
	"os"
	"path/filepath"
)

type PinataClient struct {
	JwtSecret string
	BaseURL   string
	Client    *http.Client
}

func NewClient(jwtSecret string) *PinataClient {
	return &PinataClient{
		JwtSecret: jwtSecret,
		BaseURL:   "https://uploads.pinata.cloud",
		Client:    &http.Client{},
	}
}

func (c *PinataClient) UploadFile(filePath string) (string, error) {
	file, err := os.Open(filePath)
	if err != nil {
		return "", fmt.Errorf("failed to open file: %w", err)
	}
	defer file.Close()

	body := &bytes.Buffer{}
	writer := multipart.NewWriter(body)

	part, err := writer.CreateFormFile("file", filepath.Base(filePath))
	if err != nil {
		return "", fmt.Errorf("failed to create form file: %w", err)
	}
	_, err = io.Copy(part, file)
	if err != nil {
		return "", fmt.Errorf("failed to copy file: %w", err)
	}

	// Add pinataOptions for public network
	options := map[string]interface{}{
		"cidVersion":        1,
		"wrapWithDirectory": false,
	}
	optionsJSON, err := json.Marshal(options)
	if err != nil {
		return "", fmt.Errorf("failed to marshal options: %w", err)
	}

	err = writer.WriteField("pinataOptions", string(optionsJSON))
	if err != nil {
		return "", fmt.Errorf("failed to write pinataOptions: %w", err)
	}

	// Add network: public to the form
	err = writer.WriteField("network", "public")
	if err != nil {
		return "", fmt.Errorf("failed to write network field: %w", err)
	}

	err = writer.Close()
	if err != nil {
		return "", fmt.Errorf("failed to close writer: %w", err)
	}

	req, err := http.NewRequest("POST", c.BaseURL+"/v3/files", body)
	if err != nil {
		return "", fmt.Errorf("failed to create request: %w", err)
	}

	req.Header.Add("Content-Type", writer.FormDataContentType())
	req.Header.Add("Authorization", fmt.Sprintf("Bearer %s", c.JwtSecret))

	resp, err := c.Client.Do(req)
	if err != nil {
		return "", fmt.Errorf("failed to send request: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		body, _ := io.ReadAll(resp.Body)
		return "", fmt.Errorf("pinata API error: %s - %s", resp.Status, string(body))
	}

	var result struct {
		Data struct {
			Cid string `json:"cid"`
		} `json:"data"`
	}
	if err := json.NewDecoder(resp.Body).Decode(&result); err != nil {
		return "", fmt.Errorf("failed to decode response: %w", err)
	}

	return result.Data.Cid, nil
}

func (c *PinataClient) UploadJSON(data interface{}) (string, error) {
	jsonData, err := json.Marshal(data)
	if err != nil {
		return "", fmt.Errorf("failed to marshal JSON: %w", err)
	}

	// Create a temporary file
	tmpFile, err := os.CreateTemp("", "pinata-*.json")
	if err != nil {
		return "", fmt.Errorf("failed to create temporary file: %w", err)
	}
	defer os.Remove(tmpFile.Name()) // Clean up the temp file when done

	// Write JSON data to the temporary file
	if _, err := tmpFile.Write(jsonData); err != nil {
		tmpFile.Close()
		return "", fmt.Errorf("failed to write JSON to temporary file: %w", err)
	}
	tmpFile.Close()

	// Upload the temporary file using the existing UploadFile method
	return c.UploadFile(tmpFile.Name())
}
