package handlers

import (
	"fmt"
	"io"
	"net/http"
	"strconv"

	"kubernetes-api/pkg/k8s"
	"kubernetes-api/pkg/models"
	"kubernetes-api/pkg/utils"

	"github.com/gin-gonic/gin"
	corev1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

type PodHandler struct {
	k8sClient *k8s.K8sClient
}

func NewPodHandler(client *k8s.K8sClient) *PodHandler {
	return &PodHandler{k8sClient: client}
}

func (h *PodHandler) CreatePod(c *gin.Context) {
	var req models.CreatePodRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, models.APIResponse{
			Success: false,
			Error:   err.Error(),
		})
		return
	}

	// Generate unique identifiers
	uid := utils.GenerateUID()
	podName := utils.GeneratePodName(utils.SanitizeName(req.Name))

	// Prepare labels
	labels := map[string]string{
		"app": req.Name,
		"uid": uid,
	}
	for k, v := range req.Labels {
		labels[k] = v
	}

	// Prepare environment variables
	envVars := []corev1.EnvVar{
		{Name: "POD_UID", Value: uid},
	}
	for k, v := range req.Env {
		envVars = append(envVars, corev1.EnvVar{Name: k, Value: v})
	}

	// Create pod specification
	pod := &corev1.Pod{
		ObjectMeta: metav1.ObjectMeta{
			Name:   podName,
			Labels: labels,
		},
		Spec: corev1.PodSpec{
			Containers: []corev1.Container{
				{
					Name:  req.ContainerName,
					Image: req.Image,
					Env:   envVars,
				},
			},
		},
	}

	// Add port if specified
	if req.Port > 0 {
		pod.Spec.Containers[0].Ports = []corev1.ContainerPort{
			{ContainerPort: req.Port},
		}
	}

	// Create pod in cluster
	createdPod, err := h.k8sClient.ClientSet.CoreV1().Pods("default").Create(
		h.k8sClient.Context, pod, metav1.CreateOptions{})
	if err != nil {
		c.JSON(http.StatusInternalServerError, models.APIResponse{
			Success: false,
			Error:   err.Error(),
		})
		return
	}

	response := models.PodResponse{
		UID:       uid,
		Name:      createdPod.Name,
		Namespace: createdPod.Namespace,
		Status:    string(createdPod.Status.Phase),
		Image:     req.Image,
		Labels:    createdPod.Labels,
		CreatedAt: createdPod.CreationTimestamp.Time,
	}

	c.JSON(http.StatusCreated, models.APIResponse{
		Success: true,
		Message: "Pod created successfully",
		Data:    response,
	})
}

func (h *PodHandler) GetPodByUID(c *gin.Context) {
	uid := c.Param("uid")

	pods, err := h.k8sClient.ClientSet.CoreV1().Pods("default").List(
		h.k8sClient.Context, metav1.ListOptions{
			LabelSelector: "uid=" + uid,
		})
	if err != nil {
		c.JSON(http.StatusInternalServerError, models.APIResponse{
			Success: false,
			Error:   err.Error(),
		})
		return
	}

	if len(pods.Items) == 0 {
		c.JSON(http.StatusNotFound, models.APIResponse{
			Success: false,
			Error:   "Pod not found",
		})
		return
	}

	pod := pods.Items[0]
	response := models.PodResponse{
		UID:       uid,
		Name:      pod.Name,
		Namespace: pod.Namespace,
		Status:    string(pod.Status.Phase),
		Labels:    pod.Labels,
		CreatedAt: pod.CreationTimestamp.Time,
		HostIP:    pod.Status.HostIP,
		PodIP:     pod.Status.PodIP,
	}

	// Add safety check for container statuses
	if len(pod.Status.ContainerStatuses) > 0 {
		response.RestartCount = pod.Status.ContainerStatuses[0].RestartCount
	} else {
		response.RestartCount = 0
	}

	c.JSON(http.StatusOK, models.APIResponse{
		Success: true,
		Data:    response,
	})
}

