package repository

import (
	"context"
	"database/sql"
	"fmt"
	"strings"
	"time"

	"github.com/Wei-Shaw/sub2api/internal/pkg/pagination"
	"github.com/Wei-Shaw/sub2api/internal/service"
)

type openAIConversationRetentionRepository struct {
	db *sql.DB
}

func NewOpenAIConversationRetentionRepository(db *sql.DB) service.OpenAIConversationRetentionRepository {
	return &openAIConversationRetentionRepository{db: db}
}

func (r *openAIConversationRetentionRepository) CreateTurn(ctx context.Context, input *service.OpenAIConversationRetentionRecordInput) (*service.OpenAIConversationRetentionSession, *service.OpenAIConversationRetentionTurn, error) {
	if r == nil || r.db == nil {
		return nil, nil, fmt.Errorf("nil openai conversation retention repository")
	}
	if input == nil {
		return nil, nil, fmt.Errorf("nil openai conversation retention input")
	}

	now := input.CreatedAt.UTC()
	if now.IsZero() {
		now = time.Now().UTC()
	}
	sessionKey := service.BuildOpenAIConversationSessionKey(input.UserID, input.ClientSessionID, input.RequestID)

	tx, err := r.db.BeginTx(ctx, nil)
	if err != nil {
		return nil, nil, err
	}
	defer func() {
		if err != nil {
			_ = tx.Rollback()
		}
	}()

	var session service.OpenAIConversationRetentionSession
	querySession := `
SELECT
  id,
  user_id,
  api_key_id,
  account_id,
  group_id,
  session_key,
  client_session_id,
  requested_model,
  upstream_model,
  reasoning_effort,
  first_request_id,
  last_request_id,
  first_response_id,
  last_response_id,
  turn_count,
  created_at,
  updated_at
FROM openai_retained_sessions
WHERE session_key = $1
FOR UPDATE
`
	err = tx.QueryRowContext(ctx, querySession, sessionKey).Scan(
		&session.ID,
		&session.UserID,
		&session.APIKeyID,
		&session.AccountID,
		&session.GroupID,
		&session.SessionKey,
		&session.ClientSessionID,
		&session.RequestedModel,
		&session.UpstreamModel,
		&session.ReasoningEffort,
		&session.FirstRequestID,
		&session.LastRequestID,
		&session.FirstResponseID,
		&session.LastResponseID,
		&session.TurnCount,
		&session.CreatedAt,
		&session.UpdatedAt,
	)
	if err != nil {
		if err != sql.ErrNoRows {
			return nil, nil, err
		}
		insertSession := `
INSERT INTO openai_retained_sessions (
  user_id,
  api_key_id,
  account_id,
  group_id,
  session_key,
  client_session_id,
  requested_model,
  upstream_model,
  reasoning_effort,
  first_request_id,
  last_request_id,
  first_response_id,
  last_response_id,
  turn_count,
  created_at,
  updated_at
) VALUES (
  $1,$2,$3,$4,$5,$6,$7,$8,$9,$10,$11,$12,$13,$14,$15,$16
)
RETURNING
  id,
  user_id,
  api_key_id,
  account_id,
  group_id,
  session_key,
  client_session_id,
  requested_model,
  upstream_model,
  reasoning_effort,
  first_request_id,
  last_request_id,
  first_response_id,
  last_response_id,
  turn_count,
  created_at,
  updated_at
`
		err = tx.QueryRowContext(
			ctx,
			insertSession,
			input.UserID,
			input.APIKeyID,
			input.AccountID,
			input.GroupID,
			sessionKey,
			strings.TrimSpace(input.ClientSessionID),
			strings.TrimSpace(input.RequestedModel),
			strings.TrimSpace(input.UpstreamModel),
			strings.TrimSpace(input.ReasoningEffort),
			strings.TrimSpace(input.RequestID),
			strings.TrimSpace(input.RequestID),
			strings.TrimSpace(input.ResponseID),
			strings.TrimSpace(input.ResponseID),
			0,
			now,
			now,
		).Scan(
			&session.ID,
			&session.UserID,
			&session.APIKeyID,
			&session.AccountID,
			&session.GroupID,
			&session.SessionKey,
			&session.ClientSessionID,
			&session.RequestedModel,
			&session.UpstreamModel,
			&session.ReasoningEffort,
			&session.FirstRequestID,
			&session.LastRequestID,
			&session.FirstResponseID,
			&session.LastResponseID,
			&session.TurnCount,
			&session.CreatedAt,
			&session.UpdatedAt,
		)
		if err != nil {
			return nil, nil, err
		}
	} else {
		updateSession := `
UPDATE openai_retained_sessions
SET
  api_key_id = $2,
  account_id = $3,
  group_id = $4,
  client_session_id = $5,
  requested_model = $6,
  upstream_model = $7,
  reasoning_effort = $8,
  last_request_id = $9,
  last_response_id = $10,
  updated_at = $11
WHERE id = $1
`
		if _, err = tx.ExecContext(
			ctx,
			updateSession,
			session.ID,
			input.APIKeyID,
			input.AccountID,
			input.GroupID,
			strings.TrimSpace(input.ClientSessionID),
			strings.TrimSpace(input.RequestedModel),
			strings.TrimSpace(input.UpstreamModel),
			strings.TrimSpace(input.ReasoningEffort),
			strings.TrimSpace(input.RequestID),
			strings.TrimSpace(input.ResponseID),
			now,
		); err != nil {
			return nil, nil, err
		}
		session.APIKeyID = input.APIKeyID
		session.AccountID = input.AccountID
		session.GroupID = input.GroupID
		session.ClientSessionID = strings.TrimSpace(input.ClientSessionID)
		session.RequestedModel = strings.TrimSpace(input.RequestedModel)
		session.UpstreamModel = strings.TrimSpace(input.UpstreamModel)
		session.ReasoningEffort = strings.TrimSpace(input.ReasoningEffort)
		session.LastRequestID = strings.TrimSpace(input.RequestID)
		session.LastResponseID = strings.TrimSpace(input.ResponseID)
		session.UpdatedAt = now
	}

	turn := &service.OpenAIConversationRetentionTurn{
		SessionID:           session.ID,
		UserID:              input.UserID,
		APIKeyID:            input.APIKeyID,
		AccountID:           input.AccountID,
		GroupID:             input.GroupID,
		RequestID:           strings.TrimSpace(input.RequestID),
		ResponseID:          strings.TrimSpace(input.ResponseID),
		TurnIndex:           session.TurnCount + 1,
		RequestedModel:      strings.TrimSpace(input.RequestedModel),
		UpstreamModel:       strings.TrimSpace(input.UpstreamModel),
		ReasoningEffort:     strings.TrimSpace(input.ReasoningEffort),
		Stream:              input.Stream,
		RequestType:         input.RequestType.Normalize(),
		InboundEndpoint:     strings.TrimSpace(input.InboundEndpoint),
		UpstreamEndpoint:    strings.TrimSpace(input.UpstreamEndpoint),
		UserInputText:       input.UserInputText,
		AssistantOutputText: input.AssistantText,
		CreatedAt:           now,
	}
	insertTurn := `
INSERT INTO openai_retained_turns (
  session_id,
  user_id,
  api_key_id,
  account_id,
  group_id,
  request_id,
  response_id,
  turn_index,
  requested_model,
  upstream_model,
  reasoning_effort,
  stream,
  request_type,
  inbound_endpoint,
  upstream_endpoint,
  user_input_text,
  assistant_output_text,
  created_at
) VALUES (
  $1,$2,$3,$4,$5,$6,$7,$8,$9,$10,$11,$12,$13,$14,$15,$16,$17,$18
)
RETURNING id
`
	if err = tx.QueryRowContext(
		ctx,
		insertTurn,
		turn.SessionID,
		turn.UserID,
		turn.APIKeyID,
		turn.AccountID,
		turn.GroupID,
		turn.RequestID,
		turn.ResponseID,
		turn.TurnIndex,
		turn.RequestedModel,
		turn.UpstreamModel,
		turn.ReasoningEffort,
		turn.Stream,
		int16(turn.RequestType),
		turn.InboundEndpoint,
		turn.UpstreamEndpoint,
		turn.UserInputText,
		turn.AssistantOutputText,
		turn.CreatedAt,
	).Scan(&turn.ID); err != nil {
		return nil, nil, err
	}

	if _, err = tx.ExecContext(
		ctx,
		"UPDATE openai_retained_sessions SET turn_count = $2, updated_at = $3 WHERE id = $1",
		session.ID,
		turn.TurnIndex,
		now,
	); err != nil {
		return nil, nil, err
	}
	session.TurnCount = turn.TurnIndex
	session.UpdatedAt = now

	if err = tx.Commit(); err != nil {
		return nil, nil, err
	}
	return &session, turn, nil
}

