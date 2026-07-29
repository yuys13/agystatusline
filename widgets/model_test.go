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
	ctx := types.RenderContext{
		Data: types.StatusJSON{
			Model: types.ModelInfo{
				ID:          "claude-3-5-sonnet",
				DisplayName: "Claude 3.5 Sonnet",
			},
		},
	}

	item := types.WidgetItem{
		Type: "model",
	}

	// Normal render
	title, output, err := w.Render(item, ctx, settings)
	if err != nil {
		t.Fatalf("Render error: %v", err)
	}
	if title != "" || output != "Claude 3.5 Sonnet" {
		t.Errorf("Expected title '' and body 'Claude 3.5 Sonnet', got title '%s' and body '%s'", title, output)
	}

	// RawValue render
	rawVal := true
	itemRaw := types.WidgetItem{
		Type:     "model",
		RawValue: &rawVal,
	}
	titleRaw, outputRaw, err := w.Render(itemRaw, ctx, settings)
	if err != nil {
		t.Fatalf("Render error: %v", err)
	}
	if titleRaw != "" || outputRaw != "Claude 3.5 Sonnet" {
		t.Errorf("Expected title '' and body 'Claude 3.5 Sonnet', got title '%s' and body '%s'", titleRaw, outputRaw)
	}

	// Test that parenthesized suffixes are kept as-is.
	ctxWithNew := types.RenderContext{
		Data: types.StatusJSON{
			Model: types.ModelInfo{
				ID:          "claude-3-5-sonnet-new",
				DisplayName: "Claude 3.5 Sonnet (New)",
			},
		},
	}
	titleNew, outputNew, err := w.Render(item, ctxWithNew, settings)
	if err != nil {
		t.Fatalf("Render error: %v", err)
	}
	if titleNew != "" || outputNew != "Claude 3.5 Sonnet (New)" {
		t.Errorf("Expected title '' and body 'Claude 3.5 Sonnet (New)', got title '%s' and body '%s'", titleNew, outputNew)
	}

	ctxWithMedium := types.RenderContext{
		Data: types.StatusJSON{
			Model: types.ModelInfo{
				ID:          "gemini-3.5-flash-medium",
				DisplayName: "Gemini 3.5 Flash (Medium)",
			},
		},
	}
	titleMedium, outputMedium, err := w.Render(item, ctxWithMedium, settings)
	if err != nil {
		t.Fatalf("Render error: %v", err)
	}
	if titleMedium != "" || outputMedium != "Gemini 3.5 Flash (Medium)" {
		t.Errorf("Expected title '' and body 'Gemini 3.5 Flash (Medium)', got title '%s' and body '%s'", titleMedium, outputMedium)
	}
}

func TestModelWidget_EdgeCases(t *testing.T) {
	RegisterAll()
	w := GetWidget("model")
	settings := types.DefaultSettings()

	// 1. Empty Model
	ctxEmpty := types.RenderContext{Data: types.StatusJSON{}}
	item := types.WidgetItem{Type: "model"}
	title, body, err := w.Render(item, ctxEmpty, settings)
	if err != nil || title != "" || body != "" {
		t.Errorf("Expected empty output for empty model, got title=%q, body=%q, err=%v", title, body, err)
	}

	// 2. ID only (no DisplayName)
	ctxIDOnly := types.RenderContext{
		Data: types.StatusJSON{
			Model: types.ModelInfo{ID: "claude-3-5-haiku"},
		},
	}
	_, bodyID, _ := w.Render(item, ctxIDOnly, settings)
	if bodyID != "claude-3-5-haiku" {
		t.Errorf("Expected fallback to ID 'claude-3-5-haiku', got %q", bodyID)
	}

	// 3. PreserveColors
	trueVal := true
	itemPreserve := types.WidgetItem{Type: "model", PreserveColors: &trueVal}
	_, bodyPreserve, _ := w.Render(itemPreserve, ctxIDOnly, settings)
	if !strings.Contains(bodyPreserve, "\x1b[95m") {
		t.Errorf("Expected ANSI colors in preserve mode, got %q", bodyPreserve)
	}
}

func TestModelWidget_IDOnlyAndEmpty(t *testing.T) {
	RegisterAll()
	w := GetWidget("model")
	settings := types.DefaultSettings()
	item := types.WidgetItem{Type: "model"}

	// DisplayName empty, ID present
	ctxIDOnly := types.RenderContext{
		Data: types.StatusJSON{
			Model: types.ModelInfo{ID: "gemini-flash"},
		},
	}
	_, body, err := w.Render(item, ctxIDOnly, settings)
	if err != nil || body != "gemini-flash" {
		t.Errorf("Expected 'gemini-flash', got body=%q, err=%v", body, err)
	}

	// Both DisplayName and ID empty
	ctxEmpty := types.RenderContext{
		Data: types.StatusJSON{
			Model: types.ModelInfo{},
		},
	}
	titleEmpty, bodyEmpty, err := w.Render(item, ctxEmpty, settings)
	if err != nil || titleEmpty != "" || bodyEmpty != "" {
		t.Errorf("Expected empty title/body for empty model info, got title=%q, body=%q", titleEmpty, bodyEmpty)
	}
}
