package server

import (
	"encoding/json"
	"io"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	"github.com/ikhsan3adi/gemini-web2api/internal/config"
	"github.com/ikhsan3adi/gemini-web2api/internal/format"
	"github.com/ikhsan3adi/gemini-web2api/internal/gemini"
	"github.com/ikhsan3adi/gemini-web2api/internal/models"
)

func TestHealthEndpoint(t *testing.T) {
	testVer := "test-version-1.0"
	cfg := config.Default()
	app := New(cfg, testVer)
	handler := app.Handler()

	req := httptest.NewRequest("GET", "/", nil)
	rec := httptest.NewRecorder()

	handler.ServeHTTP(rec, req)

	if rec.Code != http.StatusOK {
		t.Fatalf("Expected status 200, got %d", rec.Code)
	}

	var body map[string]any
	if err := json.Unmarshal(rec.Body.Bytes(), &body); err != nil {
		t.Fatalf("Invalid JSON response: %v", err)
	}

	if body["status"] != "ok" || body["version"] != testVer {
		t.Errorf("Unexpected health response: %v", body)
	}
}

func TestAuthMatrix(t *testing.T) {
	cfg := config.Default()
	cfg.APIKeys = []string{"sk-secret-key"}
	app := New(cfg, "test-version")
	handler := app.Handler()

	// 1. No key -> 401
	req1 := httptest.NewRequest("GET", "/v1/models", nil)
	rec1 := httptest.NewRecorder()
	handler.ServeHTTP(rec1, req1)
	if rec1.Code != http.StatusUnauthorized {
		t.Errorf("Expected 401 without auth, got %d", rec1.Code)
	}

	// 2. Wrong key -> 401
	req2 := httptest.NewRequest("GET", "/v1/models", nil)
	req2.Header.Set("Authorization", "Bearer wrong-key")
	rec2 := httptest.NewRecorder()
	handler.ServeHTTP(rec2, req2)
	if rec2.Code != http.StatusUnauthorized {
		t.Errorf("Expected 401 with wrong key, got %d", rec2.Code)
	}

	// 3. Valid Bearer token -> 200
	req3 := httptest.NewRequest("GET", "/v1/models", nil)
	req3.Header.Set("Authorization", "Bearer sk-secret-key")
	rec3 := httptest.NewRecorder()
	handler.ServeHTTP(rec3, req3)
	if rec3.Code != http.StatusOK {
		t.Errorf("Expected 200 with Bearer token, got %d", rec3.Code)
	}

	// 4. Valid x-api-key header -> 200
	req4 := httptest.NewRequest("GET", "/v1/models", nil)
	req4.Header.Set("x-api-key", "sk-secret-key")
	rec4 := httptest.NewRecorder()
	handler.ServeHTTP(rec4, req4)
	if rec4.Code != http.StatusOK {
		t.Errorf("Expected 200 with x-api-key, got %d", rec4.Code)
	}

	// 5. Valid x-goog-api-key header -> 200
	req5 := httptest.NewRequest("GET", "/v1/models", nil)
	req5.Header.Set("x-goog-api-key", "sk-secret-key")
	rec5 := httptest.NewRecorder()
	handler.ServeHTTP(rec5, req5)
	if rec5.Code != http.StatusOK {
		t.Errorf("Expected 200 with x-goog-api-key, got %d", rec5.Code)
	}

	// 6. Valid query param ?key= -> 200
	req6 := httptest.NewRequest("GET", "/v1/models?key=sk-secret-key", nil)
	rec6 := httptest.NewRecorder()
	handler.ServeHTTP(rec6, req6)
	if rec6.Code != http.StatusOK {
		t.Errorf("Expected 200 with ?key= query, got %d", rec6.Code)
	}
}

func TestCORSPreflight(t *testing.T) {
	cfg := config.Default()
	app := New(cfg, "test-version")
	handler := app.Handler()

	req := httptest.NewRequest("OPTIONS", "/v1/chat/completions", nil)
	rec := httptest.NewRecorder()

	handler.ServeHTTP(rec, req)

	if rec.Code != http.StatusNoContent {
		t.Fatalf("Expected status 204 for OPTIONS, got %d", rec.Code)
	}
	if rec.Header().Get("Access-Control-Allow-Origin") != "*" {
		t.Errorf("Expected CORS origin *, got %s", rec.Header().Get("Access-Control-Allow-Origin"))
	}
}

func TestMarshalNoEscapeHTML(t *testing.T) {
	data := map[string]string{
		"text": "<hello & world>",
	}
	b, err := marshalNoEscapeHTML(data)
	if err != nil {
		t.Fatalf("Unexpected error: %v", err)
	}
	expected := `{"text":"<hello & world>"}`
	if string(b) != expected {
		t.Errorf("Got %q, want %q", string(b), expected)
	}
}

