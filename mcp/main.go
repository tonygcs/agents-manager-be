package main

import (
	"bytes"
	"context"
	"encoding/json"
	"flag"
	"fmt"
	"io"
	"net/http"
	"os"
	"strings"

	"github.com/mark3labs/mcp-go/mcp"
	"github.com/mark3labs/mcp-go/server"
)

var baseURL string

func main() {
	flag.StringVar(&baseURL, "server", envOr("AGENTS_MANAGER_URL", "http://localhost:8080"), "agents manager base URL")
	flag.Parse()
	baseURL = strings.TrimRight(baseURL, "/")

	s := server.NewMCPServer("agents-manager", "1.0.0",
		server.WithToolCapabilities(true),
	)

	s.AddTool(
		mcp.NewTool("list_secrets",
			mcp.WithDescription("List the names of all configured secrets. Secret values are not returned."),
		),
		listSecrets,
	)

	s.AddTool(
		mcp.NewTool("list_worker_configs",
			mcp.WithDescription("List all worker configuration definitions, including image, description, command, labels and required secrets."),
		),
		listWorkerConfigs,
	)

	s.AddTool(
		mcp.NewTool("deploy_worker",
			mcp.WithDescription("Deploy a new worker instance from an existing worker configuration."),
			mcp.WithString("worker_name",
				mcp.Required(),
				mcp.Description("Name of the worker configuration to deploy (e.g. \"opencode-node\")."),
			),
			mcp.WithString("env",
				mcp.Description("Optional JSON object with environment variables to pass to the worker (e.g. {\"KEY\": \"value\"}). "+
					"A non-empty value for a key overrides the corresponding secret defined in the worker configuration. "+
					"To use the default secret value, include the key with an empty string (e.g. {\"MY_SECRET\": \"\"})."),
			),
		),
		deployWorker,
	)

	s.AddTool(
		mcp.NewTool("get_container_logs",
			mcp.WithDescription("Get the logs from a running or stopped container by its ID."),
			mcp.WithString("container_id",
				mcp.Required(),
				mcp.Description("The container ID to fetch logs from."),
			),
		),
		getContainerLogs(s),
	)

	if err := server.ServeStdio(s); err != nil {
		fmt.Fprintln(os.Stderr, "error:", err)
		os.Exit(1)
	}
}

func envOr(key, fallback string) string {
	if v := os.Getenv(key); v != "" {
		return v
	}
	return fallback
}

func get(path string) ([]byte, int, error) {
	resp, err := http.Get(baseURL + path)
	if err != nil {
		return nil, 0, err
	}
	defer resp.Body.Close()
	body, err := io.ReadAll(resp.Body)
	return body, resp.StatusCode, err
}

func listSecrets(_ context.Context, _ mcp.CallToolRequest) (*mcp.CallToolResult, error) {
	body, status, err := get("/config/secrets")
	if err != nil {
		return mcp.NewToolResultError("failed to reach agents manager: " + err.Error()), nil
	}
	if status != http.StatusOK {
		return mcp.NewToolResultError(fmt.Sprintf("server returned %d: %s", status, body)), nil
	}
	// The API returns map[string]string with blank values; return only the keys.
	var raw map[string]string
	if err := json.Unmarshal(body, &raw); err != nil {
		return mcp.NewToolResultError("invalid response: " + err.Error()), nil
	}
	keys := make([]string, 0, len(raw))
	for k := range raw {
		keys = append(keys, k)
	}
	out, _ := json.Marshal(keys)
	return mcp.NewToolResultText(string(out)), nil
}

func listWorkerConfigs(_ context.Context, _ mcp.CallToolRequest) (*mcp.CallToolResult, error) {
	body, status, err := get("/config/workers")
	if err != nil {
		return mcp.NewToolResultError("failed to reach agents manager: " + err.Error()), nil
	}
	if status != http.StatusOK {
		return mcp.NewToolResultError(fmt.Sprintf("server returned %d: %s", status, body)), nil
	}
	return mcp.NewToolResultText(string(body)), nil
}

func getContainerLogs(s *server.MCPServer) func(context.Context, mcp.CallToolRequest) (*mcp.CallToolResult, error) {
	return func(ctx context.Context, req mcp.CallToolRequest) (*mcp.CallToolResult, error) {
		containerID, err := req.RequireString("container_id")
		if err != nil {
			return mcp.NewToolResultError(err.Error()), nil
		}

		httpReq, err := http.NewRequestWithContext(ctx, http.MethodGet, baseURL+"/containers/"+containerID+"/logs", nil)
		if err != nil {
			return mcp.NewToolResultError("failed to build request: " + err.Error()), nil
		}
		resp, err := http.DefaultClient.Do(httpReq)
		if err != nil {
			return mcp.NewToolResultError("failed to reach agents manager: " + err.Error()), nil
		}
		defer resp.Body.Close()

		if resp.StatusCode != http.StatusOK {
			body, _ := io.ReadAll(resp.Body)
			return mcp.NewToolResultError(fmt.Sprintf("server returned %d: %s", resp.StatusCode, body)), nil
		}

		// Decode NDJSON lines as they arrive and forward each one as a
		// logging notification so the client sees logs in real time.
		type logLine struct {
			Stream string `json:"stream"`
			Line   string `json:"line"`
		}
		decoder := json.NewDecoder(resp.Body)
		count := 0
		for decoder.More() {
			var l logLine
			if err := decoder.Decode(&l); err != nil {
				break
			}
			_ = s.SendNotificationToClient(ctx, "notifications/message", map[string]any{
				"level":  "info",
				"logger": l.Stream,
				"data":   l.Line,
			})
			count++
		}
		return mcp.NewToolResultText(fmt.Sprintf("streamed %d log lines", count)), nil
	}
}

func deployWorker(_ context.Context, req mcp.CallToolRequest) (*mcp.CallToolResult, error) {
	workerName, err := req.RequireString("worker_name")
	if err != nil {
		return mcp.NewToolResultError(err.Error()), nil
	}

	var env map[string]string
	if envStr := req.GetString("env", ""); envStr != "" {
		if err := json.Unmarshal([]byte(envStr), &env); err != nil {
			return mcp.NewToolResultError("env must be a valid JSON object: " + err.Error()), nil
		}
	}

	payload := map[string]any{
		"workerName": workerName,
		"env":        env,
	}
	data, _ := json.Marshal(payload)

	resp, err := http.Post(baseURL+"/deploy", "application/json", bytes.NewReader(data))
	if err != nil {
		return mcp.NewToolResultError("failed to reach agents manager: " + err.Error()), nil
	}
	defer resp.Body.Close()
	body, _ := io.ReadAll(resp.Body)
	if resp.StatusCode != http.StatusOK {
		return mcp.NewToolResultError(fmt.Sprintf("deploy failed (%d): %s", resp.StatusCode, body)), nil
	}
	return mcp.NewToolResultText(string(body)), nil
}
