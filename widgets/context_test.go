package widgets

import (
	"strings"
	"testing"

	"github.com/yuys13/agystatusline/types"
)

func TestContextBarWidget(t *testing.T) {
	RegisterAll()
	w := GetWidget("context-bar")
	if w == nil {
		t.Fatalf("Context bar widget not found")
	}

	if w.GetDefaultColor() != "brightWhite" {
		t.Errorf("Expected default color %q, got %q", "brightWhite", w.GetDefaultColor())
	}
	if w.GetDisplayName() != "Context Bar" {
		t.Errorf("Expected display name %q, got %q", "Context Bar", w.GetDisplayName())
	}

	settings := types.DefaultSettings()

	pct50 := 50.0
	pct65 := 65.0
	pct95 := 95.0

	tests := []struct {
		name             string
		item             types.WidgetItem
		ctx              types.RenderContext
		expectedTitle    string
		bodyContainsText string
		expectedBody     string
		expectedColor    string
	}{
		{
			name:          "Nil context window",
			item:          types.WidgetItem{Type: "context-bar"},
			ctx:           types.RenderContext{Data: types.StatusJSON{}},
			expectedTitle: "ctx",
			expectedBody:  "",
			expectedColor: "brightWhite",
		},
		{
			name:          "Nil context window with Raw=true",
			item:          types.WidgetItem{Type: "context-bar", Raw: true},
			ctx:           types.RenderContext{Data: types.StatusJSON{}},
			expectedTitle: "",
			expectedBody:  "",
			expectedColor: "brightWhite",
		},
		{
			name:             "50% usage",
			item:             types.WidgetItem{Type: "context-bar"},
			ctx:              types.RenderContext{Data: types.StatusJSON{ContextWindow: &types.ContextWindowInfo{UsedPercentage: &pct50}}},
			expectedTitle:    "ctx",
			bodyContainsText: "50.0%",
			expectedColor:    "brightWhite",
		},
		{
			name:             "50% usage with Raw=true",
			item:             types.WidgetItem{Type: "context-bar", Raw: true},
			ctx:              types.RenderContext{Data: types.StatusJSON{ContextWindow: &types.ContextWindowInfo{UsedPercentage: &pct50}}},
			expectedTitle:    "",
			bodyContainsText: "50.0%",
			expectedColor:    "brightWhite",
		},
		{
			name:             "65% usage (yellow warning)",
			item:             types.WidgetItem{Type: "context-bar"},
			ctx:              types.RenderContext{Data: types.StatusJSON{ContextWindow: &types.ContextWindowInfo{UsedPercentage: &pct65}}},
			expectedTitle:    "ctx",
			bodyContainsText: "65.0%",
			expectedColor:    "brightYellow",
		},
		{
			name:             "95% usage (red critical)",
			item:             types.WidgetItem{Type: "context-bar"},
			ctx:              types.RenderContext{Data: types.StatusJSON{ContextWindow: &types.ContextWindowInfo{UsedPercentage: &pct95}}},
			expectedTitle:    "ctx",
			bodyContainsText: "95.0%",
			expectedColor:    "brightRed",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			title, body, err := w.Render(tt.item, tt.ctx, settings)
			if err != nil {
				t.Fatalf("Render error: %v", err)
			}
			if tt.expectedTitle != "" || title != "" {
				if title != tt.expectedTitle {
					t.Errorf("Expected title %q, got %q", tt.expectedTitle, title)
				}
			}
			if tt.bodyContainsText != "" && !strings.Contains(body, tt.bodyContainsText) {
				t.Errorf("Expected body to contain %q, got %q", tt.bodyContainsText, body)
			}
			if tt.expectedBody != "" && body != tt.expectedBody {
				t.Errorf("Expected body %q, got %q", tt.expectedBody, body)
			}
			if tt.expectedColor != "" {
				if color := w.GetBodyColor(tt.item, tt.ctx); color != tt.expectedColor {
					t.Errorf("Expected color %q, got %q", tt.expectedColor, color)
				}
			}
		})
	}
}

func TestContextBarWidget_Percentages(t *testing.T) {
	RegisterAll()
	settings := types.DefaultSettings()

	w := GetWidget("context-bar")
	if w == nil {
		t.Fatalf("context-bar widget not found")
	}

	pctNegative := -25.5
	pctZero := 0.0
	pct59 := 59.9
	pct60 := 60.0
	pct89 := 89.9
	pct90 := 90.0
	pct100 := 100.0
	pct150 := 150.0

	testCases := []struct {
		name          string
		pct           *float64
		expectedTitle string
		expectedColor string
	}{
		{"Negative percentage", &pctNegative, "ctx", "brightWhite"},
		{"Zero percentage", &pctZero, "ctx", "brightWhite"},
		{"59.9% percentage", &pct59, "ctx", "brightWhite"},
		{"60.0% percentage", &pct60, "ctx", "brightYellow"},
		{"89.9% percentage", &pct89, "ctx", "brightYellow"},
		{"90.0% percentage", &pct90, "ctx", "brightRed"},
		{"100% percentage", &pct100, "ctx", "brightRed"},
		{"150% percentage", &pct150, "ctx", "brightRed"},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			ctx := types.RenderContext{
				Data: types.StatusJSON{
					ContextWindow: &types.ContextWindowInfo{
						UsedPercentage: tc.pct,
					},
				},
			}
			title, body, err := w.Render(types.WidgetItem{Type: "context-bar"}, ctx, settings)
			if err != nil {
				t.Fatalf("Unexpected render error: %v", err)
			}
			if title != tc.expectedTitle {
				t.Errorf("Expected title %q, got %q", tc.expectedTitle, title)
			}
			if body == "" {
				t.Errorf("Expected non-empty body for pct %v", tc.pct)
			}
			color := w.GetBodyColor(types.WidgetItem{Type: "context-bar"}, ctx)
			if color != tc.expectedColor {
				t.Errorf("Expected body color %q, got %q", tc.expectedColor, color)
			}
		})
	}
}
