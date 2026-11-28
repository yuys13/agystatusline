package tui

import (
	"os"
	"sync"
	"testing"

	"github.com/yuys13/agystatusline/types"
	"github.com/yuys13/agystatusline/widgets"
)

func TestMain(m *testing.M) {
	widgets.RegisterAll()
	os.Exit(m.Run())
}

func TestInitialModel(t *testing.T) {
	settings := types.DefaultSettings()
	m := NewModel(settings, "/tmp/settings.toml")

	if len(m.settings.Lines) == 0 {
		t.Errorf("Expected non-empty default lines, got %d", len(m.settings.Lines))
	}

	if m.activeMenu != "main" {
		t.Errorf("Expected initial menu %q, got %q", "main", m.activeMenu)
	}
}

func TestTUI_PowerlineSeparator_IndexInitialization(t *testing.T) {
	tests := []struct {
		name          string
		setupSettings func() types.Settings
		wantSepName   string
		wantSepValue  string
	}{
		{
			name: "Default Separator",
			setupSettings: func() types.Settings {
				return types.DefaultSettings()
			},
			wantSepName:  "Arrow (\uE0B0)",
			wantSepValue: "\uE0B0",
		},
		{
			name: "Preset Separator (exists in list)",
			setupSettings: func() types.Settings {
				s := types.DefaultSettings()
				s.Powerline.Separator = "\uE0B4"
				return s
			},
			wantSepName:  "Round (\uE0B4)",
			wantSepValue: "\uE0B4",
		},
		{
			name: "None Separator",
			setupSettings: func() types.Settings {
				s := types.DefaultSettings()
				s.Powerline.Separator = ""
				return s
			},
			wantSepName:  "None",
			wantSepValue: "",
		},
		{
			name: "None Separator with Powerline Enabled",
			setupSettings: func() types.Settings {
				s := types.DefaultSettings()
				s.Powerline.Enabled = true
				s.Powerline.Separator = ""
				return s
			},
			wantSepName:  "None",
			wantSepValue: "",
		},
		{
			name: "Custom Separator (NOT in list)",
			setupSettings: func() types.Settings {
				s := types.DefaultSettings()
				s.Powerline.Separator = "♦"
				return s
			},
			wantSepName:  "Custom (♦)",
			wantSepValue: "♦",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			settings := tt.setupSettings()
			m := NewModel(settings, "/tmp/settings.toml")

			if m.separatorIndex < 0 || m.separatorIndex >= len(m.separators) {
				t.Fatalf("Expected valid separatorIndex [0, %d), got %d", len(m.separators), m.separatorIndex)
			}
			gotSep := m.separators[m.separatorIndex]
			if gotSep.name != tt.wantSepName {
				t.Errorf("Expected separator name %q, got %q", tt.wantSepName, gotSep.name)
			}
			if gotSep.value != tt.wantSepValue {
				t.Errorf("Expected separator value %q, got %q", tt.wantSepValue, gotSep.value)
			}
		})
	}
}

func TestTUI_PowerlineSeparator_PresetIndexes(t *testing.T) {
	tests := []struct {
		name      string
		separator string
		wantIndex int
	}{
		{name: "None", separator: "", wantIndex: 0},
		{name: "Arrow", separator: "\uE0B0", wantIndex: 1},
		{name: "Round", separator: "\uE0B4", wantIndex: 2},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			s := types.DefaultSettings()
			s.Powerline.Separator = tt.separator
			m := NewModel(s, "/tmp/settings.toml")
			if m.separatorIndex != tt.wantIndex {
				t.Errorf("Expected separatorIndex %d, got %d", tt.wantIndex, m.separatorIndex)
			}
		})
	}
}

func TestTUI_CustomCaps_Initialization(t *testing.T) {
	tests := []struct {
		name         string
		startCap     string
		endCap       string
		wantStartVal string
		wantEndVal   string
	}{
		{
			name:         "custom start and end caps initialized in model slices",
			startCap:     "[CUSTOM_START]",
			endCap:       "[CUSTOM_END]",
			wantStartVal: "[CUSTOM_START]",
			wantEndVal:   "[CUSTOM_END]",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			settings := types.DefaultSettings()
			settings.Powerline.StartCaps = tt.startCap
			settings.Powerline.EndCaps = tt.endCap

			m := NewModel(settings, "/tmp/settings.toml")

			if m.startCapIndex < 0 || m.startCapIndex >= len(m.startCaps) {
				t.Fatalf("Expected valid startCapIndex [0, %d), got %d", len(m.startCaps), m.startCapIndex)
			}
			gotStart := m.startCaps[m.startCapIndex].value
			if gotStart != tt.wantStartVal {
				t.Errorf("Expected custom start cap %q, got %q", tt.wantStartVal, gotStart)
			}

			if m.endCapIndex < 0 || m.endCapIndex >= len(m.endCaps) {
				t.Fatalf("Expected valid endCapIndex [0, %d), got %d", len(m.endCaps), m.endCapIndex)
			}
			gotEnd := m.endCaps[m.endCapIndex].value
			if gotEnd != tt.wantEndVal {
				t.Errorf("Expected custom end cap %q, got %q", tt.wantEndVal, gotEnd)
			}
		})
	}
}

