package backgroundimage

import (
	"context"
	"testing"

	"github.com/LaPingvino/recuerdo/internal/core"
)

func TestNewBackgroundImageGeneratorModule(t *testing.T) {
	module := NewBackgroundImageGeneratorModule()

	if module == nil {
		t.Fatal("NewBackgroundImageGeneratorModule returned nil")
	}

	if module.Type() != "backgroundImageGenerator" {
		t.Errorf("Expected module type 'backgroundImageGenerator', got '%s'", module.Type())
	}

	if module.Name() != "background-image-generator" {
		t.Errorf("Expected module name 'background-image-generator', got '%s'", module.Name())
	}

	if module.IsActive() {
		t.Error("Module should not be active by default")
	}
}

func TestBackgroundImageGeneratorModuleRequirements(t *testing.T) {
	module := NewBackgroundImageGeneratorModule()
	requires := module.Requires()

	expectedRequires := []string{"metadata", "ui"}
	if len(requires) != len(expectedRequires) {
		t.Errorf("Expected %d requirements, got %d", len(expectedRequires), len(requires))
	}

	requiresMap := make(map[string]bool)
	for _, req := range requires {
		requiresMap[req] = true
	}

	for _, expected := range expectedRequires {
		if !requiresMap[expected] {
			t.Errorf("Missing required dependency: %s", expected)
		}
	}
}

func TestBackgroundImageGeneratorModuleLifecycle(t *testing.T) {
	module := NewBackgroundImageGeneratorModule()
	ctx := context.Background()

	// Test enable
	err := module.Enable(ctx)
	if err != nil {
		t.Errorf("Enable failed: %v", err)
	}

	if !module.IsActive() {
		t.Error("Module should be active after enable")
	}

	// Test disable
	err = module.Disable(ctx)
	if err != nil {
		t.Errorf("Disable failed: %v", err)
	}

	if module.IsActive() {
		t.Error("Module should not be active after disable")
	}
}

func TestBackgroundImageGeneratorInit(t *testing.T) {
	module := Init()

	if module == nil {
		t.Fatal("Init returned nil")
	}

	coreModule, ok := module.(core.Module)
	if !ok {
		t.Error("Init did not return a core.Module")
	}

	if coreModule.Type() != "backgroundImageGenerator" {
		t.Errorf("Expected module type 'backgroundImageGenerator', got '%s'", coreModule.Type())
	}
}

func TestBackgroundImageGeneratorGenerate(t *testing.T) {
	module := NewBackgroundImageGeneratorModule()
	ctx := context.Background()

	// Test Generate when module is not active - should fail
	_, err := module.Generate()
	if err == nil {
		t.Error("Expected Generate to fail when module is not active")
	}

	// Enable the module
	err = module.Enable(ctx)
	if err != nil {
		t.Fatalf("Failed to enable module: %v", err)
	}

	// Test Generate when module is active - should succeed
	img, err := module.Generate()
	if err != nil {
		t.Errorf("Generate failed: %v", err)
	}

	if img == nil {
		t.Error("Generate returned nil image")
	}

	// Verify image dimensions
	if img != nil {
		bounds := img.Bounds()
		expectedWidth := 1000
		expectedHeight := 5000

		if bounds.Dx() != expectedWidth {
			t.Errorf("Expected image width %d, got %d", expectedWidth, bounds.Dx())
		}

		if bounds.Dy() != expectedHeight {
			t.Errorf("Expected image height %d, got %d", expectedHeight, bounds.Dy())
		}
	}
}

func TestBackgroundImageGeneratorColorFromHSV(t *testing.T) {
	module := NewBackgroundImageGeneratorModule()

	// Test known HSV to RGB conversions
	testCases := []struct {
		h, s, v   int
		expectedR uint8
		expectedG uint8
		expectedB uint8
		name      string
	}{
		{0, 255, 255, 255, 0, 0, "Pure Red"},
		{120, 255, 255, 0, 255, 0, "Pure Green"},
		{240, 255, 255, 0, 0, 255, "Pure Blue"},
		{0, 0, 255, 255, 255, 255, "White"},
		{0, 0, 0, 0, 0, 0, "Black"},
		{0, 255, 128, 128, 0, 0, "Half-bright Red"},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			color := module.colorFromHSV(tc.h, tc.s, tc.v)

			// Allow for small rounding errors
			tolerance := uint8(1)

			if absDiff(color.R, tc.expectedR) > tolerance {
				t.Errorf("Red component: expected %d, got %d", tc.expectedR, color.R)
			}

			if absDiff(color.G, tc.expectedG) > tolerance {
				t.Errorf("Green component: expected %d, got %d", tc.expectedG, color.G)
			}

			if absDiff(color.B, tc.expectedB) > tolerance {
				t.Errorf("Blue component: expected %d, got %d", tc.expectedB, color.B)
			}

			if color.A != 255 {
				t.Errorf("Alpha component: expected 255, got %d", color.A)
			}
		})
	}
}

func TestBackgroundImageGeneratorSetManager(t *testing.T) {
	module := NewBackgroundImageGeneratorModule()

	// Create a mock manager (nil is acceptable for this test)
	var manager *core.Manager

	// This should not panic
	module.SetManager(manager)

	// Verify the manager was set
	if module.manager != manager {
		t.Error("Manager was not set correctly")
	}
}

// Helper function to calculate absolute difference between two uint8 values
func absDiff(a, b uint8) uint8 {
	if a > b {
		return a - b
	}
	return b - a
}
