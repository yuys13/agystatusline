package widgets

import (
	"testing"

	"github.com/yuys13/agystatusline/types"
)

func TestQuotaWidget_DefaultColor(t *testing.T) {
	w := initTestWidget(t, "quota")
	if w.GetDefaultColor() != "brightWhite" {
		t.Errorf("Expected default color 'brightWhite', got '%s'", w.GetDefaultColor())
	}
}

func TestQuotaWidget_LabeledPercentageAndReset(t *testing.T) {
	w := initTestWidget(t, "quota")
	settings := types.DefaultSettings()

	remaining1 := float64(0.5019274)
	resetInSecs1 := float64(8891)

	ctx := types.RenderContext{
		Data: types.StatusJSON{
			Quota: map[string]types.QuotaInfo{
				"gemini-5h": {
					RemainingFraction: &remaining1,
					ResetTime:         "2026-06-20T11:27:27Z",
					ResetInSeconds:    &resetInSecs1,
				},
			},
		},
	}
	item := types.WidgetItem{
		Type: "quota",
		Metadata: map[string]string{
			"key": "gemini-5h",
		},
	}
	title, output, err := w.Render(item, ctx, settings)
	if err != nil {
		t.Fatalf("Render error: %v", err)
	}
	if title != "gemini-5h" || output != "50.19% (2h 28m)" {
		t.Errorf("Expected title 'gemini-5h' and body '50.19%% (2h 28m)', got title '%s' and body '%s'", title, output)
	}
}

func TestQuotaWidget_RawValue_PercentageAndReset(t *testing.T) {
	w := initTestWidget(t, "quota")
	settings := types.DefaultSettings()

	remaining1 := float64(0.5019274)
	resetInSecs1 := float64(8891)

	ctx := types.RenderContext{
		Data: types.StatusJSON{
			Quota: map[string]types.QuotaInfo{
				"gemini-5h": {
					RemainingFraction: &remaining1,
					ResetTime:         "2026-06-20T11:27:27Z",
					ResetInSeconds:    &resetInSecs1,
				},
			},
		},
	}
	rawVal := true
	item := types.WidgetItem{
		Type: "quota",
		Metadata: map[string]string{
			"key": "gemini-5h",
		},
		RawValue: &rawVal,
	}
	title, output, err := w.Render(item, ctx, settings)
	if err != nil {
		t.Fatalf("Render error: %v", err)
	}
	if title != "" || output != "50.19% (2h 28m)" {
		t.Errorf("Expected title '' and body '50.19%% (2h 28m)', got title '%s' and body '%s'", title, output)
	}
}

func TestQuotaWidget_CustomTextLabelAndReset(t *testing.T) {
	w := initTestWidget(t, "quota")
	settings := types.DefaultSettings()

	remaining1 := float64(0.5019274)
	resetInSecs1 := float64(8891)

	ctx := types.RenderContext{
		Data: types.StatusJSON{
			Quota: map[string]types.QuotaInfo{
				"gemini-5h": {
					RemainingFraction: &remaining1,
					ResetTime:         "2026-06-20T11:27:27Z",
					ResetInSeconds:    &resetInSecs1,
				},
			},
		},
	}
	item := types.WidgetItem{
		Type: "quota",
		Metadata: map[string]string{
			"key": "gemini-5h",
		},
		CustomText: "Gemini Q",
	}
	title, output, err := w.Render(item, ctx, settings)
	if err != nil {
		t.Fatalf("Render error: %v", err)
	}
	if title != "Gemini Q" || output != "50.19% (2h 28m)" {
		t.Errorf("Expected title 'Gemini Q' and body '50.19%% (2h 28m)', got title '%s' and body '%s'", title, output)
	}
}

func TestQuotaWidget_DisplayQuotaOnly_Labeled(t *testing.T) {
	w := initTestWidget(t, "quota")
	settings := types.DefaultSettings()

	remaining1 := float64(0.5019274)
	resetInSecs1 := float64(8891)

	ctx := types.RenderContext{
		Data: types.StatusJSON{
			Quota: map[string]types.QuotaInfo{
				"gemini-5h": {
					RemainingFraction: &remaining1,
					ResetTime:         "2026-06-20T11:27:27Z",
					ResetInSeconds:    &resetInSecs1,
				},
			},
		},
	}
	item := types.WidgetItem{
		Type: "quota",
		Metadata: map[string]string{
			"key":     "gemini-5h",
			"display": "quota",
		},
	}
	title, output, err := w.Render(item, ctx, settings)
	if err != nil {
		t.Fatalf("Render error: %v", err)
	}
	if title != "gemini-5h" || output != "50.19%" {
		t.Errorf("Expected title 'gemini-5h' and body '50.19%%', got title '%s' and body '%s'", title, output)
	}
}

