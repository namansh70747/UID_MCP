package main

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"time"

	"github.com/modelcontextprotocol/go-sdk/mcp"
)

// Kubernetes API client configuration
const (
	DefaultAPIBaseURL = "http://localhost:8080"
	DefaultTimeout    = 30 * time.Second
)

// Kubernetes API request/response types based on the API reference

// CreatePodRequest matches the API reference structure
type CreatePodRequest struct {
	Name          string            `json:"name"`
	Image         string            `json:"image"`
	ContainerName string            `json:"container_name"`
	Port          *int              `json:"port,omitempty"`
	Labels        map[string]string `json:"labels,omitempty"`
	Env           map[string]string `json:"env,omitempty"`
}

// CreatePodArgs for MCP tool
type CreatePodArgs struct {
	Name          string            `json:"name" mcp:"name of the pod"`
	Image         string            `json:"image" mcp:"container image to use"`
	ContainerName string            `json:"container_name" mcp:"name of the container"`
	Port          *int              `json:"port,omitempty" mcp:"port to expose (optional)"`
	Labels        map[string]string `json:"labels,omitempty" mcp:"labels to apply (optional)"`
	Env           map[string]string `json:"env,omitempty" mcp:"environment variables (optional)"`
}

// GetPodArgs for retrieving pod by UID
type GetPodArgs struct {
	UID string `json:"uid" mcp:"unique identifier of the pod"`
}

// DeletePodArgs for deleting pod by UID
type DeletePodArgs struct {
	UID string `json:"uid" mcp:"unique identifier of the pod to delete"`
}

// GetPodLogsArgs for retrieving pod logs
type GetPodLogsArgs struct {
	UID   string `json:"uid" mcp:"unique identifier of the pod"`
	Lines *int   `json:"lines,omitempty" mcp:"number of log lines to retrieve (optional)"`
}

// CreateServiceRequest matches the API reference structure
type CreateServiceRequest struct {
	Name        string `json:"name"`
	PodUID      string `json:"pod_uid"`
	Port        int    `json:"port"`
	TargetPort  int    `json:"target_port"`
	ServiceType string `json:"service_type"` // ClusterIP, NodePort, LoadBalancer
}

// CreateServiceArgs for MCP tool
type CreateServiceArgs struct {
	Name        string `json:"name" mcp:"name of the service"`
	PodUID      string `json:"pod_uid" mcp:"UID of the pod to link to"`
	Port        int    `json:"port" mcp:"service port"`
	TargetPort  int    `json:"target_port" mcp:"target port on the pod"`
	ServiceType string `json:"service_type" mcp:"service type (ClusterIP, NodePort, LoadBalancer)"`
}

// APIResponse represents the standard API response format
type APIResponse struct {
	Success bool                   `json:"success"`
	Message string                 `json:"message,omitempty"`
	Data    map[string]interface{} `json:"data,omitempty"`
	Error   string                 `json:"error,omitempty"`
}

// APIClient handles HTTP requests to the Kubernetes API
type APIClient struct {
	BaseURL    string
	HTTPClient *http.Client
}

// NewAPIClient creates a new API client
func NewAPIClient(baseURL string) *APIClient {
	if baseURL == "" {
		baseURL = DefaultAPIBaseURL
	}

	return &APIClient{
		BaseURL: baseURL,
		HTTPClient: &http.Client{
			Timeout: DefaultTimeout,
		},
	}
}

// makeRequest performs HTTP requests to the Kubernetes API
func (c *APIClient) makeRequest(method, endpoint string, payload interface{}) (*APIResponse, error) {
	url := c.BaseURL + endpoint

	var body io.Reader
	if payload != nil {
		jsonData, err := json.Marshal(payload)
		if err != nil {
			return nil, fmt.Errorf("failed to marshal request payload: %w", err)
		}
		body = bytes.NewBuffer(jsonData)
	}

	req, err := http.NewRequest(method, url, body)
	if err != nil {
		return nil, fmt.Errorf("failed to create request: %w", err)
	}

	if payload != nil {
		req.Header.Set("Content-Type", "application/json")
	}

	resp, err := c.HTTPClient.Do(req)
	if err != nil {
		return nil, fmt.Errorf("request failed: %w", err)
	}
	defer resp.Body.Close()

	respBody, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, fmt.Errorf("failed to read response: %w", err)
	}

	// For logs endpoint, return raw text
	if endpoint == "/api/v1/pods/logs" || (len(endpoint) > 20 && endpoint[len(endpoint)-5:] == "/logs") {
		return &APIResponse{
			Success: true,
			Data:    map[string]interface{}{"logs": string(respBody)},
		}, nil
	}

	var apiResp APIResponse
	if err := json.Unmarshal(respBody, &apiResp); err != nil {
		return nil, fmt.Errorf("failed to unmarshal response: %w", err)
	}

	if !apiResp.Success {
		return &apiResp, fmt.Errorf("API error: %s", apiResp.Error)
	}

	return &apiResp, nil
}

