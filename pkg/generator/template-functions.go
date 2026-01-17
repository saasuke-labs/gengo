package generator

import (
	"strings"
)

// wherePages filters pages based on flag conditions
// Syntax: "flag1,!flag2,flag3" means pages with flag1 AND flag3 AND NOT flag2
func wherePages(flagConditions string, pages []Page) []Page {
	if flagConditions == "" {
		return pages
	}

	conditions := strings.Split(flagConditions, ",")
	var result []Page

	for _, page := range pages {
		if matchesConditions(page.Flags, conditions) {
			result = append(result, page)
		}
	}

	return result
}

// matchesConditions checks if page flags match all conditions
func matchesConditions(pageFlags []string, conditions []string) bool {
	for _, condition := range conditions {
		condition = strings.TrimSpace(condition)
		if condition == "" {
			continue
		}

		isNegated := strings.HasPrefix(condition, "!")
		flag := condition
		if isNegated {
			flag = strings.TrimPrefix(condition, "!")
		}

		hasFlag := contains(pageFlags, flag)

		// If condition is negated (!flag), page should NOT have the flag
		// If condition is not negated (flag), page SHOULD have the flag
		if isNegated && hasFlag {
			return false
		}
		if !isNegated && !hasFlag {
			return false
		}
	}

	return true
}

// contains checks if a slice contains a specific string
func contains(slice []string, item string) bool {
	for _, s := range slice {
		if s == item {
			return true
		}
	}
	return false
}
