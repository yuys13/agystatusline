package widgets

import (
	"fmt"
	"strings"
	"testing"
	"unicode/utf8"

	"github.com/yuys13/agystatusline/types"
)

func TestQuotaWidget(t *testing.T) {
	RegisterAll()
	w := GetWidget("quota")
	if w == nil {
		t.Fatalf("Quota widget not found")
	}

	if w.GetDefaultColor() != "brightWhite" {
		t.Errorf("Expected default color 'brightWhite', got '%s'", w.GetDefaultColor())
	}

	remaining1 := float64(0.5019274)
	resetInSecs1 := float64(8891)
	remaining2 := float64(1.0)
	resetInSecs2 := float64(17956)

	ctx := types.RenderContext{
		Data: types.StatusJSON{
			Quota: map[string]types.QuotaInfo{
				"gemini-5h": {
					RemainingFraction: &remaining1,
					ResetTime:         "2026-06-20T11:27:27Z",
					ResetInSeconds:    &resetInSecs1,
				},
				"3p-5h": {
					RemainingFraction: &remaining2,
					ResetTime:         "2026-06-20T13:58:32Z",
					ResetInSeconds:    &resetInSecs2,
				},
			},
		},
	}
	settings := types.DefaultSettings()

	// Case 1: Labeled Percentage + Reset (default)
	item1 := types.WidgetItem{
		Type: "quota",
		Metadata: map[string]string{
			"key": "gemini-5h",
		},
	}
	title1, output1, err := w.Render(item1, ctx, settings)
	if err != nil {
		t.Fatalf("Render error: %v", err)
	}
	if title1 != "gemini-5h" || output1 != "50.19% (2h 28m)" {
		t.Errorf("Expected title 'gemini-5h' and body '50.19%% (2h 28m)', got title '%s' and body '%s'", title1, output1)
	}

	// Case 2: Raw Percentage + Reset (rawValue = true, default)
	rawVal := true
	item2 := types.WidgetItem{
		Type: "quota",
		Metadata: map[string]string{
			"key": "gemini-5h",
		},
		RawValue: &rawVal,
	}
	title2, output2, err := w.Render(item2, ctx, settings)
	if err != nil {
		t.Fatalf("Render error: %v", err)
	}
	if title2 != "" || output2 != "50.19% (2h 28m)" {
		t.Errorf("Expected title '' and body '50.19%% (2h 28m)', got title '%s' and body '%s'", title2, output2)
	}

	// Case 3: Custom Text label + Reset
	item3 := types.WidgetItem{
		Type: "quota",
		Metadata: map[string]string{
			"key": "gemini-5h",
		},
		CustomText: "Gemini Q",
	}
	title3, output3, err := w.Render(item3, ctx, settings)
	if err != nil {
		t.Fatalf("Render error: %v", err)
	}
	if title3 != "Gemini Q" || output3 != "50.19% (2h 28m)" {
		t.Errorf("Expected title 'Gemini Q' and body '50.19%% (2h 28m)', got title '%s' and body '%s'", title3, output3)
	}

	// Case 3b: display="quota" (Percentage only, labeled)
	itemQuotaOnly := types.WidgetItem{
		Type: "quota",
		Metadata: map[string]string{
			"key":     "gemini-5h",
			"display": "quota",
		},
	}
	titleQuotaOnly, outputQuotaOnly, err := w.Render(itemQuotaOnly, ctx, settings)
	if err != nil {
		t.Fatalf("Render error: %v", err)
	}
	if titleQuotaOnly != "gemini-5h" || outputQuotaOnly != "50.19%" {
		t.Errorf("Expected title 'gemini-5h' and body '50.19%%', got title '%s' and body '%s'", titleQuotaOnly, outputQuotaOnly)
	}

	// Case 3c: display="quota" (Percentage only, raw)
	itemQuotaOnlyRaw := types.WidgetItem{
		Type: "quota",
		Metadata: map[string]string{
			"key":     "gemini-5h",
			"display": "quota",
		},
		RawValue: &rawVal,
	}
	titleQuotaOnlyRaw, outputQuotaOnlyRaw, err := w.Render(itemQuotaOnlyRaw, ctx, settings)
	if err != nil {
		t.Fatalf("Render error: %v", err)
	}
	if titleQuotaOnlyRaw != "" || outputQuotaOnlyRaw != "50.19%" {
		t.Errorf("Expected title '' and body '50.19%%', got title '%s' and body '%s'", titleQuotaOnlyRaw, outputQuotaOnlyRaw)
	}

	// Case 4: Reset time labeled
	item4 := types.WidgetItem{
		Type: "quota",
		Metadata: map[string]string{
			"key":     "gemini-5h",
			"display": "reset",
		},
	}
	title4, output4, err := w.Render(item4, ctx, settings)
	if err != nil {
		t.Fatalf("Render error: %v", err)
	}
	if title4 != "gemini-5h (reset)" || output4 != "2h 28m" {
		t.Errorf("Expected title 'gemini-5h (reset)' and body '2h 28m', got title '%s' and body '%s'", title4, output4)
	}

	// Case 5: Reset time raw
	item5 := types.WidgetItem{
		Type: "quota",
		Metadata: map[string]string{
			"key":     "gemini-5h",
			"display": "reset",
		},
		RawValue: &rawVal,
	}
	title5, output5, err := w.Render(item5, ctx, settings)
	if err != nil {
		t.Fatalf("Render error: %v", err)
	}
	if title5 != "" || output5 != "2h 28m" {
		t.Errorf("Expected title '' and body '2h 28m', got title '%s' and body '%s'", title5, output5)
	}

	// Case 6: Reset time other durations
	secs45 := float64(45)
	secs125 := float64(125)
	secs567440 := float64(567440)
	ctxDur := types.RenderContext{
		Data: types.StatusJSON{
			Quota: map[string]types.QuotaInfo{
				"secs45": {
					RemainingFraction: &remaining2,
					ResetInSeconds:    &secs45,
				},
				"secs125": {
					RemainingFraction: &remaining2,
					ResetInSeconds:    &secs125,
				},
				"secs567440": {
					RemainingFraction: &remaining2,
					ResetInSeconds:    &secs567440,
				},
			},
		},
	}
	itemSecs45 := types.WidgetItem{
		Type: "quota",
		Metadata: map[string]string{
			"key":     "secs45",
			"display": "reset",
		},
		RawValue: &rawVal,
	}
	title45, output45, _ := w.Render(itemSecs45, ctxDur, settings)
	if title45 != "" || output45 != "45s" {
		t.Errorf("Expected title '' and body '45s', got title '%s' and body '%s'", title45, output45)
	}

	itemSecs125 := types.WidgetItem{
		Type: "quota",
		Metadata: map[string]string{
			"key":     "secs125",
			"display": "reset",
		},
		RawValue: &rawVal,
	}
	title125, output125, _ := w.Render(itemSecs125, ctxDur, settings)
	if title125 != "" || output125 != "2m 5s" {
		t.Errorf("Expected title '' and body '2m 5s', got title '%s' and body '%s'", title125, output125)
	}

	itemSecs567440 := types.WidgetItem{
		Type: "quota",
		Metadata: map[string]string{
			"key":     "secs567440",
			"display": "reset",
		},
		RawValue: &rawVal,
	}
	title567440, output567440, _ := w.Render(itemSecs567440, ctxDur, settings)
	if title567440 != "" || output567440 != "6d 13h" {
		t.Errorf("Expected title '' and body '6d 13h', got title '%s' and body '%s'", title567440, output567440)
	}

	// Case 7: Key not found or empty
	itemEmpty := types.WidgetItem{
		Type: "quota",
	}
	titleEmpty, outputEmpty, _ := w.Render(itemEmpty, ctx, settings)
	if titleEmpty != "" || outputEmpty != "" {
		t.Errorf("Expected empty title/body for empty key, got title '%s' and body '%s'", titleEmpty, outputEmpty)
	}

	itemInvalid := types.WidgetItem{
		Type: "quota",
		Metadata: map[string]string{
			"key": "invalid-key",
		},
	}
	titleInvalid, outputInvalid, _ := w.Render(itemInvalid, ctx, settings)
	if titleInvalid != "" || outputInvalid != "" {
		t.Errorf("Expected empty title/body for invalid key, got title '%s' and body '%s'", titleInvalid, outputInvalid)
	}

	// Case 8: Quota map is nil
	ctxNilQuota := types.RenderContext{
		Data: types.StatusJSON{},
	}
	titleNil, outputNil, _ := w.Render(item1, ctxNilQuota, settings)
	if titleNil != "" || outputNil != "" {
		t.Errorf("Expected empty title/body for nil quota map, got title '%s' and body '%s'", titleNil, outputNil)
	}
}

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

