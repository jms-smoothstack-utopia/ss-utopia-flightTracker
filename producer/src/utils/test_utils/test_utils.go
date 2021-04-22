package test_utils

import "testing"

func AssertPanics(t *testing.T, f func()) {
	defer func() {
		if r := recover(); r == nil {
			t.Error("Expected panic but no panic occurred.")
		}
	}()
	f()
}

func ErrorIf(t *testing.T, condition bool, fieldName, expect, got string) {
	if condition {
		t.Errorf("Failure on Field: %v\nExpected: %v\tGot: %v", fieldName, expect, got)
	}
}