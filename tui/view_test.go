package tui

import (
	"strings"
	"testing"

	"github.com/yuys13/agystatusline/renderer"
	"github.com/yuys13/agystatusline/types"
)

func TestBuildPreviewSettings(t *testing.T) {
	t.Run("select_theme updates preview theme", func(t *testing.T) {
		m := NewModel(types.DefaultSettings(), "/tmp/settings.toml")
		m.activeMenu = menuSelectTheme
		m.cursor = 2 // monokai

		got := m.buildPreviewSettings()
		if got.Powerline.Theme != themesList[2] {
			t.Errorf("Powerline.Theme = %q, want %q", got.Powerline.Theme, themesList[2])
		}
	})

	t.Run("select_separator updates preview separator", func(t *testing.T) {
		m := NewModel(types.DefaultSettings(), "/tmp/settings.toml")
		m.activeMenu = menuSelectSeparator
		m.cursor = 2 // Round

		got := m.buildPreviewSettings()
		if got.Powerline.Separator != m.separators[2].value {
			t.Errorf("Powerline.Separator = %q, want %q", got.Powerline.Separator, m.separators[2].value)
		}
	})

	t.Run("select_start_cap updates preview start cap", func(t *testing.T) {
		m := NewModel(types.DefaultSettings(), "/tmp/settings.toml")
		m.activeMenu = menuSelectStartCap
		m.cursor = 2

		got := m.buildPreviewSettings()
		if got.Powerline.StartCaps != m.startCaps[2].value {
			t.Errorf("Powerline.StartCaps = %q, want %q", got.Powerline.StartCaps, m.startCaps[2].value)
		}
	})

	t.Run("select_end_cap updates preview end cap", func(t *testing.T) {
		m := NewModel(types.DefaultSettings(), "/tmp/settings.toml")
		m.activeMenu = menuSelectEndCap
		m.cursor = 2

		got := m.buildPreviewSettings()
		if got.Powerline.EndCaps != m.endCaps[2].value {
			t.Errorf("Powerline.EndCaps = %q, want %q", got.Powerline.EndCaps, m.endCaps[2].value)
		}
	})

	t.Run("select_color_level updates preview color level", func(t *testing.T) {
		m := NewModel(types.DefaultSettings(), "/tmp/settings.toml")
		m.activeMenu = menuSelectColorLevel
		m.cursor = 2 // Truecolor

		got := m.buildPreviewSettings()
		if got.General.ColorLevel != colorLevelsList[2].value {
			t.Errorf("General.ColorLevel = %d, want %d", got.General.ColorLevel, colorLevelsList[2].value)
		}
	})

	t.Run("add_widget inserts preview widget without mutating model settings", func(t *testing.T) {
		s := types.DefaultSettings()
		s.Lines = [][]types.WidgetItem{{{Type: "model"}}}
		m := NewModel(s, "/tmp/settings.toml")
		m.activeMenu = menuAddWidget
		m.selectedLine = 0
		m.itemIndex = 0
		m.cursor = 2 // context-bar

		got := m.buildPreviewSettings()
		if len(got.Lines[0]) != 2 {
			t.Fatalf("len(got.Lines[0]) = %d, want 2", len(got.Lines[0]))
		}
		if got.Lines[0][1].Type != widgetTypes[2].item.Type {
			t.Errorf("got.Lines[0][1].Type = %q, want %q", got.Lines[0][1].Type, widgetTypes[2].item.Type)
		}
		if len(m.settings.Lines[0]) != 1 {
			t.Errorf("m.settings.Lines[0] was mutated: len = %d, want 1", len(m.settings.Lines[0]))
		}
	})

	t.Run("main menu preserves settings", func(t *testing.T) {
		s := types.DefaultSettings()
		s.Powerline.Theme = "nord"
		m := NewModel(s, "/tmp/settings.toml")
		m.activeMenu = menuMain

		got := m.buildPreviewSettings()
		if got.Powerline.Theme != "nord" {
			t.Errorf("Powerline.Theme = %q, want %q", got.Powerline.Theme, "nord")
		}
	})
}

func TestWriteChoiceListView(t *testing.T) {
	items := []choiceItem{
		{name: "None", value: ""},
		{name: "Arrow (\uE0B0)", value: "\uE0B0"},
		{name: "Round (\uE0B4)", value: "\uE0B4"},
	}

	tests := []struct {
		name        string
		title       string
		cursor      int
		activeValue string
		wantLines   []string
	}{
		{
			name:        "renders title cursor active marker and footer",
			title:       "Select Powerline Separator",
			cursor:      1,
			activeValue: "\uE0B4",
			wantLines: []string{
				"Select Powerline Separator",
				"  None",
				"> Arrow (\uE0B0)",
				"  Round (\uE0B4) (active)",
				"(Use arrows/jk to navigate, Enter to select, Esc: cancel)",
			},
		},
		{
			name:        "active item matches cursor",
			title:       "Select Powerline Start Cap",
			cursor:      0,
			activeValue: "",
			wantLines: []string{
				"Select Powerline Start Cap",
				"> None (active)",
				"  Arrow (\uE0B0)",
				"  Round (\uE0B4)",
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			m := NewModel(types.DefaultSettings(), "/tmp/settings.toml")
			m.cursor = tt.cursor

			var sb strings.Builder
			m.writeChoiceListView(&sb, tt.title, items, tt.activeValue)
			plain := renderer.StripAnsi(sb.String())

			for _, want := range tt.wantLines {
				if !strings.Contains(plain, want) {
					t.Errorf("output missing %q\ngot:\n%s", want, plain)
				}
			}
		})
	}
}
