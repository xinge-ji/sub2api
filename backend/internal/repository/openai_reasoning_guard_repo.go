package repository

import (
	"context"
	"database/sql"
	"fmt"
	"strings"
	"time"

	"github.com/Wei-Shaw/sub2api/internal/service"
)

type openAIReasoningGuardRepository struct {
	db *sql.DB
}

func NewOpenAIReasoningGuardRepository(db *sql.DB) service.OpenAIReasoningGuardRepository {
	return &openAIReasoningGuardRepository{db: db}
}

func (r *openAIReasoningGuardRepository) CreateEvent(ctx context.Context, event *service.OpenAIReasoningGuardEvent) error {
	if r == nil || r.db == nil {
		return fmt.Errorf("nil openai reasoning guard repository")
	}
	if event == nil {
		return fmt.Errorf("nil openai reasoning guard event")
	}
	createdAt := event.CreatedAt
	if createdAt.IsZero() {
		createdAt = time.Now()
	}
	_, err := r.db.ExecContext(ctx, `
INSERT INTO openai_reasoning_guard_events (
  user_id,
  api_key_id,
  account_id,
  group_id,
  subscription_id,
  request_id,
  requested_model,
  upstream_model,
  inbound_endpoint,
  upstream_endpoint,
  service_tier,
  reasoning_effort,
  reasoning_tokens,
  has_reasoning_tokens,
  matched_reasoning_code,
  intercepted,
  response_status_code,
  stream,
  openai_ws_mode,
  created_at
) VALUES (
  $1,$2,$3,$4,$5,$6,$7,$8,$9,$10,$11,$12,$13,$14,$15,$16,$17,$18,$19,$20
)`,
		event.UserID,
		event.APIKeyID,
		event.AccountID,
		event.GroupID,
		event.SubscriptionID,
		strings.TrimSpace(event.RequestID),
		strings.TrimSpace(event.RequestedModel),
		strings.TrimSpace(event.UpstreamModel),
		strings.TrimSpace(event.InboundEndpoint),
		strings.TrimSpace(event.UpstreamEndpoint),
		strings.TrimSpace(event.ServiceTier),
		strings.TrimSpace(event.ReasoningEffort),
		event.ReasoningTokens,
		event.HasReasoningTokens,
		event.MatchedReasoningCode,
		event.Intercepted,
		event.ResponseStatusCode,
		event.Stream,
		event.OpenAIWSMode,
		createdAt,
	)
	return err
}

func (r *openAIReasoningGuardRepository) GetUserStats(ctx context.Context, filter service.OpenAIReasoningGuardStatsFilter) (*service.OpenAIReasoningGuardUserStats, error) {
	filter.APIKeyID = 0
	filter.AccountID = 0
	filter.GroupID = 0
	filter.Model = ""
	filter.RequestType = nil
	filter.Stream = nil
	return r.GetStats(ctx, filter)
}