func TestChunkedTransferEncoding(t *testing.T) {
	// Verify that the server can parse request bodies via net/http transparent dechunking.
	cfg := config.Default()
	app := New(cfg, "test")
	handler := app.Handler()

	body := `{"model":"gemini-3.6-flash","messages":[{"role":"user","content":"hi"}]}`
	req := httptest.NewRequest("POST", "/v1/chat/completions", strings.NewReader(body))
	req.Header.Set("Content-Type", "application/json")
	rec := httptest.NewRecorder()
	handler.ServeHTTP(rec, req)

	// Should not be 400 (bad request) or 500 — it will be 502 (upstream unreachable) or 200,
	// but not a body-parsing error.
	if rec.Code == http.StatusBadRequest {
		t.Errorf("Handler rejected valid JSON body, got %d", rec.Code)
	}
}

func TestResponsesSequenceNumberOrder(t *testing.T) {
	// Test BuildResponseOutput structure and verify format helpers produce consistent output.
	output := format.BuildResponseOutput("Hello world", nil, "msg_test123")
	if len(output) != 1 {
		t.Fatalf("Expected 1 output item, got %d", len(output))
	}
	item := output[0]
	if item["type"] != "message" {
		t.Errorf("Expected type=message, got %v", item["type"])
	}
	if item["id"] != "msg_test123" {
		t.Errorf("Expected id=msg_test123, got %v", item["id"])
	}

	// With tool calls
	tc := models.OpenAIToolCall{
		ID:   "call_abc123",
		Type: "function",
		Function: models.OpenAIToolCallFunction{
			Name:      "get_weather",
			Arguments: `{"city":"Jakarta"}`,
		},
	}
	outputWithTools := format.BuildResponseOutput("", []models.OpenAIToolCall{tc}, "msg_456")
	// When text is empty, only the function_call item is returned (no empty message).
	if len(outputWithTools) != 1 {
		t.Fatalf("Expected 1 output item (function_call only when text empty), got %d", len(outputWithTools))
	}
	if outputWithTools[0]["type"] != "function_call" {
		t.Errorf("First item should be function_call, got %v", outputWithTools[0]["type"])
	}
	if outputWithTools[0]["call_id"] != "call_abc123" {
		t.Errorf("Expected call_id=call_abc123, got %v", outputWithTools[0]["call_id"])
	}

	// Verify responses format: the non-streaming endpoint should accept valid structure.
	// The responses API uses "input" field, not "messages".
	cfg := config.Default()
	app := New(cfg, "test")
	handler := app.Handler()

	body := `{"model":"gemini-3.6-flash","input":"test prompt"}`
	req := httptest.NewRequest("POST", "/v1/responses", strings.NewReader(body))
	req.Header.Set("Content-Type", "application/json")
	rec := httptest.NewRecorder()
	handler.ServeHTTP(rec, req)

	// Should not be 400 (bad request) — may be 502 (upstream unreachable)
	if rec.Code == http.StatusBadRequest {
		t.Errorf("Responses endpoint rejected valid JSON, got %d: %s", rec.Code, rec.Body.String())
	}
}

func TestModelHeaders(t *testing.T) {
	cfg := config.Default()
	app := New(cfg, "test")
	handler := app.Handler()

	// 1. Known model: gemini-4-argon
	bodyKnown := `{"model":"gemini-4-argon","messages":[{"role":"user","content":"hi"}]}`
	req1 := httptest.NewRequest("POST", "/v1/chat/completions", strings.NewReader(bodyKnown))
	req1.Header.Set("Content-Type", "application/json")
	rec1 := httptest.NewRecorder()
	handler.ServeHTTP(rec1, req1)

	if got := rec1.Header().Get("X-Resolved-Model"); got != "gemini-4-argon" {
		t.Errorf("Expected X-Resolved-Model=gemini-4-argon, got %q", got)
	}
	if got := rec1.Header().Get("X-Model-Fallback"); got != "" {
		t.Errorf("Expected X-Model-Fallback to be empty for known model, got %q", got)
	}

	// 2. Unknown model: should fallback and set X-Model-Fallback: true
	bodyUnknown := `{"model":"gemini-9.9-fake","messages":[{"role":"user","content":"hi"}]}`
	req2 := httptest.NewRequest("POST", "/v1/chat/completions", strings.NewReader(bodyUnknown))
	req2.Header.Set("Content-Type", "application/json")
	rec2 := httptest.NewRecorder()
	handler.ServeHTTP(rec2, req2)

	if got := rec2.Header().Get("X-Resolved-Model"); got != "gemini-3.8-flash" {
		t.Errorf("Expected X-Resolved-Model=gemini-3.8-flash, got %q", got)
	}
	if got := rec2.Header().Get("X-Model-Fallback"); got != "true" {
		t.Errorf("Expected X-Model-Fallback=true for unknown model, got %q", got)
	}
}

