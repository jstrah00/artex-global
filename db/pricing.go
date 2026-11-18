package db

import "strings"

// ModelPrice is the USD cost per 1,000,000 tokens for one model, split by token
// kind. These feed the per-task cost estimate shown in the UI (see TokenByModel).
//
// IMPORTANT: these are ESTIMATES. Edit them to match your provider's current
// price list. Values below are USD per 1M tokens. CacheRead is the discounted
// "cache hit" input price; CacheWrite is the cache-creation price (for Anthropic,
// ~1.25x the base input price). A model missing from the table yields cost 0 with
// Known=false, and the UI marks it as "price unknown" instead of guessing.
type ModelPrice struct {
	Input      float64
	Output     float64
	CacheRead  float64
	CacheWrite float64
}

// modelPrices maps a normalized model id (see normalizeModel) to its price.
// Keep ids lower-case and without a provider prefix. Source of the figures below:
// llm-stats.com pricing (Oct 2026) for the models this deployment uses; adjust as
// prices change.
var modelPrices = map[string]ModelPrice{
	// Anthropic
	"claude-opus-5-5":   {Input: 4.00, Output: 20.00, CacheRead: 0.40, CacheWrite: 5.00},
	"claude-sonnet-5-5": {Input: 2.00, Output: 10.00, CacheRead: 0.20, CacheWrite: 2.50},
	"claude-haiku-5-5":  {Input: 0.10, Output: 0.50, CacheRead: 0.01, CacheWrite: 0.125},
	// claude-sonnet-4-6 is not on the pricing page used; this mirrors the legacy
	// Claude Sonnet list price and is a placeholder — correct it if needed.
	"claude-sonnet-4-6": {Input: 3.00, Output: 15.00, CacheRead: 0.30, CacheWrite: 3.75},

	// OpenAI (blended figures from llm-stats; verify against your account)
	"gpt-6.1-sol": {Input: 2.00, Output: 10.00, CacheRead: 0.10, CacheWrite: 0.10},

	// Google
	"gemini-4-argon": {Input: 2.00, Output: 10.00, CacheRead: 0.10, CacheWrite: 0.10},

	// Low-cost options
	"muse-spark-1.3":      {Input: 0.10, Output: 0.20, CacheRead: 0.002, CacheWrite: 0.002},
	"deepseek-v4.1-flash": {Input: 0.22, Output: 0.66, CacheRead: 0.0, CacheWrite: 0.0},
}

// normalizeModel lower-cases a model id and strips a leading provider prefix like
// "anthropic/" or "openai:" so "anthropic/claude-sonnet-5-5" and "claude-sonnet-5-5"
// price the same.
func normalizeModel(model string) string {
	m := strings.ToLower(strings.TrimSpace(model))
	for _, sep := range []string{"/", ":"} {
		if i := strings.LastIndex(m, sep); i >= 0 {
			m = m[i+1:]
		}
	}
	return m
}

// costUSD returns the estimated USD cost for the given token counts on one model,
// and whether the model's price is known. tokens are raw counts (not per-million).
func costUSD(model string, input, output, cacheRead, cacheWrite int) (float64, bool) {
	p, ok := modelPrices[normalizeModel(model)]
	if !ok {
		return 0, false
	}
	const perM = 1_000_000.0
	cost := float64(input)/perM*p.Input +
		float64(output)/perM*p.Output +
		float64(cacheRead)/perM*p.CacheRead +
		float64(cacheWrite)/perM*p.CacheWrite
	return cost, true
}
