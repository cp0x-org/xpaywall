package handlers

import (
	"fmt"
	"net/http"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
	"github.com/jackc/pgx/v5/pgtype"

	postgres "github.com/cp0x-org/xpaywall/control-api/internal/storage/postgres/generated"
)

type periodRange struct {
	From string `json:"from"`
	To   string `json:"to"`
}

type periodStats struct {
	Range            periodRange `json:"range"`
	TotalProjects    int64       `json:"total_projects"`
	TotalRoutes      int64       `json:"total_routes"`
	TotalRequests    int64       `json:"total_requests"`
	TotalEarningsUSD float64     `json:"total_earnings_usd"`
	SuccessRate      float64     `json:"success_rate"`
}

type dashboardStatsResponse struct {
	Period         periodStats `json:"period"`
	PreviousPeriod periodStats `json:"previous_period"`
}

type dailyStatPoint struct {
	Time          string  `json:"time"`
	TotalRequests int64   `json:"total_requests"`
	EarningsUsd   float64 `json:"earnings_usd"`
}

type chartStatsResponse struct {
	Granularity string           `json:"granularity"`
	Points      []dailyStatPoint `json:"points"`
}

// getHourlyStatsSQL groups into 2-hour buckets using UTC timestamps.
const getHourlyStatsSQL = `
SELECT
    date_trunc('hour', created_at) - (EXTRACT(HOUR FROM created_at)::INT % 2) * INTERVAL '1 hour' AS bucket,
    COUNT(*)::BIGINT                                                                                AS total_requests,
    COALESCE(SUM(amount_usd) FILTER (WHERE payment_completed = TRUE), 0)::FLOAT8                  AS total_earnings_usd
FROM request_logs
WHERE created_at >= $1 AND created_at < $2
GROUP BY 1
ORDER BY 1
`

const getHourlyStatsWithProjectSQL = `
SELECT
    date_trunc('hour', created_at) - (EXTRACT(HOUR FROM created_at)::INT % 2) * INTERVAL '1 hour' AS bucket,
    COUNT(*)::BIGINT                                                                                AS total_requests,
    COALESCE(SUM(amount_usd) FILTER (WHERE payment_completed = TRUE), 0)::FLOAT8                  AS total_earnings_usd
FROM request_logs
WHERE created_at >= $1 AND created_at < $2 AND project_id = $3
GROUP BY 1
ORDER BY 1
`

const getHourlyStatsByOwnerSQL = `
SELECT
    date_trunc('hour', rl.created_at) - (EXTRACT(HOUR FROM rl.created_at)::INT % 2) * INTERVAL '1 hour' AS bucket,
    COUNT(*)::BIGINT                                                                                AS total_requests,
    COALESCE(SUM(rl.amount_usd) FILTER (WHERE rl.payment_completed = TRUE), 0)::FLOAT8            AS total_earnings_usd
FROM request_logs rl
JOIN projects p ON p.id = rl.project_id
WHERE rl.created_at >= $1 AND rl.created_at < $2 AND p.owner_user_id = $3
GROUP BY 1
ORDER BY 1
`

const getDailyStatsWithProjectSQL = `
SELECT
    DATE(created_at)                                                                                   AS day,
    COUNT(*)::BIGINT                                                                                   AS total_requests,
    COALESCE(SUM(amount_usd) FILTER (WHERE payment_completed = TRUE), 0)::FLOAT8                      AS total_earnings_usd
FROM request_logs
WHERE created_at >= $1 AND created_at < $2 AND project_id = $3
GROUP BY DATE(created_at)
ORDER BY DATE(created_at)
`

