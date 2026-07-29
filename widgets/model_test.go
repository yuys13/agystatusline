package widgets

import (
	"strings"
	"testing"

	"github.com/yuys13/agystatusline/types"
)

func TestModelWidget(t *testing.T) {
	RegisterAll()

	w := GetWidget("model")
	if w == nil {
		t.Fatalf("Model widget not found in registry")
	}

	if w.GetDefaultColor() != "brightMagenta" {
		t.Errorf("Expected default color 'brightMagenta', got '%s'", w.GetDefaultColor())
	}

	settings := types.DefaultSettings()
	rawVal := true
	trueVal := true

	tests := []struct {
		name             string
		item             types.WidgetItem
		ctx              types.RenderContext
		expectedTitle    string
		expectedBody     string
		bodyContainsANSI bool
	}{
		{
			name: "Normal DisplayName render",
			item: types.WidgetItem{Type: "model"},
			ctx: types.RenderContext{
				Data: types.StatusJSON{
					Model: types.ModelInfo{
						ID:          "claude-3-5-sonnet",
						DisplayName: "Claude 3.5 Sonnet",
					},
				},
			},
			expectedTitle: "",
			expectedBody:  "Claude 3.5 Sonnet",
		},
		{
			name: "RawValue render",
			item: types.WidgetItem{Type: "model", RawValue: &rawVal},
			ctx: types.RenderContext{
				Data: types.StatusJSON{
					Model: types.ModelInfo{
						ID:          "claude-3-5-sonnet",
						DisplayName: "Claude 3.5 Sonnet",
					},
				},
			},
			expectedTitle: "",
			expectedBody:  "Claude 3.5 Sonnet",
		},
		{
			name: "DisplayName with parenthesized suffix (New)",
			item: types.WidgetItem{Type: "model"},
			ctx: types.RenderContext{
				Data: types.StatusJSON{
					Model: types.ModelInfo{
						ID:          "claude-3-5-sonnet-new",
						DisplayName: "Claude 3.5 Sonnet (New)",
					},
				},
			},
			expectedTitle: "",
			expectedBody:  "Claude 3.5 Sonnet (New)",
		},
		{
			name: "DisplayName with parenthesized suffix (Medium)",
			item: types.WidgetItem{Type: "model"},
			ctx: types.RenderContext{
				Data: types.StatusJSON{
					Model: types.ModelInfo{
						ID:          "gemini-3.5-flash-medium",
						DisplayName: "Gemini 3.5 Flash (Medium)",
					},
				},
			},
			expectedTitle: "",
			expectedBody:  "Gemini 3.5 Flash (Medium)",
		},
		{
			name:          "Empty Model info",
			item:          types.WidgetItem{Type: "model"},
			ctx:           types.RenderContext{Data: types.StatusJSON{}},
			expectedTitle: "",
			expectedBody:  "",
		},
		{
			name:          "Both DisplayName and ID empty",
			item:          types.WidgetItem{Type: "model"},
			ctx:           types.RenderContext{Data: types.StatusJSON{Model: types.ModelInfo{}}},
			expectedTitle: "",
			expectedBody:  "",
		},
		{
			name: "ID-only fallback when DisplayName is empty (Haiku)",
			item: types.WidgetItem{Type: "model"},
			ctx: types.RenderContext{
				Data: types.StatusJSON{
					Model: types.ModelInfo{ID: "claude-3-5-haiku"},
				},
			},
			expectedTitle: "",
			expectedBody:  "claude-3-5-haiku",
		},
		{
			name: "ID-only fallback when DisplayName is empty (Flash)",
			item: types.WidgetItem{Type: "model"},
			ctx: types.RenderContext{
				Data: types.StatusJSON{
					Model: types.ModelInfo{ID: "gemini-flash"},
				},
			},
			expectedTitle: "",
			expectedBody:  "gemini-flash",
		},
		{
			name: "PreserveColors mode",
			item: types.WidgetItem{Type: "model", PreserveColors: &trueVal},
			ctx: types.RenderContext{
				Data: types.StatusJSON{
					Model: types.ModelInfo{ID: "claude-3-5-haiku"},
				},
			},
			bodyContainsANSI: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			title, body, err := w.Render(tt.item, tt.ctx, settings)
			if err != nil {
				t.Fatalf("Render error: %v", err)
			}
			if tt.bodyContainsANSI {
				if !strings.Contains(body, "\x1b[95m") {
					t.Errorf("Expected ANSI colors in body, got %q", body)
				}
			} else {
				if title != tt.expectedTitle || body != tt.expectedBody {
					t.Errorf("Expected title %q and body %q, got title %q and body %q", tt.expectedTitle, tt.expectedBody, title, body)
				}
			}
		})
	}
}