// Global API client instance
var kubeAPI = NewAPIClient("")

// MCP Tool implementations

// CreatePod creates a new pod with auto-generated UID
func CreatePod(ctx context.Context, ss *mcp.ServerSession, params *mcp.CallToolParamsFor[CreatePodArgs]) (*mcp.CallToolResultFor[interface{}], error) {
	args := params.Arguments

	req := CreatePodRequest{
		Name:          args.Name,
		Image:         args.Image,
		ContainerName: args.ContainerName,
		Labels:        args.Labels,
		Env:           args.Env,
	}

	if args.Port != nil {
		req.Port = args.Port
	}

	resp, err := kubeAPI.makeRequest("POST", "/api/v1/pods", req)
	if err != nil {
		return nil, fmt.Errorf("failed to create pod: %w", err)
	}

	return &mcp.CallToolResultFor[interface{}]{
		Content: []mcp.Content{
			&mcp.TextContent{Text: fmt.Sprintf("Pod created successfully: %s", resp.Message)},
		},
	}, nil
}

// GetPod retrieves pod details by UID
func GetPod(ctx context.Context, ss *mcp.ServerSession, params *mcp.CallToolParamsFor[GetPodArgs]) (*mcp.CallToolResultFor[interface{}], error) {
	args := params.Arguments

	resp, err := kubeAPI.makeRequest("GET", fmt.Sprintf("/api/v1/pods/%s", args.UID), nil)
	if err != nil {
		return nil, fmt.Errorf("failed to get pod: %w", err)
	}

	// Format the pod data for display
	podData, _ := json.MarshalIndent(resp.Data, "", "  ")

	return &mcp.CallToolResultFor[interface{}]{
		Content: []mcp.Content{
			&mcp.TextContent{Text: fmt.Sprintf("Pod Details:\n%s", string(podData))},
		},
	}, nil
}

// ListPods retrieves all pods managed by the API
func ListPods(ctx context.Context, ss *mcp.ServerSession, params *mcp.CallToolParamsFor[struct{}]) (*mcp.CallToolResultFor[interface{}], error) {
	resp, err := kubeAPI.makeRequest("GET", "/api/v1/pods", nil)
	if err != nil {
		return nil, fmt.Errorf("failed to list pods: %w", err)
	}

	// Format the pods list for display
	if items, ok := resp.Data["items"].([]interface{}); ok {
		result := fmt.Sprintf("Found %d pods:\n", len(items))
		for i, item := range items {
			if pod, ok := item.(map[string]interface{}); ok {
				uid, _ := pod["uid"].(string)
				name, _ := pod["name"].(string)
				status, _ := pod["status"].(string)
				result += fmt.Sprintf("%d. UID: %s, Name: %s, Status: %s\n", i+1, uid, name, status)
			}
		}

		return &mcp.CallToolResultFor[interface{}]{
			Content: []mcp.Content{
				&mcp.TextContent{Text: result},
			},
		}, nil
	}

	return &mcp.CallToolResultFor[interface{}]{
		Content: []mcp.Content{
			&mcp.TextContent{Text: "No pods found"},
		},
	}, nil
}

// DeletePod removes a pod by UID
func DeletePod(ctx context.Context, ss *mcp.ServerSession, params *mcp.CallToolParamsFor[DeletePodArgs]) (*mcp.CallToolResultFor[interface{}], error) {
	args := params.Arguments

	resp, err := kubeAPI.makeRequest("DELETE", fmt.Sprintf("/api/v1/pods/%s", args.UID), nil)
	if err != nil {
		return nil, fmt.Errorf("failed to delete pod: %w", err)
	}

	return &mcp.CallToolResultFor[interface{}]{
		Content: []mcp.Content{
			&mcp.TextContent{Text: fmt.Sprintf("Pod deleted successfully: %s", resp.Message)},
		},
	}, nil
}