const getDashboardStatsWithProjectSQL = `
SELECT
    1::BIGINT AS total_projects,
    (SELECT COUNT(*) FROM routes WHERE project_id = $4 AND created_at >= $1 AND created_at < $2)::BIGINT AS total_routes,
    (COUNT(*) FILTER (WHERE created_at >= $1 AND created_at < $2))::BIGINT AS total_requests,
    COALESCE(SUM(amount_usd) FILTER (
        WHERE payment_required = TRUE AND payment_completed = TRUE AND created_at >= $1 AND created_at < $2
    ), 0)::FLOAT8 AS total_earnings_usd,
    CASE
        WHEN (COUNT(*) FILTER (WHERE payment_required = TRUE AND created_at >= $1 AND created_at < $2)) = 0 THEN 0.0
        ELSE (COUNT(*) FILTER (WHERE payment_required = TRUE AND payment_completed = TRUE AND created_at >= $1 AND created_at < $2))::FLOAT8
            * 100.0
            / (COUNT(*) FILTER (WHERE payment_required = TRUE AND created_at >= $1 AND created_at < $2))::FLOAT8
    END::FLOAT8 AS success_rate,
    1::BIGINT AS prev_total_projects,
    (SELECT COUNT(*) FROM routes WHERE project_id = $4 AND created_at >= $3 AND created_at < $1)::BIGINT AS prev_total_routes,
    (COUNT(*) FILTER (WHERE created_at >= $3 AND created_at < $1))::BIGINT AS prev_total_requests,
    COALESCE(SUM(amount_usd) FILTER (
        WHERE payment_required = TRUE AND payment_completed = TRUE AND created_at >= $3 AND created_at < $1
    ), 0)::FLOAT8 AS prev_total_earnings_usd,
    CASE
        WHEN (COUNT(*) FILTER (WHERE payment_required = TRUE AND created_at >= $3 AND created_at < $1)) = 0 THEN 0.0
        ELSE (COUNT(*) FILTER (WHERE payment_required = TRUE AND payment_completed = TRUE AND created_at >= $3 AND created_at < $1))::FLOAT8
            * 100.0
            / (COUNT(*) FILTER (WHERE payment_required = TRUE AND created_at >= $3 AND created_at < $1))::FLOAT8
    END::FLOAT8 AS prev_success_rate
FROM request_logs
WHERE project_id = $4 AND created_at >= $3 AND created_at < $2
`