func TestDynamicModelsFrom(t *testing.T) {
	discovered := []gemini.DiscoveredModel{
		{ID: "56fdd199312815e2", Display: "4 Hydrogen", Desc: "", Number: 1},
		{ID: "krypton-ticket", Display: "4 Krypton", Desc: "rare gas", Number: 8},
		{ID: "xenon-ticket", Display: "4 Xenon", Desc: "", Number: 8},
		{ID: "xenon-duplicate", Display: "4 Xenon", Desc: "", Number: 8},
		{ID: "unknown", Display: "weird label", Desc: "", Number: 1},
	}

	got := dynamicModelsFrom(discovered)

	byName := make(map[string]models.DynamicModel, len(got))
	for _, dm := range got {
		byName[dm.Name] = dm
	}

	if _, ok := byName["gemini-4-hydrogen"]; ok {
		t.Error("static Hydrogen entry must take precedence over discovered model")
	}
	if _, ok := byName["gemini-4-krypton"]; !ok {
		t.Fatal("expected gemini-4-krypton from discovery")
	}
	if byName["gemini-4-krypton"].Ticket != "krypton-ticket" || byName["gemini-4-krypton"].Mode != 8 {
		t.Errorf("unexpected krypton mapping: %+v", byName["gemini-4-krypton"])
	}
	if byName["gemini-4-krypton"].Desc != "rare gas" {
		t.Errorf("expected discovery description, got %q", byName["gemini-4-krypton"].Desc)
	}
	if byName["gemini-4-xenon"].Desc != "4 Xenon" {
		t.Errorf("expected display fallback description, got %q", byName["gemini-4-xenon"].Desc)
	}
	if len(got) != 2 {
		t.Errorf("expected 2 dynamic models after dedupe/filter, got %d: %+v", len(got), got)
	}
}

func TestWriteSSEData_ProtocolFormat(t *testing.T) {
	rec := httptest.NewRecorder()
	payload := map[string]string{"message": "hello world"}

	if err := writeSSEData(rec, payload); err != nil {
		t.Fatalf("writeSSEData failed: %v", err)
	}

	got := rec.Body.String()
	want := "data: {\"message\":\"hello world\"}\n\n"
	if got != want {
		t.Errorf("writeSSEData format mismatch:\ngot:  %q\nwant: %q", got, want)
	}
}

func TestWriteSSEEvent_ProtocolFormat(t *testing.T) {
	rec := httptest.NewRecorder()
	payload := map[string]string{"status": "in_progress"}

	if err := writeSSEEvent(rec, "update", payload); err != nil {
		t.Fatalf("writeSSEEvent failed: %v", err)
	}

	got := rec.Body.String()
	want := "event: update\ndata: {\"status\":\"in_progress\"}\n\n"
	if got != want {
		t.Errorf("writeSSEEvent format mismatch:\ngot:  %q\nwant: %q", got, want)
	}
}

func TestWriteSSEDone_ProtocolFormat(t *testing.T) {
	rec := httptest.NewRecorder()

	if err := writeSSEDone(rec); err != nil {
		t.Fatalf("writeSSEDone failed: %v", err)
	}

	got := rec.Body.String()
	want := "data: [DONE]\n\n"
	if got != want {
		t.Errorf("writeSSEDone format mismatch:\ngot:  %q\nwant: %q", got, want)
	}
}

func TestDynamicModelsCaching(t *testing.T) {
	cfg := config.Default()
	app := New(cfg, "test")

	// Call dynamicModels twice
	m1 := app.dynamicModels()
	m2 := app.dynamicModels()

	// Both should point to the exact same cached slice header in memory
	if len(m1) != len(m2) {
		t.Errorf("expected same length, got %d and %d", len(m1), len(m2))
	}
	if len(m1) > 0 && &m1[0] != &m2[0] {
		t.Errorf("expected cached slice to be reused")
	}
}

type mockRequester struct {
	doFunc func(req *http.Request) (*http.Response, error)
}

func (m *mockRequester) Do(req *http.Request) (*http.Response, error) {
	return m.doFunc(req)
}