func TestNewModel_PresetsImmutability(t *testing.T) {
	tests := []struct {
		name       string
		customSep  string
		customCap  string
		wantSepLen int
		wantCapLen int
	}{
		{
			name:       "first custom config does not mutate presets",
			customSep:  "[CUSTOM_SEP_1]",
			customCap:  "[CUSTOM_CAP_1]",
			wantSepLen: 6,
			wantCapLen: 8,
		},
		{
			name:       "second distinct custom config does not accumulate in presets",
			customSep:  "[CUSTOM_SEP_2]",
			customCap:  "[CUSTOM_CAP_2]",
			wantSepLen: 6,
			wantCapLen: 8,
		},
		{
			name:       "third distinct custom config keeps presets pristine",
			customSep:  "[CUSTOM_SEP_3]",
			customCap:  "[CUSTOM_CAP_3]",
			wantSepLen: 6,
			wantCapLen: 8,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			s := types.DefaultSettings()
			s.Powerline.Separator = tt.customSep
			s.Powerline.StartCaps = tt.customCap
			s.Powerline.EndCaps = tt.customCap
			_ = NewModel(s, "/tmp/settings.toml")

			gotSepLen := len(defaultSeparators)
			if gotSepLen != tt.wantSepLen {
				t.Errorf("defaultSeparators length: got %d, want %d", gotSepLen, tt.wantSepLen)
			}

			gotStartLen := len(defaultStartCaps)
			if gotStartLen != tt.wantCapLen {
				t.Errorf("defaultStartCaps length: got %d, want %d", gotStartLen, tt.wantCapLen)
			}

			gotEndLen := len(defaultEndCaps)
			if gotEndLen != tt.wantCapLen {
				t.Errorf("defaultEndCaps length: got %d, want %d", gotEndLen, tt.wantCapLen)
			}
		})
	}
}

func TestNewModel_ModelInstanceIsolation_Separators(t *testing.T) {
	tests := []struct {
		name         string
		customSep    string
		wantSepCount int
		wantSepName  string
		wantSepValue string
	}{
		{
			name:         "model A with custom separator star",
			customSep:    "★",
			wantSepCount: 7, // 6 presets + 1 custom
			wantSepName:  "Custom (★)",
			wantSepValue: "★",
		},
		{
			name:         "model B with custom separator diamond does not accumulate previous custom",
			customSep:    "♦",
			wantSepCount: 7, // 6 presets + 1 custom (not 8)
			wantSepName:  "Custom (♦)",
			wantSepValue: "♦",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			s := types.DefaultSettings()
			s.Powerline.Separator = tt.customSep
			m := NewModel(s, "/tmp/settings.toml")

			gotCount := len(m.separators)
			if gotCount != tt.wantSepCount {
				t.Errorf("model separators length: got %d, want %d", gotCount, tt.wantSepCount)
			}

			gotName := m.separators[m.separatorIndex].name
			if gotName != tt.wantSepName {
				t.Errorf("model separator name: got %q, want %q", gotName, tt.wantSepName)
			}

			gotValue := m.separators[m.separatorIndex].value
			if gotValue != tt.wantSepValue {
				t.Errorf("model separator value: got %q, want %q", gotValue, tt.wantSepValue)
			}
		})
	}
}

func TestNewModel_ModelInstanceIsolation_StartCaps(t *testing.T) {
	tests := []struct {
		name           string
		customStartCap string
		wantStartCount int
		wantStartName  string
		wantStartValue string
	}{
		{
			name:           "model A with custom start cap 1",
			customStartCap: "[START_A]",
			wantStartCount: 9, // 8 presets + 1 custom
			wantStartName:  "Custom ([START_A])",
			wantStartValue: "[START_A]",
		},
		{
			name:           "model B with custom start cap 2 does not accumulate previous custom",
			customStartCap: "[START_B]",
			wantStartCount: 9, // 8 presets + 1 custom (not 10)
			wantStartName:  "Custom ([START_B])",
			wantStartValue: "[START_B]",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			s := types.DefaultSettings()
			s.Powerline.StartCaps = tt.customStartCap
			m := NewModel(s, "/tmp/settings.toml")

			gotCount := len(m.startCaps)
			if gotCount != tt.wantStartCount {
				t.Errorf("model startCaps length: got %d, want %d", gotCount, tt.wantStartCount)
			}

			gotName := m.startCaps[m.startCapIndex].name
			if gotName != tt.wantStartName {
				t.Errorf("model startCap name: got %q, want %q", gotName, tt.wantStartName)
			}

			gotValue := m.startCaps[m.startCapIndex].value
			if gotValue != tt.wantStartValue {
				t.Errorf("model startCap value: got %q, want %q", gotValue, tt.wantStartValue)
			}
		})
	}
}

