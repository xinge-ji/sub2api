package service

import (
	"context"
	"crypto/sha256"
	"encoding/hex"
	"encoding/json"
	"fmt"
	"strings"
	"time"

	"github.com/Wei-Shaw/sub2api/internal/pkg/pagination"
	"github.com/tidwall/gjson"
)

type OpenAIConversationRetentionSession struct {
	ID              int64     `json:"id"`
	UserID          int64     `json:"user_id"`
	APIKeyID        int64     `json:"api_key_id"`
	AccountID       *int64    `json:"account_id"`
	GroupID         *int64    `json:"group_id"`
	SessionKey      string    `json:"session_key"`
	ClientSessionID string    `json:"client_session_id"`
	RequestedModel  string    `json:"requested_model"`
	UpstreamModel   string    `json:"upstream_model"`
	ReasoningEffort string    `json:"reasoning_effort"`
	FirstRequestID  string    `json:"first_request_id"`
	LastRequestID   string    `json:"last_request_id"`
	FirstResponseID string    `json:"first_response_id"`
	LastResponseID  string    `json:"last_response_id"`
	TurnCount       int       `json:"turn_count"`
	CreatedAt       time.Time `json:"created_at"`
	UpdatedAt       time.Time `json:"updated_at"`
}

type OpenAIConversationRetentionTurn struct {
	ID                  int64       `json:"id"`
	SessionID           int64       `json:"session_id"`
	UserID              int64       `json:"user_id"`
	APIKeyID            int64       `json:"api_key_id"`
	AccountID           *int64      `json:"account_id"`
	GroupID             *int64      `json:"group_id"`
	RequestID           string      `json:"request_id"`
	ResponseID          string      `json:"response_id"`
	TurnIndex           int         `json:"turn_index"`
	RequestedModel      string      `json:"requested_model"`
	UpstreamModel       string      `json:"upstream_model"`
	ReasoningEffort     string      `json:"reasoning_effort"`
	Stream              bool        `json:"stream"`
	RequestType         RequestType `json:"request_type"`
	InboundEndpoint     string      `json:"inbound_endpoint"`
	UpstreamEndpoint    string      `json:"upstream_endpoint"`
	UserInputText       string      `json:"user_input_text"`
	AssistantOutputText string      `json:"assistant_output_text"`
	CreatedAt           time.Time   `json:"created_at"`
}

type OpenAIConversationRetentionRecordInput struct {
	UserID           int64
	APIKeyID         int64
	AccountID        *int64
	GroupID          *int64
	RequestID        string
	ResponseID       string
	RequestedModel   string
	UpstreamModel    string
	ReasoningEffort  string
	Stream           bool
	RequestType      RequestType
	InboundEndpoint  string
	UpstreamEndpoint string
	ClientSessionID  string
	UserInputText    string
	AssistantText    string
	CreatedAt        time.Time
}

type OpenAIConversationRetentionView struct {
	Session  *OpenAIConversationRetentionSession `json:"session"`
	Turns    []*OpenAIConversationRetentionTurn  `json:"turns"`
	Total    int64                               `json:"total"`
	Page     int                                 `json:"page"`
	PageSize int                                 `json:"page_size"`
	Pages    int                                 `json:"pages"`
}

type OpenAIConversationRetentionListFilters struct {
	UserID          int64
	APIKeyID        int64
	AccountID       int64
	GroupID         int64
	Model           string
	ReasoningEffort string
	RequestType     *int16
	Stream          *bool
	StartTime       *time.Time
	EndTime         *time.Time
}

type OpenAIConversationRetentionListItem struct {
	ID              int64                            `json:"id"`
	UserID          int64                            `json:"user_id"`
	APIKeyID        int64                            `json:"api_key_id"`
	AccountID       *int64                           `json:"account_id"`
	GroupID         *int64                           `json:"group_id"`
	SessionKey      string                           `json:"session_key"`
	ClientSessionID string                           `json:"client_session_id"`
	RequestedModel  string                           `json:"requested_model"`
	UpstreamModel   string                           `json:"upstream_model"`
	ReasoningEffort string                           `json:"reasoning_effort"`
	FirstRequestID  string                           `json:"first_request_id"`
	LastRequestID   string                           `json:"last_request_id"`
	FirstResponseID string                           `json:"first_response_id"`
	LastResponseID  string                           `json:"last_response_id"`
	TurnCount       int                              `json:"turn_count"`
	CreatedAt       time.Time                        `json:"created_at"`
	UpdatedAt       time.Time                        `json:"updated_at"`
	UserEmail       string                           `json:"user_email"`
	APIKeyName      string                           `json:"api_key_name"`
	AccountName     string                           `json:"account_name"`
	GroupName       string                           `json:"group_name"`
	LastTurn        *OpenAIConversationRetentionTurn `json:"last_turn,omitempty"`
}

