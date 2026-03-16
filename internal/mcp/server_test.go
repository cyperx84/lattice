package mcp

import (
	"bytes"
	"encoding/json"
	"io"
	"strings"
	"testing"

	"github.com/cyperx84/lattice/internal/index"
)

// testIndex builds a small index for testing.
func testIndex() *index.ModelIndex {
	return &index.ModelIndex{
		Version:     "1.0.0",
		Generated:   "2025-01-01",
		TotalModels: 3,
		Categories: map[string][]string{
			"General Thinking Tools": {"m01", "m02"},
			"Economics":              {"m03"},
		},
		Models: []index.ModelEntry{
			{ID: "m01", Name: "Inversion", Slug: "inversion", Category: "General Thinking Tools", Path: "models/m01_inversion.md", Keywords: []string{"flip", "reverse"}, Summary: "Think backwards"},
			{ID: "m02", Name: "Second-Order Thinking", Slug: "second-order-thinking", Category: "General Thinking Tools", Path: "models/m02_second-order-thinking.md", Keywords: []string{"consequences"}, Summary: "Consider second-order effects"},
			{ID: "m03", Name: "Trade-offs", Slug: "trade-offs", Category: "Economics", Path: "models/m03_trade-offs.md", Keywords: []string{"cost", "decision"}, Summary: "Map opportunity costs"},
		},
	}
}

func testModelFiles() map[string]string {
	return map[string]string{
		"models/m01_inversion.md": `## Mental Model = Inversion

**Category = General Thinking Tools**

**Description:**
Think backwards from the goal to identify what to avoid.

**When to Avoid This Model:**
Simple straightforward problems.

**Keywords for Situations:**
flip, reverse, avoid failure

**Thinking Steps:**
1. Define the desired outcome
2. Ask what would guarantee failure
3. List failure modes
4. Invert into action items
5. Prioritize

**Coaching Questions:**
- What would make this fail?
- Which assumptions are unquestioned?
- What's the worst decision?
`,
		"models/m02_second-order-thinking.md": `## Mental Model = Second-Order Thinking

**Category = General Thinking Tools**

**Description:**
Consider the consequences of the consequences.

**Thinking Steps:**
1. Identify first-order effects
2. Ask what happens next
3. Map the chain of consequences
4. Identify unintended effects
5. Weigh long-term vs short-term

**Coaching Questions:**
- What are the second-order effects?
- Who else is affected?
`,
		"models/m03_trade-offs.md": `## Mental Model = Trade-offs

**Category = Economics**

**Description:**
Every decision has an opportunity cost.

**Thinking Steps:**
1. List all options
2. Identify costs of each
3. Map opportunity costs
4. Compare trade-offs
5. Decide with full picture

**Coaching Questions:**
- What are you giving up?
- Is this the best use of resources?
`,
	}
}

// sendRequest sends a JSON-RPC request and returns the response.
func sendRequest(t *testing.T, s *Server, method string, id int, params interface{}) response {
	t.Helper()
	req := map[string]interface{}{
		"jsonrpc": "2.0",
		"id":      id,
		"method":  method,
	}
	if params != nil {
		data, _ := json.Marshal(params)
		req["params"] = json.RawMessage(data)
	}

	reqData, err := json.Marshal(req)
	if err != nil {
		t.Fatal(err)
	}

	reader := strings.NewReader(string(reqData) + "\n")
	var writer bytes.Buffer

	if err := s.Run(reader, &writer); err != nil {
		t.Fatal(err)
	}

	var resp response
	if err := json.Unmarshal(writer.Bytes(), &resp); err != nil {
		t.Fatalf("failed to parse response: %v\nraw: %s", err, writer.String())
	}
	return resp
}

func newTestServer() *Server {
	return NewServer(testIndex(), testModelFiles(), false, io.Discard)
}

func TestInitialize(t *testing.T) {
	s := newTestServer()
	resp := sendRequest(t, s, "initialize", 1, nil)

	if resp.Error != nil {
		t.Fatalf("unexpected error: %s", resp.Error.Message)
	}

	data, _ := json.Marshal(resp.Result)
	var info serverInfo
	if err := json.Unmarshal(data, &info); err != nil {
		t.Fatal(err)
	}

	if info.ServerInfo.Name != "lattice" {
		t.Errorf("expected server name 'lattice', got %q", info.ServerInfo.Name)
	}
	if info.ServerInfo.Version != "0.3.0" {
		t.Errorf("expected version '0.3.0', got %q", info.ServerInfo.Version)
	}
}

func TestToolsList(t *testing.T) {
	s := newTestServer()
	resp := sendRequest(t, s, "tools/list", 2, nil)

	if resp.Error != nil {
		t.Fatalf("unexpected error: %s", resp.Error.Message)
	}

	data, _ := json.Marshal(resp.Result)
	var result toolsListResult
	if err := json.Unmarshal(data, &result); err != nil {
		t.Fatal(err)
	}

	if len(result.Tools) != 5 {
		t.Errorf("expected 5 tools, got %d", len(result.Tools))
	}

	expected := map[string]bool{"think": false, "suggest": false, "search": false, "apply": false, "list": false}
	for _, tool := range result.Tools {
		if _, ok := expected[tool.Name]; ok {
			expected[tool.Name] = true
		}
	}
	for name, found := range expected {
		if !found {
			t.Errorf("missing tool: %s", name)
		}
	}
}

