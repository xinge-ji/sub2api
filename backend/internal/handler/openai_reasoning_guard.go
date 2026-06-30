package handler

import (
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"strings"
	"time"

	"github.com/Wei-Shaw/sub2api/internal/service"
	"github.com/gin-gonic/gin"
)

type openAIReasoningGuardCapture struct {
	writer               *openAIReasoningGuardWriter
	originalWriter       gin.ResponseWriter
	writerInstalled      bool
	responseWasCommitted bool
}

func (h *OpenAIGatewayHandler) beginOpenAIReasoningGuardCapture(c *gin.Context) *openAIReasoningGuardCapture {
	if h == nil || c == nil || c.Writer == nil || h.gatewayService == nil {
		return nil
	}
	capture := &openAIReasoningGuardCapture{
		originalWriter:       c.Writer,
		writer:               newOpenAIReasoningGuardWriter(c.Writer),
		responseWasCommitted: service.IsResponseCommitted(c),
	}
	c.Writer = capture.writer
	capture.writerInstalled = true
	return capture
}

func (h *OpenAIGatewayHandler) restoreOpenAIReasoningGuardWriter(c *gin.Context, capture *openAIReasoningGuardCapture) {
	if c == nil || capture == nil || !capture.writerInstalled {
		return
	}
	c.Writer = capture.originalWriter
	capture.writerInstalled = false
	if capture.responseWasCommitted {
		service.MarkResponseCommitted(c)
	} else {
		c.Set(service.ResponseCommittedKey, false)
	}
}

func (h *OpenAIGatewayHandler) openAIReasoningGuardClientOutputStarted(capture *openAIReasoningGuardCapture, localStarted bool) bool {
	if localStarted {
		return true
	}
	if capture == nil || capture.writer == nil {
		return false
	}
	return capture.writer.Written()
}

func (h *OpenAIGatewayHandler) openAIReasoningGuardRestoreAndReplay(c *gin.Context, capture *openAIReasoningGuardCapture) error {
	if capture == nil || capture.writer == nil {
		return nil
	}
	h.restoreOpenAIReasoningGuardWriter(c, capture)
	return capture.writer.replayToUnderlying()
}

func (h *OpenAIGatewayHandler) writeOpenAIReasoningGuardStreamIntercept(c *gin.Context, statusCode int, reasoningTokens int) error {
	if c == nil || c.Writer == nil {
		return nil
	}

	pathname := "/"
	if c.Request != nil && c.Request.URL != nil {
		pathname = c.Request.URL.Path
	}
	message := "codex retry gateway blocked suspicious reasoning response on " + pathname

	if inboundIsResponses(c) {
		payload, err := json.Marshal(gin.H{
			"type": "response.failed",
			"response": gin.H{
				"id":     synthesizeResponseID(c),
				"object": "response",
				"status": "failed",
				"error": gin.H{
					"type":             "codex_retry_gateway",
					"code":             "reasoning_guard_triggered",
					"message":          message,
					"reasoning_tokens": reasoningTokens,
					"status_code":      statusCode,
				},
			},
		})
		if err != nil {
			return err
		}
		_, err = fmt.Fprintf(c.Writer, "event: response.failed\ndata: %s\n\n", payload)
		return err
	}

	if GetInboundEndpoint(c) == EndpointMessages {
		_, err := fmt.Fprint(c.Writer, buildOpenAIReasoningGuardAnthropicStreamErrorSSE("codex_retry_gateway", message))
		return err
	}

	_, err := fmt.Fprint(c.Writer, buildOpenAIReasoningGuardChatStreamErrorSSE("reasoning_guard_triggered", message))
	return err
}

func buildOpenAIReasoningGuardAnthropicStreamErrorSSE(errType, message string) string {
	payload, err := json.Marshal(gin.H{
		"type": "error",
		"error": gin.H{
			"type":    errType,
			"message": message,
		},
	})
	if err != nil {
		return "event: error\ndata: {\"type\":\"error\",\"error\":{\"type\":\"" + errType + "\",\"message\":\"upstream error\"}}\n\n"
	}
	return "event: error\ndata: " + string(payload) + "\n\n"
}

