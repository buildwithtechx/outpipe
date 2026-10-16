package validation

import (
	"encoding/json"
	"fmt"
	"slices"
	"strings"
)

type CaptureRule struct {
	ID         string   `json:"id"`
	Methods    []string `json:"methods"`
	PathPrefix string   `json:"pathPrefix"`
	Capture    bool     `json:"capture"`
}

var captureRuleMethods = map[string]bool{
	"GET": true, "POST": true, "PUT": true, "PATCH": true,
	"DELETE": true, "HEAD": true, "OPTIONS": true,
}

func ParseCaptureRules(raw string) ([]CaptureRule, error) {
	trimmed := strings.TrimSpace(raw)
	if trimmed == "" || trimmed == "[]" {
		return nil, nil
	}
	var rules []CaptureRule
	if err := json.Unmarshal([]byte(trimmed), &rules); err != nil {
		return nil, fmt.Errorf("decode capture rules: %w", err)
	}
	if len(rules) > 64 {
		return nil, fmt.Errorf("capture policy exceeds 64 rules")
	}
	seen := map[string]bool{}
	for i := range rules {
		rule := &rules[i]
		rule.ID = strings.TrimSpace(rule.ID)
		rule.PathPrefix = strings.TrimSpace(rule.PathPrefix)
		if rule.ID == "" || len(rule.ID) > 64 {
			return nil, fmt.Errorf("capture rule id is required")
		}
		if seen[rule.ID] {
			return nil, fmt.Errorf("duplicate capture rule id %q", rule.ID)
		}
		seen[rule.ID] = true
		if !strings.HasPrefix(rule.PathPrefix, "/") || len(rule.PathPrefix) > 256 {
			return nil, fmt.Errorf("capture rule %q needs a path prefix", rule.ID)
		}
		if len(rule.Methods) == 0 || len(rule.Methods) > 7 {
			return nil, fmt.Errorf("capture rule %q needs between 1 and 7 methods", rule.ID)
		}
		methods := make([]string, 0, len(rule.Methods))
		for _, method := range rule.Methods {
			method = strings.ToUpper(strings.TrimSpace(method))
			if !captureRuleMethods[method] {
				return nil, fmt.Errorf("capture rule %q uses unsupported method %q", rule.ID, method)
			}
			methods = append(methods, method)
		}
		rule.Methods = methods
	}
	return rules, nil
}

func MarshalCaptureRules(rules []CaptureRule) (string, error) {
	if len(rules) == 0 {
		return "[]", nil
	}
	encoded, err := json.Marshal(rules)
	if err != nil {
		return "", fmt.Errorf("encode capture rules: %w", err)
	}
	return string(encoded), nil
}

func CaptureAllowed(rules []CaptureRule, method, path string) bool {
	if len(rules) == 0 {
		return true
	}
	method = strings.ToUpper(strings.TrimSpace(method))
	if !strings.HasPrefix(path, "/") {
		return false
	}
	for _, rule := range rules {
		if !strings.HasPrefix(path, rule.PathPrefix) {
			continue
		}
		matched := slices.Contains(rule.Methods, method)
		if !matched {
			continue
		}
		return rule.Capture
	}
	return false
}
