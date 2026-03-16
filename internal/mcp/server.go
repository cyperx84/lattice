package mcp

import (
	"bufio"
	"encoding/json"
	"fmt"
	"io"
	"strings"

	"github.com/cyperx84/lattice/internal/index"
	"github.com/cyperx84/lattice/internal/modelfile"
)

// JSON-RPC 2.0 types

type request struct {
	JSONRPC string          `json:"jsonrpc"`
	ID      json.RawMessage `json:"id,omitempty"`
	Method  string          `json:"method"`
	Params  json.RawMessage `json:"params,omitempty"`
}

type response struct {
	JSONRPC string      `json:"jsonrpc"`
	ID      json.RawMessage `json:"id"`
	Result  interface{} `json:"result,omitempty"`
	Error   *rpcError   `json:"error,omitempty"`
}

type rpcError struct {
	Code    int    `json:"code"`
	Message string `json:"message"`
}

// MCP types

type serverInfo struct {
	ProtocolVersion string       `json:"protocolVersion"`
	Capabilities    capabilities `json:"capabilities"`
	ServerInfo      nameVersion  `json:"serverInfo"`
}

type capabilities struct {
	Tools toolsCap `json:"tools"`
}

type toolsCap struct {
	ListChanged bool `json:"listChanged"`
}

type nameVersion struct {
	Name    string `json:"name"`
	Version string `json:"version"`
}

type toolDef struct {
	Name        string      `json:"name"`
	Description string      `json:"description"`
	InputSchema inputSchema `json:"inputSchema"`
}

type inputSchema struct {
	Type       string              `json:"type"`
	Properties map[string]property `json:"properties"`
	Required   []string            `json:"required,omitempty"`
}

type property struct {
	Type        string `json:"type"`
	Description string `json:"description"`
	Default     *int   `json:"default,omitempty"`
}

type toolsListResult struct {
	Tools []toolDef `json:"tools"`
}

type toolCallParams struct {
	Name      string          `json:"name"`
	Arguments json.RawMessage `json:"arguments"`
}

type toolContent struct {
	Type string `json:"type"`
	Text string `json:"text"`
}

type toolResult struct {
	Content []toolContent `json:"content"`
	IsError bool          `json:"isError,omitempty"`
}

// Server is the MCP server.
type Server struct {
	idx        *index.ModelIndex
	modelFiles map[string]string
	verbose    bool
	logWriter  io.Writer
}

// NewServer creates a new MCP server.
func NewServer(idx *index.ModelIndex, modelFiles map[string]string, verbose bool, logWriter io.Writer) *Server {
	return &Server{
		idx:        idx,
		modelFiles: modelFiles,
		verbose:    verbose,
		logWriter:  logWriter,
	}
}

func (s *Server) log(format string, args ...interface{}) {
	if s.verbose && s.logWriter != nil {
		fmt.Fprintf(s.logWriter, "[mcp] "+format+"\n", args...)
	}
}

// Run reads JSON-RPC 2.0 messages from reader and writes responses to writer.
func (s *Server) Run(reader io.Reader, writer io.Writer) error {
	scanner := bufio.NewScanner(reader)
	scanner.Buffer(make([]byte, 1024*1024), 1024*1024) // 1MB buffer
	enc := json.NewEncoder(writer)

	for scanner.Scan() {
		line := scanner.Text()
		if strings.TrimSpace(line) == "" {
			continue
		}

		var req request
		if err := json.Unmarshal([]byte(line), &req); err != nil {
			s.log("invalid JSON: %v", err)
			continue
		}

		s.log("method=%s", req.Method)

		resp := s.handleRequest(req)
		if resp == nil {
			// Notification — no response
			continue
		}

		if err := enc.Encode(resp); err != nil {
			s.log("write error: %v", err)
			return err
		}
	}

	return scanner.Err()
}