func TestQuotaWidget_DisplayQuotaOnly_RawValue(t *testing.T) {
	w := initTestWidget(t, "quota")
	settings := types.DefaultSettings()

	remaining1 := float64(0.5019274)
	resetInSecs1 := float64(8891)
	rawVal := true

	ctx := types.RenderContext{
		Data: types.StatusJSON{
			Quota: map[string]types.QuotaInfo{
				"gemini-5h": {
					RemainingFraction: &remaining1,
					ResetTime:         "2026-06-20T11:27:27Z",
					ResetInSeconds:    &resetInSecs1,
				},
			},
		},
	}
	item := types.WidgetItem{
		Type: "quota",
		Metadata: map[string]string{
			"key":     "gemini-5h",
			"display": "quota",
		},
		RawValue: &rawVal,
	}
	title, output, err := w.Render(item, ctx, settings)
	if err != nil {
		t.Fatalf("Render error: %v", err)
	}
	if title != "" || output != "50.19%" {
		t.Errorf("Expected title '' and body '50.19%%', got title '%s' and body '%s'", title, output)
	}
}

func TestQuotaWidget_DisplayReset_Labeled(t *testing.T) {
	w := initTestWidget(t, "quota")
	settings := types.DefaultSettings()

	remaining1 := float64(0.5019274)
	resetInSecs1 := float64(8891)

	ctx := types.RenderContext{
		Data: types.StatusJSON{
			Quota: map[string]types.QuotaInfo{
				"gemini-5h": {
					RemainingFraction: &remaining1,
					ResetTime:         "2026-06-20T11:27:27Z",
					ResetInSeconds:    &resetInSecs1,
				},
			},
		},
	}
	item := types.WidgetItem{
		Type: "quota",
		Metadata: map[string]string{
			"key":     "gemini-5h",
			"display": "reset",
		},
	}
	title, output, err := w.Render(item, ctx, settings)
	if err != nil {
		t.Fatalf("Render error: %v", err)
	}
	if title != "gemini-5h (reset)" || output != "2h 28m" {
		t.Errorf("Expected title 'gemini-5h (reset)' and body '2h 28m', got title '%s' and body '%s'", title, output)
	}
}

func TestQuotaWidget_DisplayReset_RawValue(t *testing.T) {
	w := initTestWidget(t, "quota")
	settings := types.DefaultSettings()

	remaining1 := float64(0.5019274)
	resetInSecs1 := float64(8891)
	rawVal := true

	ctx := types.RenderContext{
		Data: types.StatusJSON{
			Quota: map[string]types.QuotaInfo{
				"gemini-5h": {
					RemainingFraction: &remaining1,
					ResetTime:         "2026-06-20T11:27:27Z",
					ResetInSeconds:    &resetInSecs1,
				},
			},
		},
	}
	item := types.WidgetItem{
		Type: "quota",
		Metadata: map[string]string{
			"key":     "gemini-5h",
			"display": "reset",
		},
		RawValue: &rawVal,
	}
	title, output, err := w.Render(item, ctx, settings)
	if err != nil {
		t.Fatalf("Render error: %v", err)
	}
	if title != "" || output != "2h 28m" {
		t.Errorf("Expected title '' and body '2h 28m', got title '%s' and body '%s'", title, output)
	}
}

func TestQuotaWidget_DisplayReset_OtherDurations_45s(t *testing.T) {
	w := initTestWidget(t, "quota")
	settings := types.DefaultSettings()
	rawVal := true
	remaining2 := float64(1.0)
	secs45 := float64(45)

	ctxDur := types.RenderContext{
		Data: types.StatusJSON{
			Quota: map[string]types.QuotaInfo{
				"secs45": {
					RemainingFraction: &remaining2,
					ResetInSeconds:    &secs45,
				},
			},
		},
	}
	itemSecs45 := types.WidgetItem{
		Type: "quota",
		Metadata: map[string]string{
			"key":     "secs45",
			"display": "reset",
		},
		RawValue: &rawVal,
	}
	title45, output45, err := w.Render(itemSecs45, ctxDur, settings)
	if err != nil || title45 != "" || output45 != "45s" {
		t.Errorf("Expected title '' and body '45s', got title '%s' and body '%s', err=%v", title45, output45, err)
	}
}

