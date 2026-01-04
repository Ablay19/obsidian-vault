package ai

import (
	"obsidian-automation/internal/config"
	"sort"
)

// select_provider selects the best AI provider based on task complexity and cost constraints.
func select_provider(task_tokens int, task_depth int, max_cost float64, profiles map[string]config.ProviderConfig, rules config.SwitchingRules) string {
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

	// Sort candidates: 1. cost (asc), 2. latency (asc), 3. accuracy (desc)
	sort.Slice(candidates, func(i, j int) bool {
		if candidates[i].cost != candidates[j].cost {
			return candidates[i].cost < candidates[j].cost
		}
		if candidates[i].latency != candidates[j].latency {
			return candidates[i].latency < candidates[j].latency
		}
		return candidates[i].accuracy > candidates[j].accuracy
	})

	return candidates[0].name
}
