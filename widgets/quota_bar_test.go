package widgets

import (
	"fmt"
	"strings"
	"testing"
	"unicode/utf8"

	"github.com/yuys13/agystatusline/types"
)

func TestQuotaBarWidget_NoData(t *testing.T) {
	w := initTestWidget(t, "quota-bar")
	settings := types.DefaultSettings()

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
}

func TestQuotaBarWidget_Render_80Percent(t *testing.T) {
	w := initTestWidget(t, "quota-bar")
	settings := types.DefaultSettings()

	pct := 0.8
	ctx := types.RenderContext{
		Data: types.StatusJSON{
			Quota: map[string]types.QuotaInfo{
				"gemini-5h": {
					RemainingFraction: &pct,
				},
			},
		},
	}
	item := types.WidgetItem{
		Type:     "quota-bar",
		Metadata: map[string]string{"key": "gemini-5h"},
	}

	color := w.GetBodyColor(item, ctx)
	if color != "brightGreen" {
		t.Errorf("For pct 80.0, expected body color brightGreen, got %s", color)
	}

	title, output, err := w.Render(item, ctx, settings)
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
}

func TestQuotaBarWidget_Render_30Percent(t *testing.T) {
	w := initTestWidget(t, "quota-bar")
	settings := types.DefaultSettings()

	pct := 0.3
	ctx := types.RenderContext{
		Data: types.StatusJSON{
			Quota: map[string]types.QuotaInfo{
				"gemini-5h": {
					RemainingFraction: &pct,
				},
			},
		},
	}
	item := types.WidgetItem{
		Type:     "quota-bar",
		Metadata: map[string]string{"key": "gemini-5h"},
	}

	color := w.GetBodyColor(item, ctx)
	if color != "brightYellow" {
		t.Errorf("For pct 30.0, expected body color brightYellow, got %s", color)
	}

	title, output, err := w.Render(item, ctx, settings)
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
}

func TestQuotaBarWidget_Render_5Percent(t *testing.T) {
	w := initTestWidget(t, "quota-bar")
	settings := types.DefaultSettings()

	pct := 0.05
	ctx := types.RenderContext{
		Data: types.StatusJSON{
			Quota: map[string]types.QuotaInfo{
				"gemini-5h": {
					RemainingFraction: &pct,
				},
			},
		},
	}
	item := types.WidgetItem{
		Type:     "quota-bar",
		Metadata: map[string]string{"key": "gemini-5h"},
	}

	color := w.GetBodyColor(item, ctx)
	if color != "brightRed" {
		t.Errorf("For pct 5.0, expected body color brightRed, got %s", color)
	}

	title, output, err := w.Render(item, ctx, settings)
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
}

func TestQuotaBarWidget_TitleMapping_GeminiWeekly(t *testing.T) {
	w := initTestWidget(t, "quota-bar")
	settings := types.DefaultSettings()
	pct := 0.8

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
}

func TestQuotaBarWidget_TitleMapping_3PWeekly(t *testing.T) {
	w := initTestWidget(t, "quota-bar")
	settings := types.DefaultSettings()
	pct := 0.8

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
}

func TestQuotaBarWidget_RawValue(t *testing.T) {
	w := initTestWidget(t, "quota-bar")
	settings := types.DefaultSettings()
	pct := 0.8
	ctx := types.RenderContext{
		Data: types.StatusJSON{
			Quota: map[string]types.QuotaInfo{
				"gemini-5h": {
					RemainingFraction: &pct,
				},
			},
		},
	}

	trueVal := true
	itemRaw := types.WidgetItem{
		Type:     "quota-bar",
		Metadata: map[string]string{"key": "gemini-5h"},
		RawValue: &trueVal,
	}
	titleR, outputR, err := w.Render(itemRaw, ctx, settings)
	if err != nil {
		t.Fatalf("Render error: %v", err)
	}
	expectedPctStr := fmt.Sprintf("%.1f%%", pct*100)
	if titleR != "" || !strings.Contains(outputR, expectedPctStr) {
		t.Errorf("Expected empty title and body containing %q, got title %q and body %q", expectedPctStr, titleR, outputR)
	}
}