func (r *openAIConversationRetentionRepository) GetConversationByRequestID(ctx context.Context, requestID string) (*service.OpenAIConversationRetentionView, error) {
	if r == nil || r.db == nil {
		return nil, fmt.Errorf("nil openai conversation retention repository")
	}
	requestID = strings.TrimSpace(requestID)
	if requestID == "" {
		return nil, nil
	}

	querySession := `
SELECT
  s.id,
  s.user_id,
  s.api_key_id,
  s.account_id,
  s.group_id,
  s.session_key,
  s.client_session_id,
  s.requested_model,
  s.upstream_model,
  s.reasoning_effort,
  s.first_request_id,
  s.last_request_id,
  s.first_response_id,
  s.last_response_id,
  s.turn_count,
  s.created_at,
  s.updated_at
FROM openai_retained_sessions s
JOIN openai_retained_turns t ON t.session_id = s.id
WHERE t.request_id = $1
LIMIT 1
`
	session := &service.OpenAIConversationRetentionSession{}
	err := r.db.QueryRowContext(ctx, querySession, requestID).Scan(
		&session.ID,
		&session.UserID,
		&session.APIKeyID,
		&session.AccountID,
		&session.GroupID,
		&session.SessionKey,
		&session.ClientSessionID,
		&session.RequestedModel,
		&session.UpstreamModel,
		&session.ReasoningEffort,
		&session.FirstRequestID,
		&session.LastRequestID,
		&session.FirstResponseID,
		&session.LastResponseID,
		&session.TurnCount,
		&session.CreatedAt,
		&session.UpdatedAt,
	)
	if err != nil {
		if err == sql.ErrNoRows {
			return nil, nil
		}
		return nil, err
	}

	rows, err := r.db.QueryContext(ctx, `
SELECT
  id,
  session_id,
  user_id,
  api_key_id,
  account_id,
  group_id,
  request_id,
  response_id,
  turn_index,
  requested_model,
  upstream_model,
  reasoning_effort,
  stream,
  request_type,
  inbound_endpoint,
  upstream_endpoint,
  user_input_text,
  assistant_output_text,
  created_at
FROM openai_retained_turns
WHERE session_id = $1
ORDER BY turn_index ASC, created_at ASC
`, session.ID)
	if err != nil {
		return nil, err
	}
	defer func() { _ = rows.Close() }()

	turns := make([]*service.OpenAIConversationRetentionTurn, 0, session.TurnCount)
	for rows.Next() {
		turn := &service.OpenAIConversationRetentionTurn{}
		var requestType int16
		if err := rows.Scan(
			&turn.ID,
			&turn.SessionID,
			&turn.UserID,
			&turn.APIKeyID,
			&turn.AccountID,
			&turn.GroupID,
			&turn.RequestID,
			&turn.ResponseID,
			&turn.TurnIndex,
			&turn.RequestedModel,
			&turn.UpstreamModel,
			&turn.ReasoningEffort,
			&turn.Stream,
			&requestType,
			&turn.InboundEndpoint,
			&turn.UpstreamEndpoint,
			&turn.UserInputText,
			&turn.AssistantOutputText,
			&turn.CreatedAt,
		); err != nil {
			return nil, err
		}
		turn.RequestType = service.RequestTypeFromInt16(requestType)
		turns = append(turns, turn)
	}
	if err := rows.Err(); err != nil {
		return nil, err
	}
	return &service.OpenAIConversationRetentionView{
		Session: session,
		Turns:   turns,
	}, nil
}

