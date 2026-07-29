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
	secs45 := float64(45)
	secs125 := float64(125)
	secs567440 := float64(567440)
	val80 := float64(0.8)
	rawVal := true

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
			item:          types.WidgetItem{Type: "quota", Metadata: map[string]string{"key": "gemini-5h"}},
			ctx:           standardCtx,
			expectedTitle: "gemini-5h",
			expectedBody:  "50.19% (2h 28m)",
		},
		{
			name:          "Raw Percentage + Reset (rawValue = true)",
			item:          types.WidgetItem{Type: "quota", Metadata: map[string]string{"key": "gemini-5h"}, RawValue: &rawVal},
			ctx:           standardCtx,
			expectedTitle: "",
			expectedBody:  "50.19% (2h 28m)",
		},
		{
			name:          "Custom Text label + Reset",
			item:          types.WidgetItem{Type: "quota", Metadata: map[string]string{"key": "gemini-5h"}, CustomText: "Gemini Q"},
			ctx:           standardCtx,
			expectedTitle: "Gemini Q",
			expectedBody:  "50.19% (2h 28m)",
		},
		{
			name:          "display=quota (Percentage only, labeled)",
			item:          types.WidgetItem{Type: "quota", Metadata: map[string]string{"key": "gemini-5h", "display": "quota"}},
			ctx:           standardCtx,
			expectedTitle: "gemini-5h",
			expectedBody:  "50.19%",
		},
		{
			name:          "display=quota (Percentage only, raw)",
			item:          types.WidgetItem{Type: "quota", Metadata: map[string]string{"key": "gemini-5h", "display": "quota"}, RawValue: &rawVal},
			ctx:           standardCtx,
			expectedTitle: "",
			expectedBody:  "50.19%",
		},
		{
			name:          "display=reset (Reset time labeled)",
			item:          types.WidgetItem{Type: "quota", Metadata: map[string]string{"key": "gemini-5h", "display": "reset"}},
			ctx:           standardCtx,
			expectedTitle: "gemini-5h (reset)",
			expectedBody:  "2h 28m",
		},
		{
			name:          "display=reset (Reset time raw)",
			item:          types.WidgetItem{Type: "quota", Metadata: map[string]string{"key": "gemini-5h", "display": "reset"}, RawValue: &rawVal},
			ctx:           standardCtx,
			expectedTitle: "",
			expectedBody:  "2h 28m",
		},
		{
			name:          "display=reset (45s duration)",
			item:          types.WidgetItem{Type: "quota", Metadata: map[string]string{"key": "secs45", "display": "reset"}, RawValue: &rawVal},
			ctx:           standardCtx,
			expectedTitle: "",
			expectedBody:  "45s",
		},
		{
			name:          "display=reset (2m 5s duration)",
			item:          types.WidgetItem{Type: "quota", Metadata: map[string]string{"key": "secs125", "display": "reset"}, RawValue: &rawVal},
			ctx:           standardCtx,
			expectedTitle: "",
			expectedBody:  "2m 5s",
		},
		{
			name:          "display=reset (6d 13h duration)",
			item:          types.WidgetItem{Type: "quota", Metadata: map[string]string{"key": "secs567440", "display": "reset"}, RawValue: &rawVal},
			ctx:           standardCtx,
			expectedTitle: "",
			expectedBody:  "6d 13h",
		},
		{
			name:          "display=reset when ResetInSeconds is nil",
			item:          types.WidgetItem{Type: "quota", Metadata: map[string]string{"key": "no-reset-secs", "display": "reset"}},
			ctx:           standardCtx,
			expectedTitle: "",
			expectedBody:  "",
		},
		{
			name:          "Missing metadata key",
			item:          types.WidgetItem{Type: "quota"},
			ctx:           standardCtx,
			expectedTitle: "",
			expectedBody:  "",
		},
		{
			name:          "Unknown quota key",
			item:          types.WidgetItem{Type: "quota", Metadata: map[string]string{"key": "invalid-key"}},
			ctx:           standardCtx,
			expectedTitle: "",
			expectedBody:  "",
		},
		{
			name:          "Nil quota map",
			item:          types.WidgetItem{Type: "quota", Metadata: map[string]string{"key": "gemini-5h"}},
			ctx:           types.RenderContext{Data: types.StatusJSON{}},
			expectedTitle: "",
			expectedBody:  "",
		},
		{
			name:          "Nil fields in QuotaInfo",
			item:          types.WidgetItem{Type: "quota", Metadata: map[string]string{"key": "empty-info", "display": "quota"}},
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
		})
	}
}

func TestQuotaBarWidget(t *testing.T) {
	RegisterAll()
	w := GetWidget("quota-bar")
	if w == nil {
		t.Fatalf("Quota bar widget not found")
	}

	settings := types.DefaultSettings()

	t.Run("Missing or nil quota data", func(t *testing.T) {
		item := types.WidgetItem{Type: "quota-bar", Metadata: map[string]string{"key": "gemini-5h"}}
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
			{0.10, "brightYellow"},
			{0.09, "brightRed"},
			{0.05, "brightRed"},
		}
		item := types.WidgetItem{Type: "quota-bar", Metadata: map[string]string{"key": "gemini-5h"}}
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
					t.Errorf("For fraction %.2f expected color %s, got %s", tt.fraction, tt.expectedColor, color)
				}
			})
		}
	})

	t.Run("Standard rendering and bar width", func(t *testing.T) {
		pcts := []float64{0.8, 0.3, 0.05}
		expectedColors := []string{"brightGreen", "brightYellow", "brightRed"}
		item := types.WidgetItem{Type: "quota-bar", Metadata: map[string]string{"key": "gemini-5h"}}

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
					t.Errorf("Expected body color %s, got %s", expectedColors[i], color)
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
		rawVal := true
		tests := []struct {
			name          string
			item          types.WidgetItem
			ctx           types.RenderContext
			expectedTitle string
		}{
			{
				name:          "gemini-weekly maps to 7d",
				item:          types.WidgetItem{Type: "quota-bar", Metadata: map[string]string{"key": "gemini-weekly"}},
				ctx:           types.RenderContext{Data: types.StatusJSON{Quota: map[string]types.QuotaInfo{"gemini-weekly": {RemainingFraction: &pct}}}},
				expectedTitle: "7d",
			},
			{
				name:          "3p-weekly maps to 3p-7d",
				item:          types.WidgetItem{Type: "quota-bar", Metadata: map[string]string{"key": "3p-weekly"}},
				ctx:           types.RenderContext{Data: types.StatusJSON{Quota: map[string]types.QuotaInfo{"3p-weekly": {RemainingFraction: &pct}}}},
				expectedTitle: "3p-7d",
			},
			{
				name:          "RawValue true returns empty title",
				item:          types.WidgetItem{Type: "quota-bar", Metadata: map[string]string{"key": "gemini-5h"}, RawValue: &rawVal},
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

		item := types.WidgetItem{Type: "quota-bar", Metadata: map[string]string{"key": "gemini-5h"}}
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
					t.Fatalf("Render error for %s: %v", tt.desc, err)
				}
				parts := strings.Split(output, " ")
				if parts[0] != tt.expectedBar {
					t.Errorf("For %s expected bar %q, got %q", tt.desc, tt.expectedBar, parts[0])
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
