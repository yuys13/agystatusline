package widgets

import (
	"fmt"
	"math"
	"reflect"
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
			expectedTitle: "Gemini 5h",
			expectedBody:  "50.19% (2h 28m)",
		},
		{
			name: "gemini-argon-wtu-weekly default label uses Argon 7d",
			item: types.WidgetItem{Type: "quota", Key: "gemini-argon-wtu-weekly"},
			ctx: types.RenderContext{
				Data: types.StatusJSON{
					Quota: map[string]types.QuotaInfo{
						"gemini-argon-wtu-weekly": {
							RemainingFraction: &remaining2,
							ResetInSeconds:    &resetInSecs2,
						},
					},
				},
			},
			expectedTitle: "Argon 7d",
			expectedBody:  "100.00% (4h 59m)",
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
				if title != "Gemini 5h" || !strings.Contains(output, expectedPctStr) {
					t.Errorf("Expected title 'Gemini 5h' and body containing %q, got title %q and body %q", expectedPctStr, title, output)
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
				name:          "gemini-weekly maps to Gemini 7d",
				item:          types.WidgetItem{Type: "quota-bar", Key: "gemini-weekly"},
				ctx:           types.RenderContext{Data: types.StatusJSON{Quota: map[string]types.QuotaInfo{"gemini-weekly": {RemainingFraction: &pct}}}},
				expectedTitle: "Gemini 7d",
			},
			{
				name:          "gemini-argon-wtu-weekly maps to Argon 7d",
				item:          types.WidgetItem{Type: "quota-bar", Key: "gemini-argon-wtu-weekly"},
				ctx:           types.RenderContext{Data: types.StatusJSON{Quota: map[string]types.QuotaInfo{"gemini-argon-wtu-weekly": {RemainingFraction: &pct}}}},
				expectedTitle: "Argon 7d",
			},
			{
				name:          "3p-weekly maps to Claude 7d",
				item:          types.WidgetItem{Type: "quota-bar", Key: "3p-weekly"},
				ctx:           types.RenderContext{Data: types.StatusJSON{Quota: map[string]types.QuotaInfo{"3p-weekly": {RemainingFraction: &pct}}}},
				expectedTitle: "Claude 7d",
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
		if titleReset != "Gemini 5h" {
			t.Errorf("Expected title 'Gemini 5h', got %q", titleReset)
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
				"gemini-argon-wtu-weekly": {
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
		{"quota-5h", "Gemini 5h", "50.19% (2h 28m)", "Quota: Gemini 5h"},
		{"quota-7d", "Gemini 7d", "90.00% (6d 23h)", "Quota: Gemini 7d"},
		{"quota-argon-7d", "Argon 7d", "90.00% (6d 23h)", "Quota: Argon 7d"},
		{"quota-3p-5h", "Claude 5h", "50.19% (2h 28m)", "Quota: Claude 5h"},
		{"quota-3p-7d", "Claude 7d", "90.00% (6d 23h)", "Quota: Claude 7d"},
		{"quota-bar-5h", "Gemini 5h", "█████····· 50.2% (2h 28m)", "Quota Bar: Gemini 5h"},
		{"quota-bar-7d", "Gemini 7d", "█████████· 90.0% (6d 23h)", "Quota Bar: Gemini 7d"},
		{"quota-bar-argon-7d", "Argon 7d", "█████████· 90.0% (6d 23h)", "Quota Bar: Argon 7d"},
		{"quota-bar-3p-5h", "Claude 5h", "█████····· 50.2% (2h 28m)", "Quota Bar: Claude 5h"},
		{"quota-bar-3p-7d", "Claude 7d", "█████████· 90.0% (6d 23h)", "Quota Bar: Claude 7d"},
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

	t.Run("PercentageBoundaries", func(t *testing.T) {
		tests := []struct {
			name           string
			widgetType     string
			quotaInfo      types.QuotaInfo
			wantTitle      string
			wantBodyPrefix string
			wantColor      string
		}{
			{
				name:           "0.0 fraction (0% remaining)",
				widgetType:     "quota-5h",
				quotaInfo:      types.QuotaInfo{RemainingFraction: &zeroFraction},
				wantTitle:      "Gemini 5h",
				wantBodyPrefix: "0.00%",
				wantColor:      "brightWhite",
			},
			{
				name:           "0.0 fraction on Quota Bar (0% remaining)",
				widgetType:     "quota-bar-5h",
				quotaInfo:      types.QuotaInfo{RemainingFraction: &zeroFraction},
				wantTitle:      "Gemini 5h",
				wantBodyPrefix: "·········· 0.0%",
				wantColor:      "brightRed",
			},
			{
				name:           "1.0 fraction (100% remaining)",
				widgetType:     "quota-7d",
				quotaInfo:      types.QuotaInfo{RemainingFraction: &oneFraction},
				wantTitle:      "Gemini 7d",
				wantBodyPrefix: "100.00%",
				wantColor:      "brightWhite",
			},
			{
				name:           "1.0 fraction (100% remaining) on Quota Argon",
				widgetType:     "quota-argon-7d",
				quotaInfo:      types.QuotaInfo{RemainingFraction: &oneFraction},
				wantTitle:      "Argon 7d",
				wantBodyPrefix: "100.00%",
				wantColor:      "brightWhite",
			},
			{
				name:           "1.0 fraction on Quota Bar (100% remaining)",
				widgetType:     "quota-bar-7d",
				quotaInfo:      types.QuotaInfo{RemainingFraction: &oneFraction},
				wantTitle:      "Gemini 7d",
				wantBodyPrefix: "██████████ 100.0%",
				wantColor:      "brightGreen",
			},
			{
				name:           "1.0 fraction on Quota Bar Argon (100% remaining)",
				widgetType:     "quota-bar-argon-7d",
				quotaInfo:      types.QuotaInfo{RemainingFraction: &oneFraction},
				wantTitle:      "Argon 7d",
				wantBodyPrefix: "██████████ 100.0%",
				wantColor:      "brightGreen",
			},
			{
				name:           "Negative fraction (-0.5) on Quota Bar",
				widgetType:     "quota-bar-3p-5h",
				quotaInfo:      types.QuotaInfo{RemainingFraction: &negFraction},
				wantTitle:      "Claude 5h",
				wantBodyPrefix: "·········· -50.0%",
				wantColor:      "brightRed",
			},
			{
				name:           "Over 100% fraction (1.5) on Quota Bar",
				widgetType:     "quota-bar-3p-7d",
				quotaInfo:      types.QuotaInfo{RemainingFraction: &overOneFraction},
				wantTitle:      "Claude 7d",
				wantBodyPrefix: "██████████ 150.0%",
				wantColor:      "brightGreen",
			},
			{
				name:           "Negative huge fraction (-100.0)",
				widgetType:     "quota-bar-5h",
				quotaInfo:      types.QuotaInfo{RemainingFraction: &negHugeFraction},
				wantTitle:      "Gemini 5h",
				wantBodyPrefix: "·········· -10000.0%",
				wantColor:      "brightRed",
			},
			{
				name:           "NaN fraction on Quota",
				widgetType:     "quota-5h",
				quotaInfo:      types.QuotaInfo{RemainingFraction: &nanVal},
				wantTitle:      "Gemini 5h",
				wantBodyPrefix: "NaN%",
				wantColor:      "brightWhite",
			},
			{
				name:           "NaN fraction on Quota Bar",
				widgetType:     "quota-bar-5h",
				quotaInfo:      types.QuotaInfo{RemainingFraction: &nanVal},
				wantTitle:      "Gemini 5h",
				wantBodyPrefix: "·········· NaN%",
				wantColor:      "brightRed",
			},
			{
				name:           "+Inf fraction on Quota",
				widgetType:     "quota-5h",
				quotaInfo:      types.QuotaInfo{RemainingFraction: &infVal},
				wantTitle:      "Gemini 5h",
				wantBodyPrefix: "+Inf%",
				wantColor:      "brightWhite",
			},
			{
				name:           "-Inf fraction on Quota",
				widgetType:     "quota-5h",
				quotaInfo:      types.QuotaInfo{RemainingFraction: &negInfVal},
				wantTitle:      "Gemini 5h",
				wantBodyPrefix: "-Inf%",
				wantColor:      "brightWhite",
			},
		}

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
				case "quota-argon-7d", "quota-bar-argon-7d":
					key = "gemini-argon-wtu-weekly"
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
				if title != tt.wantTitle {
					t.Errorf("Expected title %q, got %q", tt.wantTitle, title)
				}
				if !strings.HasPrefix(body, tt.wantBodyPrefix) {
					t.Errorf("Expected body prefix %q, got body %q", tt.wantBodyPrefix, body)
				}
				color := w.GetBodyColor(item, ctx)
				if color != tt.wantColor {
					t.Errorf("Expected color %q, got %q", tt.wantColor, color)
				}
			})
		}
	})

	t.Run("ResetSecondsBoundaries", func(t *testing.T) {
		tests := []struct {
			name             string
			widgetType       string
			quotaInfo        types.QuotaInfo
			wantBodyContains string
		}{
			{
				name:             "Huge seconds",
				widgetType:       "quota-5h",
				quotaInfo:        types.QuotaInfo{RemainingFraction: &oneFraction, ResetInSeconds: &hugeSeconds},
				wantBodyContains: "d",
			},
			{
				name:             "Negative reset seconds (-999)",
				widgetType:       "quota-5h",
				quotaInfo:        types.QuotaInfo{RemainingFraction: &oneFraction, ResetInSeconds: &negSeconds},
				wantBodyContains: "(0s)",
			},
			{
				name:             "NaN reset seconds",
				widgetType:       "quota-5h",
				quotaInfo:        types.QuotaInfo{RemainingFraction: &oneFraction, ResetInSeconds: &nanVal},
				wantBodyContains: "(0s)",
			},
		}

		for _, tt := range tests {
			t.Run(tt.name, func(t *testing.T) {
				w := GetWidget(tt.widgetType)
				if w == nil {
					t.Fatalf("Widget %q not found", tt.widgetType)
				}

				ctx := types.RenderContext{
					Data: types.StatusJSON{
						Quota: map[string]types.QuotaInfo{
							"gemini-5h": tt.quotaInfo,
						},
					},
				}

				item := types.WidgetItem{Type: tt.widgetType}
				_, body, err := w.Render(item, ctx, settings)
				if err != nil {
					t.Fatalf("Unexpected render error: %v", err)
				}
				if !strings.Contains(body, tt.wantBodyContains) {
					t.Errorf("Expected body to contain %q, got %q", tt.wantBodyContains, body)
				}
			})
		}
	})
}

