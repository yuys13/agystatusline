package widgets

import (
	"fmt"
	"math"
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
		t.Errorf("Expected default color %q, got %q", "brightWhite", w.GetDefaultColor())
	}
	if w.GetDisplayName() != "Quota" {
		t.Errorf("Expected display name %q, got %q", "Quota", w.GetDisplayName())
	}

	remaining1 := float64(0.5019274)
	resetInSecs1 := float64(8891)
	remaining2 := float64(1.0)
	resetInSecs2 := float64(17956)
	secs45 := float64(45)
	secs125 := float64(125)
	secs567440 := float64(567440)
	val80 := float64(0.8)

	standardCtx := types.RenderContext{
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
				"no-reset-secs": {
					RemainingFraction: &val80,
				},
				"only-reset-secs": {
					ResetInSeconds: &secs45,
				},
				"empty-info": {},
			},
		},
	}
	settings := types.DefaultSettings()

	tests := []struct {
		name          string
		item          types.WidgetItem
		ctx           types.RenderContext
		expectedTitle string
		expectedBody  string
	}{
		{
			name:          "Labeled Percentage + Reset (default)",
			item:          types.WidgetItem{Type: "quota", Key: "gemini-5h"},
			ctx:           standardCtx,
			expectedTitle: "gemini-5h",
			expectedBody:  "50.19% (2h 28m)",
		},
		{
			name:          "Raw Percentage + Reset (Raw = true)",
			item:          types.WidgetItem{Type: "quota", Key: "gemini-5h", Raw: true},
			ctx:           standardCtx,
			expectedTitle: "",
			expectedBody:  "50.19% (2h 28m)",
		},
		{
			name:          "Custom Text label + Reset",
			item:          types.WidgetItem{Type: "quota", Key: "gemini-5h", Text: "Gemini Q"},
			ctx:           standardCtx,
			expectedTitle: "Gemini Q",
			expectedBody:  "50.19% (2h 28m)",
		},
		{
			name:          "Percentage only (no reset seconds in info)",
			item:          types.WidgetItem{Type: "quota", Key: "no-reset-secs"},
			ctx:           standardCtx,
			expectedTitle: "no-reset-secs",
			expectedBody:  "80.00%",
		},
		{
			name:          "Reset only (no percentage in info)",
			item:          types.WidgetItem{Type: "quota", Key: "only-reset-secs"},
			ctx:           standardCtx,
			expectedTitle: "only-reset-secs",
			expectedBody:  "45s",
		},
		{
			name:          "Missing Key",
			item:          types.WidgetItem{Type: "quota"},
			ctx:           standardCtx,
			expectedTitle: "",
			expectedBody:  "",
		},
		{
			name:          "Unknown quota key",
			item:          types.WidgetItem{Type: "quota", Key: "invalid-key"},
			ctx:           standardCtx,
			expectedTitle: "",
			expectedBody:  "",
		},
		{
			name:          "Nil quota map",
			item:          types.WidgetItem{Type: "quota", Key: "gemini-5h"},
			ctx:           types.RenderContext{Data: types.StatusJSON{}},
			expectedTitle: "",
			expectedBody:  "",
		},
		{
			name:          "Nil fields in QuotaInfo",
			item:          types.WidgetItem{Type: "quota", Key: "empty-info"},
			ctx:           standardCtx,
			expectedTitle: "",
			expectedBody:  "",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			title, body, err := w.Render(tt.item, tt.ctx, settings)
			if err != nil {
				t.Fatalf("Render error: %v", err)
			}
			if title != tt.expectedTitle || body != tt.expectedBody {
				t.Errorf("Expected title %q and body %q, got title %q and body %q", tt.expectedTitle, tt.expectedBody, title, body)
			}
			if color := w.GetBodyColor(tt.item, tt.ctx); color != "brightWhite" {
				t.Errorf("Expected body color 'brightWhite', got %q", color)
			}
		})
	}
}

