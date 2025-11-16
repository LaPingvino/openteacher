package modules

import (
	"context"
	"fmt"
	"sync"

	"github.com/LaPingvino/recuerdo/internal/core"
)

// Button represents a registered button in the UI
type Button struct {
	name     string
	text     string
	size     string
	priority int
	clicked  core.Event
	enabled  bool
	visible  bool
}

// NewButton creates a new button
func NewButton(name, text string, clickedEvent core.Event) *Button {
	return &Button{
		name:     name,
		text:     text,
		size:     "medium",
		priority: 100,
		clicked:  clickedEvent,
		enabled:  true,
		visible:  true,
	}
}

// ChangePriority changes the button priority
func (b *Button) ChangePriority(priority int) {
	b.priority = priority
}

// ChangeSize changes the button size
func (b *Button) ChangeSize(size string) {
	b.size = size
}

// SetEnabled sets the button enabled state
func (b *Button) SetEnabled(enabled bool) {
	b.enabled = enabled
}

// SetVisible sets the button visibility
func (b *Button) SetVisible(visible bool) {
	b.visible = visible
}

// Name returns the button name
func (b *Button) Name() string {
	return b.name
}

// Text returns the button text
func (b *Button) Text() string {
	return b.text
}

// Size returns the button size
func (b *Button) Size() string {
	return b.size
}

// Priority returns the button priority
func (b *Button) Priority() int {
	return b.priority
}

// IsEnabled returns whether the button is enabled
func (b *Button) IsEnabled() bool {
	return b.enabled
}

// IsVisible returns whether the button is visible
func (b *Button) IsVisible() bool {
	return b.visible
}

// Click triggers the button click event
func (b *Button) Click() error {
	if b.clicked != nil {
		return b.clicked.Trigger(b)
	}
	return nil
}

// ButtonRegisterModule manages UI buttons registration and interaction
type ButtonRegisterModule struct {
	*core.BaseModule
	manager *core.Manager

	// Button management
	buttons   map[string]*Button
	buttonsMu sync.RWMutex

	// Event handling
	eventModule core.EventModule
}

// NewButtonRegisterModule creates a new button register module
func NewButtonRegisterModule() *ButtonRegisterModule {
	base := core.NewBaseModule("buttonRegister", "button-register-module")
	base.SetRequires("event")
	base.SetPriority(1600) // High priority - UI components depend on this

	return &ButtonRegisterModule{
		BaseModule: base,
		buttons:    make(map[string]*Button),
	}
}

// Enable initializes the button register module
func (br *ButtonRegisterModule) Enable(ctx context.Context) error {
	if err := br.BaseModule.Enable(ctx); err != nil {
		return err
	}

	// Get event module
	if br.manager != nil {
		if eventMod, exists := br.manager.GetDefaultModule("event"); exists {
			if em, ok := eventMod.(core.EventModule); ok {
				br.eventModule = em
			}
		}
	}

	fmt.Println("Button register module enabled - ready to register buttons")
	return nil
}

// Disable shuts down the button register module
func (br *ButtonRegisterModule) Disable(ctx context.Context) error {
	br.buttonsMu.Lock()
	defer br.buttonsMu.Unlock()

	// Clear all buttons
	br.buttons = make(map[string]*Button)

	fmt.Println("Button register module disabled - all buttons cleared")
	return br.BaseModule.Disable(ctx)
}

// SetManager sets the module manager
func (br *ButtonRegisterModule) SetManager(manager *core.Manager) {
	br.manager = manager
}

// RegisterButton registers a new button and returns it
func (br *ButtonRegisterModule) RegisterButton(name string) *Button {
	br.buttonsMu.Lock()
	defer br.buttonsMu.Unlock()

	// Check if button already exists
	if existing, exists := br.buttons[name]; exists {
		return existing
	}

	// Create click event
	var clickedEvent core.Event
	if br.eventModule != nil {
		clickedEvent = br.eventModule.CreateEvent(fmt.Sprintf("button.%s.clicked", name))
	}

	// Create button
	button := NewButton(name, name, clickedEvent)
	br.buttons[name] = button

	fmt.Printf("Registered button: %s\n", name)
	return button
}

// UnregisterButton removes a button from the registry
func (br *ButtonRegisterModule) UnregisterButton(name string) bool {
	br.buttonsMu.Lock()
	defer br.buttonsMu.Unlock()

	if _, exists := br.buttons[name]; exists {
		delete(br.buttons, name)
		fmt.Printf("Unregistered button: %s\n", name)
		return true
	}
	return false
}

// GetButton retrieves a button by name
func (br *ButtonRegisterModule) GetButton(name string) (*Button, bool) {
	br.buttonsMu.RLock()
	defer br.buttonsMu.RUnlock()

	button, exists := br.buttons[name]
	return button, exists
}

// ListButtons returns all registered button names
func (br *ButtonRegisterModule) ListButtons() []string {
	br.buttonsMu.RLock()
	defer br.buttonsMu.RUnlock()

	names := make([]string, 0, len(br.buttons))
	for name := range br.buttons {
		names = append(names, name)
	}
	return names
}

// ButtonCount returns the number of registered buttons
func (br *ButtonRegisterModule) ButtonCount() int {
	br.buttonsMu.RLock()
	defer br.buttonsMu.RUnlock()

	return len(br.buttons)
}
