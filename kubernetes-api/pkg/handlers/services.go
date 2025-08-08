package handlers

import (
	"net/http"

	"kubernetes-api/pkg/k8s"
	"kubernetes-api/pkg/models"
	"kubernetes-api/pkg/utils"

	"github.com/gin-gonic/gin"
	corev1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/util/intstr"
)

type ServiceHandler struct {
	k8sClient *k8s.K8sClient
}

func NewServiceHandler(client *k8s.K8sClient) *ServiceHandler {
	return &ServiceHandler{k8sClient: client}
}

func (h *ServiceHandler) CreateService(c *gin.Context) {
	var req models.CreateServiceRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, models.APIResponse{
			Success: false,
			Error:   err.Error(),
		})
		return
	}

	uid := utils.GenerateUID()
	serviceName := utils.GeneratePodName(utils.SanitizeName(req.Name))

	serviceType := corev1.ServiceTypeClusterIP
	if req.ServiceType != "" {
		serviceType = corev1.ServiceType(req.ServiceType)
	}

	service := &corev1.Service{
		ObjectMeta: metav1.ObjectMeta{
			Name: serviceName,
			Labels: map[string]string{
				"uid": uid,
			},
		},
		Spec: corev1.ServiceSpec{
			Selector: map[string]string{
				"uid": req.PodUID,
			},
			Ports: []corev1.ServicePort{
				{
					Port:       req.Port,
					TargetPort: intstr.FromInt(int(req.TargetPort)),
				},
			},
			Type: serviceType,
		},
	}

	createdService, err := h.k8sClient.ClientSet.CoreV1().Services("default").Create(
		h.k8sClient.Context, service, metav1.CreateOptions{})
	if err != nil {
		c.JSON(http.StatusInternalServerError, models.APIResponse{
			Success: false,
			Error:   err.Error(),
		})
		return
	}

	response := models.ServiceResponse{
		UID:         uid,
		Name:        createdService.Name,
		Namespace:   createdService.Namespace,
		ServiceType: string(createdService.Spec.Type),
		ClusterIP:   createdService.Spec.ClusterIP,
		Port:        createdService.Spec.Ports[0].Port,
		TargetPort:  createdService.Spec.Ports[0].TargetPort.IntVal,
	}

	c.JSON(http.StatusCreated, models.APIResponse{
		Success: true,
		Message: "Service created successfully",
		Data:    response,
	})
}

func (h *ServiceHandler) ListServices(c *gin.Context) {
	services, err := h.k8sClient.ClientSet.CoreV1().Services("default").List(
		h.k8sClient.Context, metav1.ListOptions{})
	if err != nil {
		c.JSON(http.StatusInternalServerError, models.APIResponse{
			Success: false,
			Error:   err.Error(),
		})
		return
	}

	var serviceResponses []models.ServiceResponse
	for _, service := range services.Items {
		if service.Labels["uid"] != "" {
			serviceResponse := models.ServiceResponse{
				UID:         service.Labels["uid"],
				Name:        service.Name,
				Namespace:   service.Namespace,
				ServiceType: string(service.Spec.Type),
				ClusterIP:   service.Spec.ClusterIP,
			}
			if len(service.Spec.Ports) > 0 {
				serviceResponse.Port = service.Spec.Ports[0].Port
				serviceResponse.TargetPort = service.Spec.Ports[0].TargetPort.IntVal
			}
			serviceResponses = append(serviceResponses, serviceResponse)
		}
	}

	// FIX: Convert to []interface{} properly
	var items []interface{}
	for _, svc := range serviceResponses {
		items = append(items, svc)
	}

	c.JSON(http.StatusOK, models.APIResponse{
		Success: true,
		Data: models.ListResponse{
			Items: items,
			Count: len(serviceResponses),
		},
	})
}