func TestQuotaBarWidget(t *testing.T) {
	RegisterAll()
	w := GetWidget("quota-bar")
	if w == nil {
		t.Fatalf("Quota bar widget not found")
	}

	if w.GetDefaultColor() != "brightGreen" {
		t.Errorf("Expected default color %q, got %q", "brightGreen", w.GetDefaultColor())
	}
	if w.GetDisplayName() != "Quota Bar" {
		t.Errorf("Expected display name %q, got %q", "Quota Bar", w.GetDisplayName())
	}

	settings := types.DefaultSettings()

	t.Run("Missing or nil quota data", func(t *testing.T) {
		item := types.WidgetItem{Type: "quota-bar", Key: "gemini-5h"}
		ctxs := []types.RenderContext{
			{Data: types.StatusJSON{}},
			{Data: types.StatusJSON{Quota: nil}},
		}
		for i, ctx := range ctxs {
			t.Run(fmt.Sprintf("Case %d", i+1), func(t *testing.T) {
				title, body, err := w.Render(item, ctx, settings)
				if err != nil {
					t.Fatalf("Render error: %v", err)
				}
				if title != "" || body != "" {
					t.Errorf("Expected empty output, got title %q, body %q", title, body)
				}
				if color := w.GetBodyColor(item, ctx); color != "brightGreen" {
					t.Errorf("Expected default color 'brightGreen', got %q", color)
				}
			})
		}
	})

	t.Run("Color thresholds", func(t *testing.T) {
		boundaryTests := []struct {
			fraction      float64
			expectedColor string
		}{
			{0.80, "brightGreen"},
			{0.50, "brightGreen"},
			{0.49, "brightYellow"},
			{0.30, "brightYellow"},
			{0.20, "brightYellow"},
			{0.19, "brightRed"},
			{0.10, "brightRed"},
			{0.05, "brightRed"},
		}
		item := types.WidgetItem{Type: "quota-bar", Key: "gemini-5h"}
		for _, tt := range boundaryTests {
			t.Run(fmt.Sprintf("Fraction %.2f", tt.fraction), func(t *testing.T) {
				ctx := types.RenderContext{
					Data: types.StatusJSON{
						Quota: map[string]types.QuotaInfo{
							"gemini-5h": {RemainingFraction: &tt.fraction},
						},
					},
				}
				if color := w.GetBodyColor(item, ctx); color != tt.expectedColor {
					t.Errorf("For fraction %.2f expected color %q, got %q", tt.fraction, tt.expectedColor, color)
				}
			})
		}
	})

	t.Run("Standard rendering and bar width", func(t *testing.T) {
		pcts := []float64{0.8, 0.3, 0.05}
		expectedColors := []string{"brightGreen", "brightYellow", "brightRed"}
		item := types.WidgetItem{Type: "quota-bar", Key: "gemini-5h"}

		for i, pct := range pcts {
			pctVal := pct
			t.Run(fmt.Sprintf("Pct %.0f%%", pctVal*100), func(t *testing.T) {
				ctx := types.RenderContext{
					Data: types.StatusJSON{
						Quota: map[string]types.QuotaInfo{
							"gemini-5h": {RemainingFraction: &pctVal},
						},
					},
				}

				color := w.GetBodyColor(item, ctx)
				if color != expectedColors[i] {
					t.Errorf("Expected body color %q, got %q", expectedColors[i], color)
				}

				title, output, err := w.Render(item, ctx, settings)
				if err != nil {
					t.Fatalf("Render error: %v", err)
				}
				expectedPctStr := fmt.Sprintf("%.1f%%", pctVal*100)
				if title != "5h" || !strings.Contains(output, expectedPctStr) {
					t.Errorf("Expected title '5h' and body containing %q, got title %q and body %q", expectedPctStr, title, output)
				}
				parts := strings.Split(output, " ")
				barRunes := utf8.RuneCountInString(parts[0])
				if barRunes != 10 {
					t.Errorf("Expected bar width of 10 characters, got %d (bar: %q)", barRunes, parts[0])
				}
			})
		}
	})

	t.Run("Title mapping variations", func(t *testing.T) {
		pct := 0.8
		tests := []struct {
			name          string
			item          types.WidgetItem
			ctx           types.RenderContext
			expectedTitle string
		}{
			{
				name:          "gemini-weekly maps to 7d",
				item:          types.WidgetItem{Type: "quota-bar", Key: "gemini-weekly"},
				ctx:           types.RenderContext{Data: types.StatusJSON{Quota: map[string]types.QuotaInfo{"gemini-weekly": {RemainingFraction: &pct}}}},
				expectedTitle: "7d",
			},
			{
				name:          "3p-weekly maps to 3p-7d",
				item:          types.WidgetItem{Type: "quota-bar", Key: "3p-weekly"},
				ctx:           types.RenderContext{Data: types.StatusJSON{Quota: map[string]types.QuotaInfo{"3p-weekly": {RemainingFraction: &pct}}}},
				expectedTitle: "3p-7d",
			},
			{
				name:          "Custom text label override",
				item:          types.WidgetItem{Type: "quota-bar", Key: "gemini-5h", Text: "My5h"},
				ctx:           types.RenderContext{Data: types.StatusJSON{Quota: map[string]types.QuotaInfo{"gemini-5h": {RemainingFraction: &pct}}}},
				expectedTitle: "My5h",
			},
			{
				name:          "Raw true returns empty title",
				item:          types.WidgetItem{Type: "quota-bar", Key: "gemini-5h", Raw: true},
				ctx:           types.RenderContext{Data: types.StatusJSON{Quota: map[string]types.QuotaInfo{"gemini-5h": {RemainingFraction: &pct}}}},
				expectedTitle: "",
			},
		}

		for _, tt := range tests {
			t.Run(tt.name, func(t *testing.T) {
				title, _, err := w.Render(tt.item, tt.ctx, settings)
				if err != nil {
					t.Fatalf("Render error: %v", err)
				}
				if title != tt.expectedTitle {
					t.Errorf("Expected title %q, got %q", tt.expectedTitle, title)
				}
			})
		}
	})

	t.Run("Fractional progress block characters", func(t *testing.T) {
		fractionalTests := []struct {
			pct         float64
			expectedBar string
			desc        string
		}{
			{pct: 0.28, expectedBar: "██▓·······", desc: "remainder >= 75 (28%)"},
			{pct: 0.27, expectedBar: "██▒·······", desc: "remainder >= 50 (27%)"},
			{pct: 0.23, expectedBar: "██░·······", desc: "remainder >= 25 (23%)"},
		}

		item := types.WidgetItem{Type: "quota-bar", Key: "gemini-5h"}
		for _, tt := range fractionalTests {
			t.Run(tt.desc, func(t *testing.T) {
				ctx := types.RenderContext{
					Data: types.StatusJSON{
						Quota: map[string]types.QuotaInfo{
							"gemini-5h": {RemainingFraction: &tt.pct},
						},
					},
				}
				_, output, err := w.Render(item, ctx, settings)
				if err != nil {
					t.Fatalf("Render error for %q: %v", tt.desc, err)
				}
				parts := strings.Split(output, " ")
				if parts[0] != tt.expectedBar {
					t.Errorf("For %q expected bar %q, got %q", tt.desc, tt.expectedBar, parts[0])
				}
			})
		}
	})

	t.Run("Reset time inclusion", func(t *testing.T) {
		pct := 0.5019
		resetSecs := 750.0
		ctxWithReset := types.RenderContext{
			Data: types.StatusJSON{
				Quota: map[string]types.QuotaInfo{
					"gemini-5h": {
						RemainingFraction: &pct,
						ResetInSeconds:    &resetSecs,
					},
				},
			},
		}
		itemWithReset := types.WidgetItem{
			Type: "quota-bar",
			Key:  "gemini-5h",
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
	})
}

func TestFormatResetInSeconds(t *testing.T) {
	ptr := func(f float64) *float64 { return &f }

	tests := []struct {
		name     string
		secs     *float64
		expected string
	}{
		{"Nil pointer", nil, ""},
		{"Negative seconds", ptr(-10), "0s"},
		{"Zero seconds", ptr(0), "0s"},
		{"45 seconds", ptr(45), "45s"},
		{"60 seconds (1m)", ptr(60), "1m"},
		{"90 seconds (1m 30s)", ptr(90), "1m 30s"},
		{"3600 seconds (1h)", ptr(3600), "1h"},
		{"3660 seconds (1h 1m)", ptr(3660), "1h 1m"},
		{"86400 seconds (1d)", ptr(86400), "1d"},
		{"90000 seconds (1d 1h)", ptr(90000), "1d 1h"},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			res := formatResetInSeconds(tt.secs)
			if res != tt.expected {
				t.Errorf("For secs=%v expected %q, got %q", tt.secs, tt.expected, res)
			}
		})
	}
}