func TestQuotaBarWidget_GetBodyColor_Boundary50Percent_Green(t *testing.T) {
	w := initTestWidget(t, "quota-bar")
	fraction := 0.50
	item := types.WidgetItem{
		Type:     "quota-bar",
		Metadata: map[string]string{"key": "gemini-5h"},
	}
	ctx := types.RenderContext{
		Data: types.StatusJSON{
			Quota: map[string]types.QuotaInfo{
				"gemini-5h": {
					RemainingFraction: &fraction,
				},
			},
		},
	}
	color := w.GetBodyColor(item, ctx)
	if color != "brightGreen" {
		t.Errorf("For boundary fraction 0.50, expected color brightGreen, got %s", color)
	}
}

func TestQuotaBarWidget_GetBodyColor_Boundary49Percent_Yellow(t *testing.T) {
	w := initTestWidget(t, "quota-bar")
	fraction := 0.49
	item := types.WidgetItem{
		Type:     "quota-bar",
		Metadata: map[string]string{"key": "gemini-5h"},
	}
	ctx := types.RenderContext{
		Data: types.StatusJSON{
			Quota: map[string]types.QuotaInfo{
				"gemini-5h": {
					RemainingFraction: &fraction,
				},
			},
		},
	}
	color := w.GetBodyColor(item, ctx)
	if color != "brightYellow" {
		t.Errorf("For boundary fraction 0.49, expected color brightYellow, got %s", color)
	}
}

func TestQuotaBarWidget_GetBodyColor_Boundary10Percent_Yellow(t *testing.T) {
	w := initTestWidget(t, "quota-bar")
	fraction := 0.10
	item := types.WidgetItem{
		Type:     "quota-bar",
		Metadata: map[string]string{"key": "gemini-5h"},
	}
	ctx := types.RenderContext{
		Data: types.StatusJSON{
			Quota: map[string]types.QuotaInfo{
				"gemini-5h": {
					RemainingFraction: &fraction,
				},
			},
		},
	}
	color := w.GetBodyColor(item, ctx)
	if color != "brightYellow" {
		t.Errorf("For boundary fraction 0.10, expected color brightYellow, got %s", color)
	}
}

func TestQuotaBarWidget_GetBodyColor_Boundary09Percent_Red(t *testing.T) {
	w := initTestWidget(t, "quota-bar")
	fraction := 0.09
	item := types.WidgetItem{
		Type:     "quota-bar",
		Metadata: map[string]string{"key": "gemini-5h"},
	}
	ctx := types.RenderContext{
		Data: types.StatusJSON{
			Quota: map[string]types.QuotaInfo{
				"gemini-5h": {
					RemainingFraction: &fraction,
				},
			},
		},
	}
	color := w.GetBodyColor(item, ctx)
	if color != "brightRed" {
		t.Errorf("For boundary fraction 0.09, expected color brightRed, got %s", color)
	}
}

func TestQuotaBarWidget_FractionalBlock_Remainder75(t *testing.T) {
	w := initTestWidget(t, "quota-bar")
	settings := types.DefaultSettings()
	pct := 0.28
	expectedBar := "██▓·······"
	item := types.WidgetItem{
		Type:     "quota-bar",
		Metadata: map[string]string{"key": "gemini-5h"},
	}
	ctx := types.RenderContext{
		Data: types.StatusJSON{
			Quota: map[string]types.QuotaInfo{
				"gemini-5h": {RemainingFraction: &pct},
			},
		},
	}
	_, output, err := w.Render(item, ctx, settings)
	if err != nil {
		t.Fatalf("Render error for remainder >= 75 (28%%): %v", err)
	}
	parts := strings.Split(output, " ")
	if parts[0] != expectedBar {
		t.Errorf("For remainder >= 75 (28%%), expected bar %q, got %q", expectedBar, parts[0])
	}
}

