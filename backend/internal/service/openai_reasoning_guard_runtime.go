package service

import (
	"context"
	"strings"
)

type OpenAIReasoningGuardDecision struct {
	Enabled              bool
	Matched              bool
	Intercepted          bool
	MatchedReasoningCode *int
	InterceptStatusCode  int
}

func (s *OpenAIGatewayService) EvaluateOpenAIReasoningGuard(ctx context.Context, result *OpenAIForwardResult) (*OpenAIReasoningGuardSettings, *OpenAIReasoningGuardDecision, error) {
	settings := DefaultOpenAIReasoningGuardSettings()
	if s != nil && s.settingService != nil {
		fetched, err := s.settingService.GetOpenAIReasoningGuardSettings(ctx)
		if err != nil {
			return nil, nil, err
		}
		if fetched != nil {
			settings = fetched
		}
	}
	decision := &OpenAIReasoningGuardDecision{
		Enabled:             settings.Enabled,
		InterceptStatusCode: settings.InterceptStatusCode,
	}
	if result != nil && result.Usage.HasReasoningTokens {
		model := strings.ToLower(strings.TrimSpace(firstNonEmpty(result.Model, result.UpstreamModel)))
		for _, rule := range settings.Rules {
			if strings.ToLower(strings.TrimSpace(rule.Model)) != model {
				continue
			}
			for _, code := range rule.Codes {
				if code == result.Usage.ReasoningTokens {
					matched := code
					decision.MatchedReasoningCode = &matched
					decision.Matched = true
					break
				}
			}
			break
		}
	}
	decision.Intercepted = decision.Enabled && decision.Matched
	return settings, decision, nil
}

func (s *OpenAIGatewayService) RecordOpenAIReasoningGuardEvent(ctx context.Context, event *OpenAIReasoningGuardEvent) error {
	if s == nil || s.reasoningGuardRepo == nil || event == nil {
		return nil
	}
	return s.reasoningGuardRepo.CreateEvent(ctx, event)
}