func TestSpecificQuotaWidgets(t *testing.T) {
	RegisterAll()

	remaining1 := float64(0.5019274)
	resetInSecs1 := float64(8891)
	remaining2 := float64(0.90)
	resetInSecs2 := float64(604756)

	ctx := types.RenderContext{
		Data: types.StatusJSON{
			Quota: map[string]types.QuotaInfo{
				"gemini-5h": {
					RemainingFraction: &remaining1,
					ResetInSeconds:    &resetInSecs1,
				},
				"gemini-weekly": {
					RemainingFraction: &remaining2,
					ResetInSeconds:    &resetInSecs2,
				},
				"3p-5h": {
					RemainingFraction: &remaining1,
					ResetInSeconds:    &resetInSecs1,
				},
				"3p-weekly": {
					RemainingFraction: &remaining2,
					ResetInSeconds:    &resetInSecs2,
				},
			},
		},
	}
	settings := types.DefaultSettings()

	tests := []struct {
		widgetType    string
		expectedTitle string
		expectedBody  string
		expectedName  string
	}{
		{"quota-5h", "5h", "50.19% (2h 28m)", "Quota: 5h"},
		{"quota-7d", "7d", "90.00% (6d 23h)", "Quota: 7d"},
		{"quota-3p-5h", "3p-5h", "50.19% (2h 28m)", "Quota: 3P 5h"},
		{"quota-3p-7d", "3p-7d", "90.00% (6d 23h)", "Quota: 3P 7d"},
		{"quota-bar-5h", "5h", "█████····· 50.2% (2h 28m)", "Quota Bar: 5h"},
		{"quota-bar-7d", "7d", "█████████· 90.0% (6d 23h)", "Quota Bar: 7d"},
		{"quota-bar-3p-5h", "3p-5h", "█████····· 50.2% (2h 28m)", "Quota Bar: 3P 5h"},
		{"quota-bar-3p-7d", "3p-7d", "█████████· 90.0% (6d 23h)", "Quota Bar: 3P 7d"},
	}

	for _, tt := range tests {
		t.Run(tt.widgetType, func(t *testing.T) {
			w := GetWidget(tt.widgetType)
			if w == nil {
				t.Fatalf("Widget %q not found in registry", tt.widgetType)
			}
			if w.GetDisplayName() != tt.expectedName {
				t.Errorf("Expected display name %q, got %q", tt.expectedName, w.GetDisplayName())
			}
			title, body, err := w.Render(types.WidgetItem{Type: tt.widgetType}, ctx, settings)
			if err != nil {
				t.Fatalf("Render error: %v", err)
			}
			if title != tt.expectedTitle {
				t.Errorf("Expected title %q, got %q", tt.expectedTitle, title)
			}
			if body != tt.expectedBody {
				t.Errorf("Expected body %q, got %q", tt.expectedBody, body)
			}
		})
	}
}

