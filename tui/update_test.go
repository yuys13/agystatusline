package tui

import (
	"slices"
	"testing"

	"github.com/yuys13/agystatusline/types"
)

func TestInsertWidget(t *testing.T) {
	tests := []struct {
		name            string
		widgets         []types.WidgetItem
		itemIndex       int
		newWidget       types.WidgetItem
		wantInsertIndex int
		wantTypes       []string
	}{
		{
			name:            "empty slice inserts at index 0",
			widgets:         []types.WidgetItem{},
			itemIndex:       0,
			newWidget:       types.WidgetItem{Type: "model"},
			wantInsertIndex: 0,
			wantTypes:       []string{"model"},
		},
		{
			name: "insert after index 0",
			widgets: []types.WidgetItem{
				{Type: "agent-state"},
				{Type: "context-bar"},
			},
			itemIndex:       0,
			newWidget:       types.WidgetItem{Type: "model"},
			wantInsertIndex: 1,
			wantTypes:       []string{"agent-state", "model", "context-bar"},
		},
		{
			name: "insert after middle index",
			widgets: []types.WidgetItem{
				{Type: "agent-state"},
				{Type: "model"},
				{Type: "sandbox"},
			},
			itemIndex:       1,
			newWidget:       types.WidgetItem{Type: "context-bar"},
			wantInsertIndex: 2,
			wantTypes:       []string{"agent-state", "model", "context-bar", "sandbox"},
		},
		{
			name: "insert after last index appends at end",
			widgets: []types.WidgetItem{
				{Type: "agent-state"},
				{Type: "model"},
			},
			itemIndex:       1,
			newWidget:       types.WidgetItem{Type: "sandbox"},
			wantInsertIndex: 2,
			wantTypes:       []string{"agent-state", "model", "sandbox"},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			gotWidgets, gotIdx := insertWidget(tt.widgets, tt.itemIndex, tt.newWidget)
			if gotIdx != tt.wantInsertIndex {
				t.Fatalf("insertIndex = %d, want %d", gotIdx, tt.wantInsertIndex)
			}
			if len(gotWidgets) != len(tt.wantTypes) {
				t.Fatalf("len(gotWidgets) = %d, want %d", len(gotWidgets), len(tt.wantTypes))
			}
			for i, wantType := range tt.wantTypes {
				if gotWidgets[i].Type != wantType {
					t.Errorf("gotWidgets[%d].Type = %q, want %q", i, gotWidgets[i].Type, wantType)
				}
			}
		})
	}
}

func TestInsertWidget_DoesNotMutateOriginalSlice(t *testing.T) {
	original := make([]types.WidgetItem, 2, 8)
	original[0] = types.WidgetItem{Type: "model"}
	original[1] = types.WidgetItem{Type: "sandbox"}
	snapshot := slices.Clone(original)

	gotWidgets, gotIdx := insertWidget(original, 0, types.WidgetItem{Type: "context-bar"})
	if gotIdx != 1 || len(gotWidgets) != 3 {
		t.Fatalf("unexpected insert result: len=%d, idx=%d", len(gotWidgets), gotIdx)
	}
	if !slices.Equal(original, snapshot) {
		t.Errorf("original slice was mutated: got %+v, want %+v", original, snapshot)
	}
}