type OpenAIConversationRetentionRepository interface {
	CreateTurn(ctx context.Context, input *OpenAIConversationRetentionRecordInput) (*OpenAIConversationRetentionSession, *OpenAIConversationRetentionTurn, error)
	GetConversationByRequestID(ctx context.Context, requestID string, params pagination.PaginationParams) (*OpenAIConversationRetentionView, error)
	ListConversations(ctx context.Context, filters OpenAIConversationRetentionListFilters, params pagination.PaginationParams) ([]*OpenAIConversationRetentionListItem, *pagination.PaginationResult, error)
	DeleteExpiredTurns(ctx context.Context, cutoff time.Time, limit int) (int64, error)
	DeleteEmptySessions(ctx context.Context, limit int) (int64, error)
}

func BuildOpenAIConversationSessionKey(userID int64, clientSessionID, requestID string) string {
	normalizedClientSessionID := strings.TrimSpace(clientSessionID)
	if normalizedClientSessionID == "" {
		normalizedClientSessionID = strings.TrimSpace(requestID)
	}
	sum := sha256.Sum256([]byte(fmt.Sprintf("%d|%s", userID, normalizedClientSessionID)))
	return hex.EncodeToString(sum[:])
}

func extractOpenAIClientSessionID(headers sessionHeaderReader, body []byte) string {
	if headers != nil {
		if v := strings.TrimSpace(headers.GetHeader("session_id")); v != "" {
			return v
		}
		if v := strings.TrimSpace(headers.GetHeader("conversation_id")); v != "" {
			return v
		}
	}
	if len(body) == 0 {
		return ""
	}
	return strings.TrimSpace(gjson.GetBytes(body, "prompt_cache_key").String())
}

type sessionHeaderReader interface {
	GetHeader(string) string
}

func extractOpenAIUserInputText(body []byte, inboundEndpoint string) string {
	if len(body) == 0 || !gjson.ValidBytes(body) {
		return ""
	}
	if strings.Contains(inboundEndpoint, "/chat/completions") {
		return normalizeConversationText(extractLastChatCompletionsUserText(body))
	}
	return normalizeConversationText(extractResponsesUserText(body))
}

func extractResponsesUserText(body []byte) string {
	var parts []string

	input := gjson.GetBytes(body, "input")
	if !input.Exists() {
		return ""
	}
	if input.Type == gjson.String {
		return input.String()
	}
	if !input.IsArray() {
		return ""
	}

	for _, item := range input.Array() {
		role := strings.TrimSpace(item.Get("role").String())
		if role != "" && role != "user" {
			continue
		}
		if text := extractTextFromResponsesContent(item.Get("content")); text != "" {
			parts = append(parts, text)
			continue
		}
		if text := extractTextFromResponsesInputItem(item); text != "" {
			parts = append(parts, text)
		}
	}
	return strings.Join(parts, "\n\n")
}

func extractTextFromResponsesInputItem(item gjson.Result) string {
	itemType := strings.TrimSpace(item.Get("type").String())
	switch itemType {
	case "input_text":
		return strings.TrimSpace(item.Get("text").String())
	case "message":
		return extractTextFromResponsesContent(item.Get("content"))
	default:
		return ""
	}
}

func extractTextFromResponsesContent(content gjson.Result) string {
	if !content.Exists() {
		return ""
	}
	if content.Type == gjson.String {
		return content.String()
	}
	if !content.IsArray() {
		return ""
	}
	parts := make([]string, 0)
	for _, part := range content.Array() {
		partType := strings.TrimSpace(part.Get("type").String())
		switch partType {
		case "text", "input_text":
			if text := strings.TrimSpace(part.Get("text").String()); text != "" {
				parts = append(parts, text)
			}
		}
	}
	return strings.Join(parts, "\n\n")
}

