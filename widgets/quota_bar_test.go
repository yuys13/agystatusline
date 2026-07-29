package widgets

import (
	"fmt"
	"github.com/yuys13/agystatusline/types"
	"strings"
	"testing"
	"unicode/utf8"
)

func TestQuotaBarWidget(t *testing.T) {
	RegisterAll()
	w := GetWidget("quota-bar")
	if w == nil {
		t.Fatalf("Quota bar widget not found")
	}

	settings := types.DefaultSettings()

	// Test 1: No quota data
	ctxNoData := types.RenderContext{
		Data: types.StatusJSON{},
	}
	item := types.WidgetItem{
		Type:     "quota-bar",
		Metadata: map[string]string{"key": "gemini-5h"},
	}
	title, output, err := w.Render(item, ctxNoData, settings)
	if err != nil {
		t.Fatalf("Render error: %v", err)
	}
	if title != "" || output != "" {
		t.Errorf("Expected empty output for missing quota data, got title %q and body %q", title, output)
	}

	// Test 2: Standard rendering with different percentages
	pcts := []float64{0.8, 0.3, 0.05} // 80%, 30%, 5%
	expectedColors := []string{"brightGreen", "brightYellow", "brightRed"}

	for i, pct := range pcts {
		ctx := types.RenderContext{
			Data: types.StatusJSON{
				Quota: map[string]types.QuotaInfo{
					"gemini-5h": {
						RemainingFraction: &pct,
					},
				},
			},
		}

		// Verify GetBodyColor
		color := w.GetBodyColor(item, ctx)
		if color != expectedColors[i] {
			t.Errorf("For pct %.1f, expected body color %s, got %s", pct*100, expectedColors[i], color)
		}

		// Verify normal Render
		title, output, err = w.Render(item, ctx, settings)
		if err != nil {
			t.Fatalf("Render error: %v", err)
		}
		expectedPctStr := fmt.Sprintf("%.1f%%", pct*100)
		if title != "5h" || !strings.Contains(output, expectedPctStr) {
			t.Errorf("Expected title '5h' and body containing %q, got title %q and body %q", expectedPctStr, title, output)
		}
		parts := strings.Split(output, " ")
		barRunes := utf8.RuneCountInString(parts[0])
		if barRunes != 10 {
			t.Errorf("Expected quota bar width of 10 characters, got %d (bar: %q)", barRunes, parts[0])
		}
		// Verify gemini-weekly maps to '7d'
		itemWeekly := types.WidgetItem{
			Type:     "quota-bar",
			Metadata: map[string]string{"key": "gemini-weekly"},
		}
		ctxWeekly := types.RenderContext{
			Data: types.StatusJSON{
				Quota: map[string]types.QuotaInfo{
					"gemini-weekly": {
						RemainingFraction: &pct,
					},
				},
			},
		}
		titleW, _, err := w.Render(itemWeekly, ctxWeekly, settings)
		if err != nil {
			t.Fatalf("Render error: %v", err)
		}
		if titleW != "7d" {
			t.Errorf("Expected title '7d', got %q", titleW)
		}

		// Verify 3p-weekly maps to '3p-7d'
		item3PWeekly := types.WidgetItem{
			Type:     "quota-bar",
			Metadata: map[string]string{"key": "3p-weekly"},
		}
		ctx3PWeekly := types.RenderContext{
			Data: types.StatusJSON{
				Quota: map[string]types.QuotaInfo{
					"3p-weekly": {
						RemainingFraction: &pct,
					},
				},
			},
		}
		title3PW, _, err := w.Render(item3PWeekly, ctx3PWeekly, settings)
		if err != nil {
			t.Fatalf("Render error: %v", err)
		}
		if title3PW != "3p-7d" {
			t.Errorf("Expected title '3p-7d', got %q", title3PW)
		}

		// Verify RawValue
		itemRaw := types.WidgetItem{
			Type:     "quota-bar",
			Metadata: map[string]string{"key": "gemini-5h"},
			RawValue: func(b bool) *bool { return &b }(true),
		}
		titleR, outputR, err := w.Render(itemRaw, ctx, settings)
		if err != nil {
			t.Fatalf("Render error: %v", err)
		}
		if titleR != "" || !strings.Contains(outputR, expectedPctStr) {
			t.Errorf("Expected empty title and body containing %q, got title %q and body %q", expectedPctStr, titleR, outputR)
		}
	}

	// Test 3: Specific boundaries (50% should be green, 49% yellow, 10% yellow, 9% red)
	boundaryTests := []struct {
		fraction float64
		expected string
	}{
		{0.50, "brightGreen"},
		{0.49, "brightYellow"},
		{0.10, "brightYellow"},
		{0.09, "brightRed"},
	}
	for _, tc := range boundaryTests {
		ctx := types.RenderContext{
			Data: types.StatusJSON{
				Quota: map[string]types.QuotaInfo{
					"gemini-5h": {
						RemainingFraction: &tc.fraction,
					},
				},
			},
		}
		color := w.GetBodyColor(item, ctx)
		if color != tc.expected {
			t.Errorf("For boundary fraction %.2f, expected color %s, got %s", tc.fraction, tc.expected, color)
		}
	}

	// Test 3: Fractional progress block characters (▓, ▒, ░)
	fractionalTests := []struct {
		pct         float64
		expectedBar string
		desc        string
	}{
		{pct: 0.28, expectedBar: "██▓·······", desc: "remainder >= 75 (28%)"},
		{pct: 0.27, expectedBar: "██▒·······", desc: "remainder >= 50 (27%)"},
		{pct: 0.23, expectedBar: "██░·······", desc: "remainder >= 25 (23%)"},
	}

	for _, tt := range fractionalTests {
		ctx := types.RenderContext{
			Data: types.StatusJSON{
				Quota: map[string]types.QuotaInfo{
					"gemini-5h": {RemainingFraction: &tt.pct},
				},
			},
		}
		_, output, err := w.Render(item, ctx, settings)
		if err != nil {
			t.Fatalf("Render error for %s: %v", tt.desc, err)
		}
		parts := strings.Split(output, " ")
		if parts[0] != tt.expectedBar {
			t.Errorf("For %s, expected bar %q, got %q", tt.desc, tt.expectedBar, parts[0])
		}
	}

	// Test 4: Reset time inclusion
	resetSecs := 750.0 // 12m 30s
	ctxWithReset := types.RenderContext{
		Data: types.StatusJSON{
			Quota: map[string]types.QuotaInfo{
				"gemini-5h": {
					RemainingFraction: func(f float64) *float64 { return &f }(0.5019),
					ResetInSeconds:    &resetSecs,
				},
			},
		},
	}
	itemWithReset := types.WidgetItem{
		Type:     "quota-bar",
		Metadata: map[string]string{"key": "gemini-5h"},
	}
	titleReset, outputReset, err := w.Render(itemWithReset, ctxWithReset, settings)
	if err != nil {
		t.Fatalf("Render error with reset: %v", err)
	}
	if titleReset != "5h" {
		t.Errorf("Expected title '5h', got %q", titleReset)
	}
	expectedOutput := "█████····· 50.2% (12m 30s)"
	if outputReset != expectedOutput {
		t.Errorf("Expected body %q, got %q", expectedOutput, outputReset)
	}
}