func (r *openAIReasoningGuardRepository) GetStats(ctx context.Context, filter service.OpenAIReasoningGuardStatsFilter) (*service.OpenAIReasoningGuardUserStats, error) {
	if r == nil || r.db == nil {
		return nil, fmt.Errorf("nil openai reasoning guard repository")
	}
	stats := &service.OpenAIReasoningGuardUserStats{}
	whereClause, args := buildOpenAIReasoningGuardFilterQuery(filter)

	if err := scanSingleRow(
		ctx,
		r.db,
		fmt.Sprintf(`
SELECT
  COALESCE(COUNT(*), 0) AS request_count,
  COALESCE(COUNT(*) FILTER (WHERE matched_reasoning_code IS NOT NULL), 0) AS match_count,
  COALESCE(
    COUNT(*) FILTER (WHERE matched_reasoning_code IS NOT NULL)::double precision / NULLIF(COUNT(*), 0),
    0
  ) AS match_ratio,
  COALESCE(COUNT(*) FILTER (WHERE intercepted), 0) AS intercept_count,
  COALESCE(
    COUNT(*) FILTER (WHERE intercepted)::double precision / NULLIF(COUNT(*), 0),
    0
  ) AS intercept_ratio
FROM openai_reasoning_guard_events
WHERE %s
`, whereClause),
		args,
		&stats.Summary.RequestCount,
		&stats.Summary.MatchCount,
		&stats.Summary.MatchRatio,
		&stats.Summary.InterceptCount,
		&stats.Summary.InterceptRatio,
	); err != nil {
		return nil, err
	}

	dateFormat := safeDateFormat(filter.Granularity)
	trendRows, err := r.db.QueryContext(ctx, fmt.Sprintf(`
SELECT
  TO_CHAR(created_at, '%s') AS date,
  COUNT(*) AS request_count,
  COALESCE(COUNT(*) FILTER (WHERE matched_reasoning_code IS NOT NULL), 0) AS match_count,
  COALESCE(
    COUNT(*) FILTER (WHERE matched_reasoning_code IS NOT NULL)::double precision / NULLIF(COUNT(*), 0),
    0
  ) AS match_ratio,
  COALESCE(COUNT(*) FILTER (WHERE intercepted), 0) AS intercept_count,
  COALESCE(
    COUNT(*) FILTER (WHERE intercepted)::double precision / NULLIF(COUNT(*), 0),
    0
  ) AS intercept_ratio
FROM openai_reasoning_guard_events
WHERE %s
GROUP BY date
ORDER BY date ASC
`, dateFormat, whereClause), args...)
	if err != nil {
		return nil, err
	}
	defer func() { _ = trendRows.Close() }()

	for trendRows.Next() {
		var point service.OpenAIReasoningGuardTrendPoint
		if err := trendRows.Scan(&point.Date, &point.RequestCount, &point.MatchCount, &point.MatchRatio, &point.InterceptCount, &point.InterceptRatio); err != nil {
			return nil, err
		}
		stats.Trend = append(stats.Trend, point)
	}
	if err := trendRows.Err(); err != nil {
		return nil, err
	}

	models, err := r.queryBreakdown(ctx, filter, "COALESCE(NULLIF(requested_model, ''), NULLIF(upstream_model, ''), 'unknown')")
	if err != nil {
		return nil, err
	}
	stats.Models = models

	reasoningEfforts, err := r.queryBreakdown(ctx, filter, "COALESCE(NULLIF(reasoning_effort, ''), 'unknown')")
	if err != nil {
		return nil, err
	}
	stats.ReasoningEfforts = reasoningEfforts

	modelEfforts, err := r.queryModelEffortBreakdown(ctx, filter)
	if err != nil {
		return nil, err
	}
	stats.ModelEfforts = modelEfforts

	modelEffortTrend, err := r.queryModelEffortTrend(ctx, filter)
	if err != nil {
		return nil, err
	}
	stats.ModelEffortTrend = modelEffortTrend

	return stats, nil
}

func (r *openAIReasoningGuardRepository) queryBreakdown(
	ctx context.Context,
	filter service.OpenAIReasoningGuardStatsFilter,
	expr string,
) ([]service.OpenAIReasoningGuardBreakdownItem, error) {
	rows, err := r.db.QueryContext(ctx, fmt.Sprintf(`
SELECT
  %s AS key,
  COUNT(*) AS request_count,
  COALESCE(COUNT(*) FILTER (WHERE matched_reasoning_code IS NOT NULL), 0) AS match_count,
  COALESCE(
    COUNT(*) FILTER (WHERE matched_reasoning_code IS NOT NULL)::double precision / NULLIF(COUNT(*), 0),
    0
  ) AS match_ratio,
  COALESCE(COUNT(*) FILTER (WHERE intercepted), 0) AS intercept_count,
  COALESCE(
    COUNT(*) FILTER (WHERE intercepted)::double precision / NULLIF(COUNT(*), 0),
    0
  ) AS intercept_ratio
FROM openai_reasoning_guard_events
WHERE %s
GROUP BY key
ORDER BY request_count DESC, intercept_count DESC, key ASC
`, expr, buildOpenAIReasoningGuardFilterPlaceholder(filter)), buildOpenAIReasoningGuardFilterArgs(filter)...)
	if err != nil {
		return nil, err
	}
	defer func() { _ = rows.Close() }()

	result := make([]service.OpenAIReasoningGuardBreakdownItem, 0)
	for rows.Next() {
		var item service.OpenAIReasoningGuardBreakdownItem
		if err := rows.Scan(&item.Key, &item.RequestCount, &item.MatchCount, &item.MatchRatio, &item.InterceptCount, &item.InterceptRatio); err != nil {
			return nil, err
		}
		result = append(result, item)
	}
	if err := rows.Err(); err != nil {
		return nil, err
	}
	return result, nil
}

