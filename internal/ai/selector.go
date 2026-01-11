package ai

import (
	"obsidian-automation/internal/config"
	"sort"
	"strings"
)

// select_provider selects the best AI provider based on task complexity and cost constraints.
// If activeProvider is specified, it will be preferred if it's enabled.
func select_provider(task_tokens int, task_depth int, max_cost float64, profiles map[string]config.ProviderConfig, rules config.SwitchingRules, activeProvider string) string {
	// First, check if activeProvider is specified and is a valid/known provider
	if activeProvider != "" {
		// Normalize case for lookup
		normalizedActive := strings.ToLower(activeProvider)
		for name := range profiles {
			if strings.ToLower(name) == normalizedActive {
				return activeProvider
			}
		}
		// If activeProvider is not in profiles, still return it - it will be validated later
		if activeProvider != "" {
			return activeProvider
		}
	}

	type candidate struct {
		name     string
		cost     float64
		latency  int
		accuracy float64
	}

	var candidates []candidate

	for name, profile := range profiles {
		// Estimate cost
		estimated_cost := float64(task_tokens) * profile.InputCostPerToken
		if estimated_cost > max_cost {
			continue
		}

		// Check thresholds
		if task_tokens > profile.MaxInputTokens {
			continue
		}
		if profile.LatencyMsThreshold > rules.LatencyTarget {
			continue
		}
		if profile.AccuracyPctThreshold < rules.AccuracyThreshold {
			continue
		}

		candidates = append(candidates, candidate{
			name:     name,
			cost:     estimated_cost,
			latency:  profile.LatencyMsThreshold,
			accuracy: profile.AccuracyPctThreshold,
		})
	}

	if len(candidates) == 0 {
		return rules.DefaultProvider
	}

	// Sort candidates: 1. free providers first (cost == 0), 2. cost (asc), 3. latency (asc), 4. accuracy (desc)
	sort.Slice(candidates, func(i, j int) bool {
		// Prioritize free providers
		iFree := candidates[i].cost == 0
		jFree := candidates[j].cost == 0
		if iFree != jFree {
			return iFree
		}
		// Then by cost
		if candidates[i].cost != candidates[j].cost {
			return candidates[i].cost < candidates[j].cost
		}
		// Then by latency
		if candidates[i].latency != candidates[j].latency {
			return candidates[i].latency < candidates[j].latency
		}
		// Finally by accuracy
		return candidates[i].accuracy > candidates[j].accuracy
	})

	return candidates[0].name
}