func TestQuotaBarWidget_EdgeCases(t *testing.T) {
	RegisterAll()
	w := GetWidget("quota-bar")
	settings := types.DefaultSettings()

	// Nil Quota map
	ctxNil := types.RenderContext{Data: types.StatusJSON{Quota: nil}}
	item := types.WidgetItem{
		Type:     "quota-bar",
		Metadata: map[string]string{"key": "api"},
	}
	title, body, err := w.Render(item, ctxNil, settings)
	if err != nil || title != "" || body != "" {
		t.Errorf("Expected empty result for nil quota map, got title=%q, body=%q", title, body)
	}
	if color := w.GetBodyColor(item, ctxNil); color != "brightGreen" {
		t.Errorf("Expected default 'brightGreen' body color for nil quota, got %q", color)
	}

	// Test color thresholds (RemainingFraction: 80% -> brightGreen, 30% -> brightYellow, 5% -> brightRed)
	val80 := 0.8
	val30 := 0.3
	val05 := 0.05

	ctx80 := types.RenderContext{Data: types.StatusJSON{Quota: map[string]types.QuotaInfo{"api": {RemainingFraction: &val80}}}}
	ctx30 := types.RenderContext{Data: types.StatusJSON{Quota: map[string]types.QuotaInfo{"api": {RemainingFraction: &val30}}}}
	ctx05 := types.RenderContext{Data: types.StatusJSON{Quota: map[string]types.QuotaInfo{"api": {RemainingFraction: &val05}}}}

	if c := w.GetBodyColor(item, ctx80); c != "brightGreen" {
		t.Errorf("Expected brightGreen for 80%%, got %q", c)
	}
	if c := w.GetBodyColor(item, ctx30); c != "brightYellow" {
		t.Errorf("Expected brightYellow for 30%%, got %q", c)
	}
	if c := w.GetBodyColor(item, ctx05); c != "brightRed" {
		t.Errorf("Expected brightRed for 5%%, got %q", c)
	}
}
