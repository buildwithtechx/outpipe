package validation

import "testing"

func TestStruct(t *testing.T) {

	if err := Struct(struct {
		Name string `validate:"required"`
	}{}); err == nil {
		t.Fatal("expected validation error")
	}
}
