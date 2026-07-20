// SPDX-License-Identifier: GPL-3.0-or-later
// SPDX-FileCopyrightText: 2026 The MetaBoard authors

package ui

import (
	"runtime/debug"
	"testing"
)

func TestTruncateTitle(t *testing.T) {
	tests := []struct {
		input  string
		budget int
		want   string
	}{
		{"Short", 10, "Short"},
		{"Standard String", 10, "Standar..."},
		{"Multibyte string containing 🚀 and emoji", 15, "Multibyte st..."},
		{"こんにちは", 8, "こん..."}, // CJK double-width: "こ"(2) + "ん"(2) + "..."(3) = 7 (budget is 8)
		{"こんにちは", 3, "こんにちは"}, // budget <= 3 returns original
	}

	for _, tt := range tests {
		got := truncateTitle(tt.input, tt.budget)
		if got != tt.want {
			t.Errorf("truncateTitle(%q, %d) = %q; want %q", tt.input, tt.budget, got, tt.want)
		}
	}
}

func TestWrapText(t *testing.T) {
	tests := []struct {
		input string
		limit int
		want  string
	}{
		{"Short line", 10, "Short line"},
		{"This is a longer line of text", 10, "This is a\nlonger\nline of\ntext"},
		{"🚀 emoji text wrapper test", 12, "🚀 emoji\ntext wrapper\ntest"},
		{"CJK こんにちは 世界", 10, "CJK\nこんにちは\n世界"}, // CJK character width: 2 columns each
	}

	for _, tt := range tests {
		got := WrapText(tt.input, tt.limit)
		if got != tt.want {
			t.Errorf("WrapText(%q, %d) = %q; want %q", tt.input, tt.limit, got, tt.want)
		}
	}
}

func TestGetVersionStringDoesNotPanic(t *testing.T) {
	defer func() {
		if r := recover(); r != nil {
			t.Errorf("GetVersionString panicked: %v\nStack trace:\n%s", r, debug.Stack())
		}
	}()

	// Call it with a standard width
	res := GetVersionString(80)
	if len(res) == 0 {
		t.Error("expected non-empty version string")
	}

	// Call it with a very narrow width (boundary testing)
	resNarrow := GetVersionString(10)
	if len(resNarrow) == 0 {
		t.Error("expected non-empty version string at narrow width")
	}
}