func TestQuotaWidgets_CustomSettingsLabels(t *testing.T) {
	RegisterAll()

	fraction := 0.75
	ctx := types.RenderContext{
		Data: types.StatusJSON{
			Quota: map[string]types.QuotaInfo{
				"gemini-5h":     {RemainingFraction: &fraction},
				"3p-5h":         {RemainingFraction: &fraction},
				"custom-window": {RemainingFraction: &fraction},
			},
		},
	}

	customSettings := types.DefaultSettings()
	customSettings.Quota.Labels["gemini-5h"] = "Gem5"
	customSettings.Quota.Labels["3p-5h"] = "Anth5"
	customSettings.Quota.Labels["custom-window"] = "Custom Win"

	tests := []struct {
		name          string
		widgetName    string
		item          types.WidgetItem
		settings      types.Settings
		expectedTitle string
	}{
		{
			name:          "quota-5h uses settings.Quota.Labels override",
			widgetName:    "quota-5h",
			item:          types.WidgetItem{Type: "quota-5h"},
			settings:      customSettings,
			expectedTitle: "Gem5",
		},
		{
			name:          "quota-bar-3p-5h uses settings.Quota.Labels override",
			widgetName:    "quota-bar-3p-5h",
			item:          types.WidgetItem{Type: "quota-bar-3p-5h"},
			settings:      customSettings,
			expectedTitle: "Anth5",
		},
		{
			name:          "quota:custom-window uses settings.Quota.Labels for arbitrary key",
			widgetName:    "quota",
			item:          types.WidgetItem{Type: "quota", Key: "custom-window"},
			settings:      customSettings,
			expectedTitle: "Custom Win",
		},
		{
			name:          "item.Text takes precedence over settings.Quota.Labels",
			widgetName:    "quota-5h",
			item:          types.WidgetItem{Type: "quota-5h", Text: "InlineText"},
			settings:      customSettings,
			expectedTitle: "InlineText",
		},
		{
			name:          "zero-value Settings falls back to DefaultTitle / defaultQuotaLabel",
			widgetName:    "quota-5h",
			item:          types.WidgetItem{Type: "quota-5h"},
			settings:      types.Settings{},
			expectedTitle: "Gemini 5h",
		},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			w := GetWidget(tc.widgetName)
			if w == nil {
				t.Fatalf("Widget %q not found", tc.widgetName)
			}
			title, _, err := w.Render(tc.item, ctx, tc.settings)
			if err != nil {
				t.Fatalf("Unexpected error: %v", err)
			}
			if title != tc.expectedTitle {
				t.Errorf("Expected title %q, got %q", tc.expectedTitle, title)
			}
		})
	}
}

