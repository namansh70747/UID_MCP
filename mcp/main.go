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
	transport := &IOTransport{
		r: bufio.NewReader(os.Stdin),
		w: os.Stdout,
	}
	err := server.Run(context.Background(), transport)
	if err != nil {
		log.Println("[ERROR]: Failed to run server:", err)
	}
}
