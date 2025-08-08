# Complete Kubernetes Management API Reference

## 🌟 Overview

This API provides complete Kubernetes pod and service management with unique identifier tracking. All operations return consistent JSON responses and integrate directly with your Kubernetes cluster.

**Base URL:** `http://localhost:8080`  
**API Version:** `v1`  
**Content-Type:** `application/json`

## 📊 Response Format

All endpoints return responses in this format:

```json
{
  "success": true/false,
  "message": "Human readable message",
  "data": { ... },           // Present on success
  "error": "Error message"   // Present on failure
}
```

## 🚀 Core Endpoints

### 1. Health Check

**Endpoint:** `GET /health`  
**Purpose:** Verify API availability

**Response:**

```json
{
  "success": true,
  "message": "API is healthy"
}
```

---

### 2. Create Pod

**Endpoint:** `POST /api/v1/pods`  
**Purpose:** Create a new pod with auto-generated UID

**Request Body:**

```json
{
  "name": "my-app",
  "image": "nginx:latest",
  "container_name": "nginx",
  "port": 80,                    // Optional
  "labels": {                    // Optional
    "environment": "production"
  },
  "env": {                      // Optional
    "DATABASE_URL": "mysql://..."
  }
}
```

**Response:**

```json
{
  "success": true,
  "message": "Pod created successfully",
  "data": {
    "uid": "a495eff8",
    "name": "my-app-a495eff8",
    "namespace": "default",
    "status": "Pending",
    "image": "nginx:latest",
    "labels": {
      "app": "my-app",
      "uid": "a495eff8",
      "environment": "production"
    },
    "created_at": "2025-08-08T16:30:00Z"
  }
}
```

---

### 3. Get Pod by UID

**Endpoint:** `GET /api/v1/pods/{uid}`  
**Purpose:** Retrieve pod details and status

**Response:**

```json
{
  "success": true,
  "data": {
    "uid": "a495eff8",
    "name": "my-app-a495eff8",
    "namespace": "default",
    "status": "Running",
    "labels": {...},
    "created_at": "2025-08-08T16:30:00Z",
    "restart_count": 0,
    "host_ip": "192.168.1.10",
    "pod_ip": "10.244.0.5"
  }
}
```

**Pod Status Values:**

- `Pending` - Pod is being created
- `Running` - Pod is active and running
- `Succeeded` - Pod completed successfully
- `Failed` - Pod failed to run
- `Unknown` - Status cannot be determined

---

### 4. List All Pods

**Endpoint:** `GET /api/v1/pods`  
**Purpose:** Get all pods managed by this API

**Response:**

```json
{
  "success": true,
  "data": {
    "items": [
      {
        "uid": "a495eff8",
        "name": "my-app-a495eff8",
        "namespace": "default",
        "status": "Running"
      }
    ],
    "count": 1
  }
}
```

---

### 5. Get Pod Logs

**Endpoint:** `GET /api/v1/pods/{uid}/logs?lines=100`  
**Purpose:** Stream pod logs

**Query Parameters:**

- `lines` (optional): Number of log lines to retrieve (default: 100)

**Response:** Plain text logs

```text
2025/08/08 16:35:12 [notice] 1#1: nginx/1.25.2
2025/08/08 16:35:12 [notice] 1#1: start worker processes
```

---

### 6. Delete Pod

**Endpoint:** `DELETE /api/v1/pods/{uid}`  
**Purpose:** Remove pod from cluster

**Response:**

```json
{
  "success": true,
  "message": "Pod deleted successfully"
}
```

---

### 7. Create Service

**Endpoint:** `POST /api/v1/services`  
**Purpose:** Create service linked to a pod

**Request Body:**

```json
{
  "name": "my-service",
  "pod_uid": "a495eff8",
  "port": 80,
  "target_port": 80,
  "service_type": "ClusterIP"    // ClusterIP, NodePort, LoadBalancer
}
```

**Response:**

```json
{
  "success": true,
  "message": "Service created successfully",
  "data": {
    "uid": "s1e2r3v4",
    "name": "my-service-s1e2r3v4",
    "namespace": "default",
    "service_type": "ClusterIP",
    "cluster_ip": "10.96.150.123",
    "port": 80,
    "target_port": 80
  }
}
```

---

### 8. List Services

**Endpoint:** `GET /api/v1/services`  
**Purpose:** Get all services managed by this API

**Response:**

```json
{
  "success": true,
  "data": {
    "items": [
      {
        "uid": "s1e2r3v4",
        "name": "my-service-s1e2r3v4"
      }
    ],
    "count": 1
  }
}
```

---

### 9. Cluster Information