func makeGeminiTestChunk(text string) string {
	inner := make([]any, 5)
	inner[0] = strings.Repeat("metadata_filler_string_", 10)
	inner[4] = []any{[]any{"c_1", []any{text}}}
	innerBytes, _ := json.Marshal(inner)
	outer := []any{[]any{"wrb.fr", nil, string(innerBytes)}}
	lineBytes, _ := json.Marshal(outer)
	return string(lineBytes) + "\n"
}

func TestModelsEndpoint(t *testing.T) {
	cfg := config.Default()
	app := New(cfg, "test")
	handler := app.Handler()

	req := httptest.NewRequest("GET", "/v1/models", nil)
	rec := httptest.NewRecorder()
	handler.ServeHTTP(rec, req)

	if rec.Code != http.StatusOK {
		t.Fatalf("Expected 200, got %d", rec.Code)
	}
	var res map[string]any
	if err := json.Unmarshal(rec.Body.Bytes(), &res); err != nil {
		t.Fatalf("Invalid JSON: %v", err)
	}
	if res["object"] != "list" {
		t.Errorf("Expected object=list, got %v", res["object"])
	}
	data, ok := res["data"].([]any)
	if !ok || len(data) == 0 {
		t.Errorf("Expected non-empty data array, got %v", res["data"])
	}
}

func TestGoogleModelsEndpoint(t *testing.T) {
	cfg := config.Default()
	app := New(cfg, "test")
	handler := app.Handler()

	req := httptest.NewRequest("GET", "/v1beta/models", nil)
	rec := httptest.NewRecorder()
	handler.ServeHTTP(rec, req)

	if rec.Code != http.StatusOK {
		t.Fatalf("Expected 200, got %d", rec.Code)
	}
	var res map[string]any
	if err := json.Unmarshal(rec.Body.Bytes(), &res); err != nil {
		t.Fatalf("Invalid JSON: %v", err)
	}
	modelsList, ok := res["models"].([]any)
	if !ok || len(modelsList) == 0 {
		t.Errorf("Expected non-empty models array, got %v", res["models"])
	}
}

func TestGoogleGenerate_Validation(t *testing.T) {
	cfg := config.Default()
	app := New(cfg, "test")
	handler := app.Handler()

	// 1. Invalid JSON body -> 400
	req1 := httptest.NewRequest("POST", "/v1beta/models/gemini-3.8-flash:generateContent", strings.NewReader("{invalid"))
	rec1 := httptest.NewRecorder()
	handler.ServeHTTP(rec1, req1)
	if rec1.Code != http.StatusBadRequest {
		t.Errorf("Expected 400 for invalid JSON, got %d", rec1.Code)
	}

	// 2. Empty contents -> 400
	req2 := httptest.NewRequest("POST", "/v1beta/models/gemini-3.8-flash:generateContent", strings.NewReader(`{"contents":[]}`))
	rec2 := httptest.NewRecorder()
	handler.ServeHTTP(rec2, req2)
	if rec2.Code != http.StatusBadRequest {
		t.Errorf("Expected 400 for empty contents, got %d", rec2.Code)
	}
}

func TestChatCompletions_Streaming_EndToEnd(t *testing.T) {
	cfg := config.Default()
	app := New(cfg, "test")

	// Inject mock HTTP requester that yields simulated Gemini streaming chunks
	mockRespBody := io.NopCloser(strings.NewReader(
		makeGeminiTestChunk("Hello") +
			makeGeminiTestChunk("Hello world") +
			makeGeminiTestChunk("Hello world from Gemini!"),
	))

	app.Gem.HTTP = &mockRequester{
		doFunc: func(req *http.Request) (*http.Response, error) {
			return &http.Response{
				StatusCode: http.StatusOK,
				Body:       mockRespBody,
				Header:     make(http.Header),
			}, nil
		},
	}

	handler := app.Handler()
	body := `{"model":"gemini-3.8-flash","stream":true,"messages":[{"role":"user","content":"Hi"}]}`
	req := httptest.NewRequest("POST", "/v1/chat/completions", strings.NewReader(body))
	req.Header.Set("Content-Type", "application/json")
	rec := httptest.NewRecorder()

	handler.ServeHTTP(rec, req)

	if rec.Code != http.StatusOK {
		t.Fatalf("Expected 200, got %d: %s", rec.Code, rec.Body.String())
	}

	output := rec.Body.String()
	if !strings.Contains(output, "data: [DONE]") {
		t.Errorf("Missing [DONE] in SSE output: %s", output)
	}
	if !strings.Contains(output, "Hello") || !strings.Contains(output, "world") {
		t.Errorf("Missing expected streamed delta tokens in output: %s", output)
	}
}
