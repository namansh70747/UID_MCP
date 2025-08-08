package models

import "time"

type APIResponse struct {
	Success bool        `json:"success"`
	Message string      `json:"message"`
	Data    interface{} `json:"data,omitempty"`
	Error   string      `json:"error,omitempty"`
}

type PodResponse struct {
	UID          string            `json:"uid"`
	Name         string            `json:"name"`
	Namespace    string            `json:"namespace"`
	Status       string            `json:"status"`
	Image        string            `json:"image"`
	Labels       map[string]string `json:"labels"`
	CreatedAt    time.Time         `json:"created_at"`
	RestartCount int32             `json:"restart_count"`
	HostIP       string            `json:"host_ip"`
	PodIP        string            `json:"pod_ip"`
}

type ServiceResponse struct {
	UID         string `json:"uid"`
	Name        string `json:"name"`
	Namespace   string `json:"namespace"`
	ServiceType string `json:"service_type"`
	ClusterIP   string `json:"cluster_ip"`
	Port        int32  `json:"port"`
	TargetPort  int32  `json:"target_port"`
}

type ListResponse struct {
	Items []interface{} `json:"items"`
	Count int           `json:"count"`
}
