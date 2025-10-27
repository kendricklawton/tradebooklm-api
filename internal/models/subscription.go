package models

import (
	"time"
)

type Status string

const (
	StatusActive    Status = "active"
	StatusSuspended Status = "suspended"
	StatusPending   Status = "pending"
	StatusClosed    Status = "closed"
)

type Tier string

const (
	TierHobby Tier = "hobby"
	TierPro   Tier = "pro"
	TierUltra Tier = "ultra"
)

type Subscription struct {
	WorkspaceID        string     `json:"workspace_id" db:"workspace_id"`
	Tier               Tier       `json:"tier" db:"tier"`
	Status             Status     `json:"status" db:"status"`
	CurrentPeriodStart *time.Time `json:"current_period_start,omitempty" db:"current_period_start"`
	CurrentPeriodEnd   *time.Time `json:"current_period_end,omitempty" db:"current_period_end"`
	TrialEnd           *time.Time `json:"trial_end,omitempty" db:"trial_end"`
	CancelAt           *time.Time `json:"cancel_at,omitempty" db:"cancel_at"`
	CanceledAt         *time.Time `json:"canceled_at,omitempty" db:"canceled_at"`
	LastSyncedAt       *time.Time `json:"last_synced_at,omitempty" db:"last_synced_at"`
	CreatedAt          time.Time  `json:"created_at" db:"created_at"`
	UpdatedAt          time.Time  `json:"updated_at" db:"updated_at"`

	StripeCustomerID     *string `json:"stripe_customer_id,omitempty" db:"stripe_customer_id"`
	StripeSubscriptionID *string `json:"stripe_subscription_id,omitempty" db:"stripe_subscription_id"`
}

type BillingPeriod struct {
	ID              string     `json:"id" db:"id"`
	WorkspaceID     string     `json:"workspaceId" db:"workspace_id"`
	StripeInvoiceID *string    `json:"stripeInvoiceId,omitempty" db:"stripe_invoice_id"`
	StartDate       time.Time  `json:"startDate" db:"start_date"`
	EndDate         *time.Time `json:"endDate,omitempty" db:"end_date"`
	CreatedAt       time.Time  `json:"createdAt" db:"created_at"`
}

type ModelUsage struct {
	BillingPeriodID   string     `json:"billing_period_id" db:"billing_period_id"`
	ModelID           string     `json:"model_id" db:"model_id"`
	ModelName         *string    `json:"model_name,omitempty" db:"model_name"`
	ModelVersion      *string    `json:"model_version,omitempty" db:"model_version"`
	InputTokens       int64      `json:"input_tokens" db:"input_tokens"`
	OutputTokens      int64      `json:"output_tokens" db:"output_tokens"`
	InputCostCents    int64      `json:"input_cost_cents" db:"input_cost_cents"`
	OutputCostCents   int64      `json:"output_cost_cents" db:"output_cost_cents"`
	APIRequestCount   int64      `json:"api_request_count" db:"api_request_count"`
	FailedRequests    int64      `json:"failed_requests" db:"failed_requests"`
	AverageLatencyMs  int64      `json:"average_latency_ms" db:"average_latency_ms"`
	MaxLatencyMs      int64      `json:"max_latency_ms" db:"max_latency_ms"`
	CacheHits         int64      `json:"cache_hits" db:"cache_hits"`
	CacheMisses       int64      `json:"cache_misses" db:"cache_misses"`
	AverageInputSize  int64      `json:"average_input_size" db:"average_input_size"`
	AverageOutputSize int64      `json:"average_output_size" db:"average_output_size"`
	MaxInputSize      int64      `json:"max_input_size" db:"max_input_size"`
	MaxOutputSize     int64      `json:"max_output_size" db:"max_output_size"`
	WebAppRequests    int64      `json:"web_app_requests" db:"web_app_requests"`
	APIRequests       int64      `json:"api_requests" db:"api_requests"`
	MobileAppRequests int64      `json:"mobile_app_requests" db:"mobile_app_requests"`
	FirstUsageTime    *time.Time `json:"first_usage_time,omitempty" db:"first_usage_time"`
	LastUsageTime     *time.Time `json:"last_usage_time,omitempty" db:"last_usage_time"`
}