func (r *openAIReasoningGuardRepository) queryModelEffortBreakdown(
	ctx context.Context,
	filter service.OpenAIReasoningGuardStatsFilter,
) ([]service.OpenAIReasoningGuardComboBreakdownItem, error) {
	whereClause := buildOpenAIReasoningGuardFilterPlaceholder(filter)
	args := buildOpenAIReasoningGuardFilterArgs(filter)
	rows, err := r.db.QueryContext(ctx, `
SELECT
  COALESCE(NULLIF(requested_model, ''), NULLIF(upstream_model, ''), 'unknown') AS model,
  COALESCE(NULLIF(reasoning_effort, ''), 'unknown') AS reasoning_effort,
  COUNT(*) AS request_count,
  COALESCE(COUNT(*) FILTER (WHERE matched_reasoning_code IS NOT NULL), 0) AS match_count,
  COALESCE(
    COUNT(*) FILTER (WHERE matched_reasoning_code IS NOT NULL)::double precision / NULLIF(COUNT(*), 0),
    0
  ) AS match_ratio,
  COALESCE(COUNT(*) FILTER (WHERE intercepted), 0) AS intercept_count,
  COALESCE(
    COUNT(*) FILTER (WHERE intercepted)::double precision / NULLIF(COUNT(*), 0),
    0
  ) AS intercept_ratio
FROM openai_reasoning_guard_events
WHERE `+whereClause+`
GROUP BY model, reasoning_effort
ORDER BY request_count DESC, intercept_count DESC, model ASC, reasoning_effort ASC
`, args...)
	if err != nil {
		return nil, err
	}
	defer func() { _ = rows.Close() }()

	result := make([]service.OpenAIReasoningGuardComboBreakdownItem, 0)
	for rows.Next() {
		var item service.OpenAIReasoningGuardComboBreakdownItem
		if err := rows.Scan(
			&item.Model,
			&item.ReasoningEffort,
			&item.RequestCount,
			&item.MatchCount,
			&item.MatchRatio,
			&item.InterceptCount,
			&item.InterceptRatio,
		); err != nil {
			return nil, err
		}
		item.Key = buildOpenAIReasoningGuardComboKey(item.Model, item.ReasoningEffort)
		result = append(result, item)
	}
	if err := rows.Err(); err != nil {
		return nil, err
	}
	return result, nil
}

func (r *openAIReasoningGuardRepository) queryModelEffortTrend(
	ctx context.Context,
	filter service.OpenAIReasoningGuardStatsFilter,
) ([]service.OpenAIReasoningGuardComboTrendPoint, error) {
	dateFormat := safeDateFormat(filter.Granularity)
	whereClause := buildOpenAIReasoningGuardFilterPlaceholder(filter)
	args := buildOpenAIReasoningGuardFilterArgs(filter)
	rows, err := r.db.QueryContext(ctx, fmt.Sprintf(`
SELECT
  TO_CHAR(created_at, '%s') AS date,
  COALESCE(NULLIF(requested_model, ''), NULLIF(upstream_model, ''), 'unknown') AS model,
  COALESCE(NULLIF(reasoning_effort, ''), 'unknown') AS reasoning_effort,
  COUNT(*) AS request_count,
  COALESCE(COUNT(*) FILTER (WHERE matched_reasoning_code IS NOT NULL), 0) AS match_count,
  COALESCE(
    COUNT(*) FILTER (WHERE matched_reasoning_code IS NOT NULL)::double precision / NULLIF(COUNT(*), 0),
    0
  ) AS match_ratio,
  COALESCE(COUNT(*) FILTER (WHERE intercepted), 0) AS intercept_count,
  COALESCE(
    COUNT(*) FILTER (WHERE intercepted)::double precision / NULLIF(COUNT(*), 0),
    0
  ) AS intercept_ratio
FROM openai_reasoning_guard_events
WHERE %s
GROUP BY 1, 2, 3
ORDER BY 1 ASC, 4 DESC, 5 DESC, 2 ASC, 3 ASC
`, dateFormat, whereClause), args...)
	if err != nil {
		return nil, err
	}
	defer func() { _ = rows.Close() }()

	result := make([]service.OpenAIReasoningGuardComboTrendPoint, 0)
	for rows.Next() {
		var point service.OpenAIReasoningGuardComboTrendPoint
		if err := rows.Scan(
			&point.Date,
			&point.Model,
			&point.ReasoningEffort,
			&point.RequestCount,
			&point.MatchCount,
			&point.MatchRatio,
			&point.InterceptCount,
			&point.InterceptRatio,
		); err != nil {
			return nil, err
		}
		point.Key = buildOpenAIReasoningGuardComboKey(point.Model, point.ReasoningEffort)
		result = append(result, point)
	}
	if err := rows.Err(); err != nil {
		return nil, err
	}
	return result, nil
}

