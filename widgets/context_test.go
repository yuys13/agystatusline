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

	settings := types.DefaultSettings()

	pct50 := 50.0
	pct65 := 65.0
	pct95 := 95.0

	tests := []struct {
		name             string
		ctx              types.RenderContext
		expectedTitle    string
		bodyContainsText string
		expectedBody     string
		expectedColor    string
	}{
		{
			name:          "Nil context window",
			ctx:           types.RenderContext{Data: types.StatusJSON{}},
			expectedTitle: "ctx",
			expectedBody:  "",
			expectedColor: "brightWhite",
		},
		{
			name:             "50% usage",
			ctx:              types.RenderContext{Data: types.StatusJSON{ContextWindow: &types.ContextWindowInfo{UsedPercentage: &pct50}}},
			expectedTitle:    "ctx",
			bodyContainsText: "50.0%",
			expectedColor:    "brightWhite",
		},
		{
			name:             "65% usage (yellow warning)",
			ctx:              types.RenderContext{Data: types.StatusJSON{ContextWindow: &types.ContextWindowInfo{UsedPercentage: &pct65}}},
			expectedTitle:    "ctx",
			bodyContainsText: "65.0%",
			expectedColor:    "brightYellow",
		},
		{
			name:             "95% usage (red critical)",
			ctx:              types.RenderContext{Data: types.StatusJSON{ContextWindow: &types.ContextWindowInfo{UsedPercentage: &pct95}}},
			expectedTitle:    "ctx",
			bodyContainsText: "95.0%",
			expectedColor:    "brightRed",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			item := types.WidgetItem{Type: "context-bar"}
			title, body, err := w.Render(item, tt.ctx, settings)
			if err != nil {
				t.Fatalf("Render error: %v", err)
			}
			if tt.expectedTitle != "" && title != tt.expectedTitle {
				t.Errorf("Expected title %q, got %q", tt.expectedTitle, title)
			}
			if tt.bodyContainsText != "" && !strings.Contains(body, tt.bodyContainsText) {
				t.Errorf("Expected body to contain %q, got %q", tt.bodyContainsText, body)
			}
			if tt.expectedBody != "" && body != tt.expectedBody {
				t.Errorf("Expected body %q, got %q", tt.expectedBody, body)
			}
			if tt.expectedColor != "" {
				if color := w.GetBodyColor(item, tt.ctx); color != tt.expectedColor {
					t.Errorf("Expected color %q, got %q", tt.expectedColor, color)
				}
			}
		})
	}
}
