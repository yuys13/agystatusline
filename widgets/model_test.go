package widgets

import (
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
		t.Errorf("Expected default color %q, got %q", "brightMagenta", w.GetDefaultColor())
	}

	settings := types.DefaultSettings()

	tests := []struct {
		name          string
		item          types.WidgetItem
		ctx           types.RenderContext
		expectedTitle string
		expectedBody  string
		expectedColor string
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
			expectedColor: "brightMagenta",
		},
		{
			name: "Raw render",
			item: types.WidgetItem{Type: "model", Raw: true},
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
			expectedColor: "brightMagenta",
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
			expectedColor: "brightMagenta",
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
			expectedColor: "brightMagenta",
		},
		{
			name:          "Empty Model info fallback to no-model",
			item:          types.WidgetItem{Type: "model"},
			ctx:           types.RenderContext{Data: types.StatusJSON{}},
			expectedTitle: "",
			expectedBody:  "no-model",
			expectedColor: "brightMagenta",
		},
		{
			name:          "Both DisplayName and ID empty fallback to no-model",
			item:          types.WidgetItem{Type: "model"},
			ctx:           types.RenderContext{Data: types.StatusJSON{Model: types.ModelInfo{}}},
			expectedTitle: "",
			expectedBody:  "no-model",
			expectedColor: "brightMagenta",
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
			expectedColor: "brightMagenta",
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
			expectedColor: "brightMagenta",
		},
		{
			name: "Whitespace only DisplayName fallback to no-model",
			item: types.WidgetItem{Type: "model"},
			ctx: types.RenderContext{
				Data: types.StatusJSON{
					Model: types.ModelInfo{DisplayName: "   "},
				},
			},
			expectedTitle: "",
			expectedBody:  "no-model",
			expectedColor: "brightMagenta",
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
			if color := w.GetBodyColor(tt.item, tt.ctx); color != tt.expectedColor {
				t.Errorf("Expected body color %q, got %q", tt.expectedColor, color)
			}
		})
	}
}
