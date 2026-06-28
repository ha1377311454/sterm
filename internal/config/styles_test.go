package config

import "testing"

func TestBuiltInSolarizedSkin(t *testing.T) {
	styles := NewStyles()
	if err := styles.Load("solarized"); err != nil {
		t.Fatalf("load solarized skin: %v", err)
	}

	if got := styles.Body().BgColor; got != "#002b36" {
		t.Fatalf("solarized body bg = %q, want %q", got, "#002b36")
	}
	if !containsSkin(AvailableSkins(), "solarized") {
		t.Fatal("expected solarized in available skins")
	}
}

func TestTealThemeAlias(t *testing.T) {
	styles := NewStyles()
	if err := styles.Load("teal"); err != nil {
		t.Fatalf("load teal alias: %v", err)
	}
	if got := styles.Body().BgColor; got != "#002b36" {
		t.Fatalf("teal alias body bg = %q, want %q", got, "#002b36")
	}
}