func TestQuotaWidgets_ExtremeBoundaryValues(t *testing.T) {
	RegisterAll()
	settings := types.DefaultSettings()

	nanVal := math.NaN()
	infVal := math.Inf(1)
	negInfVal := math.Inf(-1)
	negFraction := -0.5
	negHugeFraction := -100.0
	zeroFraction := 0.0
	oneFraction := 1.0
	overOneFraction := 1.5
	hugeSeconds := 1e12
	negSeconds := -999.0

	tests := []struct {
		name          string
		widgetType    string
		quotaInfo     types.QuotaInfo
		expectNoPanic bool
		verify        func(t *testing.T, title, body, color string)
	}{
		{
			name:       "0.0 fraction (0% remaining)",
			widgetType: "quota-5h",
			quotaInfo:  types.QuotaInfo{RemainingFraction: &zeroFraction},
			verify: func(t *testing.T, title, body, color string) {
				if title != "5h" || body != "0.00%" {
					t.Errorf("Expected 5h and 0.00%%, got title=%q, body=%q", title, body)
				}
				if color != "brightWhite" {
					t.Errorf("Expected brightWhite, got %q", color)
				}
			},
		},
		{
			name:       "0.0 fraction on Quota Bar (0% remaining)",
			widgetType: "quota-bar-5h",
			quotaInfo:  types.QuotaInfo{RemainingFraction: &zeroFraction},
			verify: func(t *testing.T, title, body, color string) {
				if title != "5h" || !strings.HasPrefix(body, "·········· 0.0%") {
					t.Errorf("Expected empty bar and 0.0%%, got title=%q, body=%q", title, body)
				}
				if color != "brightRed" {
					t.Errorf("Expected brightRed for 0%% bar, got %q", color)
				}
			},
		},
		{
			name:       "1.0 fraction (100% remaining)",
			widgetType: "quota-7d",
			quotaInfo:  types.QuotaInfo{RemainingFraction: &oneFraction},
			verify: func(t *testing.T, title, body, color string) {
				if title != "7d" || body != "100.00%" {
					t.Errorf("Expected 7d and 100.00%%, got title=%q, body=%q", title, body)
				}
				if color != "brightWhite" {
					t.Errorf("Expected brightWhite, got %q", color)
				}
			},
		},
		{
			name:       "1.0 fraction on Quota Bar (100% remaining)",
			widgetType: "quota-bar-7d",
			quotaInfo:  types.QuotaInfo{RemainingFraction: &oneFraction},
			verify: func(t *testing.T, title, body, color string) {
				if title != "7d" || !strings.HasPrefix(body, "██████████ 100.0%") {
					t.Errorf("Expected full bar and 100.0%%, got title=%q, body=%q", title, body)
				}
				if color != "brightGreen" {
					t.Errorf("Expected brightGreen for 100%% bar, got %q", color)
				}
			},
		},
		{
			name:       "Negative fraction (-0.5) on Quota Bar",
			widgetType: "quota-bar-3p-5h",
			quotaInfo:  types.QuotaInfo{RemainingFraction: &negFraction},
			verify: func(t *testing.T, title, body, color string) {
				if title != "3p-5h" || !strings.HasPrefix(body, "·········· -50.0%") {
					t.Errorf("Expected clamped empty bar and -50.0%%, got title=%q, body=%q", title, body)
				}
				if color != "brightRed" {
					t.Errorf("Expected brightRed for negative bar, got %q", color)
				}
			},
		},
		{
			name:       "Over 100% fraction (1.5) on Quota Bar",
			widgetType: "quota-bar-3p-7d",
			quotaInfo:  types.QuotaInfo{RemainingFraction: &overOneFraction},
			verify: func(t *testing.T, title, body, color string) {
				if title != "3p-7d" || !strings.HasPrefix(body, "██████████ 150.0%") {
					t.Errorf("Expected clamped full bar and 150.0%%, got title=%q, body=%q", title, body)
				}
				if color != "brightGreen" {
					t.Errorf("Expected brightGreen for over 100%% bar, got %q", color)
				}
			},
		},
		{
			name:       "Negative huge fraction (-100.0)",
			widgetType: "quota-bar-5h",
			quotaInfo:  types.QuotaInfo{RemainingFraction: &negHugeFraction},
			verify: func(t *testing.T, title, body, color string) {
				if title != "5h" || !strings.HasPrefix(body, "·········· -10000.0%") {
					t.Errorf("Expected clamped empty bar and -10000.0%%, got title=%q, body=%q", title, body)
				}
			},
		},
		{
			name:       "NaN fraction on Quota",
			widgetType: "quota-5h",
			quotaInfo:  types.QuotaInfo{RemainingFraction: &nanVal},
			verify: func(t *testing.T, title, body, color string) {
				if title != "5h" || !strings.Contains(body, "NaN%") {
					t.Errorf("Expected 5h and NaN%%, got title=%q, body=%q", title, body)
				}
			},
		},
		{
			name:       "NaN fraction on Quota Bar",
			widgetType: "quota-bar-5h",
			quotaInfo:  types.QuotaInfo{RemainingFraction: &nanVal},
			verify: func(t *testing.T, title, body, color string) {
				if title != "5h" || !strings.Contains(body, "NaN%") {
					t.Errorf("Expected 5h and NaN%%, got title=%q, body=%q", title, body)
				}
				if color != "brightRed" {
					t.Errorf("Expected brightRed for NaN, got %q", color)
				}
			},
		},
		{
			name:       "+Inf and -Inf fraction",
			widgetType: "quota-5h",
			quotaInfo:  types.QuotaInfo{RemainingFraction: &infVal},
			verify: func(t *testing.T, title, body, color string) {
				if title != "5h" || !strings.Contains(body, "+Inf%") {
					t.Errorf("Expected +Inf%%, got body=%q", body)
				}
			},
		},
		{
			name:       "Huge seconds and negative seconds",
			widgetType: "quota-5h",
			quotaInfo:  types.QuotaInfo{RemainingFraction: &oneFraction, ResetInSeconds: &hugeSeconds},
			verify: func(t *testing.T, title, body, color string) {
				if !strings.Contains(body, "d") {
					t.Errorf("Expected days in huge reset, got body=%q", body)
				}
			},
		},
		{
			name:       "Negative reset seconds (-999)",
			widgetType: "quota-5h",
			quotaInfo:  types.QuotaInfo{RemainingFraction: &oneFraction, ResetInSeconds: &negSeconds},
			verify: func(t *testing.T, title, body, color string) {
				if !strings.Contains(body, "(0s)") {
					t.Errorf("Expected (0s) for negative reset seconds, got body=%q", body)
				}
			},
		},
		{
			name:       "NaN reset seconds",
			widgetType: "quota-5h",
			quotaInfo:  types.QuotaInfo{RemainingFraction: &oneFraction, ResetInSeconds: &nanVal},
			verify: func(t *testing.T, title, body, color string) {
				if !strings.Contains(body, "(0s)") {
					t.Errorf("Expected (0s) for NaN reset seconds, got body=%q", body)
				}
			},
		},
	}

	_ = negInfVal

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			w := GetWidget(tt.widgetType)
			if w == nil {
				t.Fatalf("Widget %q not found", tt.widgetType)
			}

			key := "gemini-5h"
			switch tt.widgetType {
			case "quota-7d", "quota-bar-7d":
				key = "gemini-weekly"
			case "quota-3p-5h", "quota-bar-3p-5h":
				key = "3p-5h"
			case "quota-3p-7d", "quota-bar-3p-7d":
				key = "3p-weekly"
			}

			ctx := types.RenderContext{
				Data: types.StatusJSON{
					Quota: map[string]types.QuotaInfo{
						key: tt.quotaInfo,
					},
				},
			}

			item := types.WidgetItem{Type: tt.widgetType}
			title, body, err := w.Render(item, ctx, settings)
			if err != nil {
				t.Fatalf("Unexpected render error: %v", err)
			}
			color := w.GetBodyColor(item, ctx)
			if tt.verify != nil {
				tt.verify(t, title, body, color)
			}
		})
	}
}