func TestNewModel_ModelInstanceIsolation_EndCaps(t *testing.T) {
	tests := []struct {
		name         string
		customEndCap string
		wantEndCount int
		wantEndName  string
		wantEndValue string
	}{
		{
			name:         "model A with custom end cap 1",
			customEndCap: "[END_A]",
			wantEndCount: 9, // 8 presets + 1 custom
			wantEndName:  "Custom ([END_A])",
			wantEndValue: "[END_A]",
		},
		{
			name:         "model B with custom end cap 2 does not accumulate previous custom",
			customEndCap: "[END_B]",
			wantEndCount: 9, // 8 presets + 1 custom (not 10)
			wantEndName:  "Custom ([END_B])",
			wantEndValue: "[END_B]",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			s := types.DefaultSettings()
			s.Powerline.EndCaps = tt.customEndCap
			m := NewModel(s, "/tmp/settings.toml")

			gotCount := len(m.endCaps)
			if gotCount != tt.wantEndCount {
				t.Errorf("model endCaps length: got %d, want %d", gotCount, tt.wantEndCount)
			}

			gotName := m.endCaps[m.endCapIndex].name
			if gotName != tt.wantEndName {
				t.Errorf("model endCap name: got %q, want %q", gotName, tt.wantEndName)
			}

			gotValue := m.endCaps[m.endCapIndex].value
			if gotValue != tt.wantEndValue {
				t.Errorf("model endCap value: got %q, want %q", gotValue, tt.wantEndValue)
			}
		})
	}
}

func TestNewModel_ConcurrentSafety(t *testing.T) {
	tests := []struct {
		name          string
		wantSeparator string
		wantStartCap  string
		wantEndCap    string
	}{
		{name: "routine 0", wantSeparator: "SEP_0", wantStartCap: "START_0", wantEndCap: "END_0"},
		{name: "routine 1", wantSeparator: "SEP_1", wantStartCap: "START_1", wantEndCap: "END_1"},
		{name: "routine 2", wantSeparator: "SEP_2", wantStartCap: "START_2", wantEndCap: "END_2"},
		{name: "routine 3", wantSeparator: "SEP_3", wantStartCap: "START_3", wantEndCap: "END_3"},
		{name: "routine 4", wantSeparator: "SEP_4", wantStartCap: "START_4", wantEndCap: "END_4"},
		{name: "routine 5", wantSeparator: "SEP_5", wantStartCap: "START_5", wantEndCap: "END_5"},
		{name: "routine 6", wantSeparator: "SEP_6", wantStartCap: "START_6", wantEndCap: "END_6"},
		{name: "routine 7", wantSeparator: "SEP_7", wantStartCap: "START_7", wantEndCap: "END_7"},
	}

	results := make([]Model, len(tests))
	var wg sync.WaitGroup
	for i, tt := range tests {
		wg.Add(1)
		go func(idx int, sep, start, end string) {
			defer wg.Done()
			s := types.DefaultSettings()
			s.Powerline.Separator = sep
			s.Powerline.StartCaps = start
			s.Powerline.EndCaps = end
			results[idx] = NewModel(s, "/tmp/settings.toml")
		}(i, tt.wantSeparator, tt.wantStartCap, tt.wantEndCap)
	}
	wg.Wait()

	for i, tt := range tests {
		m := results[i]
		gotSep := m.separators[m.separatorIndex].value
		if gotSep != tt.wantSeparator {
			t.Errorf("[%s] separator value: got %q, want %q", tt.name, gotSep, tt.wantSeparator)
		}
		gotStart := m.startCaps[m.startCapIndex].value
		if gotStart != tt.wantStartCap {
			t.Errorf("[%s] startCap value: got %q, want %q", tt.name, gotStart, tt.wantStartCap)
		}
		gotEnd := m.endCaps[m.endCapIndex].value
		if gotEnd != tt.wantEndCap {
			t.Errorf("[%s] endCap value: got %q, want %q", tt.name, gotEnd, tt.wantEndCap)
		}
	}
}
