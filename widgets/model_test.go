package widgets

import (
	"strings"
	"testing"

	"github.com/yuys13/agystatusline/types"
)

func TestModelWidget_DefaultColor(t *testing.T) {
	w := initTestWidget(t, "model")
	if w.GetDefaultColor() != "brightMagenta" {
		t.Errorf("Expected default color 'brightMagenta', got '%s'", w.GetDefaultColor())
	}
}

func TestModelWidget_NormalRender(t *testing.T) {
	w := initTestWidget(t, "model")
	settings := types.DefaultSettings()
	ctx := types.RenderContext{
		Data: types.StatusJSON{
			Model: types.ModelInfo{
				ID:          "claude-3-5-sonnet",
				DisplayName: "Claude 3.5 Sonnet",
			},
		},
	}
	item := types.WidgetItem{Type: "model"}

	title, output, err := w.Render(item, ctx, settings)
	if err != nil {
		t.Fatalf("Render error: %v", err)
	}
	if title != "" || output != "Claude 3.5 Sonnet" {
		t.Errorf("Expected title '' and body 'Claude 3.5 Sonnet', got title '%s' and body '%s'", title, output)
	}
}

func TestModelWidget_RawValueRender(t *testing.T) {
	w := initTestWidget(t, "model")
	settings := types.DefaultSettings()
	ctx := types.RenderContext{
		Data: types.StatusJSON{
			Model: types.ModelInfo{
				ID:          "claude-3-5-sonnet",
				DisplayName: "Claude 3.5 Sonnet",
			},
		},
	}
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
}

func TestModelWidget_ParenthesizedSuffix_SonnetNew(t *testing.T) {
	w := initTestWidget(t, "model")
	settings := types.DefaultSettings()
	item := types.WidgetItem{Type: "model"}
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
}

func TestModelWidget_ParenthesizedSuffix_FlashMedium(t *testing.T) {
	w := initTestWidget(t, "model")
	settings := types.DefaultSettings()
	item := types.WidgetItem{Type: "model"}
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

func TestModelWidget_EdgeCase_EmptyModel(t *testing.T) {
	w := initTestWidget(t, "model")
	settings := types.DefaultSettings()
	ctxEmpty := types.RenderContext{Data: types.StatusJSON{}}
	item := types.WidgetItem{Type: "model"}

	title, body, err := w.Render(item, ctxEmpty, settings)
	if err != nil || title != "" || body != "" {
		t.Errorf("Expected empty output for empty model, got title=%q, body=%q, err=%v", title, body, err)
	}
}

func TestModelWidget_EdgeCase_IDOnly(t *testing.T) {
	w := initTestWidget(t, "model")
	settings := types.DefaultSettings()
	ctxIDOnly := types.RenderContext{
		Data: types.StatusJSON{
			Model: types.ModelInfo{ID: "claude-3-5-haiku"},
		},
	}
	item := types.WidgetItem{Type: "model"}

	_, bodyID, err := w.Render(item, ctxIDOnly, settings)
	if err != nil || bodyID != "claude-3-5-haiku" {
		t.Errorf("Expected fallback to ID 'claude-3-5-haiku', got %q, err=%v", bodyID, err)
	}
}

func TestModelWidget_EdgeCase_PreserveColors(t *testing.T) {
	w := initTestWidget(t, "model")
	settings := types.DefaultSettings()
	ctxIDOnly := types.RenderContext{
		Data: types.StatusJSON{
			Model: types.ModelInfo{ID: "claude-3-5-haiku"},
		},
	}
	trueVal := true
	itemPreserve := types.WidgetItem{Type: "model", PreserveColors: &trueVal}

	_, bodyPreserve, err := w.Render(itemPreserve, ctxIDOnly, settings)
	if err != nil || !strings.Contains(bodyPreserve, "\x1b[95m") {
		t.Errorf("Expected ANSI colors in preserve mode, got %q, err=%v", bodyPreserve, err)
	}
}

func TestModelWidget_DisplayNameEmpty_IDPresent(t *testing.T) {
	w := initTestWidget(t, "model")
	settings := types.DefaultSettings()
	item := types.WidgetItem{Type: "model"}
	ctxIDOnly := types.RenderContext{
		Data: types.StatusJSON{
			Model: types.ModelInfo{ID: "gemini-flash"},
		},
	}

	_, body, err := w.Render(item, ctxIDOnly, settings)
	if err != nil || body != "gemini-flash" {
		t.Errorf("Expected 'gemini-flash', got body=%q, err=%v", body, err)
	}
}

func TestModelWidget_DisplayNameAndIDEmpty(t *testing.T) {
	w := initTestWidget(t, "model")
	settings := types.DefaultSettings()
	item := types.WidgetItem{Type: "model"}
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

func TestModelWidget_Interface(t *testing.T) {
	w := initTestWidget(t, "model")
	ctx := types.RenderContext{Data: types.StatusJSON{}}
	item := types.WidgetItem{Type: "model"}

	if nameStr := w.GetDisplayName(); nameStr == "" {
		t.Errorf("GetDisplayName() returned empty for model")
	}
	if defaultColor := w.GetDefaultColor(); defaultColor == "" {
		t.Errorf("GetDefaultColor() returned empty for model")
	}
	if bodyColor := w.GetBodyColor(item, ctx); bodyColor == "" {
		t.Errorf("GetBodyColor() returned empty for model")
	}
}