// GetPodLogs retrieves logs from a specific pod
func GetPodLogs(ctx context.Context, ss *mcp.ServerSession, params *mcp.CallToolParamsFor[GetPodLogsArgs]) (*mcp.CallToolResultFor[interface{}], error) {
	args := params.Arguments

	endpoint := fmt.Sprintf("/api/v1/pods/%s/logs", args.UID)
	if args.Lines != nil {
		endpoint += fmt.Sprintf("?lines=%d", *args.Lines)
	}

	resp, err := kubeAPI.makeRequest("GET", endpoint, nil)
	if err != nil {
		return nil, fmt.Errorf("failed to get pod logs: %w", err)
	}

	logs, _ := resp.Data["logs"].(string)

	return &mcp.CallToolResultFor[interface{}]{
		Content: []mcp.Content{
			&mcp.TextContent{Text: fmt.Sprintf("Pod Logs for %s:\n%s", args.UID, logs)},
		},
	}, nil
}

// CreateService creates a service linked to a pod
func CreateService(ctx context.Context, ss *mcp.ServerSession, params *mcp.CallToolParamsFor[CreateServiceArgs]) (*mcp.CallToolResultFor[interface{}], error) {
	args := params.Arguments

	req := CreateServiceRequest{
		Name:        args.Name,
		PodUID:      args.PodUID,
		Port:        args.Port,
		TargetPort:  args.TargetPort,
		ServiceType: args.ServiceType,
	}

	resp, err := kubeAPI.makeRequest("POST", "/api/v1/services", req)
	if err != nil {
		return nil, fmt.Errorf("failed to create service: %w", err)
	}

	return &mcp.CallToolResultFor[interface{}]{
		Content: []mcp.Content{
			&mcp.TextContent{Text: fmt.Sprintf("Service created successfully: %s", resp.Message)},
		},
	}, nil
}

// ListServices retrieves all services managed by the API
func ListServices(ctx context.Context, ss *mcp.ServerSession, params *mcp.CallToolParamsFor[struct{}]) (*mcp.CallToolResultFor[interface{}], error) {
	resp, err := kubeAPI.makeRequest("GET", "/api/v1/services", nil)
	if err != nil {
		return nil, fmt.Errorf("failed to list services: %w", err)
	}

	// Format the services list for display
	if items, ok := resp.Data["items"].([]interface{}); ok {
		result := fmt.Sprintf("Found %d services:\n", len(items))
		for i, item := range items {
			if svc, ok := item.(map[string]interface{}); ok {
				uid, _ := svc["uid"].(string)
				name, _ := svc["name"].(string)
				result += fmt.Sprintf("%d. UID: %s, Name: %s\n", i+1, uid, name)
			}
		}

		return &mcp.CallToolResultFor[interface{}]{
			Content: []mcp.Content{
				&mcp.TextContent{Text: result},
			},
		}, nil
	}

	return &mcp.CallToolResultFor[interface{}]{
		Content: []mcp.Content{
			&mcp.TextContent{Text: "No services found"},
		},
	}, nil
}

// GetClusterInfo retrieves cluster status and node information
func GetClusterInfo(ctx context.Context, ss *mcp.ServerSession, params *mcp.CallToolParamsFor[struct{}]) (*mcp.CallToolResultFor[interface{}], error) {
	resp, err := kubeAPI.makeRequest("GET", "/api/v1/cluster/info", nil)
	if err != nil {
		return nil, fmt.Errorf("failed to get cluster info: %w", err)
	}

	// Format cluster info for display
	clusterData, _ := json.MarshalIndent(resp.Data, "", "  ")

	return &mcp.CallToolResultFor[interface{}]{
		Content: []mcp.Content{
			&mcp.TextContent{Text: fmt.Sprintf("Cluster Information:\n%s", string(clusterData))},
		},
	}, nil
}

// HealthCheck verifies API availability
func HealthCheck(ctx context.Context, ss *mcp.ServerSession, params *mcp.CallToolParamsFor[struct{}]) (*mcp.CallToolResultFor[interface{}], error) {
	resp, err := kubeAPI.makeRequest("GET", "/health", nil)
	if err != nil {
		return nil, fmt.Errorf("health check failed: %w", err)
	}

	return &mcp.CallToolResultFor[interface{}]{
		Content: []mcp.Content{
			&mcp.TextContent{Text: fmt.Sprintf("Health Status: %s", resp.Message)},
		},
	}, nil
}
