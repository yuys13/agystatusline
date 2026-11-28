package tui

import (
	"slices"
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