func (r *openAIConversationRetentionRepository) ListConversations(ctx context.Context, filters service.OpenAIConversationRetentionListFilters, params pagination.PaginationParams) ([]*service.OpenAIConversationRetentionListItem, *pagination.PaginationResult, error) {
	if r == nil || r.db == nil {
		return nil, nil, fmt.Errorf("nil openai conversation retention repository")
	}

	whereParts := make([]string, 0, 12)
	args := make([]any, 0, 16)
	argIndex := 1

	addWhere := func(clause string, value any) {
		whereParts = append(whereParts, clause)
		args = append(args, value)
		argIndex++
	}

	if filters.UserID > 0 {
		addWhere(fmt.Sprintf("s.user_id = $%d", argIndex), filters.UserID)
	}
	if filters.APIKeyID > 0 {
		addWhere(fmt.Sprintf("s.api_key_id = $%d", argIndex), filters.APIKeyID)
	}
	if filters.AccountID > 0 {
		addWhere(fmt.Sprintf("COALESCE(s.account_id, 0) = $%d", argIndex), filters.AccountID)
	}
	if filters.GroupID > 0 {
		addWhere(fmt.Sprintf("COALESCE(s.group_id, 0) = $%d", argIndex), filters.GroupID)
	}
	if model := strings.TrimSpace(filters.Model); model != "" {
		addWhere(fmt.Sprintf("(s.requested_model = $%d OR s.upstream_model = $%d)", argIndex, argIndex), model)
	}
	if effort := strings.TrimSpace(filters.ReasoningEffort); effort != "" {
		addWhere(fmt.Sprintf("s.reasoning_effort = $%d", argIndex), effort)
	}
	if filters.RequestType != nil {
		addWhere(fmt.Sprintf(`EXISTS (
  SELECT 1
  FROM openai_retained_turns rt
  WHERE rt.session_id = s.id AND rt.request_type = $%d
)`, argIndex), *filters.RequestType)
	}
	if filters.Stream != nil {
		addWhere(fmt.Sprintf(`EXISTS (
  SELECT 1
  FROM openai_retained_turns rt
  WHERE rt.session_id = s.id AND rt.stream = $%d
)`, argIndex), *filters.Stream)
	}
	if filters.StartTime != nil {
		addWhere(fmt.Sprintf("s.updated_at >= $%d", argIndex), filters.StartTime.UTC())
	}
	if filters.EndTime != nil {
		addWhere(fmt.Sprintf("s.updated_at < $%d", argIndex), filters.EndTime.UTC())
	}

	whereSQL := ""
	if len(whereParts) > 0 {
		whereSQL = "WHERE " + strings.Join(whereParts, " AND ")
	}

	countQuery := `
SELECT COUNT(*)
FROM openai_retained_sessions s
` + whereSQL

	var total int64
	if err := r.db.QueryRowContext(ctx, countQuery, args...).Scan(&total); err != nil {
		return nil, nil, err
	}

	limit := params.Limit()
	offset := params.Offset()
	sortOrder := params.NormalizedSortOrder(pagination.SortOrderDesc)
	sortColumn := "s.updated_at"
	switch strings.TrimSpace(params.SortBy) {
	case "created_at":
		sortColumn = "s.created_at"
	case "turn_count":
		sortColumn = "s.turn_count"
	case "requested_model":
		sortColumn = "s.requested_model"
	case "reasoning_effort":
		sortColumn = "s.reasoning_effort"
	}

	listArgs := append([]any{}, args...)
	listArgs = append(listArgs, limit, offset)
	limitPos := len(args) + 1
	offsetPos := len(args) + 2

	query := `
SELECT
  s.id,
  s.user_id,
  s.api_key_id,
  s.account_id,
  s.group_id,
  s.session_key,
  s.client_session_id,
  s.requested_model,
  s.upstream_model,
  s.reasoning_effort,
  s.first_request_id,
  s.last_request_id,
  s.first_response_id,
  s.last_response_id,
  s.turn_count,
  s.created_at,
  s.updated_at,
  COALESCE(u.email, ''),
  COALESCE(k.name, ''),
  COALESCE(a.name, ''),
  COALESCE(g.name, ''),
  lt.id,
  lt.session_id,
  lt.user_id,
  lt.api_key_id,
  lt.account_id,
  lt.group_id,
  lt.request_id,
  lt.response_id,
  lt.turn_index,
  lt.requested_model,
  lt.upstream_model,
  lt.reasoning_effort,
  lt.stream,
  lt.request_type,
  lt.inbound_endpoint,
  lt.upstream_endpoint,
  lt.user_input_text,
  lt.assistant_output_text,
  lt.created_at
FROM openai_retained_sessions s
LEFT JOIN users u ON u.id = s.user_id
LEFT JOIN api_keys k ON k.id = s.api_key_id
LEFT JOIN accounts a ON a.id = s.account_id
LEFT JOIN groups g ON g.id = s.group_id
LEFT JOIN LATERAL (
  SELECT
    t.id,
    t.session_id,
    t.user_id,
    t.api_key_id,
    t.account_id,
    t.group_id,
    t.request_id,
    t.response_id,
    t.turn_index,
    t.requested_model,
    t.upstream_model,
    t.reasoning_effort,
    t.stream,
    t.request_type,
    t.inbound_endpoint,
    t.upstream_endpoint,
    t.user_input_text,
    t.assistant_output_text,
    t.created_at
  FROM openai_retained_turns t
  WHERE t.session_id = s.id
  ORDER BY t.turn_index DESC, t.created_at DESC, t.id DESC
  LIMIT 1
) lt ON TRUE
` + whereSQL + `
ORDER BY ` + sortColumn + ` ` + sortOrder + `, s.id DESC
LIMIT $` + fmt.Sprintf("%d", limitPos) + ` OFFSET $` + fmt.Sprintf("%d", offsetPos)

	rows, err := r.db.QueryContext(ctx, query, listArgs...)
	if err != nil {
		return nil, nil, err
	}
	defer func() { _ = rows.Close() }()

	items := make([]*service.OpenAIConversationRetentionListItem, 0, limit)
	for rows.Next() {
		item := &service.OpenAIConversationRetentionListItem{}
		lastTurn := &service.OpenAIConversationRetentionTurn{}
		var lastTurnRequestType sql.NullInt16
		var lastTurnID sql.NullInt64
		if err := rows.Scan(
			&item.ID,
			&item.UserID,
			&item.APIKeyID,
			&item.AccountID,
			&item.GroupID,
			&item.SessionKey,
			&item.ClientSessionID,
			&item.RequestedModel,
			&item.UpstreamModel,
			&item.ReasoningEffort,
			&item.FirstRequestID,
			&item.LastRequestID,
			&item.FirstResponseID,
			&item.LastResponseID,
			&item.TurnCount,
			&item.CreatedAt,
			&item.UpdatedAt,
			&item.UserEmail,
			&item.APIKeyName,
			&item.AccountName,
			&item.GroupName,
			&lastTurnID,
			&lastTurn.SessionID,
			&lastTurn.UserID,
			&lastTurn.APIKeyID,
			&lastTurn.AccountID,
			&lastTurn.GroupID,
			&lastTurn.RequestID,
			&lastTurn.ResponseID,
			&lastTurn.TurnIndex,
			&lastTurn.RequestedModel,
			&lastTurn.UpstreamModel,
			&lastTurn.ReasoningEffort,
			&lastTurn.Stream,
			&lastTurnRequestType,
			&lastTurn.InboundEndpoint,
			&lastTurn.UpstreamEndpoint,
			&lastTurn.UserInputText,
			&lastTurn.AssistantOutputText,
			&lastTurn.CreatedAt,
		); err != nil {
			return nil, nil, err
		}
		if lastTurnID.Valid {
			lastTurn.ID = lastTurnID.Int64
			if lastTurnRequestType.Valid {
				lastTurn.RequestType = service.RequestTypeFromInt16(lastTurnRequestType.Int16)
			}
			item.LastTurn = lastTurn
		}
		items = append(items, item)
	}
	if err := rows.Err(); err != nil {
		return nil, nil, err
	}

	pages := 1
	if total > 0 {
		pages = int((total + int64(limit) - 1) / int64(limit))
	}
	return items, &pagination.PaginationResult{
		Total:    total,
		Page:     params.Page,
		PageSize: limit,
		Pages:    pages,
	}, nil
}