// GetDailyStats returns time-series request and earnings data.
// @Summary     Get daily/hourly stats
// @Tags        stats
// @Produce     json
// @Param       period query string false "Period: day, week, month, custom (default: 7 days)"
// @Param       from query string false "Start date (YYYY-MM-DD or DD.MM.YYYY) — required for period=custom"
// @Param       to query string false "End date (YYYY-MM-DD or DD.MM.YYYY) — required for period=custom"
// @Param       days query int false "Number of days (used when period is not set, default 7, max 90)"
// @Param       project_id query string false "Filter by Project ID (UUID)"
// @Success     200 {object} chartStatsResponse
// @Failure     400 {object} errorResponse
// @Failure     500 {object} errorResponse
// @Security    BearerAuth
// @Router      /api/v1/stats/daily [get]
func (h *Handler) GetDailyStats(c *gin.Context) {
	pt := c.DefaultQuery("period", "")
	fromStr := c.Query("from")
	toStr := c.Query("to")

	var from, to time.Time

	if pt != "" || fromStr != "" || toStr != "" {
		periodStart, periodEnd, _, err := computeDashboardPeriod(pt, fromStr, toStr)
		if err != nil {
			c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
			return
		}
		from = periodStart
		to = periodEnd
	} else {
		daysStr := c.DefaultQuery("days", "7")
		days := 7
		if n, err := parseInt(daysStr); err == nil && n > 0 && n <= 90 {
			days = n
		}
		now := time.Now().UTC()
		from = now.AddDate(0, 0, -days).Truncate(24 * time.Hour)
		to = now.Add(24 * time.Hour).Truncate(24 * time.Hour)
	}

	callerID, ok := callerUserID(c)
	if !ok {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "unauthorized"})
		return
	}

	var projectID *uuid.UUID
	if pidStr := c.Query("project_id"); pidStr != "" {
		pid, err := uuid.Parse(pidStr)
		if err != nil {
			c.JSON(http.StatusBadRequest, gin.H{"error": "invalid project_id"})
			return
		}
		if !h.requireProjectOwner(c, pid) {
			return
		}
		projectID = &pid
	}

	useHourly := to.Sub(from) < 48*time.Hour

	if useHourly {
		// Default to the caller's projects; narrow to one project when requested.
		sql := getHourlyStatsByOwnerSQL
		args := []any{
			pgtype.Timestamp{Time: from, Valid: true},
			pgtype.Timestamp{Time: to, Valid: true},
			&callerID,
		}
		if projectID != nil {
			sql = getHourlyStatsWithProjectSQL
			args = []any{
				pgtype.Timestamp{Time: from, Valid: true},
				pgtype.Timestamp{Time: to, Valid: true},
				*projectID,
			}
		}
		pgRows, err := h.db.Query(c.Request.Context(), sql, args...)
		if err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
			return
		}
		defer pgRows.Close()

		points := make([]dailyStatPoint, 0)
		for pgRows.Next() {
			var bucket pgtype.Timestamp
			var totalReqs int64
			var earnings float64
			if err := pgRows.Scan(&bucket, &totalReqs, &earnings); err != nil {
				c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
				return
			}
			points = append(points, dailyStatPoint{
				Time:          bucket.Time.UTC().Format(time.RFC3339),
				TotalRequests: totalReqs,
				EarningsUsd:   earnings,
			})
		}
		if err := pgRows.Err(); err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
			return
		}
		c.JSON(http.StatusOK, chartStatsResponse{Granularity: "hour", Points: points})
		return
	}

	if projectID != nil {
		pgRows, err := h.db.Query(c.Request.Context(), getDailyStatsWithProjectSQL,
			pgtype.Timestamp{Time: from, Valid: true},
			pgtype.Timestamp{Time: to, Valid: true},
			*projectID,
		)
		if err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
			return
		}
		defer pgRows.Close()

		points := make([]dailyStatPoint, 0)
		for pgRows.Next() {
			var day pgtype.Date
			var totalReqs int64
			var earnings float64
			if err := pgRows.Scan(&day, &totalReqs, &earnings); err != nil {
				c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
				return
			}
			points = append(points, dailyStatPoint{
				Time:          day.Time.UTC().Format(time.RFC3339),
				TotalRequests: totalReqs,
				EarningsUsd:   earnings,
			})
		}
		if err := pgRows.Err(); err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
			return
		}
		c.JSON(http.StatusOK, chartStatsResponse{Granularity: "day", Points: points})
		return
	}

	rows, err := h.q.GetDailyStatsByOwner(c.Request.Context(), postgres.GetDailyStatsByOwnerParams{
		Owner:       &callerID,
		PeriodStart: pgtype.Timestamp{Time: from, Valid: true},
		PeriodEnd:   pgtype.Timestamp{Time: to, Valid: true},
	})
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	points := make([]dailyStatPoint, 0, len(rows))
	for _, r := range rows {
		points = append(points, dailyStatPoint{
			Time:          r.Day.Time.UTC().Format(time.RFC3339),
			TotalRequests: r.TotalRequests,
			EarningsUsd:   r.TotalEarningsUsd,
		})
	}
	c.JSON(http.StatusOK, chartStatsResponse{Granularity: "day", Points: points})
}

func parseInt(s string) (int, error) {
	var n int
	_, err := fmt.Sscanf(s, "%d", &n)
	return n, err
}

// parseDate accepts YYYY-MM-DD or DD.MM.YYYY.
func parseDate(s string) (time.Time, error) {
	for _, layout := range []string{"2006-01-02", "02.01.2006"} {
		if t, e := time.ParseInLocation(layout, s, time.UTC); e == nil {
			return t, nil
		}
	}
	return time.Time{}, fmt.Errorf("invalid date %q: use YYYY-MM-DD or DD.MM.YYYY", s)
}