func TestFormatResetInSeconds(t *testing.T) {
	tests := []struct {
		secs     *float64
		expected string
	}{
		{nil, ""},
		{func(f float64) *float64 { return &f }(-10), "0s"},
		{func(f float64) *float64 { return &f }(0), "0s"},
		{func(f float64) *float64 { return &f }(45), "45s"},
		{func(f float64) *float64 { return &f }(60), "1m"},
		{func(f float64) *float64 { return &f }(90), "1m 30s"},
		{func(f float64) *float64 { return &f }(3600), "1h"},
		{func(f float64) *float64 { return &f }(3660), "1h 1m"},
		{func(f float64) *float64 { return &f }(86400), "1d"},
		{func(f float64) *float64 { return &f }(90000), "1d 1h"},
	}

	for _, tc := range tests {
		res := formatResetInSeconds(tc.secs)
		if res != tc.expected {
			t.Errorf("For secs=%v expected %q, got %q", tc.secs, tc.expected, res)
		}
	}
}

func TestQuotaWidget_EdgeCases(t *testing.T) {
	RegisterAll()
	w := GetWidget("quota")
	settings := types.DefaultSettings()

	// Nil Quota
	ctxNil := types.RenderContext{Data: types.StatusJSON{}}
	item := types.WidgetItem{Type: "quota", Metadata: map[string]string{"key": "gemini-5h"}}
	title, body, err := w.Render(item, ctxNil, settings)
	if err != nil || title != "" || body != "" {
		t.Errorf("Expected empty result for nil Quota map")
	}

	// Missing Key in Metadata
	ctxData := types.RenderContext{
		Data: types.StatusJSON{
			Quota: map[string]types.QuotaInfo{
				"gemini-5h": {RemainingFraction: func(f float64) *float64 { return &f }(0.8)},
			},
		},
	}
	itemNoKey := types.WidgetItem{Type: "quota"}
	_, bodyNoKey, _ := w.Render(itemNoKey, ctxData, settings)
	if bodyNoKey != "" {
		t.Errorf("Expected empty body when metadata key is missing")
	}

	// Unknown Key
	itemBadKey := types.WidgetItem{Type: "quota", Metadata: map[string]string{"key": "unknown"}}
	_, bodyBadKey, _ := w.Render(itemBadKey, ctxData, settings)
	if bodyBadKey != "" {
		t.Errorf("Expected empty body for unknown quota key")
	}

	// Display mode "quota"
	itemDisplayQuota := types.WidgetItem{Type: "quota", Metadata: map[string]string{"key": "gemini-5h", "display": "quota"}}
	titleQ, bodyQ, _ := w.Render(itemDisplayQuota, ctxData, settings)
	if titleQ != "gemini-5h" || bodyQ != "80.00%" {
		t.Errorf("Expected 'gemini-5h' and '80.00%%', got title=%q body=%q", titleQ, bodyQ)
	}

	// Display mode "reset"
	itemDisplayReset := types.WidgetItem{Type: "quota", Metadata: map[string]string{"key": "gemini-5h", "display": "reset"}}
	_, bodyReset, _ := w.Render(itemDisplayReset, ctxData, settings)
	if bodyReset != "" {
		t.Errorf("Expected empty reset body when ResetInSeconds is nil")
	}
}

