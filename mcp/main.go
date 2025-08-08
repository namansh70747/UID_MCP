package main

import (
	"bufio"
	"context"
	"encoding/json"
	"errors"
	"io"
	"log"
	"os"

	"github.com/modelcontextprotocol/go-sdk/jsonrpc"
	"github.com/modelcontextprotocol/go-sdk/mcp"
)

// HiArgs defines arguments for the greeting tool.
type HiArgs struct {
	Name string `json:"name"`
}

// CreateEntitiesArgs defines the create entities tool parameters.
type CreateEntitiesArgs struct {
	Entities []Entity `json:"entities" mcp:"entities to create"`
}

// CreateEntitiesResult returns newly created entities.
type CreateEntitiesResult struct {
	Entities []Entity `json:"entities"`
}

// CreateRelationsArgs defines the create relations tool parameters.
type CreateRelationsArgs struct {
	Relations []Relation `json:"relations" mcp:"relations to create"`
}

// CreateRelationsResult returns newly created relations.
type CreateRelationsResult struct {
	Relations []Relation `json:"relations"`
}

// AddObservationsArgs defines the add observations tool parameters.
type AddObservationsArgs struct {
	Observations []Observation `json:"observations" mcp:"observations to add"`
}

// AddObservationsResult returns newly added observations.
type AddObservationsResult struct {
	Observations []Observation `json:"observations"`
}

// DeleteEntitiesArgs defines the delete entities tool parameters.
type DeleteEntitiesArgs struct {
	EntityNames []string `json:"entityNames" mcp:"entities to delete"`
}

// DeleteObservationsArgs defines the delete observations tool parameters.
type DeleteObservationsArgs struct {
	Deletions []Observation `json:"deletions" mcp:"obeservations to delete"`
}

// DeleteRelationsArgs defines the delete relations tool parameters.
type DeleteRelationsArgs struct {
	Relations []Relation `json:"relations" mcp:"relations to delete"`
}

// SearchNodesArgs defines the search nodes tool parameters.
type SearchNodesArgs struct {
	Query string `json:"query" mcp:"query string"`
}

// OpenNodesArgs defines the open nodes tool parameters.
type OpenNodesArgs struct {
	Names []string `json:"names" mcp:"names of nodes to open"`
}

type IOTransport struct {
	r *bufio.Reader
	w io.Writer
}

func NewIOTransport(r io.Reader, w io.Writer) *IOTransport {
	return &IOTransport{
		r: bufio.NewReader(r),
		w: w,
	}
}

type ioConn struct {
	r *bufio.Reader
	w io.Writer
}

func (t *IOTransport) Connect(ctx context.Context) (mcp.Connection, error) {
	return &ioConn{
		r: t.r,
		w: t.w,
	}, nil
}

// problem with an import for decodemsg fnc so used json unmarshal
func (t *ioConn) Read(context.Context) (jsonrpc.Message, error) {
	data, err := t.r.ReadBytes('\n')
	if err != nil {
		return nil, err
	}

	var msg jsonrpc.Message
	err = json.Unmarshal(data[:len(data)-1], &msg)
	if err != nil {
		return nil, err
	}
	return msg, nil
}

// problem with an import for encodemsg fnc so used json marshal
func (t *ioConn) Write(_ context.Context, msg jsonrpc.Message) error {
	data, err := json.Marshal(msg)
	if err != nil {
		return err
	}

	_, err1 := t.w.Write(data)
	_, err2 := t.w.Write([]byte{'\n'})
	return errors.Join(err1, err2)
}

func (t *ioConn) Close() error {
	// We need to clean resources here before close
	return nil
}

// constant session id for our local setup for now
func (t *ioConn) SessionID() string {
	return "kubernetes-1"
}

func main() {
	server := mcp.NewServer(&mcp.Implementation{Name: "kubernetes-uuid"}, nil)

	// add tools for k8s here
	
	// sequential thinking
	mcp.AddTool(server, &mcp.Tool{
		Name:        "start_thinking",
		Description: "Begin a new sequential thinking session for a complex problem",
	}, func(ctx context.Context, ss *mcp.ServerSession, params *mcp.CallToolParamsFor[StartThinkingArgs]) (*mcp.CallToolResultFor[any], error) {
		return StartThinking(ctx, ss, params)
	})
	mcp.AddTool(server, &mcp.Tool{
		Name:        "continue_thinking",
		Description: "Add the next thought step, revise a previous step, or create a branch",
	}, func(ctx context.Context, ss *mcp.ServerSession, params *mcp.CallToolParamsFor[ContinueThinkingArgs]) (*mcp.CallToolResultFor[any], error) {
		return ContinueThinking(ctx, ss, params)
	})
	mcp.AddTool(server, &mcp.Tool{
		Name:        "review_thinking",
		Description: "Review the complete thinking process for a session",
	}, func(ctx context.Context, ss *mcp.ServerSession, params *mcp.CallToolParamsFor[ReviewThinkingArgs]) (*mcp.CallToolResultFor[any], error) {
		return ReviewThinking(ctx, ss, params)
	})
	server.AddResource(&mcp.Resource{
		Name:        "thinking_sessions",
		Description: "Access thinking session data and history",
		URI:         "thinking://sessions",
		MIMEType:    "application/json",
	}, func(ctx context.Context, ss *mcp.ServerSession, params *mcp.ReadResourceParams) (*mcp.ReadResourceResult, error) {
		return ThinkingHistory(ctx, ss, params)
	})

	// Memory Store 
	kb := knowledgeBase{s: &memoryStore{}}
	mcp.AddTool(server, &mcp.Tool{
		Name:        "create_entities",
		Description: "Create multiple new entities in the knowledge graph",
	}, kb.CreateEntities)
	mcp.AddTool(server, &mcp.Tool{
		Name:        "create_relations",
		Description: "Create multiple new relations between entities",
	}, kb.CreateRelations)
	mcp.AddTool(server, &mcp.Tool{
		Name:        "add_observations",
		Description: "Add new observations to existing entities",
	}, kb.AddObservations)
	mcp.AddTool(server, &mcp.Tool{
		Name:        "delete_entities",
		Description: "Remove entities and their relations",
	}, kb.DeleteEntities)
	mcp.AddTool(server, &mcp.Tool{
		Name:        "delete_observations",
		Description: "Remove specific observations from entities",
	}, kb.DeleteObservations)
	mcp.AddTool(server, &mcp.Tool{
		Name:        "delete_relations",
		Description: "Remove specific relations from the graph",
	}, kb.DeleteRelations)
	mcp.AddTool(server, &mcp.Tool{
		Name:        "read_graph",
		Description: "Read the entire knowledge graph",
	}, kb.ReadGraph)
	mcp.AddTool(server, &mcp.Tool{
		Name:        "search_nodes",
		Description: "Search for nodes based on query",
	}, kb.SearchNodes)
	mcp.AddTool(server, &mcp.Tool{
		Name:        "open_nodes",
		Description: "Retrieve specific nodes by name",
	}, kb.OpenNodes)

	transport := &IOTransport{
		r: bufio.NewReader(os.Stdin),
		w: os.Stdout,
	}
	err := server.Run(context.Background(), transport)
	if err != nil {
		log.Println("[ERROR]: Failed to run server:", err)
	}
}