func (h *PodHandler) ListPods(c *gin.Context) {
	pods, err := h.k8sClient.ClientSet.CoreV1().Pods("default").List(
		h.k8sClient.Context, metav1.ListOptions{})
	if err != nil {
		c.JSON(http.StatusInternalServerError, models.APIResponse{
			Success: false,
			Error:   err.Error(),
		})
		return
	}

	var podResponses []models.PodResponse
	for _, pod := range pods.Items {
		podResponse := models.PodResponse{
			UID:       pod.Labels["uid"],
			Name:      pod.Name,
			Namespace: pod.Namespace,
			Status:    string(pod.Status.Phase),
			Labels:    pod.Labels,
			CreatedAt: pod.CreationTimestamp.Time,
			HostIP:    pod.Status.HostIP,
			PodIP:     pod.Status.PodIP,
		}
		if len(pod.Status.ContainerStatuses) > 0 {
			podResponse.RestartCount = pod.Status.ContainerStatuses[0].RestartCount
		}
		podResponses = append(podResponses, podResponse)
	}

	// Convert to []interface{} properly
	var items []interface{}
	for _, pod := range podResponses {
		items = append(items, pod)
	}

	c.JSON(http.StatusOK, models.APIResponse{
		Success: true,
		Data: models.ListResponse{
			Items: items,
			Count: len(podResponses),
		},
	})
}

func (h *PodHandler) DeletePodByUID(c *gin.Context) {
	uid := c.Param("uid")

	pods, err := h.k8sClient.ClientSet.CoreV1().Pods("default").List(
		h.k8sClient.Context, metav1.ListOptions{
			LabelSelector: "uid=" + uid,
		})
	if err != nil {
		c.JSON(http.StatusInternalServerError, models.APIResponse{
			Success: false,
			Error:   err.Error(),
		})
		return
	}

	if len(pods.Items) == 0 {
		c.JSON(http.StatusNotFound, models.APIResponse{
			Success: false,
			Error:   "Pod not found",
		})
		return
	}

	pod := pods.Items[0]
	err = h.k8sClient.ClientSet.CoreV1().Pods("default").Delete(
		h.k8sClient.Context, pod.Name, metav1.DeleteOptions{})
	if err != nil {
		c.JSON(http.StatusInternalServerError, models.APIResponse{
			Success: false,
			Error:   err.Error(),
		})
		return
	}

	c.JSON(http.StatusOK, models.APIResponse{
		Success: true,
		Message: "Pod deleted successfully",
	})
}

func (h *PodHandler) GetPodLogs(c *gin.Context) {
	uid := c.Param("uid")
	lines := c.DefaultQuery("lines", "100")

	lineCount, _ := strconv.ParseInt(lines, 10, 64)

	pods, err := h.k8sClient.ClientSet.CoreV1().Pods("default").List(
		h.k8sClient.Context, metav1.ListOptions{
			LabelSelector: "uid=" + uid,
		})
	if err != nil {
		c.JSON(http.StatusInternalServerError, models.APIResponse{
			Success: false,
			Error:   err.Error(),
		})
		return
	}

	if len(pods.Items) == 0 {
		c.JSON(http.StatusNotFound, models.APIResponse{
			Success: false,
			Error:   "Pod not found",
		})
		return
	}

	pod := pods.Items[0]

	// Check if pod is running
	if pod.Status.Phase != corev1.PodRunning {
		c.JSON(http.StatusBadRequest, models.APIResponse{
			Success: false,
			Error:   fmt.Sprintf("Pod is not running (status: %s)", pod.Status.Phase),
		})
		return
	}

	podLogOpts := corev1.PodLogOptions{
		TailLines: &lineCount,
	}

	req := h.k8sClient.ClientSet.CoreV1().Pods(pod.Namespace).GetLogs(pod.Name, &podLogOpts)
	logs, err := req.Stream(h.k8sClient.Context)
	if err != nil {
		c.JSON(http.StatusInternalServerError, models.APIResponse{
			Success: false,
			Error:   fmt.Sprintf("Failed to get logs: %v", err),
		})
		return
	}
	defer logs.Close()

	// Read logs into buffer first to check if empty
	logBytes, err := io.ReadAll(logs)
	if err != nil {
		c.JSON(http.StatusInternalServerError, models.APIResponse{
			Success: false,
			Error:   fmt.Sprintf("Failed to read logs: %v", err),
		})
		return
	}

	c.Header("Content-Type", "text/plain")
	c.Status(http.StatusOK)
	c.Writer.Write(logBytes)
}