func (s *Server) handleRequest(req request) *response {
	switch req.Method {
	case "initialize":
		return s.handleInitialize(req)
	case "notifications/initialized":
		return nil // notification, no response
	case "tools/list":
		return s.handleToolsList(req)
	case "tools/call":
		return s.handleToolsCall(req)
	default:
		s.log("unknown method: %s", req.Method)
		return &response{
			JSONRPC: "2.0",
			ID:      req.ID,
			Error:   &rpcError{Code: -32601, Message: "method not found: " + req.Method},
		}
	}
}

func (s *Server) handleInitialize(req request) *response {
	return &response{
		JSONRPC: "2.0",
		ID:      req.ID,
		Result: serverInfo{
			ProtocolVersion: "2024-11-05",
			Capabilities:    capabilities{Tools: toolsCap{ListChanged: false}},
			ServerInfo:      nameVersion{Name: "lattice", Version: "0.1.0"},
		},
	}
}

func (s *Server) handleToolsList(req request) *response {
	defaultThinkCount := 3
	defaultSuggestCount := 5

	tools := []toolDef{
		{
			Name:        "think",
			Description: "Surface and apply the most relevant mental models to a problem. Returns structured analysis with thinking steps and synthesis.",
			InputSchema: inputSchema{
				Type: "object",
				Properties: map[string]property{
					"problem": {Type: "string", Description: "The problem or decision to analyze"},
					"models":  {Type: "string", Description: "Optional comma-separated model slugs to force (e.g. 'inversion,second-order_thinking')"},
					"count":   {Type: "integer", Description: "Number of models to apply (default 3)", Default: &defaultThinkCount},
				},
				Required: []string{"problem"},
			},
		},
		{
			Name:        "suggest",
			Description: "Recommend which mental models to use for a given situation, without applying them.",
			InputSchema: inputSchema{
				Type: "object",
				Properties: map[string]property{
					"situation": {Type: "string", Description: "Describe your situation or decision"},
					"count":     {Type: "integer", Description: "Number of suggestions (default 5)", Default: &defaultSuggestCount},
				},
				Required: []string{"situation"},
			},
		},
		{
			Name:        "search",
			Description: "Search the mental models index by keyword.",
			InputSchema: inputSchema{
				Type: "object",
				Properties: map[string]property{
					"query": {Type: "string", Description: "Search keyword"},
				},
				Required: []string{"query"},
			},
		},
		{
			Name:        "apply",
			Description: "Apply a specific mental model's thinking steps to a context.",
			InputSchema: inputSchema{
				Type: "object",
				Properties: map[string]property{
					"model":   {Type: "string", Description: "Model slug (e.g. 'inversion', 'first-principle_thinking')"},
					"context": {Type: "string", Description: "The context or problem to apply the model to"},
				},
				Required: []string{"model", "context"},
			},
		},
		{
			Name:        "list",
			Description: "List all available mental models, optionally filtered by category.",
			InputSchema: inputSchema{
				Type: "object",
				Properties: map[string]property{
					"category": {Type: "string", Description: "Filter by category name"},
				},
			},
		},
	}

	return &response{
		JSONRPC: "2.0",
		ID:      req.ID,
		Result:  toolsListResult{Tools: tools},
	}
}

func (s *Server) handleToolsCall(req request) *response {
	var params toolCallParams
	if err := json.Unmarshal(req.Params, &params); err != nil {
		return &response{
			JSONRPC: "2.0",
			ID:      req.ID,
			Result:  toolResult{Content: []toolContent{{Type: "text", Text: "Error: invalid tool call params"}}, IsError: true},
		}
	}

	s.log("tool=%s", params.Name)

	var result string
	var isError bool

	switch params.Name {
	case "think":
		result, isError = s.toolThink(params.Arguments)
	case "suggest":
		result, isError = s.toolSuggest(params.Arguments)
	case "search":
		result, isError = s.toolSearch(params.Arguments)
	case "apply":
		result, isError = s.toolApply(params.Arguments)
	case "list":
		result, isError = s.toolList(params.Arguments)
	default:
		result = "Error: unknown tool: " + params.Name
		isError = true
	}

	return &response{
		JSONRPC: "2.0",
		ID:      req.ID,
		Result:  toolResult{Content: []toolContent{{Type: "text", Text: result}}, IsError: isError},
	}
}

