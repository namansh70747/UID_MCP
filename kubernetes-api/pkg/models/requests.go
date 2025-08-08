package models

type CreatePodRequest struct {
	Name          string            `json:"name"`
	Image         string            `json:"image"`
	ContainerName string            `json:"container_name"`
	Port          int32             `json:"port,omitempty"`
	Labels        map[string]string `json:"labels,omitempty"`
	Env           map[string]string `json:"env,omitempty"`
}

type CreateServiceRequest struct {
	Name        string `json:"name"`
	PodUID      string `json:"pod_uid"`
	Port        int32  `json:"port"`
	TargetPort  int32  `json:"target_port"`
	ServiceType string `json:"service_type,omitempty"`
}

type CreateDeploymentRequest struct {
	Name          string            `json:"name"`
	Image         string            `json:"image"`
	ContainerName string            `json:"container_name"`
	Replicas      int32             `json:"replicas"`
	Port          int32             `json:"port,omitempty"`
	Labels        map[string]string `json:"labels,omitempty"`
}

type PodOperationRequest struct {
	UID       string `json:"uid"`
	Operation string `json:"operation"` // start, stop, restart, delete
}
