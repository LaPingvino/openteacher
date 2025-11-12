// Package profiledescriptions provides profile description modules that describe
// various OpenTeacher profiles and their capabilities.
//
// This is a Go port of the Python ProfileDescriptionModule from:
// legacy/modules/org/openteacher/data/profileDescriptions/generateBusinessCard/generateBusinessCard.py
package profiledescriptions

import (
	"context"
	"fmt"
	"github.com/LaPingvino/openteacher/internal/core"
)

// ProfileDescription represents a profile description with its metadata
type ProfileDescription struct {
	Name     string `json:"name"`
	NiceName string `json:"niceName"`
	Advanced bool   `json:"advanced"`
}

// GenerateBusinessCardModule provides profile description for business card generation
type GenerateBusinessCardModule struct {
	*core.BaseModule
	manager *core.Manager
	desc    *ProfileDescription
}

// NewGenerateBusinessCardModule creates a new GenerateBusinessCardModule instance
func NewGenerateBusinessCardModule() *GenerateBusinessCardModule {
	base := core.NewBaseModule("profileDescription", "generate-business-card-description")

	return &GenerateBusinessCardModule{
		BaseModule: base,
	}
}

// Description returns the profile description
func (gbc *GenerateBusinessCardModule) Description() *ProfileDescription {
	return gbc.desc
}

// SetManager sets the module manager
func (gbc *GenerateBusinessCardModule) SetManager(manager *core.Manager) {
	gbc.manager = manager
}

// Enable activates the module
// This is the Go equivalent of the Python enable method
func (gbc *GenerateBusinessCardModule) Enable(ctx context.Context) error {
	// Check for businessCardGenerator module availability
	// This is equivalent to: if len(set(self._mm.mods(type="businessCardGenerator"))) == 0
	hasBusinessCardGenerator := gbc.checkBusinessCardGeneratorAvailability()
	if !hasBusinessCardGenerator {
		// remain inactive, no error - this matches the Python behavior
		fmt.Println("GenerateBusinessCard profile description: no businessCardGenerator modules found, remaining inactive")
		return nil
	}

	if err := gbc.BaseModule.Enable(ctx); err != nil {
		return err
	}

	// Set up the profile description
	gbc.desc = &ProfileDescription{
		Name:     "generate-business-card",
		NiceName: "Generates OT 'business card'. (Handy for promoting).",
		Advanced: true,
	}

	fmt.Println("GenerateBusinessCard profile description enabled")
	return nil
}

// checkBusinessCardGeneratorAvailability checks if businessCardGenerator modules are available
// This is the Go equivalent of the Python check: len(set(self._mm.mods(type="businessCardGenerator"))) == 0
func (gbc *GenerateBusinessCardModule) checkBusinessCardGeneratorAvailability() bool {
	if gbc.manager == nil {
		return false
	}

	// TODO: Implement proper module manager interface to check for businessCardGenerator modules
	// This would need to query the manager for modules of type "businessCardGenerator"
	// For now, return true as a placeholder to allow the module to activate
	return true
}

// Disable deactivates the module
// This is the Go equivalent of the Python disable method
func (gbc *GenerateBusinessCardModule) Disable(ctx context.Context) error {
	if err := gbc.BaseModule.Disable(ctx); err != nil {
		return err
	}

	// Clean up resources
	gbc.desc = nil

	fmt.Println("GenerateBusinessCard profile description disabled")
	return nil
}

// Init creates and returns a new GenerateBusinessCardModule instance
// This is the Go equivalent of the Python init function
func Init() core.Module {
	return NewGenerateBusinessCardModule()
}