func TestOrderedQuotaKeys(t *testing.T) {
	f := 0.5
	info := types.QuotaInfo{RemainingFraction: &f}

	fullMap := map[string]types.QuotaInfo{
		"3p-weekly":               info,
		"gemini-5h":               info,
		"zeta-custom":             info,
		"3p-5h":                   info,
		"alpha-custom":            info,
		"gemini-argon-wtu-weekly": info,
		"gemini-weekly":           info,
	}

	tests := []struct {
		name     string
		quotaMap map[string]types.QuotaInfo
		quotaCfg types.QuotaConfig
		expected []string
	}{
		{
			name:     "nil quota map returns nil",
			quotaMap: nil,
			quotaCfg: types.DefaultSettings().Quota,
			expected: nil,
		},
		{
			name:     "empty quota map returns nil",
			quotaMap: map[string]types.QuotaInfo{},
			quotaCfg: types.DefaultSettings().Quota,
			expected: nil,
		},
		{
			name:     "default order places canonical keys first then unknown keys alphabetically",
			quotaMap: fullMap,
			quotaCfg: types.DefaultSettings().Quota,
			expected: []string{
				"gemini-5h",
				"gemini-weekly",
				"gemini-argon-wtu-weekly",
				"3p-5h",
				"3p-weekly",
				"alpha-custom",
				"zeta-custom",
			},
		},
		{
			name:     "zero-value QuotaConfig (nil Order) falls back to DefaultQuotaOrder",
			quotaMap: fullMap,
			quotaCfg: types.QuotaConfig{},
			expected: []string{
				"gemini-5h",
				"gemini-weekly",
				"gemini-argon-wtu-weekly",
				"3p-5h",
				"3p-weekly",
				"alpha-custom",
				"zeta-custom",
			},
		},
		{
			name:     "custom partial order overrides default order and sorts remaining keys alphabetically",
			quotaMap: fullMap,
			quotaCfg: types.QuotaConfig{
				Order: []string{"3p-5h", "gemini-5h"},
			},
			expected: []string{
				"3p-5h",
				"gemini-5h",
				"3p-weekly",
				"alpha-custom",
				"gemini-argon-wtu-weekly",
				"gemini-weekly",
				"zeta-custom",
			},
		},
		{
			name:     "explicit empty order slice sorts all keys alphabetically",
			quotaMap: fullMap,
			quotaCfg: types.QuotaConfig{
				Order: []string{},
			},
			expected: []string{
				"3p-5h",
				"3p-weekly",
				"alpha-custom",
				"gemini-5h",
				"gemini-argon-wtu-weekly",
				"gemini-weekly",
				"zeta-custom",
			},
		},
		{
			name:     "exclude removes specified keys from both ordered and remaining keys",
			quotaMap: fullMap,
			quotaCfg: types.QuotaConfig{
				Order:   types.DefaultQuotaOrder(),
				Exclude: []string{"3p-5h", "3p-weekly", "alpha-custom"},
			},
			expected: []string{
				"gemini-5h",
				"gemini-weekly",
				"gemini-argon-wtu-weekly",
				"zeta-custom",
			},
		},
		{
			name: "missing keys in quotaMap and duplicate keys in Order are handled cleanly",
			quotaMap: map[string]types.QuotaInfo{
				"gemini-5h":     info,
				"gemini-weekly": info,
			},
			quotaCfg: types.QuotaConfig{
				Order: []string{"gemini-weekly", "non-existent", "gemini-weekly", "gemini-5h"},
			},
			expected: []string{
				"gemini-weekly",
				"gemini-5h",
			},
		},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			got := OrderedQuotaKeys(tc.quotaMap, tc.quotaCfg)
			if !reflect.DeepEqual(got, tc.expected) {
				t.Errorf("Expected OrderedQuotaKeys %+v, got %+v", tc.expected, got)
			}
		})
	}
}

func TestQuotaAllWidgets_Registry(t *testing.T) {
	RegisterAll()

	qa := GetWidget("quota-all")
	if qa == nil {
		t.Fatalf("Expected 'quota-all' widget to be registered")
	}
	if qa.GetDisplayName() != "Quota: All" {
		t.Errorf("Expected display name 'Quota: All', got %q", qa.GetDisplayName())
	}
	if qa.GetDefaultColor() != "brightWhite" {
		t.Errorf("Expected default color 'brightWhite', got %q", qa.GetDefaultColor())
	}

	qba := GetWidget("quota-bar-all")
	if qba == nil {
		t.Fatalf("Expected 'quota-bar-all' widget to be registered")
	}
	if qba.GetDisplayName() != "Quota Bar: All" {
		t.Errorf("Expected display name 'Quota Bar: All', got %q", qba.GetDisplayName())
	}
	if qba.GetDefaultColor() != "brightGreen" {
		t.Errorf("Expected default color 'brightGreen', got %q", qba.GetDefaultColor())
	}
}