func extractLastChatCompletionsUserText(body []byte) string {
	messages := gjson.GetBytes(body, "messages")
	if !messages.Exists() || !messages.IsArray() {
		return ""
	}
	last := ""
	for _, item := range messages.Array() {
		if strings.TrimSpace(item.Get("role").String()) != "user" {
			continue
		}
		if text := extractChatMessageText(item.Get("content")); text != "" {
			last = text
		}
	}
	return last
}

func extractChatMessageText(content gjson.Result) string {
	if !content.Exists() {
		return ""
	}
	if content.Type == gjson.String {
		return content.String()
	}
	if !content.IsArray() {
		return ""
	}
	parts := make([]string, 0)
	for _, part := range content.Array() {
		if strings.TrimSpace(part.Get("type").String()) != "text" {
			continue
		}
		if text := strings.TrimSpace(part.Get("text").String()); text != "" {
			parts = append(parts, text)
		}
	}
	return strings.Join(parts, "\n\n")
}

func extractOpenAIAssistantFinalTextFromJSON(body []byte, inboundEndpoint string) string {
	if len(body) == 0 || !gjson.ValidBytes(body) {
		return ""
	}
	if strings.Contains(inboundEndpoint, "/chat/completions") {
		return normalizeConversationText(extractChatCompletionsAssistantText(body))
	}
	return normalizeConversationText(extractResponsesAssistantText(body))
}

func ExtractOpenAIAssistantFinalTextFromJSON(body []byte, inboundEndpoint string) string {
	return extractOpenAIAssistantFinalTextFromJSON(body, inboundEndpoint)
}

func AnalyzeOpenAIResponseShapeFromJSON(body []byte, inboundEndpoint string) OpenAIResponseShape {
	if len(body) == 0 || !gjson.ValidBytes(body) {
		return OpenAIResponseShape{}
	}
	if strings.Contains(inboundEndpoint, "/chat/completions") {
		return analyzeChatCompletionsResponseShape(body)
	}
	return analyzeResponsesResponseShape(body)
}

func ExtractOpenAIClientSessionID(headers sessionHeaderReader, body []byte) string {
	return extractOpenAIClientSessionID(headers, body)
}

func ExtractOpenAIUserInputText(body []byte, inboundEndpoint string) string {
	return extractOpenAIUserInputText(body, inboundEndpoint)
}

func extractChatCompletionsAssistantText(body []byte) string {
	choices := gjson.GetBytes(body, "choices")
	if !choices.Exists() || !choices.IsArray() {
		return ""
	}
	parts := make([]string, 0)
	for _, choice := range choices.Array() {
		message := choice.Get("message")
		if !message.Exists() {
			continue
		}
		text := extractChatMessageText(message.Get("content"))
		if text == "" {
			text = strings.TrimSpace(message.Get("content").String())
		}
		if text != "" {
			parts = append(parts, text)
		}
	}
	return strings.Join(parts, "\n\n")
}

func analyzeChatCompletionsResponseShape(body []byte) OpenAIResponseShape {
	var shape OpenAIResponseShape
	choices := gjson.GetBytes(body, "choices")
	if !choices.Exists() || !choices.IsArray() {
		return shape
	}
	for _, choice := range choices.Array() {
		message := choice.Get("message")
		if !message.Exists() {
			continue
		}
		if extractChatMessageText(message.Get("content")) != "" || strings.TrimSpace(message.Get("content").String()) != "" {
			shape.HasFinalAnswer = true
		}
		if strings.TrimSpace(message.Get("reasoning_content").String()) != "" {
			shape.HasCommentary = true
		}
		if toolCalls := message.Get("tool_calls"); toolCalls.Exists() && toolCalls.IsArray() && len(toolCalls.Array()) > 0 {
			shape.HasToolCall = true
		}
	}
	return shape
}

func extractResponsesAssistantText(body []byte) string {
	output := gjson.GetBytes(body, "output")
	if !output.Exists() || !output.IsArray() {
		return ""
	}
	parts := make([]string, 0)
	for _, item := range output.Array() {
		if strings.TrimSpace(item.Get("type").String()) != "message" {
			continue
		}
		if strings.TrimSpace(item.Get("role").String()) != "assistant" {
			continue
		}
		content := item.Get("content")
		if !content.Exists() || !content.IsArray() {
			continue
		}
		for _, part := range content.Array() {
			if strings.TrimSpace(part.Get("type").String()) != "output_text" {
				continue
			}
			if text := strings.TrimSpace(part.Get("text").String()); text != "" {
				parts = append(parts, text)
			}
		}
	}
	return strings.Join(parts, "\n\n")
}

