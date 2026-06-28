package config

import "testing"

func TestBuiltInTealSkin(t *testing.T) {
	styles := NewStyles()
	if err := styles.Load("teal"); err != nil {
		t.Fatalf("load teal skin: %v", err)
	}

	if got := styles.Body().BgColor; got != "#00464d" {
		t.Fatalf("teal body bg = %q, want %q", got, "#00464d")
	}
	if !containsSkin(AvailableSkins(), "teal") {
		t.Fatal("expected teal in available skins")
	}
}