func TestToolsCallThink(t *testing.T) {
	s := newTestServer()
	params := map[string]interface{}{
		"name":      "think",
		"arguments": map[string]interface{}{"problem": "reverse flip inversion decision cost"},
	}
	resp := sendRequest(t, s, "tools/call", 3, params)

	if resp.Error != nil {
		t.Fatalf("unexpected error: %s", resp.Error.Message)
	}

	data, _ := json.Marshal(resp.Result)
	var result toolResult
	if err := json.Unmarshal(data, &result); err != nil {
		t.Fatal(err)
	}

	if result.IsError {
		t.Fatalf("tool returned error: %s", result.Content[0].Text)
	}
	if len(result.Content) == 0 {
		t.Fatal("expected content in response")
	}
	if !strings.Contains(result.Content[0].Text, "Thinking:") {
		t.Errorf("expected 'Thinking:' in output, got: %s", result.Content[0].Text)
	}
}

func TestToolsCallSearch(t *testing.T) {
	s := newTestServer()
	params := map[string]interface{}{
		"name":      "search",
		"arguments": map[string]interface{}{"query": "inversion"},
	}
	resp := sendRequest(t, s, "tools/call", 4, params)

	if resp.Error != nil {
		t.Fatalf("unexpected error: %s", resp.Error.Message)
	}

	data, _ := json.Marshal(resp.Result)
	var result toolResult
	if err := json.Unmarshal(data, &result); err != nil {
		t.Fatal(err)
	}

	if result.IsError {
		t.Fatalf("tool returned error: %s", result.Content[0].Text)
	}
	if !strings.Contains(result.Content[0].Text, "Inversion") {
		t.Error("expected 'Inversion' in search results")
	}
}

func TestToolsCallSuggest(t *testing.T) {
	s := newTestServer()
	params := map[string]interface{}{
		"name":      "suggest",
		"arguments": map[string]interface{}{"situation": "making a big decision"},
	}
	resp := sendRequest(t, s, "tools/call", 5, params)

	if resp.Error != nil {
		t.Fatalf("unexpected error: %s", resp.Error.Message)
	}

	data, _ := json.Marshal(resp.Result)
	var result toolResult
	if err := json.Unmarshal(data, &result); err != nil {
		t.Fatal(err)
	}

	if result.IsError {
		t.Fatalf("tool returned error: %s", result.Content[0].Text)
	}
	if !strings.Contains(result.Content[0].Text, "Suggested models") {
		t.Error("expected 'Suggested models' in output")
	}
}

func TestToolsCallApply(t *testing.T) {
	s := newTestServer()
	params := map[string]interface{}{
		"name":      "apply",
		"arguments": map[string]interface{}{"model": "inversion", "context": "designing a new API"},
	}
	resp := sendRequest(t, s, "tools/call", 6, params)

	if resp.Error != nil {
		t.Fatalf("unexpected error: %s", resp.Error.Message)
	}

	data, _ := json.Marshal(resp.Result)
	var result toolResult
	if err := json.Unmarshal(data, &result); err != nil {
		t.Fatal(err)
	}

	if result.IsError {
		t.Fatalf("tool returned error: %s", result.Content[0].Text)
	}
	if !strings.Contains(result.Content[0].Text, "Inversion") {
		t.Error("expected 'Inversion' in apply output")
	}
	if !strings.Contains(result.Content[0].Text, "Thinking Steps") {
		t.Error("expected 'Thinking Steps' in apply output")
	}
}

func TestToolsCallList(t *testing.T) {
	s := newTestServer()
	params := map[string]interface{}{
		"name":      "list",
		"arguments": map[string]interface{}{},
	}
	resp := sendRequest(t, s, "tools/call", 7, params)

	if resp.Error != nil {
		t.Fatalf("unexpected error: %s", resp.Error.Message)
	}

	data, _ := json.Marshal(resp.Result)
	var result toolResult
	if err := json.Unmarshal(data, &result); err != nil {
		t.Fatal(err)
	}

	if result.IsError {
		t.Fatalf("tool returned error: %s", result.Content[0].Text)
	}
	if !strings.Contains(result.Content[0].Text, "3 total") {
		t.Errorf("expected '3 total' in list output, got: %s", result.Content[0].Text)
	}
}

func TestUnknownMethod(t *testing.T) {
	s := newTestServer()
	resp := sendRequest(t, s, "nonexistent/method", 8, nil)

	if resp.Error == nil {
		t.Fatal("expected error for unknown method")
	}
	if resp.Error.Code != -32601 {
		t.Errorf("expected error code -32601, got %d", resp.Error.Code)
	}
}

func TestUnknownTool(t *testing.T) {
	s := newTestServer()
	params := map[string]interface{}{
		"name":      "nonexistent_tool",
		"arguments": map[string]interface{}{},
	}
	resp := sendRequest(t, s, "tools/call", 9, params)

	if resp.Error != nil {
		t.Fatalf("unexpected JSON-RPC error: %s", resp.Error.Message)
	}

	data, _ := json.Marshal(resp.Result)
	var result toolResult
	if err := json.Unmarshal(data, &result); err != nil {
		t.Fatal(err)
	}

	if !result.IsError {
		t.Error("expected tool error for unknown tool")
	}
	if !strings.Contains(result.Content[0].Text, "unknown tool") {
		t.Errorf("expected 'unknown tool' in error message, got: %s", result.Content[0].Text)
	}
}