func analyzeResponsesResponseShape(body []byte) OpenAIResponseShape {
	var shape OpenAIResponseShape
	output := gjson.GetBytes(body, "output")
	if !output.Exists() || !output.IsArray() {
		return shape
	}
	for _, item := range output.Array() {
		itemType := strings.TrimSpace(item.Get("type").String())
		switch itemType {
		case "message":
			role := strings.TrimSpace(item.Get("role").String())
			if role != "" && role != "assistant" {
				continue
			}
			if strings.TrimSpace(item.Get("phase").String()) == "commentary" {
				shape.HasCommentary = true
			}
			if responsesMessageHasText(item.Get("content")) {
				shape.HasFinalAnswer = true
			}
		case "reasoning":
			shape.HasReasoningItem = true
			if responsesReasoningItemHasVisibleText(item) {
				shape.HasCommentary = true
			}
		case "function_call", "custom_tool_call", "web_search_call", "file_search_call", "computer_call", "mcp_call", "mcp_tool_call":
			shape.HasToolCall = true
		}
	}
	return shape
}

func responsesMessageHasText(content gjson.Result) bool {
	if !content.Exists() {
		return false
	}
	if content.Type == gjson.String {
		return strings.TrimSpace(content.String()) != ""
	}
	if !content.IsArray() {
		return false
	}
	for _, part := range content.Array() {
		partType := strings.TrimSpace(part.Get("type").String())
		if partType != "output_text" && partType != "text" {
			continue
		}
		if strings.TrimSpace(part.Get("text").String()) != "" {
			return true
		}
	}
	return false
}

func responsesReasoningItemHasVisibleText(item gjson.Result) bool {
	for _, path := range []string{"summary", "content"} {
		parts := item.Get(path)
		if !parts.Exists() || !parts.IsArray() {
			continue
		}
		for _, part := range parts.Array() {
			if strings.TrimSpace(part.Get("text").String()) != "" {
				return true
			}
		}
	}
	return strings.TrimSpace(item.Get("text").String()) != ""
}

func normalizeConversationText(raw string) string {
	raw = strings.ReplaceAll(raw, "\r\n", "\n")
	raw = strings.ReplaceAll(raw, "\r", "\n")
	lines := strings.Split(raw, "\n")
	out := make([]string, 0, len(lines))
	blank := false
	for _, line := range lines {
		trimmedRight := strings.TrimRight(line, " \t")
		trimmedLeft := strings.TrimLeft(trimmedRight, " \t")
		if trimmedLeft == "" {
			if blank {
				continue
			}
			out = append(out, "")
			blank = true
			continue
		}
		out = append(out, trimmedRight)
		blank = false
	}
	return strings.TrimSpace(strings.Join(out, "\n"))
}

func cloneJSONBytes(body []byte) []byte {
	if len(body) == 0 {
		return nil
	}
	out := make([]byte, len(body))
	copy(out, body)
	return out
}

func marshalOpenAIResponseBody(v any) []byte {
	if v == nil {
		return nil
	}
	b, err := json.Marshal(v)
	if err != nil {
		return nil
	}
	return b
}

func shouldRecordOpenAIConversationRetention(inboundEndpoint string) bool {
	path := strings.TrimRight(strings.TrimSpace(inboundEndpoint), "/")
	if path == "" {
		return false
	}
	if strings.HasSuffix(path, "/responses") || strings.Contains(path, "/responses/") {
		return true
	}
	if strings.HasSuffix(path, "/chat/completions") || strings.Contains(path, "/chat/completions/") {
		return true
	}
	return false
}

func normalizeOpenAIConversationRetentionEndpoint(inboundEndpoint string) string {
	path := strings.TrimSpace(inboundEndpoint)
	switch {
	case strings.HasSuffix(path, "/chat/completions") || strings.Contains(path, "/chat/completions/"):
		return "/v1/chat/completions"
	case strings.HasSuffix(path, "/responses") || strings.Contains(path, "/responses/"):
		return "/v1/responses"
	default:
		return path
	}
}