// Tool implementations

func (s *Server) toolThink(args json.RawMessage) (string, bool) {
	var p struct {
		Problem string `json:"problem"`
		Models  string `json:"models"`
		Count   int    `json:"count"`
	}
	if err := json.Unmarshal(args, &p); err != nil {
		return "Error: " + err.Error(), true
	}
	if p.Problem == "" {
		return "Error: problem is required", true
	}
	if p.Count <= 0 {
		p.Count = 3
	}

	var entries []index.ModelEntry
	if p.Models != "" {
		for _, slug := range strings.Split(p.Models, ",") {
			slug = strings.TrimSpace(slug)
			entry := s.idx.FindBySlug(slug)
			if entry == nil {
				entry = s.idx.FindByID(slug)
			}
			if entry != nil {
				entries = append(entries, *entry)
			}
		}
	} else {
		entries = s.idx.TopNForQuery(p.Problem, p.Count)
	}

	if len(entries) == 0 {
		return "No relevant models found for: " + p.Problem, false
	}

	var b strings.Builder
	b.WriteString(fmt.Sprintf("# Thinking: %s\n\n", p.Problem))
	b.WriteString(fmt.Sprintf("Models applied: %d\n\n", len(entries)))

	for _, entry := range entries {
		content, ok := s.modelFiles[entry.Path]
		if !ok {
			continue
		}
		model := modelfile.Parse(content)
		if model.Name == "" {
			model.Name = entry.Name
		}

		b.WriteString(fmt.Sprintf("## %s (%s)\n\n", model.Name, entry.Category))

		if len(model.ThinkingSteps) > 0 {
			b.WriteString("### Thinking Steps\n")
			for i, step := range model.ThinkingSteps {
				b.WriteString(fmt.Sprintf("%d. %s\n", i+1, step))
			}
			b.WriteString("\n")
		}

		if len(model.CoachingQuestions) > 0 {
			b.WriteString("### Coaching Questions\n")
			for _, q := range model.CoachingQuestions {
				b.WriteString(fmt.Sprintf("- %s\n", q))
			}
			b.WriteString("\n")
		}

		b.WriteString("---\n\n")
	}

	return b.String(), false
}

func (s *Server) toolSuggest(args json.RawMessage) (string, bool) {
	var p struct {
		Situation string `json:"situation"`
		Count     int    `json:"count"`
	}
	if err := json.Unmarshal(args, &p); err != nil {
		return "Error: " + err.Error(), true
	}
	if p.Situation == "" {
		return "Error: situation is required", true
	}
	if p.Count <= 0 {
		p.Count = 5
	}

	matches := s.idx.TopNForQuery(p.Situation, p.Count)
	if len(matches) == 0 {
		return "No relevant models found for: " + p.Situation, false
	}

	var b strings.Builder
	b.WriteString(fmt.Sprintf("Suggested models for: \"%s\"\n\n", p.Situation))
	for i, m := range matches {
		b.WriteString(fmt.Sprintf("%d. %s (%s)\n", i+1, m.Name, m.Category))
		if m.Summary != "" {
			b.WriteString(fmt.Sprintf("   %s\n", m.Summary))
		}
		b.WriteString("\n")
	}

	return b.String(), false
}