func buildOpenAIReasoningGuardComboKey(model, reasoningEffort string) string {
	return model + " / " + reasoningEffort
}

func buildOpenAIReasoningGuardFilterQuery(filter service.OpenAIReasoningGuardStatsFilter) (string, []any) {
	return buildOpenAIReasoningGuardFilterPlaceholder(filter), buildOpenAIReasoningGuardFilterArgs(filter)
}

func buildOpenAIReasoningGuardFilterPlaceholder(filter service.OpenAIReasoningGuardStatsFilter) string {
	parts := []string{"created_at >= $1", "created_at < $2"}
	argIndex := 3
	if filter.UserID > 0 {
		parts = append(parts, fmt.Sprintf("user_id = $%d", argIndex))
		argIndex++
	}
	if filter.APIKeyID > 0 {
		parts = append(parts, fmt.Sprintf("api_key_id = $%d", argIndex))
		argIndex++
	}
	if filter.AccountID > 0 {
		parts = append(parts, fmt.Sprintf("account_id = $%d", argIndex))
		argIndex++
	}
	if filter.GroupID > 0 {
		parts = append(parts, fmt.Sprintf("group_id = $%d", argIndex))
		argIndex++
	}
	if model := strings.TrimSpace(filter.Model); model != "" {
		parts = append(parts, fmt.Sprintf("(requested_model = $%d OR upstream_model = $%d)", argIndex, argIndex))
		argIndex++
	}
	if filter.RequestType != nil {
		requestType := service.RequestTypeFromInt16(*filter.RequestType)
		switch requestType {
		case service.RequestTypeSync:
			parts = append(parts, "openai_ws_mode = false", "stream = false")
		case service.RequestTypeStream:
			parts = append(parts, "openai_ws_mode = false", "stream = true")
		case service.RequestTypeWSV2:
			parts = append(parts, "openai_ws_mode = true")
		case service.RequestTypeCyberBlocked:
			parts = append(parts, "1 = 0")
		}
	} else if filter.Stream != nil {
		if *filter.Stream {
			parts = append(parts, "stream = true")
		} else {
			parts = append(parts, "stream = false", "openai_ws_mode = false")
		}
	}
	return strings.Join(parts, " AND ")
}

func buildOpenAIReasoningGuardFilterArgs(filter service.OpenAIReasoningGuardStatsFilter) []any {
	args := []any{filter.StartTime, filter.EndTime}
	if filter.UserID > 0 {
		args = append(args, filter.UserID)
	}
	if filter.APIKeyID > 0 {
		args = append(args, filter.APIKeyID)
	}
	if filter.AccountID > 0 {
		args = append(args, filter.AccountID)
	}
	if filter.GroupID > 0 {
		args = append(args, filter.GroupID)
	}
	if model := strings.TrimSpace(filter.Model); model != "" {
		args = append(args, model)
	}
	return args
}