// computeDashboardPeriod returns (periodStart, periodEnd, prevStart).
// If from+to are provided, always treated as custom regardless of period param.
// period=day|week|month|custom. Default: current week (Mon–Sun).
func computeDashboardPeriod(periodType, fromStr, toStr string) (periodStart, periodEnd, prevStart time.Time, err error) {
	now := time.Now().UTC()
	y, m, d := now.Date()

	// auto-detect custom when from+to are present
	if fromStr != "" && toStr != "" {
		periodType = "custom"
	}

	switch periodType {
	case "day":
		periodStart = time.Date(y, m, d, 0, 0, 0, 0, time.UTC)
		periodEnd = periodStart.AddDate(0, 0, 1)
		prevStart = periodStart.AddDate(0, 0, -1)

	case "month":
		periodStart = time.Date(y, m, 1, 0, 0, 0, 0, time.UTC)
		periodEnd = periodStart.AddDate(0, 1, 0)
		prevStart = periodStart.AddDate(0, -1, 0)

	case "custom":
		if fromStr == "" || toStr == "" {
			err = fmt.Errorf("from and to are required for custom period")
			return
		}
		periodStart, err = parseDate(fromStr)
		if err != nil {
			return
		}
		var toTime time.Time
		toTime, err = parseDate(toStr)
		if err != nil {
			return
		}
		periodEnd = toTime.AddDate(0, 0, 1)
		prevStart = periodStart.Add(-(periodEnd.Sub(periodStart)))

	default: // "week"
		// Monday of current week
		weekday := int(now.Weekday())
		if weekday == 0 {
			weekday = 7 // Sunday → 7 so Monday offset = 6
		}
		periodStart = time.Date(y, m, d-(weekday-1), 0, 0, 0, 0, time.UTC)
		periodEnd = periodStart.AddDate(0, 0, 7)
		prevStart = periodStart.AddDate(0, 0, -7)
	}
	return
}

