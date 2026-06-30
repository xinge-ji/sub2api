package service

import (
	"context"
	"time"
)

// OpenAIReasoningGuardEvent records one inspected OpenAI HTTP response.
type OpenAIReasoningGuardEvent struct {
	UserID               int64
	APIKeyID             int64
	AccountID            int64
	GroupID              *int64
	SubscriptionID       *int64
	RequestID            string
	RequestedModel       string
	UpstreamModel        string
	InboundEndpoint      string
	UpstreamEndpoint     string
	ServiceTier          string
	ReasoningEffort      string
	ReasoningTokens      int
	HasReasoningTokens   bool
	MatchedReasoningCode *int
	Intercepted          bool
	ResponseStatusCode   int
	Stream               bool
	OpenAIWSMode         bool
	CreatedAt            time.Time
}

type OpenAIReasoningGuardStatsFilter struct {
	UserID      int64
	APIKeyID    int64
	AccountID   int64
	GroupID     int64
	Model       string
	RequestType *int16
	Stream      *bool
	StartTime   time.Time
	EndTime     time.Time
	Granularity string
}

type OpenAIReasoningGuardSummary struct {
	RequestCount   int64   `json:"request_count"`
	MatchCount     int64   `json:"match_count"`
	MatchRatio     float64 `json:"match_ratio"`
	InterceptCount int64   `json:"intercept_count"`
	InterceptRatio float64 `json:"intercept_ratio"`
}

type OpenAIReasoningGuardTrendPoint struct {
	Date           string  `json:"date"`
	RequestCount   int64   `json:"request_count"`
	MatchCount     int64   `json:"match_count"`
	MatchRatio     float64 `json:"match_ratio"`
	InterceptCount int64   `json:"intercept_count"`
	InterceptRatio float64 `json:"intercept_ratio"`
}

type OpenAIReasoningGuardBreakdownItem struct {
	Key            string  `json:"key"`
	RequestCount   int64   `json:"request_count"`
	MatchCount     int64   `json:"match_count"`
	MatchRatio     float64 `json:"match_ratio"`
	InterceptCount int64   `json:"intercept_count"`
	InterceptRatio float64 `json:"intercept_ratio"`
}

type OpenAIReasoningGuardComboBreakdownItem struct {
	Model           string  `json:"model"`
	ReasoningEffort string  `json:"reasoning_effort"`
	Key             string  `json:"key"`
	RequestCount    int64   `json:"request_count"`
	MatchCount      int64   `json:"match_count"`
	MatchRatio      float64 `json:"match_ratio"`
	InterceptCount  int64   `json:"intercept_count"`
	InterceptRatio  float64 `json:"intercept_ratio"`
}

type OpenAIReasoningGuardComboTrendPoint struct {
	Date            string  `json:"date"`
	Model           string  `json:"model"`
	ReasoningEffort string  `json:"reasoning_effort"`
	Key             string  `json:"key"`
	RequestCount    int64   `json:"request_count"`
	MatchCount      int64   `json:"match_count"`
	MatchRatio      float64 `json:"match_ratio"`
	InterceptCount  int64   `json:"intercept_count"`
	InterceptRatio  float64 `json:"intercept_ratio"`
}

type OpenAIReasoningGuardRuntimeView struct {
	Enabled             bool                           `json:"enabled"`
	InterceptStatusCode int                            `json:"intercept_status_code"`
	Rules               []OpenAIReasoningGuardModelRule `json:"rules"`
}

type OpenAIReasoningGuardUserStats struct {
	Summary          OpenAIReasoningGuardSummary              `json:"summary"`
	Trend            []OpenAIReasoningGuardTrendPoint         `json:"trend"`
	Models           []OpenAIReasoningGuardBreakdownItem      `json:"models"`
	ReasoningEfforts []OpenAIReasoningGuardBreakdownItem      `json:"reasoning_efforts"`
	ModelEfforts     []OpenAIReasoningGuardComboBreakdownItem `json:"model_efforts"`
	ModelEffortTrend []OpenAIReasoningGuardComboTrendPoint    `json:"model_effort_trend"`
	Runtime          OpenAIReasoningGuardRuntimeView          `json:"runtime"`
}

type OpenAIReasoningGuardRepository interface {
	CreateEvent(ctx context.Context, event *OpenAIReasoningGuardEvent) error
	GetUserStats(ctx context.Context, filter OpenAIReasoningGuardStatsFilter) (*OpenAIReasoningGuardUserStats, error)
	GetStats(ctx context.Context, filter OpenAIReasoningGuardStatsFilter) (*OpenAIReasoningGuardUserStats, error)
}
