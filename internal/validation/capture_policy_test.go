package validation

import (
	"testing"
)

func TestParseCaptureRules(t *testing.T) {
	rules, err := ParseCaptureRules(`[{"id":"api","methods":["get","POST"],"pathPrefix":"/api","capture":true}]`)
	if err != nil {
		t.Fatal(err)
	}
	if len(rules) != 1 || rules[0].Methods[0] != "GET" || rules[0].Methods[1] != "POST" {
		t.Fatalf("methods were not normalized: %+v", rules)
	}
	if _, err := ParseCaptureRules(`[{"id":"","methods":["GET"],"pathPrefix":"/","capture":true}]`); err == nil {
		t.Fatal("rule without id accepted")
	}
	if _, err := ParseCaptureRules(`[{"id":"a","methods":["BREW"],"pathPrefix":"/","capture":true}]`); err == nil {
		t.Fatal("rule with unknown method accepted")
	}
	if _, err := ParseCaptureRules(`[{"id":"a","methods":["GET"],"pathPrefix":"noslash","capture":true}]`); err == nil {
		t.Fatal("rule without path prefix accepted")
	}
	if _, err := ParseCaptureRules(`[{"id":"a","methods":["GET"],"pathPrefix":"/","capture":true},{"id":"a","methods":["POST"],"pathPrefix":"/","capture":false}]`); err == nil {
		t.Fatal("duplicate rule id accepted")
	}
	if _, err := ParseCaptureRules(`not json`); err == nil {
		t.Fatal("malformed rules accepted")
	}
	if rules, err := ParseCaptureRules(""); err != nil || len(rules) != 0 {
		t.Fatal("empty rules must parse to no rules")
	}
}

func TestCaptureAllowed(t *testing.T) {
	if !CaptureAllowed(nil, "GET", "/anything") {
		t.Fatal("absent policy must allow")
	}
	rules, err := ParseCaptureRules(`[{"id":"deny-admin","methods":["GET","POST"],"pathPrefix":"/admin","capture":false},{"id":"allow-api","methods":["GET"],"pathPrefix":"/api","capture":true}]`)
	if err != nil {
		t.Fatal(err)
	}
	if CaptureAllowed(rules, "GET", "/admin/users") {
		t.Fatal("denied prefix was allowed")
	}
	if !CaptureAllowed(rules, "GET", "/api/users") {
		t.Fatal("allowed prefix was denied")
	}
	if CaptureAllowed(rules, "POST", "/api/users") {
		t.Fatal("unlisted method on matched prefix was allowed instead of denied")
	}
	if CaptureAllowed(rules, "GET", "/other") {
		t.Fatal("unmatched path was allowed with rules present")
	}
	if CaptureAllowed(rules, "GET", "not-a-path") {
		t.Fatal("path without slash was allowed")
	}
}