// GetDashboardStats returns aggregated stats for the dashboard.
// @Summary     Get dashboard stats
// @Tags        stats
// @Produce     json
// @Param       period query string false "Period: day, week (default), month, custom"
// @Param       from query string false "Start date (YYYY-MM-DD or DD.MM.YYYY) — required for period=custom"
// @Param       to query string false "End date (YYYY-MM-DD or DD.MM.YYYY) — required for period=custom"
// @Param       project_id query string false "Filter by Project ID (UUID)"
// @Success     200 {object} dashboardStatsResponse
// @Failure     400 {object} errorResponse
// @Failure     500 {object} errorResponse
// @Security    BearerAuth
// @Router      /api/v1/stats/dashboard [get]
func (h *Handler) GetDashboardStats(c *gin.Context) {
	pt := c.DefaultQuery("period", "week")
	fromStr := c.Query("from")
	toStr := c.Query("to")

	periodStart, periodEnd, prevStart, err := computeDashboardPeriod(pt, fromStr, toStr)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	callerID, ok := callerUserID(c)
	if !ok {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "unauthorized"})
		return
	}

	var projectID *uuid.UUID
	if pidStr := c.Query("project_id"); pidStr != "" {
		pid, parseErr := uuid.Parse(pidStr)
		if parseErr != nil {
			c.JSON(http.StatusBadRequest, gin.H{"error": "invalid project_id"})
			return
		}
		if !h.requireProjectOwner(c, pid) {
			return
		}
		projectID = &pid
	}

	const dateFmt = "2006-01-02"

	if projectID != nil {
		pgRow := h.db.QueryRow(c.Request.Context(), getDashboardStatsWithProjectSQL,
			pgtype.Timestamp{Time: periodStart, Valid: true},
			pgtype.Timestamp{Time: periodEnd, Valid: true},
			pgtype.Timestamp{Time: prevStart, Valid: true},
			*projectID,
		)
		var r postgres.GetDashboardStatsRow
		if err := pgRow.Scan(
			&r.TotalProjects, &r.TotalRoutes, &r.TotalRequests, &r.TotalEarningsUsd, &r.SuccessRate,
			&r.PrevTotalProjects, &r.PrevTotalRoutes, &r.PrevTotalRequests, &r.PrevTotalEarningsUsd, &r.PrevSuccessRate,
		); err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
			return
		}
		c.JSON(http.StatusOK, dashboardStatsResponse{
			Period: periodStats{
				Range:            periodRange{From: periodStart.Format(dateFmt), To: periodEnd.AddDate(0, 0, -1).Format(dateFmt)},
				TotalProjects:    r.TotalProjects,
				TotalRoutes:      r.TotalRoutes,
				TotalRequests:    r.TotalRequests,
				TotalEarningsUSD: r.TotalEarningsUsd,
				SuccessRate:      r.SuccessRate,
			},
			PreviousPeriod: periodStats{
				Range:            periodRange{From: prevStart.Format(dateFmt), To: periodStart.AddDate(0, 0, -1).Format(dateFmt)},
				TotalProjects:    r.PrevTotalProjects,
				TotalRoutes:      r.PrevTotalRoutes,
				TotalRequests:    r.PrevTotalRequests,
				TotalEarningsUSD: r.PrevTotalEarningsUsd,
				SuccessRate:      r.PrevSuccessRate,
			},
		})
		return
	}

	row, err := h.q.GetDashboardStatsByOwner(c.Request.Context(), postgres.GetDashboardStatsByOwnerParams{
		Owner:       &callerID,
		PeriodStart: pgtype.Timestamp{Time: periodStart, Valid: true},
		PeriodEnd:   pgtype.Timestamp{Time: periodEnd, Valid: true},
		PrevStart:   pgtype.Timestamp{Time: prevStart, Valid: true},
	})
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, dashboardStatsResponse{
		Period: periodStats{
			Range:            periodRange{From: periodStart.Format(dateFmt), To: periodEnd.AddDate(0, 0, -1).Format(dateFmt)},
			TotalProjects:    row.TotalProjects,
			TotalRoutes:      row.TotalRoutes,
			TotalRequests:    row.TotalRequests,
			TotalEarningsUSD: row.TotalEarningsUsd,
			SuccessRate:      row.SuccessRate,
		},
		PreviousPeriod: periodStats{
			Range:            periodRange{From: prevStart.Format(dateFmt), To: periodStart.AddDate(0, 0, -1).Format(dateFmt)},
			TotalProjects:    row.PrevTotalProjects,
			TotalRoutes:      row.PrevTotalRoutes,
			TotalRequests:    row.PrevTotalRequests,
			TotalEarningsUSD: row.PrevTotalEarningsUsd,
			SuccessRate:      row.PrevSuccessRate,
		},
	})
}

type topRouteItem struct {
	PathPattern   string  `json:"path_pattern"`
	PriceUsd      string  `json:"price_usd"`
	TotalRequests int64   `json:"total_requests"`
	RevenueUsd    float64 `json:"revenue_usd"`
}