func (r *openAIConversationRetentionRepository) DeleteExpiredTurns(ctx context.Context, cutoff time.Time, limit int) (int64, error) {
	if r == nil || r.db == nil {
		return 0, fmt.Errorf("nil openai conversation retention repository")
	}
	if limit <= 0 {
		limit = 5000
	}
	query := `
WITH victims AS (
  SELECT id
  FROM openai_retained_turns
  WHERE created_at < $1
  ORDER BY created_at ASC, id ASC
  LIMIT $2
)
DELETE FROM openai_retained_turns
WHERE id IN (SELECT id FROM victims)
`
	res, err := r.db.ExecContext(ctx, query, cutoff, limit)
	if err != nil {
		return 0, err
	}
	return res.RowsAffected()
}

func (r *openAIConversationRetentionRepository) DeleteEmptySessions(ctx context.Context, limit int) (int64, error) {
	if r == nil || r.db == nil {
		return 0, fmt.Errorf("nil openai conversation retention repository")
	}
	if limit <= 0 {
		limit = 5000
	}
	query := `
WITH victims AS (
  SELECT s.id
  FROM openai_retained_sessions s
  WHERE NOT EXISTS (
    SELECT 1
    FROM openai_retained_turns t
    WHERE t.session_id = s.id
  )
  ORDER BY s.updated_at ASC, s.id ASC
  LIMIT $1
)
DELETE FROM openai_retained_sessions
WHERE id IN (SELECT id FROM victims)
`
	res, err := r.db.ExecContext(ctx, query, limit)
	if err != nil {
		return 0, err
	}
	return res.RowsAffected()
}
