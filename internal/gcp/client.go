package gcp

import (
	"encoding/json"
	"fmt"
	"net/http"
)

type GCPClient struct {
	client *http.Client
}

func NewClient(client *http.Client) *GCPClient {
	return &GCPClient{client: client}
}

type GCPProject struct {
	ProjectID string `json:"projectId"`
	Name      string `json:"name"`
}

type GCPAPIKey struct {
	Name        string `json:"name"`
	UID         string `json:"uid"`
	DisplayName string `json:"displayName"`
	KeyString   string `json:"keyString"`
}

func (c *GCPClient) ListProjects() ([]GCPProject, error) {
	resp, err := c.client.Get("https://cloudresourcemanager.googleapis.com/v1/projects")
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("failed to list projects: status %d", resp.StatusCode)
	}

	var data struct {
		Projects []GCPProject `json:"projects"`
	}
	if err := json.NewDecoder(resp.Body).Decode(&data); err != nil {
		return nil, err
	}

	return data.Projects, nil
}

func (c *GCPClient) ListAPIKeys(projectID string) ([]GCPAPIKey, error) {
	url := fmt.Sprintf("https://apikeys.googleapis.com/v2/projects/%s/locations/global/keys", projectID)
	resp, err := c.client.Get(url)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("failed to list keys: status %d", resp.StatusCode)
	}

	var data struct {
		Keys []GCPAPIKey `json:"keys"`
	}
	if err := json.NewDecoder(resp.Body).Decode(&data); err != nil {
		return nil, err
	}

	return data.Keys, nil
}
