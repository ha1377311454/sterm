package ui

import "testing"

func TestIsDeleteConfirm(t *testing.T) {
	tests := []struct {
		label string
		want  bool
	}{
		{label: "Delete", want: true},
		{label: "Cancel", want: false},
		{label: "", want: false},
	}

	for _, tt := range tests {
		if got := isDeleteConfirm(tt.label); got != tt.want {
			t.Fatalf("isDeleteConfirm(%q) = %v, want %v", tt.label, got, tt.want)
		}
	}
}