func (s *Server) toolSearch(args json.RawMessage) (string, bool) {
	var p struct {
		Query string `json:"query"`
	}
	if err := json.Unmarshal(args, &p); err != nil {
		return "Error: " + err.Error(), true
	}
	if p.Query == "" {
		return "Error: query is required", true
	}

	results := s.idx.Search(p.Query)
	if len(results) == 0 {
		return "No models found for: " + p.Query, false
	}

	var b strings.Builder
	b.WriteString(fmt.Sprintf("Found %d model(s) matching \"%s\":\n\n", len(results), p.Query))
	for _, m := range results {
		b.WriteString(fmt.Sprintf("  %-30s [%s] %s\n", m.Name, m.Category, m.Slug))
		if m.Summary != "" {
			b.WriteString(fmt.Sprintf("  %s\n\n", m.Summary))
		}
	}

	return b.String(), false
}

func (s *Server) toolApply(args json.RawMessage) (string, bool) {
	var p struct {
		Model   string `json:"model"`
		Context string `json:"context"`
	}
	if err := json.Unmarshal(args, &p); err != nil {
		return "Error: " + err.Error(), true
	}
	if p.Model == "" {
		return "Error: model is required", true
	}
	if p.Context == "" {
		return "Error: context is required", true
	}

	entry := s.idx.FindBySlug(p.Model)
	if entry == nil {
		entry = s.idx.FindByID(p.Model)
	}
	if entry == nil {
		return "Error: model not found: " + p.Model, true
	}

	content, ok := s.modelFiles[entry.Path]
	if !ok {
		return "Error: model file not found: " + entry.Path, true
	}

	model := modelfile.Parse(content)
	if model.Name == "" {
		model.Name = entry.Name
	}

	var b strings.Builder
	b.WriteString(fmt.Sprintf("## %s\n", model.Name))
	b.WriteString(fmt.Sprintf("Category: %s\n\n", entry.Category))

	if model.Description != "" {
		b.WriteString(fmt.Sprintf("### Description\n%s\n\n", model.Description))
	}

	if len(model.ThinkingSteps) > 0 {
		b.WriteString("### Thinking Steps\n")
		for i, step := range model.ThinkingSteps {
			b.WriteString(fmt.Sprintf("%d. %s\n", i+1, step))
		}
		b.WriteString("\n")
	}

	if len(model.CoachingQuestions) > 0 {
		b.WriteString("### Coaching Questions\n")
		for _, q := range model.CoachingQuestions {
			b.WriteString(fmt.Sprintf("- %s\n", q))
		}
		b.WriteString("\n")
	}

	if model.WhenToAvoid != "" {
		b.WriteString(fmt.Sprintf("### When to Avoid\n%s\n\n", model.WhenToAvoid))
	}

	b.WriteString(fmt.Sprintf("### Context\n%s\n", p.Context))

	return b.String(), false
}

func (s *Server) toolList(args json.RawMessage) (string, bool) {
	var p struct {
		Category string `json:"category"`
	}
	if args != nil {
		_ = json.Unmarshal(args, &p)
	}

	var models []index.ModelEntry
	if p.Category != "" {
		models = s.idx.FilterByCategory(p.Category)
		if len(models) == 0 {
			var b strings.Builder
			b.WriteString(fmt.Sprintf("No models found in category: %s\n\nAvailable categories:\n", p.Category))
			for cat := range s.idx.Categories {
				b.WriteString(fmt.Sprintf("  - %s\n", cat))
			}
			return b.String(), false
		}
	} else {
		models = s.idx.Models
	}

	var b strings.Builder
	if p.Category != "" {
		b.WriteString(fmt.Sprintf("%s (%d models):\n\n", p.Category, len(models)))
	} else {
		b.WriteString(fmt.Sprintf("All models (%d total):\n\n", len(models)))
	}

	currentCat := ""
	for _, m := range models {
		if m.Category != currentCat {
			currentCat = m.Category
			if p.Category == "" {
				b.WriteString(fmt.Sprintf("\n  %s\n", currentCat))
			}
		}
		b.WriteString(fmt.Sprintf("    %-4s %-40s %s\n", m.ID, m.Name, m.Slug))
	}

	return b.String(), false
}
