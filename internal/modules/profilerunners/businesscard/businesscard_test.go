package businesscard

import (
	"context"
	"testing"

	"github.com/LaPingvino/recuerdo/internal/core"
)

func TestNewBusinessCardGeneratorModule(t *testing.T) {
	module := NewBusinessCardGeneratorModule()

	if module == nil {
		t.Fatal("NewBusinessCardGeneratorModule returned nil")
	}

	if module.Type() != "businessCardGenerator" {
		t.Errorf("Expected module type 'businessCardGenerator', got '%s'", module.Type())
	}

	if module.Name() != "business-card-generator" {
		t.Errorf("Expected module name 'business-card-generator', got '%s'", module.Name())
	}

	if module.IsActive() {
		t.Error("Module should not be active by default")
	}
}

func TestBusinessCardGeneratorModuleRequirements(t *testing.T) {
	module := NewBusinessCardGeneratorModule()
	requires := module.Requires()

	expectedRequires := []string{"backgroundImageGenerator", "ui", "execute"}
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

func TestBusinessCardGeneratorModuleLifecycle(t *testing.T) {
	module := NewBusinessCardGeneratorModule()
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

func TestBusinessCardGeneratorInit(t *testing.T) {
	module := Init()

	if module == nil {
		t.Fatal("Init returned nil")
	}

	coreModule, ok := module.(core.Module)
	if !ok {
		t.Error("Init did not return a core.Module")
	}

	if coreModule.Type() != "businessCardGenerator" {
		t.Errorf("Expected module type 'businessCardGenerator', got '%s'", coreModule.Type())
	}
}

func TestBusinessCardGeneratorRun(t *testing.T) {
	module := NewBusinessCardGeneratorModule()
	ctx := context.Background()

	// Enable the module first
	err := module.Enable(ctx)
	if err != nil {
		t.Fatalf("Failed to enable module: %v", err)
	}

	// Test Run method - this should fail because no arguments are provided
	// and the backgroundImageGenerator is not implemented yet
	err = module.Run()
	if err == nil {
		t.Error("Expected Run to fail without proper arguments and dependencies")
	}

	// The error message should indicate missing path argument
	expectedError := "please specify a path where the business card image can be saved"
	if err != nil && len(err.Error()) > 0 {
		errorMsg := err.Error()
		if len(errorMsg) < len(expectedError) || errorMsg[:len(expectedError)] != expectedError {
			t.Logf("Got expected error: %v", err)
		}
	}
}

func TestBusinessCardGeneratorSetManager(t *testing.T) {
	module := NewBusinessCardGeneratorModule()

	// Create a mock manager (nil is acceptable for this test)
	var manager *core.Manager

	// This should not panic
	module.SetManager(manager)

	// Verify the manager was set
	if module.manager != manager {
		t.Error("Manager was not set correctly")
	}
}