func TestQuotaWidget_DisplayReset_OtherDurations_2m5s(t *testing.T) {
	w := initTestWidget(t, "quota")
	settings := types.DefaultSettings()
	rawVal := true
	remaining2 := float64(1.0)
	secs125 := float64(125)

	ctxDur := types.RenderContext{
		Data: types.StatusJSON{
			Quota: map[string]types.QuotaInfo{
				"secs125": {
					RemainingFraction: &remaining2,
					ResetInSeconds:    &secs125,
				},
			},
		},
	}
	itemSecs125 := types.WidgetItem{
		Type: "quota",
		Metadata: map[string]string{
			"key":     "secs125",
			"display": "reset",
		},
		RawValue: &rawVal,
	}
	title125, output125, err := w.Render(itemSecs125, ctxDur, settings)
	if err != nil || title125 != "" || output125 != "2m 5s" {
		t.Errorf("Expected title '' and body '2m 5s', got title '%s' and body '%s', err=%v", title125, output125, err)
	}
}

func TestQuotaWidget_DisplayReset_OtherDurations_6d13h(t *testing.T) {
	w := initTestWidget(t, "quota")
	settings := types.DefaultSettings()
	rawVal := true
	remaining2 := float64(1.0)
	secs567440 := float64(567440)

	ctxDur := types.RenderContext{
		Data: types.StatusJSON{
			Quota: map[string]types.QuotaInfo{
				"secs567440": {
					RemainingFraction: &remaining2,
					ResetInSeconds:    &secs567440,
				},
			},
		},
	}
	itemSecs567440 := types.WidgetItem{
		Type: "quota",
		Metadata: map[string]string{
			"key":     "secs567440",
			"display": "reset",
		},
		RawValue: &rawVal,
	}
	title567440, output567440, err := w.Render(itemSecs567440, ctxDur, settings)
	if err != nil || title567440 != "" || output567440 != "6d 13h" {
		t.Errorf("Expected title '' and body '6d 13h', got title '%s' and body '%s', err=%v", title567440, output567440, err)
	}
}

func TestQuotaWidget_EmptyKey(t *testing.T) {
	w := initTestWidget(t, "quota")
	settings := types.DefaultSettings()
	ctx := types.RenderContext{Data: types.StatusJSON{}}
	itemEmpty := types.WidgetItem{Type: "quota"}

	titleEmpty, outputEmpty, err := w.Render(itemEmpty, ctx, settings)
	if err != nil || titleEmpty != "" || outputEmpty != "" {
		t.Errorf("Expected empty title/body for empty key, got title '%s' and body '%s', err=%v", titleEmpty, outputEmpty, err)
	}
}

func TestQuotaWidget_InvalidKey(t *testing.T) {
	w := initTestWidget(t, "quota")
	settings := types.DefaultSettings()
	ctx := types.RenderContext{Data: types.StatusJSON{}}
	itemInvalid := types.WidgetItem{
		Type: "quota",
		Metadata: map[string]string{
			"key": "invalid-key",
		},
	}

	titleInvalid, outputInvalid, err := w.Render(itemInvalid, ctx, settings)
	if err != nil || titleInvalid != "" || outputInvalid != "" {
		t.Errorf("Expected empty title/body for invalid key, got title '%s' and body '%s', err=%v", titleInvalid, outputInvalid, err)
	}
}

func TestQuotaWidget_NilQuotaMap(t *testing.T) {
	w := initTestWidget(t, "quota")
	settings := types.DefaultSettings()
	ctxNilQuota := types.RenderContext{Data: types.StatusJSON{}}
	item1 := types.WidgetItem{
		Type: "quota",
		Metadata: map[string]string{
			"key": "gemini-5h",
		},
	}

	titleNil, outputNil, err := w.Render(item1, ctxNilQuota, settings)
	if err != nil || titleNil != "" || outputNil != "" {
		t.Errorf("Expected empty title/body for nil quota map, got title '%s' and body '%s', err=%v", titleNil, outputNil, err)
	}
}

