package main

import (
	"context"
	"fmt"
	"log"
	"os"
	"os/signal"
	"syscall"

	"github.com/LaPingvino/openteacher/internal/core"
	"github.com/LaPingvino/openteacher/internal/modules"
	"github.com/LaPingvino/openteacher/internal/modules/data/profiledescriptions"
	startwidget "github.com/LaPingvino/openteacher/internal/modules/interfaces/qt/startWidget"
	"github.com/LaPingvino/openteacher/internal/modules/profilerunners/backgroundimage"
	"github.com/LaPingvino/openteacher/internal/modules/profilerunners/businesscard"
	"github.com/LaPingvino/openteacher/internal/modules/interfaces/qt/qtApp"
)

const (
	appName    = "OpenTeacher"
	appVersion = "4.0.0-alpha"
)

func main() {
	// Setup logging
	log.SetFlags(log.LstdFlags | log.Lshortfile)

	fmt.Printf("%s %s - Starting...\n", appName, appVersion)

	// Create module manager
	manager := core.NewManager()

	// Register all modules
	if err := registerAllModules(manager); err != nil {
		log.Fatalf("Failed to register modules: %v", err)
	}

	fmt.Printf("Registered %d modules of %d types\n", manager.ModuleCount(), manager.TypeCount())

	// Setup context with cancellation
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	// Setup signal handling for graceful shutdown
	sigChan := make(chan os.Signal, 1)
	signal.Notify(sigChan, syscall.SIGINT, syscall.SIGTERM)

	go func() {
		sig := <-sigChan
		fmt.Printf("\nReceived signal: %v\n", sig)
		fmt.Println("Shutting down gracefully...")
		cancel()
	}()

	// Enable all modules
	fmt.Println("Enabling modules...")
	if err := manager.EnableAll(ctx); err != nil {
		log.Fatalf("Failed to enable modules: %v", err)
	}

	fmt.Println("All modules enabled successfully")

	// Start the main application
	if err := runApplication(ctx, manager); err != nil {
		log.Fatalf("Application error: %v", err)
	}

	// Shutdown
	fmt.Println("Disabling modules...")
	if err := manager.DisableAll(context.Background()); err != nil {
		log.Printf("Error during shutdown: %v", err)
	}

	fmt.Println("OpenTeacher shutdown complete")
}

func registerAllModules(manager *core.Manager) error {
	// Create and register essential modules
	executeModule := modules.NewExecuteModule()
	if err := manager.Register(executeModule); err != nil {
		return fmt.Errorf("failed to register execute module: %w", err)
	}

	eventModule := modules.NewEventModule()
	if err := manager.Register(eventModule); err != nil {
		return fmt.Errorf("failed to register event module: %w", err)
	}

	settingsModule := modules.NewSettingsModule()
	if err := manager.Register(settingsModule); err != nil {
		return fmt.Errorf("failed to register settings module: %w", err)
	}

	metadataModule := modules.NewMetadataModule()
	if err := manager.Register(metadataModule); err != nil {
		return fmt.Errorf("failed to register metadata module: %w", err)
	}

	buttonRegisterModule := modules.NewButtonRegisterModule()
	if err := manager.Register(buttonRegisterModule); err != nil {
		return fmt.Errorf("failed to register button register module: %w", err)
	}

	startWidgetModule := startwidget.InitStartwidgetModule()
	if err := manager.Register(startWidgetModule); err != nil {
		return fmt.Errorf("failed to register start widget module: %w", err)
	}

	uiModule := modules.NewUIModule()
	if err := manager.Register(uiModule); err != nil {
		return fmt.Errorf("failed to register ui module: %w", err)
	}

	// Register business card related modules
	businessCardModule := businesscard.Init()
	if err := manager.Register(businessCardModule); err != nil {
		return fmt.Errorf("failed to register business card module: %w", err)
	}

	backgroundImageModule := backgroundimage.Init()
	if err := manager.Register(backgroundImageModule); err != nil {
		return fmt.Errorf("failed to register background image module: %w", err)
	}

	profileDescModule := profiledescriptions.Init()
	if err := manager.Register(profileDescModule); err != nil {
		return fmt.Errorf("failed to register profile description module: %w", err)
	}


	// Register qtapp module
	qtappModule := qtapp.NewQtAppModule()
	if err := manager.Register(qtappModule); err != nil {
		return fmt.Errorf("failed to register qtapp module: %w", err)
	}

	return nil
}

func runApplication(ctx context.Context, manager *core.Manager) error {
	// Get the execute module and start the main loop
	executeModule, exists := manager.GetDefaultModule("execute")
	if !exists {
		return fmt.Errorf("no execute module found")
	}

	execMod, ok := executeModule.(core.ExecuteModule)
	if !ok {
		return fmt.Errorf("execute module does not implement ExecuteModule interface")
	}

	fmt.Println("Starting main application loop...")

	// Start running - this will block until context is cancelled
	if err := execMod.StartRunning(ctx); err != nil {
		return fmt.Errorf("execute module failed: %w", err)
	}

	return nil
}
