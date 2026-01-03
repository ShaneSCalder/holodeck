package account

import (
	"fmt"
	"time"
)

// ==================== ACCOUNT MANAGER ====================

// Account manages all account-level operations and state
type Account struct {
	// Identification
	AccountID   string
	Name        string
	Description string

	// Initial Setup
	InitialBalance float64
	Currency       string
	Leverage       float64

	// Current State
	CurrentBalance     float64
	UsedMargin         float64
	AvailableMargin    float64
	BuyingPower        float64
	TotalRealizedPnL   float64
	TotalUnrealizedPnL float64
	CommissionPaid     float64

	// Trade Statistics
	TotalTrades       int
	WinningTrades     int
	LosingTrades      int
	BreakevenTrades   int
	ConsecutiveWins   int
	ConsecutiveLosses int

	// Risk Management
	MaxDrawdownPercent     float64
	MaxDrawdownExperienced float64
	MaxDrawdownAmount      float64
	MaxPositionSize        float64
	MaxPositionsOpen       int
	MaxLeverageAllowed     float64
	RiskPerTradePercent    float64

	// Account Status
	Status         string // ACTIVE, BLOWN, AT_LIMIT, CLOSED
	HighWaterMark  float64
	LowWaterMark   float64
	CreatedTime    time.Time
	LastUpdateTime time.Time
	UpdateHistory  []*BalanceUpdate
}

// ==================== BALANCE UPDATE ====================

// BalanceUpdate records a balance change event
type BalanceUpdate struct {
	Timestamp      time.Time
	BalanceBefore  float64
	BalanceAfter   float64
	Change         float64
	Reason         string
	TransactionID  string
	RelatedTradeID string
}

// ==================== CONSTRUCTORS ====================

// NewAccount creates a new account
func NewAccount(id, name string, initialBalance float64, currency string) *Account {
	now := time.Now()
	account := &Account{
		AccountID:           id,
		Name:                name,
		InitialBalance:      initialBalance,
		CurrentBalance:      initialBalance,
		Currency:            currency,
		Leverage:            1.0,
		Status:              "ACTIVE",
		CreatedTime:         now,
		LastUpdateTime:      now,
		HighWaterMark:       initialBalance,
		LowWaterMark:        initialBalance,
		MaxDrawdownPercent:  20.0,
		MaxPositionSize:     initialBalance * 0.1,
		MaxPositionsOpen:    10,
		MaxLeverageAllowed:  50.0,
		RiskPerTradePercent: 2.0,
		UpdateHistory:       make([]*BalanceUpdate, 0),
	}
	return account
}

// ==================== STATUS CHECKS ====================

// IsActive checks if account is active
func (a *Account) IsActive() bool {
	return a.Status == "ACTIVE"
}

// IsBlown checks if account is blown
func (a *Account) IsBlown() bool {
	return a.Status == "BLOWN"
}

// IsAtLimit checks if account is at margin limit
func (a *Account) IsAtLimit() bool {
	return a.Status == "AT_LIMIT"
}

// CanTrade checks if account can execute trades
func (a *Account) CanTrade() bool {
	return a.Status == "ACTIVE" && a.CurrentBalance > 0
}

// ==================== SUMMARY ====================

// String returns a string representation
func (a *Account) String() string {
	return fmt.Sprintf(
		"Account: %s (%s)\n"+
			"Balance: %.2f %s | P&L: %.2f | Commission: %.2f\n"+
			"Margin: %.2f / %.2f (%.1f%%) | Leverage: %.1fx\n"+
			"Trades: %d (W:%d L:%d B:%d) | Win Rate: %.1f%%\n"+
			"Drawdown: %.1f%% (%.2f) | Status: %s",
		a.Name, a.AccountID,
		a.CurrentBalance, a.Currency, a.TotalRealizedPnL, a.CommissionPaid,
		a.UsedMargin, a.BuyingPower, a.GetAvailableMarginPercent(), a.Leverage,
		a.TotalTrades, a.WinningTrades, a.LosingTrades, a.BreakevenTrades, a.GetWinRate(),
		a.GetDrawdownPercent(), a.MaxDrawdownAmount, a.Status,
	)
}

// RecordBalanceUpdate adds an update to history
func (a *Account) RecordBalanceUpdate(before, after, change float64, reason, transactionID string) {
	update := &BalanceUpdate{
		Timestamp:     time.Now(),
		BalanceBefore: before,
		BalanceAfter:  after,
		Change:        change,
		Reason:        reason,
		TransactionID: transactionID,
	}
	a.UpdateHistory = append(a.UpdateHistory, update)
	a.LastUpdateTime = time.Now()
}