func TestQuotaWidget_MoreEdgeCases(t *testing.T) {
	RegisterAll()
	w := GetWidget("quota")
	settings := types.DefaultSettings()

	// Missing metadata key
	itemNoKey := types.WidgetItem{Type: "quota"}
	ctx := types.RenderContext{
		Data: types.StatusJSON{
			Quota: map[string]types.QuotaInfo{
				"api": {ResetTime: "2026-06-20T12:00:00Z"},
			},
		},
	}
	title, body, err := w.Render(itemNoKey, ctx, settings)
	if err != nil || title != "" || body != "" {
		t.Errorf("Expected empty result when metadata key is missing, got title=%q, body=%q", title, body)
	}

	// Key not in Quota map
	itemKey := types.WidgetItem{
		Type:     "quota",
		Metadata: map[string]string{"key": "nonexistent"},
	}
	title, body, err = w.Render(itemKey, ctx, settings)
	if err != nil || title != "" || body != "" {
		t.Errorf("Expected empty result when key is non-existent, got title=%q, body=%q", title, body)
	}

	// QuotaInfo nil fields (RemainingFraction nil, ResetTime empty)
	itemApi := types.WidgetItem{
		Type:     "quota",
		Metadata: map[string]string{"key": "api", "display": "quota"},
	}
	ctxNilFields := types.RenderContext{
		Data: types.StatusJSON{
			Quota: map[string]types.QuotaInfo{
				"api": {},
			},
		},
	}
	title, body, err = w.Render(itemApi, ctxNilFields, settings)
	if err != nil || title != "" || body != "" {
		t.Errorf("Expected empty result when QuotaInfo has nil fields, got title=%q, body=%q", title, body)
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
