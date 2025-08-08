package main

import (
	"log"
	"net/http"

	"kubernetes-api/pkg/handlers"
	"kubernetes-api/pkg/k8s"
	"kubernetes-api/pkg/models"

	"github.com/gin-gonic/gin"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

func main() {
	// Initialize Kubernetes client
	k8sClient, err := k8s.NewK8sClient()
	if err != nil {
		log.Fatalf("Failed to create Kubernetes client: %v", err)
	}

	// Initialize handlers
	podHandler := handlers.NewPodHandler(k8sClient)
	serviceHandler := handlers.NewServiceHandler(k8sClient)

	// Setup Gin router
	r := gin.Default()

	// CORS middleware
	r.Use(func(c *gin.Context) {
		c.Header("Access-Control-Allow-Origin", "*")
		c.Header("Access-Control-Allow-Methods", "GET, POST, PUT, DELETE, OPTIONS")
		c.Header("Access-Control-Allow-Headers", "Content-Type, Authorization")

		if c.Request.Method == "OPTIONS" {
			c.AbortWithStatus(204)
			return
		}

		c.Next()
	})

	// Health check endpoint
	r.GET("/health", func(c *gin.Context) {
		c.JSON(http.StatusOK, models.APIResponse{
			Success: true,
			Message: "API is healthy",
		})
	})

	// API versioning
	v1 := r.Group("/api/v1")
	{
		// Pod endpoints - Remove the group and add routes directly
		v1.POST("/pods", podHandler.CreatePod)
		v1.GET("/pods", podHandler.ListPods)
		v1.GET("/pods/:uid", podHandler.GetPodByUID)
		v1.DELETE("/pods/:uid", podHandler.DeletePodByUID)
		v1.GET("/pods/:uid/logs", podHandler.GetPodLogs)

		// Service endpoints - Remove the group and add routes directly
		v1.POST("/services", serviceHandler.CreateService)
		v1.GET("/services", serviceHandler.ListServices)

		// Cluster info endpoint
		v1.GET("/cluster/info", func(c *gin.Context) {
			nodes, err := k8sClient.ClientSet.CoreV1().Nodes().List(
				k8sClient.Context, metav1.ListOptions{})
			if err != nil {
				c.JSON(http.StatusInternalServerError, models.APIResponse{
					Success: false,
					Error:   err.Error(),
				})
				return
			}

			clusterInfo := map[string]interface{}{
				"node_count": len(nodes.Items),
				"nodes":      []string{},
			}

			for _, node := range nodes.Items {
				clusterInfo["nodes"] = append(clusterInfo["nodes"].([]string), node.Name)
			}

			c.JSON(http.StatusOK, models.APIResponse{
				Success: true,
				Data:    clusterInfo,
			})
		})
	}

	log.Println("Starting Kubernetes API server on :8080")
	log.Fatal(r.Run(":8080"))
}