func TestQuotaBarWidget_FractionalBlock_Remainder50(t *testing.T) {
	w := initTestWidget(t, "quota-bar")
	settings := types.DefaultSettings()
	pct := 0.27
	expectedBar := "██▒·······"
	item := types.WidgetItem{
		Type:     "quota-bar",
		Metadata: map[string]string{"key": "gemini-5h"},
	}
	ctx := types.RenderContext{
		Data: types.StatusJSON{
			Quota: map[string]types.QuotaInfo{
				"gemini-5h": {RemainingFraction: &pct},
			},
		},
	}
	_, output, err := w.Render(item, ctx, settings)
	if err != nil {
		t.Fatalf("Render error for remainder >= 50 (27%%): %v", err)
	}
	parts := strings.Split(output, " ")
	if parts[0] != expectedBar {
		t.Errorf("For remainder >= 50 (27%%), expected bar %q, got %q", expectedBar, parts[0])
	}
}

func TestQuotaBarWidget_FractionalBlock_Remainder25(t *testing.T) {
	w := initTestWidget(t, "quota-bar")
	settings := types.DefaultSettings()
	pct := 0.23
	expectedBar := "██░·······"
	item := types.WidgetItem{
		Type:     "quota-bar",
		Metadata: map[string]string{"key": "gemini-5h"},
	}
	ctx := types.RenderContext{
		Data: types.StatusJSON{
			Quota: map[string]types.QuotaInfo{
				"gemini-5h": {RemainingFraction: &pct},
			},
		},
	}
	_, output, err := w.Render(item, ctx, settings)
	if err != nil {
		t.Fatalf("Render error for remainder >= 25 (23%%): %v", err)
	}
	parts := strings.Split(output, " ")
	if parts[0] != expectedBar {
		t.Errorf("For remainder >= 25 (23%%), expected bar %q, got %q", expectedBar, parts[0])
	}
}

func TestQuotaBarWidget_WithResetInSeconds(t *testing.T) {
	w := initTestWidget(t, "quota-bar")
	settings := types.DefaultSettings()

	resetSecs := 750.0 // 12m 30s
	remFraction := 0.5019
	ctxWithReset := types.RenderContext{
		Data: types.StatusJSON{
			Quota: map[string]types.QuotaInfo{
				"gemini-5h": {
					RemainingFraction: &remFraction,
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

func TestQuotaBarWidget_EdgeCase_NilQuotaMap(t *testing.T) {
	w := initTestWidget(t, "quota-bar")
	settings := types.DefaultSettings()

	ctxNil := types.RenderContext{Data: types.StatusJSON{Quota: nil}}
	item := types.WidgetItem{
		Type:     "quota-bar",
		Metadata: map[string]string{"key": "api"},
	}
	title, body, err := w.Render(item, ctxNil, settings)
	if err != nil || title != "" || body != "" {
		t.Errorf("Expected empty result for nil quota map, got title=%q, body=%q, err=%v", title, body, err)
	}
	if color := w.GetBodyColor(item, ctxNil); color != "brightGreen" {
		t.Errorf("Expected default 'brightGreen' body color for nil quota, got %q", color)
	}
}

func TestQuotaBarWidget_EdgeCase_GetBodyColor_Thresholds(t *testing.T) {
	w := initTestWidget(t, "quota-bar")
	item := types.WidgetItem{
		Type:     "quota-bar",
		Metadata: map[string]string{"key": "api"},
	}

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

func TestQuotaBarWidget_Interface(t *testing.T) {
	w := initTestWidget(t, "quota-bar")
	ctx := types.RenderContext{Data: types.StatusJSON{}}
	item := types.WidgetItem{Type: "quota-bar"}

	if nameStr := w.GetDisplayName(); nameStr == "" {
		t.Errorf("GetDisplayName() returned empty for quota-bar")
	}
	if defaultColor := w.GetDefaultColor(); defaultColor == "" {
		t.Errorf("GetDefaultColor() returned empty for quota-bar")
	}
	if bodyColor := w.GetBodyColor(item, ctx); bodyColor == "" {
		t.Errorf("GetBodyColor() returned empty for quota-bar")
	}
}
