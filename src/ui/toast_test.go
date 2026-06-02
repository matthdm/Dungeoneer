package ui

import "testing"

func TestToast_NewToastHasPositiveTTL(t *testing.T) {
	toast := NewToast("hello")
	if toast.TTL <= 0 {
		t.Errorf("new toast should have positive TTL, got %f", toast.TTL)
	}
	if toast.Message != "hello" {
		t.Errorf("message not stored: want %q, got %q", "hello", toast.Message)
	}
}

func TestToast_UpdateDecreasesTTL(t *testing.T) {
	toast := NewToast("msg")
	before := toast.TTL
	toast.Update(0.5)
	if toast.TTL >= before {
		t.Error("TTL should decrease after Update")
	}
}

func TestToast_UpdateReturnsTrueWhenExpired(t *testing.T) {
	toast := NewToast("msg")
	// Drain all but a small sliver.
	toast.Update(toastDuration - 0.01)
	if toast.Update(0.02) == false {
		t.Error("Update should return true once TTL reaches zero")
	}
}

func TestToast_UpdateReturnsFalseBeforeExpiry(t *testing.T) {
	toast := NewToast("msg")
	if toast.Update(0.1) {
		t.Error("Update should return false before toastDuration elapses")
	}
}

func TestToast_FullDurationExpiry(t *testing.T) {
	toast := NewToast("msg")
	elapsed := 0.0
	expired := false
	for i := 0; i < 1000; i++ {
		elapsed += 0.05
		if toast.Update(0.05) {
			expired = true
			break
		}
	}
	if !expired {
		t.Errorf("toast never expired after %.1f seconds", elapsed)
	}
	if elapsed < toastDuration-0.1 {
		t.Errorf("toast expired too early at %.2fs (expected ~%.1fs)", elapsed, toastDuration)
	}
}
