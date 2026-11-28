package tui

import (
	"slices"
	"strings"
	"testing"
)

func TestInitChoiceList(t *testing.T) {
	defaults := []choiceItem{
		{name: "None", value: ""},
		{name: "Arrow (\uE0B0)", value: "\uE0B0"},
		{name: "Flame (\uE0C2)", value: "\uE0C2 "},
	}

	tests := []struct {
		name           string
		current        string
		trimRightSpace bool
		wantLen        int
		wantIndex      int
		wantItem       choiceItem
	}{
		{
			name:           "exact preset match without trimming",
			current:        "\uE0B0",
			trimRightSpace: false,
			wantLen:        3,
			wantIndex:      1,
			wantItem:       choiceItem{name: "Arrow (\uE0B0)", value: "\uE0B0"},
		},
		{
			name:           "preset match with trailing space trimmed",
			current:        "\uE0C2",
			trimRightSpace: true,
			wantLen:        3,
			wantIndex:      2,
			wantItem:       choiceItem{name: "Flame (\uE0C2)", value: "\uE0C2 "},
		},
		{
			name:           "trailing space mismatch when trimming is disabled appends custom",
			current:        "\uE0C2",
			trimRightSpace: false,
			wantLen:        4,
			wantIndex:      3,
			wantItem:       choiceItem{name: "Custom (\uE0C2)", value: "\uE0C2"},
		},
		{
			name:           "unlisted custom value appends custom item",
			current:        "★",
			trimRightSpace: true,
			wantLen:        4,
			wantIndex:      3,
			wantItem:       choiceItem{name: "Custom (★)", value: "★"},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			items, idx := initChoiceList(defaults, tt.current, tt.trimRightSpace)
			if len(items) != tt.wantLen {
				t.Fatalf("len(items) = %d, want %d", len(items), tt.wantLen)
			}
			if idx != tt.wantIndex {
				t.Fatalf("idx = %d, want %d", idx, tt.wantIndex)
			}
			if items[idx] != tt.wantItem {
				t.Errorf("items[%d] = %+v, want %+v", idx, items[idx], tt.wantItem)
			}
		})
	}
}

func TestInitChoiceList_DoesNotMutateDefaults(t *testing.T) {
	backing := make([]choiceItem, 2, 8)
	backing[0] = choiceItem{name: "None", value: ""}
	backing[1] = choiceItem{name: "Arrow", value: "\uE0B0"}
	original := slices.Clone(backing)

	items, idx := initChoiceList(backing, "CUSTOM", false)
	if len(items) != 3 || idx != 2 {
		t.Fatalf("unexpected result: len=%d, idx=%d", len(items), idx)
	}
	if !slices.Equal(backing, original) {
		t.Errorf("defaults slice was mutated: got %+v, want %+v", backing, original)
	}
}

func TestTUI_NoASCIIInSeparators(t *testing.T) {
	for _, sep := range defaultSeparators {
		if strings.ContainsAny(sep.value, "/|") {
			t.Errorf("Separator %q contains forbidden ASCII characters: %q", sep.name, sep.value)
		}
	}
}

func TestTUI_SeparatorsList_Presets(t *testing.T) {
	expected := []struct {
		name  string
		value string
	}{
		{name: "None", value: ""},
		{name: "Arrow (\uE0B0)", value: "\uE0B0"},
		{name: "Round (\uE0B4)", value: "\uE0B4"},
		{name: "Flame (\uE0C0)", value: "\uE0C0"},
		{name: "Hexagon (\uE0C6)", value: "\uE0C6"},
		{name: "Slanted (\uE0C8)", value: "\uE0C8"},
	}

	if len(defaultSeparators) != len(expected) {
		t.Fatalf("Expected exactly %d preset separators, got %d", len(expected), len(defaultSeparators))
	}

	for i, want := range expected {
		t.Run(want.name, func(t *testing.T) {
			got := defaultSeparators[i]
			if got.name != want.name {
				t.Errorf("Expected preset name %q, got %q", want.name, got.name)
			}
			if got.value != want.value {
				t.Errorf("Expected preset value %q, got %q", want.value, got.value)
			}
		})
	}
}

func TestTUI_WidgetTypesOrdering(t *testing.T) {
	expectedTypes := []string{
		"agent-state",
		"model",
		"context-bar",
		"artifacts",
		"subagents",
		"tasks",
		"sandbox",
		"git-repo",
		"git-branch",
		"git-ahead-behind",
		"git-changes",
		"quota-all",
		"quota-5h",
		"quota-7d",
		"quota-argon-7d",
		"quota-3p-5h",
		"quota-3p-7d",
		"quota-bar-all",
		"quota-bar-5h",
		"quota-bar-7d",
		"quota-bar-argon-7d",
		"quota-bar-3p-5h",
		"quota-bar-3p-7d",
		"custom-text",
	}

	if len(widgetTypes) != len(expectedTypes) {
		t.Fatalf("Expected %d widget types, got %d", len(expectedTypes), len(widgetTypes))
	}

	for i, expected := range expectedTypes {
		if widgetTypes[i].item.Type != expected {
			t.Errorf("At index %d: expected widget type %q, got %q", i, expected, widgetTypes[i].item.Type)
		}
	}
}