// GetDashboardTopRoutes returns the top routes by request volume.
// @Summary     Get top routes
// @Tags        stats
// @Produce     json
// @Param       project_id query string false "Filter by Project ID (UUID)"
// @Success     200 {array} topRouteItem
// @Failure     400 {object} errorResponse
// @Failure     500 {object} errorResponse
// @Security    BearerAuth
// @Router      /api/v1/stats/top-routes [get]
func (h *Handler) GetDashboardTopRoutes(c *gin.Context) {
	callerID, ok := callerUserID(c)
	if !ok {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "unauthorized"})
		return
	}

	var projectID *uuid.UUID
	if pidStr := c.Query("project_id"); pidStr != "" {
		pid, err := uuid.Parse(pidStr)
		if err != nil {
			c.JSON(http.StatusBadRequest, gin.H{"error": "invalid project_id"})
			return
		}
		if !h.requireProjectOwner(c, pid) {
			return
		}
		projectID = &pid
	}

	items := make([]topRouteItem, 0, 5)

	if projectID != nil {
		rows, err := h.q.GetTopRoutesForDashboardByProject(c.Request.Context(), *projectID)
		if err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
			return
		}
		for _, r := range rows {
			items = append(items, topRouteItem{PathPattern: r.PathPattern, PriceUsd: r.PriceUsd, TotalRequests: r.TotalRequests, RevenueUsd: r.RevenueUsd})
		}
	} else {
		rows, err := h.q.GetTopRoutesForDashboardByOwner(c.Request.Context(), &callerID)
		if err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
			return
		}
		for _, r := range rows {
			items = append(items, topRouteItem{PathPattern: r.PathPattern, PriceUsd: r.PriceUsd, TotalRequests: r.TotalRequests, RevenueUsd: r.RevenueUsd})
		}
	}

	c.JSON(http.StatusOK, items)
}

type recentRequestItem struct {
	ID             string  `json:"id"`
	Path           string  `json:"path"`
	Method         string  `json:"method"`
	CreatedAt      string  `json:"created_at"`
	StatusCode     int32   `json:"status_code"`
	PaymentChannel *string `json:"payment_channel"`
	AmountUsd      *string `json:"amount_usd"`
}

func toRecentItem(id uuid.UUID, path, method string, createdAt pgtype.Timestamp, statusCode int32, paymentChannel pgtype.Text, amountUsd pgtype.Numeric) recentRequestItem {
	item := recentRequestItem{
		ID:         id.String(),
		Path:       path,
		Method:     method,
		CreatedAt:  createdAt.Time.UTC().Format(time.RFC3339),
		StatusCode: statusCode,
	}
	if paymentChannel.Valid {
		item.PaymentChannel = &paymentChannel.String
	}
	if amountUsd.Valid {
		if f, err := amountUsd.Float64Value(); err == nil && f.Valid {
			s := fmt.Sprintf("%.4f", f.Float64)
			item.AmountUsd = &s
		}
	}
	return item
}

// GetDashboardRecentRequests returns the most recent requests for the dashboard.
// @Summary     Get recent requests
// @Tags        stats
// @Produce     json
// @Param       project_id query string false "Filter by Project ID (UUID)"
// @Success     200 {array} recentRequestItem
// @Failure     400 {object} errorResponse
// @Failure     500 {object} errorResponse
// @Security    BearerAuth
// @Router      /api/v1/stats/recent-requests [get]
func (h *Handler) GetDashboardRecentRequests(c *gin.Context) {
	callerID, ok := callerUserID(c)
	if !ok {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "unauthorized"})
		return
	}

	var projectID *uuid.UUID
	if pidStr := c.Query("project_id"); pidStr != "" {
		pid, err := uuid.Parse(pidStr)
		if err != nil {
			c.JSON(http.StatusBadRequest, gin.H{"error": "invalid project_id"})
			return
		}
		if !h.requireProjectOwner(c, pid) {
			return
		}
		projectID = &pid
	}

	items := make([]recentRequestItem, 0, 5)

	if projectID != nil {
		rows, err := h.q.GetRecentRequestsForDashboardByProject(c.Request.Context(), *projectID)
		if err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
			return
		}
		for _, r := range rows {
			items = append(items, toRecentItem(r.ID, r.Path, r.Method, r.CreatedAt, r.StatusCode, r.PaymentChannel, r.AmountUsd))
		}
	} else {
		rows, err := h.q.GetRecentRequestsForDashboardByOwner(c.Request.Context(), &callerID)
		if err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
			return
		}
		for _, r := range rows {
			items = append(items, toRecentItem(r.ID, r.Path, r.Method, r.CreatedAt, r.StatusCode, r.PaymentChannel, r.AmountUsd))
		}
	}

	c.JSON(http.StatusOK, items)
}