func buildOpenAIReasoningGuardChatStreamErrorSSE(code, message string) string {
	payload, err := json.Marshal(gin.H{
		"error": gin.H{
			"type":    "invalid_request_error",
			"code":    code,
			"message": message,
		},
	})
	if err != nil {
		return "data: {\"error\":{\"type\":\"invalid_request_error\",\"code\":\"" + code + "\",\"message\":\"upstream error\"}}\n\n"
	}
	return "data: " + string(payload) + "\n\n"
}

func (h *OpenAIGatewayHandler) finalizeOpenAIReasoningGuardHTTP(
	ctx context.Context,
	c *gin.Context,
	capture *openAIReasoningGuardCapture,
	result *service.OpenAIForwardResult,
	apiKey *service.APIKey,
	account *service.Account,
	subscription *service.UserSubscription,
	inboundEndpoint string,
	upstreamEndpoint string,
) error {
	if capture == nil || capture.writer == nil || c == nil {
		return nil
	}

	settings, decision, err := h.gatewayService.EvaluateOpenAIReasoningGuard(ctx, result)
	if err != nil || decision == nil {
		_ = h.openAIReasoningGuardRestoreAndReplay(c, capture)
		return err
	}

	event := &service.OpenAIReasoningGuardEvent{
		UserID:             0,
		APIKeyID:           0,
		AccountID:          0,
		RequestID:          "",
		RequestedModel:     "",
		UpstreamModel:      "",
		InboundEndpoint:    strings.TrimSpace(inboundEndpoint),
		UpstreamEndpoint:   strings.TrimSpace(upstreamEndpoint),
		ServiceTier:        "",
		ReasoningEffort:    "",
		ResponseStatusCode: capture.writer.Status(),
		CreatedAt:          time.Now(),
	}
	if apiKey != nil {
		event.APIKeyID = apiKey.ID
		event.UserID = apiKey.UserID
		event.GroupID = apiKey.GroupID
	}
	if account != nil {
		event.AccountID = account.ID
	}
	if subscription != nil {
		event.SubscriptionID = &subscription.ID
	}
	if result != nil {
		event.RequestID = strings.TrimSpace(result.RequestID)
		event.RequestedModel = strings.TrimSpace(result.Model)
		event.UpstreamModel = strings.TrimSpace(result.UpstreamModel)
		event.Stream = result.Stream
		event.OpenAIWSMode = result.OpenAIWSMode
		event.ReasoningTokens = result.Usage.ReasoningTokens
		event.HasReasoningTokens = result.Usage.HasReasoningTokens
		if result.ServiceTier != nil {
			event.ServiceTier = strings.TrimSpace(*result.ServiceTier)
		}
		if result.ReasoningEffort != nil {
			event.ReasoningEffort = strings.TrimSpace(*result.ReasoningEffort)
		}
	}
	event.MatchedReasoningCode = decision.MatchedReasoningCode
	event.Intercepted = decision.Intercepted

	h.submitUsageRecordTask(ctx, func(taskCtx context.Context) {
		_ = h.gatewayService.RecordOpenAIReasoningGuardEvent(taskCtx, event)
	})

	if !decision.Intercepted || !settings.Enabled {
		return h.openAIReasoningGuardRestoreAndReplay(c, capture)
	}

	body := gin.H{
		"error": gin.H{
			"message":          "codex retry gateway blocked suspicious reasoning response on " + c.Request.URL.Path,
			"type":             "codex_retry_gateway",
			"code":             "reasoning_guard_triggered",
			"reasoning_tokens": result.Usage.ReasoningTokens,
			"status_code":      decision.InterceptStatusCode,
		},
	}
	payload, marshalErr := json.Marshal(body)
	if marshalErr != nil {
		return marshalErr
	}
	h.restoreOpenAIReasoningGuardWriter(c, capture)
	if capture.responseWasCommitted || (c.Writer != nil && c.Writer.Written()) {
		if err := h.writeOpenAIReasoningGuardStreamIntercept(c, decision.InterceptStatusCode, result.Usage.ReasoningTokens); err != nil {
			return err
		}
		if flusher, ok := c.Writer.(http.Flusher); ok {
			flusher.Flush()
		}
		service.MarkResponseCommitted(c)
		return nil
	}
	header := c.Writer.Header()
	for k := range header {
		delete(header, k)
	}
	header.Set("Content-Type", "application/json; charset=utf-8")
	c.Writer.WriteHeader(decision.InterceptStatusCode)
	service.MarkResponseCommitted(c)
	_, writeErr := c.Writer.Write(payload)
	return writeErr
}
