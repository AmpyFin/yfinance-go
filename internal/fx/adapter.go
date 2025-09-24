package fx

import (
	"context"
	"time"

	"github.com/AmpyFin/yfinance-go/internal/norm"
)

// Adapter adapts the FX Manager to the norm.FXConverter interface
type Adapter struct {
	manager *Manager
}

// NewAdapter creates a new FX adapter
func NewAdapter(manager *Manager) *Adapter {
	return &Adapter{
		manager: manager,
	}
}

// ConvertValue converts a monetary value using FX rates
func (a *Adapter) ConvertValue(ctx context.Context, value norm.ScaledDecimal, fromCurrency, toCurrency string, at time.Time) (norm.ScaledDecimal, *norm.FXMeta, error) {
	converted, fxMeta, err := a.manager.ConvertValue(ctx, value, fromCurrency, toCurrency, at)
	if err != nil {
		return norm.ScaledDecimal{}, nil, err
	}

	// Convert fx.FXMeta to norm.FXMeta
	normMeta := &norm.FXMeta{
		Provider:       fxMeta.Provider,
		Base:           fxMeta.Base,
		Symbols:        fxMeta.Symbols,
		AsOf:           fxMeta.AsOf,
		RateScale:      fxMeta.RateScale,
		CacheHit:       fxMeta.CacheHit,
		Attempts:       fxMeta.Attempts,
		BackoffProfile: fxMeta.BackoffProfile,
		Stale:          fxMeta.Stale,
	}

	return converted, normMeta, nil
}
