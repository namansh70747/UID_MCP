# MCP Server (Kubernetes)

The MCP server provides various basic tools to interact with a simple Kubernetes cluster

To Quickly Test it -
1. Start Docker Desktop
2. ```bash
   cd cluster
   ```
3. Start the cluster
   ```bash
   kind create cluster --name example --config example-cluster.yml
   ```
4. Start a simple Kubernetes client to interact with your cluster
   ```bash
   cd ..
   cd kubernetes-api
   go run main.go
   ```
5.  Build the MCP server binary
   ```bash
   go build .
   ```
6. Reference binary location to Claude desktop in `claude_desktop_config.json`

```json
   {
  "mcpServers": {
    "k8s-mcp": {
      "command": "C:\\mcp\\k8s mcp\\mcp\\mcp_server.exe"
    }
  }
}
 ```

---

`Note` - The MCP server is written by [Vaidik](https://github.com/vaidikcode) and the kuberenetes api to interact with cluster using uuid is written by Naman