func TestQuotaWidget_EdgeCase_NilQuotaMap(t *testing.T) {
	w := initTestWidget(t, "quota")
	settings := types.DefaultSettings()
	ctxNil := types.RenderContext{Data: types.StatusJSON{}}
	item := types.WidgetItem{Type: "quota", Metadata: map[string]string{"key": "gemini-5h"}}

	title, body, err := w.Render(item, ctxNil, settings)
	if err != nil || title != "" || body != "" {
		t.Errorf("Expected empty result for nil Quota map, got title=%q, body=%q, err=%v", title, body, err)
	}
}

func TestQuotaWidget_EdgeCase_MissingMetadataKey(t *testing.T) {
	w := initTestWidget(t, "quota")
	settings := types.DefaultSettings()
	rem := 0.8
	ctxData := types.RenderContext{
		Data: types.StatusJSON{
			Quota: map[string]types.QuotaInfo{
				"gemini-5h": {RemainingFraction: &rem},
			},
		},
	}
	itemNoKey := types.WidgetItem{Type: "quota"}

	_, bodyNoKey, err := w.Render(itemNoKey, ctxData, settings)
	if err != nil || bodyNoKey != "" {
		t.Errorf("Expected empty body when metadata key is missing, got %q, err=%v", bodyNoKey, err)
	}
}

func TestQuotaWidget_EdgeCase_UnknownKey(t *testing.T) {
	w := initTestWidget(t, "quota")
	settings := types.DefaultSettings()
	rem := 0.8
	ctxData := types.RenderContext{
		Data: types.StatusJSON{
			Quota: map[string]types.QuotaInfo{
				"gemini-5h": {RemainingFraction: &rem},
			},
		},
	}
	itemBadKey := types.WidgetItem{Type: "quota", Metadata: map[string]string{"key": "unknown"}}

	_, bodyBadKey, err := w.Render(itemBadKey, ctxData, settings)
	if err != nil || bodyBadKey != "" {
		t.Errorf("Expected empty body for unknown quota key, got %q, err=%v", bodyBadKey, err)
	}
}

func TestQuotaWidget_DisplayQuota(t *testing.T) {
	w := initTestWidget(t, "quota")
	settings := types.DefaultSettings()
	rem := 0.8
	ctxData := types.RenderContext{
		Data: types.StatusJSON{
			Quota: map[string]types.QuotaInfo{
				"gemini-5h": {RemainingFraction: &rem},
			},
		},
	}
	itemDisplayQuota := types.WidgetItem{Type: "quota", Metadata: map[string]string{"key": "gemini-5h", "display": "quota"}}

	titleQ, bodyQ, err := w.Render(itemDisplayQuota, ctxData, settings)
	if err != nil || titleQ != "gemini-5h" || bodyQ != "80.00%" {
		t.Errorf("Expected 'gemini-5h' and '80.00%%', got title=%q body=%q, err=%v", titleQ, bodyQ, err)
	}
}

func TestQuotaWidget_DisplayReset_NilResetInSeconds(t *testing.T) {
	w := initTestWidget(t, "quota")
	settings := types.DefaultSettings()
	rem := 0.8
	ctxData := types.RenderContext{
		Data: types.StatusJSON{
			Quota: map[string]types.QuotaInfo{
				"gemini-5h": {RemainingFraction: &rem},
			},
		},
	}
	itemDisplayReset := types.WidgetItem{Type: "quota", Metadata: map[string]string{"key": "gemini-5h", "display": "reset"}}

	_, bodyReset, err := w.Render(itemDisplayReset, ctxData, settings)
	if err != nil || bodyReset != "" {
		t.Errorf("Expected empty reset body when ResetInSeconds is nil, got %q, err=%v", bodyReset, err)
	}
}

func TestQuotaWidget_MoreEdgeCases_MissingMetadataKey(t *testing.T) {
	w := initTestWidget(t, "quota")
	settings := types.DefaultSettings()
	itemNoKey := types.WidgetItem{Type: "quota"}
	ctx := types.RenderContext{
		Data: types.StatusJSON{
			Quota: map[string]types.QuotaInfo{
				"api": {ResetTime: "2026-06-20T12:00:00Z"},
			},
		},
	}

	title, body, err := w.Render(itemNoKey, ctx, settings)
	if err != nil || title != "" || body != "" {
		t.Errorf("Expected empty result when metadata key is missing, got title=%q, body=%q, err=%v", title, body, err)
	}
}