**Endpoint:** `GET /api/v1/cluster/info`  
**Purpose:** Get cluster status and node information

**Response:**

```json
{
  "success": true,
  "data": {
    "node_count": 3,
    "nodes": [
      "example-control-plane",
      "example-worker",
      "example-worker2"
    ]
  }
}
```

## 🔧 Integration Examples

### Python Integration

```python
import requests
import time

class KubernetesAPI:
    def __init__(self, base_url="http://localhost:8080"):
        self.base_url = base_url
        self.api_url = f"{base_url}/api/v1"
    
    def create_pod(self, name, image, container_name, port=None):
        payload = {
            "name": name,
            "image": image,
            "container_name": container_name
        }
        if port:
            payload["port"] = port
            
        response = requests.post(f"{self.api_url}/pods", json=payload)
        return response.json()
    
    def get_pod_status(self, uid):
        response = requests.get(f"{self.api_url}/pods/{uid}")
        return response.json()
    
    def wait_for_pod_running(self, uid, timeout=120):
        start_time = time.time()
        while time.time() - start_time < timeout:
            status = self.get_pod_status(uid)
            if status["success"] and status["data"]["status"] == "Running":
                return True
            time.sleep(5)
        return False
    
    def delete_pod(self, uid):
        response = requests.delete(f"{self.api_url}/pods/{uid}")
        return response.json()

# Usage Example
api = KubernetesAPI()

# Create pod
result = api.create_pod("web-app", "nginx:latest", "nginx", 80)
pod_uid = result["data"]["uid"]

# Wait for pod to be ready
if api.wait_for_pod_running(pod_uid):
    print(f"Pod {pod_uid} is running!")
    
    # Get status
    status = api.get_pod_status(pod_uid)
    print(f"Pod IP: {status['data']['pod_ip']}")
    
    # Clean up
    api.delete_pod(pod_uid)
```

### Node.js Integration

```javascript
const axios = require('axios');

class KubernetesAPI {
    constructor(baseUrl = 'http://localhost:8080') {
        this.baseUrl = baseUrl;
        this.apiUrl = `${baseUrl}/api/v1`;
    }
    
    async createPod(name, image, containerName, port = null) {
        const payload = { name, image, container_name: containerName };
        if (port) payload.port = port;
        
        const response = await axios.post(`${this.apiUrl}/pods`, payload);
        return response.data;
    }
    
    async getPodStatus(uid) {
        const response = await axios.get(`${this.apiUrl}/pods/${uid}`);
        return response.data;
    }
    
    async deletePod(uid) {
        const response = await axios.delete(`${this.apiUrl}/pods/${uid}`);
        return response.data;
    }
}

// Usage
const api = new KubernetesAPI();

async function deployApp() {
    try {
        const result = await api.createPod('my-app', 'nginx:latest', 'nginx', 80);
        const podUid = result.data.uid;
        console.log(`Pod created with UID: ${podUid}`);
        
        // Check status
        const status = await api.getPodStatus(podUid);
        console.log(`Pod status: ${status.data.status}`);
        
    } catch (error) {
        console.error('Error:', error.response.data);
    }
}

deployApp();
```

## 🚨 Error Handling

### Common Error Responses

#### 404 - Resource Not Found

```json
{
  "success": false,
  "error": "Pod not found"
}
```

#### 400 - Bad Request

```json
{
  "success": false,
  "error": "Key: 'CreatePodRequest.Name' Error:Field validation for 'Name' failed on the 'required' tag"
}
```

#### 500 - Internal Server Error

```json
{
  "success": false,
  "error": "Failed to create pod: pods \"test-pod\" already exists"
}
```

## 📈 Best Practices for MCP Integration

1. **Always check `success` field** in responses
2. **Store UIDs** for future operations on resources
3. **Implement retry logic** for pod status checks
4. **Handle timeouts** when waiting for pod readiness
5. **Clean up resources** when done (delete pods/services)
6. **Monitor pod status** before performing operations
7. **Use meaningful names** for easier debugging

## 🔍 Testing Your Integration

Use this curl command to test connectivity:

```bash
curl -s http://localhost:8080/health | jq '.'
```

Expected response:

```json
{
  "success": true,
  "message": "API is healthy"
}
```

## 📞 Support

If you encounter issues:

1. Check API server logs
2. Verify Kubernetes cluster is running (`kubectl get nodes`)
3. Test individual endpoints with curl
4. Check pod status in Kubernetes (`kubectl get pods`)

---

**API Version:** 1.0  
**Last Updated:** August 8, 2025  
**Cluster Compatibility:** Kubernetes 1.27+
