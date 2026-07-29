package widgets

import (
	"testing"
)

// initTestWidget initializes the widget registry and retrieves the named widget,
// failing the test immediately if the widget is not found.
func initTestWidget(t *testing.T, name string) Widget {
	t.Helper()
	RegisterAll()
	w := GetWidget(name)
	if w == nil {
		t.Fatalf("Widget %q not found in registry", name)
	}
	return w
}