func TestQuotaWidget_MoreEdgeCases_NonExistentKey(t *testing.T) {
	w := initTestWidget(t, "quota")
	settings := types.DefaultSettings()
	ctx := types.RenderContext{
		Data: types.StatusJSON{
			Quota: map[string]types.QuotaInfo{
				"api": {ResetTime: "2026-06-20T12:00:00Z"},
			},
		},
	}
	itemKey := types.WidgetItem{
		Type:     "quota",
		Metadata: map[string]string{"key": "nonexistent"},
	}

	title, body, err := w.Render(itemKey, ctx, settings)
	if err != nil || title != "" || body != "" {
		t.Errorf("Expected empty result when key is non-existent, got title=%q, body=%q, err=%v", title, body, err)
	}
}

func TestQuotaWidget_MoreEdgeCases_NilFields(t *testing.T) {
	w := initTestWidget(t, "quota")
	settings := types.DefaultSettings()
	itemApi := types.WidgetItem{
		Type:     "quota",
		Metadata: map[string]string{"key": "api", "display": "quota"},
	}
	ctxNilFields := types.RenderContext{
		Data: types.StatusJSON{
			Quota: map[string]types.QuotaInfo{
				"api": {},
			},
		},
	}

	title, body, err := w.Render(itemApi, ctxNilFields, settings)
	if err != nil || title != "" || body != "" {
		t.Errorf("Expected empty result when QuotaInfo has nil fields, got title=%q, body=%q, err=%v", title, body, err)
	}
}

func TestFormatResetInSeconds_Nil(t *testing.T) {
	res := formatResetInSeconds(nil)
	if res != "" {
		t.Errorf("For secs=nil expected %q, got %q", "", res)
	}
}

func TestFormatResetInSeconds_Negative(t *testing.T) {
	v := float64(-10)
	res := formatResetInSeconds(&v)
	if res != "0s" {
		t.Errorf("For secs=-10 expected '0s', got %q", res)
	}
}

func TestFormatResetInSeconds_Zero(t *testing.T) {
	v := float64(0)
	res := formatResetInSeconds(&v)
	if res != "0s" {
		t.Errorf("For secs=0 expected '0s', got %q", res)
	}
}

func TestFormatResetInSeconds_SecondsOnly(t *testing.T) {
	v := float64(45)
	res := formatResetInSeconds(&v)
	if res != "45s" {
		t.Errorf("For secs=45 expected '45s', got %q", res)
	}
}

func TestFormatResetInSeconds_ExactMinutes(t *testing.T) {
	v := float64(60)
	res := formatResetInSeconds(&v)
	if res != "1m" {
		t.Errorf("For secs=60 expected '1m', got %q", res)
	}
}

func TestFormatResetInSeconds_MinutesAndSeconds(t *testing.T) {
	v := float64(90)
	res := formatResetInSeconds(&v)
	if res != "1m 30s" {
		t.Errorf("For secs=90 expected '1m 30s', got %q", res)
	}
}

func TestFormatResetInSeconds_ExactHours(t *testing.T) {
	v := float64(3600)
	res := formatResetInSeconds(&v)
	if res != "1h" {
		t.Errorf("For secs=3600 expected '1h', got %q", res)
	}
}

func TestFormatResetInSeconds_HoursAndMinutes(t *testing.T) {
	v := float64(3660)
	res := formatResetInSeconds(&v)
	if res != "1h 1m" {
		t.Errorf("For secs=3660 expected '1h 1m', got %q", res)
	}
}

func TestFormatResetInSeconds_ExactDays(t *testing.T) {
	v := float64(86400)
	res := formatResetInSeconds(&v)
	if res != "1d" {
		t.Errorf("For secs=86400 expected '1d', got %q", res)
	}
}

func TestFormatResetInSeconds_DaysAndHours(t *testing.T) {
	v := float64(90000)
	res := formatResetInSeconds(&v)
	if res != "1d 1h" {
		t.Errorf("For secs=90000 expected '1d 1h', got %q", res)
	}
}

func TestQuotaWidget_Interface(t *testing.T) {
	w := initTestWidget(t, "quota")
	ctx := types.RenderContext{Data: types.StatusJSON{}}
	item := types.WidgetItem{Type: "quota"}

	if nameStr := w.GetDisplayName(); nameStr == "" {
		t.Errorf("GetDisplayName() returned empty for quota")
	}
	if defaultColor := w.GetDefaultColor(); defaultColor == "" {
		t.Errorf("GetDefaultColor() returned empty for quota")
	}
	if bodyColor := w.GetBodyColor(item, ctx); bodyColor == "" {
		t.Errorf("GetBodyColor() returned empty for quota")
	}
}
