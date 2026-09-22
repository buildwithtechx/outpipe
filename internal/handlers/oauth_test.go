package handlers

import "testing"

func TestValidatedReturnPath(t *testing.T) {
	tests := []struct {
		name  string
		input string
		want  string
		valid bool
	}{
		{name: "default", want: "/", valid: true},
		{name: "relative dashboard path", input: "/acme/tunnels?view=active", want: "/acme/tunnels?view=active", valid: true},
		{name: "absolute url", input: "https://attacker.example", valid: false},
		{name: "protocol relative url", input: "//attacker.example", valid: false},
		{name: "relative path", input: "login", valid: false},
	}

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			got, err := validatedReturnPath(test.input)

			if test.valid && err != nil {
				t.Fatalf("validatedReturnPath(%q) returned error: %v", test.input, err)
			}

			if !test.valid && err == nil {
				t.Fatalf("validatedReturnPath(%q) succeeded", test.input)
			}

			if got != test.want {
				t.Fatalf("validatedReturnPath(%q) = %q, want %q", test.input, got, test.want)
			}

		})
	}
}
