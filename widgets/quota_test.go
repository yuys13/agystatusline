package widgets

import (
	"github.com/yuys13/agystatusline/types"
	"testing"
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

	ctx := types.RenderContext{
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
			},
		},
	}
	settings := types.DefaultSettings()

	// Case 1: Labeled Percentage + Reset (default)
	item1 := types.WidgetItem{
		Type: "quota",
		Metadata: map[string]string{
			"key": "gemini-5h",
		},
	}
	title1, output1, err := w.Render(item1, ctx, settings)
	if err != nil {
		t.Fatalf("Render error: %v", err)
	}
	if title1 != "gemini-5h" || output1 != "50.19% (2h 28m)" {
		t.Errorf("Expected title 'gemini-5h' and body '50.19%% (2h 28m)', got title '%s' and body '%s'", title1, output1)
	}

	// Case 2: Raw Percentage + Reset (rawValue = true, default)
	rawVal := true
	item2 := types.WidgetItem{
		Type: "quota",
		Metadata: map[string]string{
			"key": "gemini-5h",
		},
		RawValue: &rawVal,
	}
	title2, output2, err := w.Render(item2, ctx, settings)
	if err != nil {
		t.Fatalf("Render error: %v", err)
	}
	if title2 != "" || output2 != "50.19% (2h 28m)" {
		t.Errorf("Expected title '' and body '50.19%% (2h 28m)', got title '%s' and body '%s'", title2, output2)
	}

	// Case 3: Custom Text label + Reset
	item3 := types.WidgetItem{
		Type: "quota",
		Metadata: map[string]string{
			"key": "gemini-5h",
		},
		CustomText: "Gemini Q",
	}
	title3, output3, err := w.Render(item3, ctx, settings)
	if err != nil {
		t.Fatalf("Render error: %v", err)
	}
	if title3 != "Gemini Q" || output3 != "50.19% (2h 28m)" {
		t.Errorf("Expected title 'Gemini Q' and body '50.19%% (2h 28m)', got title '%s' and body '%s'", title3, output3)
	}

	// Case 3b: display="quota" (Percentage only, labeled)
	itemQuotaOnly := types.WidgetItem{
		Type: "quota",
		Metadata: map[string]string{
			"key":     "gemini-5h",
			"display": "quota",
		},
	}
	titleQuotaOnly, outputQuotaOnly, err := w.Render(itemQuotaOnly, ctx, settings)
	if err != nil {
		t.Fatalf("Render error: %v", err)
	}
	if titleQuotaOnly != "gemini-5h" || outputQuotaOnly != "50.19%" {
		t.Errorf("Expected title 'gemini-5h' and body '50.19%%', got title '%s' and body '%s'", titleQuotaOnly, outputQuotaOnly)
	}

	// Case 3c: display="quota" (Percentage only, raw)
	itemQuotaOnlyRaw := types.WidgetItem{
		Type: "quota",
		Metadata: map[string]string{
			"key":     "gemini-5h",
			"display": "quota",
		},
		RawValue: &rawVal,
	}
	titleQuotaOnlyRaw, outputQuotaOnlyRaw, err := w.Render(itemQuotaOnlyRaw, ctx, settings)
	if err != nil {
		t.Fatalf("Render error: %v", err)
	}
	if titleQuotaOnlyRaw != "" || outputQuotaOnlyRaw != "50.19%" {
		t.Errorf("Expected title '' and body '50.19%%', got title '%s' and body '%s'", titleQuotaOnlyRaw, outputQuotaOnlyRaw)
	}

	// Case 4: Reset time labeled
	item4 := types.WidgetItem{
		Type: "quota",
		Metadata: map[string]string{
			"key":     "gemini-5h",
			"display": "reset",
		},
	}
	title4, output4, err := w.Render(item4, ctx, settings)
	if err != nil {
		t.Fatalf("Render error: %v", err)
	}
	if title4 != "gemini-5h (reset)" || output4 != "2h 28m" {
		t.Errorf("Expected title 'gemini-5h (reset)' and body '2h 28m', got title '%s' and body '%s'", title4, output4)
	}

	// Case 5: Reset time raw
	item5 := types.WidgetItem{
		Type: "quota",
		Metadata: map[string]string{
			"key":     "gemini-5h",
			"display": "reset",
		},
		RawValue: &rawVal,
	}
	title5, output5, err := w.Render(item5, ctx, settings)
	if err != nil {
		t.Fatalf("Render error: %v", err)
	}
	if title5 != "" || output5 != "2h 28m" {
		t.Errorf("Expected title '' and body '2h 28m', got title '%s' and body '%s'", title5, output5)
	}

	// Case 6: Reset time other durations
	secs45 := float64(45)
	secs125 := float64(125)
	secs567440 := float64(567440)
	ctxDur := types.RenderContext{
		Data: types.StatusJSON{
			Quota: map[string]types.QuotaInfo{
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
	title45, output45, _ := w.Render(itemSecs45, ctxDur, settings)
	if title45 != "" || output45 != "45s" {
		t.Errorf("Expected title '' and body '45s', got title '%s' and body '%s'", title45, output45)
	}

	itemSecs125 := types.WidgetItem{
		Type: "quota",
		Metadata: map[string]string{
			"key":     "secs125",
			"display": "reset",
		},
		RawValue: &rawVal,
	}
	title125, output125, _ := w.Render(itemSecs125, ctxDur, settings)
	if title125 != "" || output125 != "2m 5s" {
		t.Errorf("Expected title '' and body '2m 5s', got title '%s' and body '%s'", title125, output125)
	}

	itemSecs567440 := types.WidgetItem{
		Type: "quota",
		Metadata: map[string]string{
			"key":     "secs567440",
			"display": "reset",
		},
		RawValue: &rawVal,
	}
	title567440, output567440, _ := w.Render(itemSecs567440, ctxDur, settings)
	if title567440 != "" || output567440 != "6d 13h" {
		t.Errorf("Expected title '' and body '6d 13h', got title '%s' and body '%s'", title567440, output567440)
	}

	// Case 7: Key not found or empty
	itemEmpty := types.WidgetItem{
		Type: "quota",
	}
	titleEmpty, outputEmpty, _ := w.Render(itemEmpty, ctx, settings)
	if titleEmpty != "" || outputEmpty != "" {
		t.Errorf("Expected empty title/body for empty key, got title '%s' and body '%s'", titleEmpty, outputEmpty)
	}

	itemInvalid := types.WidgetItem{
		Type: "quota",
		Metadata: map[string]string{
			"key": "invalid-key",
		},
	}
	titleInvalid, outputInvalid, _ := w.Render(itemInvalid, ctx, settings)
	if titleInvalid != "" || outputInvalid != "" {
		t.Errorf("Expected empty title/body for invalid key, got title '%s' and body '%s'", titleInvalid, outputInvalid)
	}

	// Case 8: Quota map is nil
	ctxNilQuota := types.RenderContext{
		Data: types.StatusJSON{},
	}
	titleNil, outputNil, _ := w.Render(item1, ctxNilQuota, settings)
	if titleNil != "" || outputNil != "" {
		t.Errorf("Expected empty title/body for nil quota map, got title '%s' and body '%s'", titleNil, outputNil)
	}
}

func TestFormatResetInSeconds(t *testing.T) {
	tests := []struct {
		secs     *float64
		expected string
	}{
		{nil, ""},
		{func(f float64) *float64 { return &f }(-10), "0s"},
		{func(f float64) *float64 { return &f }(0), "0s"},
		{func(f float64) *float64 { return &f }(45), "45s"},
		{func(f float64) *float64 { return &f }(60), "1m"},
		{func(f float64) *float64 { return &f }(90), "1m 30s"},
		{func(f float64) *float64 { return &f }(3600), "1h"},
		{func(f float64) *float64 { return &f }(3660), "1h 1m"},
		{func(f float64) *float64 { return &f }(86400), "1d"},
		{func(f float64) *float64 { return &f }(90000), "1d 1h"},
	}

	for _, tc := range tests {
		res := formatResetInSeconds(tc.secs)
		if res != tc.expected {
			t.Errorf("For secs=%v expected %q, got %q", tc.secs, tc.expected, res)
		}
	}
}

func TestQuotaWidget_EdgeCases(t *testing.T) {
	RegisterAll()
	w := GetWidget("quota")
	settings := types.DefaultSettings()

	// Nil Quota
	ctxNil := types.RenderContext{Data: types.StatusJSON{}}
	item := types.WidgetItem{Type: "quota", Metadata: map[string]string{"key": "gemini-5h"}}
	title, body, err := w.Render(item, ctxNil, settings)
	if err != nil || title != "" || body != "" {
		t.Errorf("Expected empty result for nil Quota map")
	}

	// Missing Key in Metadata
	ctxData := types.RenderContext{
		Data: types.StatusJSON{
			Quota: map[string]types.QuotaInfo{
				"gemini-5h": {RemainingFraction: func(f float64) *float64 { return &f }(0.8)},
			},
		},
	}
	itemNoKey := types.WidgetItem{Type: "quota"}
	_, bodyNoKey, _ := w.Render(itemNoKey, ctxData, settings)
	if bodyNoKey != "" {
		t.Errorf("Expected empty body when metadata key is missing")
	}

	// Unknown Key
	itemBadKey := types.WidgetItem{Type: "quota", Metadata: map[string]string{"key": "unknown"}}
	_, bodyBadKey, _ := w.Render(itemBadKey, ctxData, settings)
	if bodyBadKey != "" {
		t.Errorf("Expected empty body for unknown quota key")
	}

	// Display mode "quota"
	itemDisplayQuota := types.WidgetItem{Type: "quota", Metadata: map[string]string{"key": "gemini-5h", "display": "quota"}}
	titleQ, bodyQ, _ := w.Render(itemDisplayQuota, ctxData, settings)
	if titleQ != "gemini-5h" || bodyQ != "80.00%" {
		t.Errorf("Expected 'gemini-5h' and '80.00%%', got title=%q body=%q", titleQ, bodyQ)
	}

	// Display mode "reset"
	itemDisplayReset := types.WidgetItem{Type: "quota", Metadata: map[string]string{"key": "gemini-5h", "display": "reset"}}
	_, bodyReset, _ := w.Render(itemDisplayReset, ctxData, settings)
	if bodyReset != "" {
		t.Errorf("Expected empty reset body when ResetInSeconds is nil")
	}
}

func TestQuotaWidget_MoreEdgeCases(t *testing.T) {
	RegisterAll()
	w := GetWidget("quota")
	settings := types.DefaultSettings()

	// Missing metadata key
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
		t.Errorf("Expected empty result when metadata key is missing, got title=%q, body=%q", title, body)
	}

	// Key not in Quota map
	itemKey := types.WidgetItem{
		Type:     "quota",
		Metadata: map[string]string{"key": "nonexistent"},
	}
	title, body, err = w.Render(itemKey, ctx, settings)
	if err != nil || title != "" || body != "" {
		t.Errorf("Expected empty result when key is non-existent, got title=%q, body=%q", title, body)
	}

	// QuotaInfo nil fields (RemainingFraction nil, ResetTime empty)
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
	title, body, err = w.Render(itemApi, ctxNilFields, settings)
	if err != nil || title != "" || body != "" {
		t.Errorf("Expected empty result when QuotaInfo has nil fields, got title=%q, body=%q", title, body)
	}
}
