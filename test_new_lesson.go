package main

import (
	"context"
	"fmt"
	"log"
	"os"

	"github.com/LaPingvino/recuerdo/internal/core"
	"github.com/LaPingvino/recuerdo/internal/modules/interfaces/qt/gui"
	"github.com/LaPingvino/recuerdo/internal/modules/interfaces/qt/lessonDialogs"
	qtapp "github.com/LaPingvino/recuerdo/internal/modules/interfaces/qt/qtApp"
	qtcore "github.com/therecipe/qt/core"
	"github.com/therecipe/qt/widgets"
)

func main() {
	fmt.Println("🧪 Testing New Lesson Dialog Functionality")
	fmt.Println("===========================================")

	// Initialize Qt application
	app := widgets.NewQApplication(len(os.Args), os.Args)
	app.SetApplicationName("Recuerdo Test")
	app.SetApplicationVersion("1.0.0")

	// Create module manager
	manager := core.NewManager()

	// Register required modules
	fmt.Println("Registering modules...")

	// Qt App module (required for GUI)
	qtAppModule := qtapp.NewQtAppModule()
	if err := manager.Register(qtAppModule); err != nil {
		log.Fatalf("Failed to register qtApp module: %v", err)
	}

	// Lesson Dialogs module
	lessonDialogsModule := lessonDialogs.NewLessonDialogsModule()
	if err := manager.Register(lessonDialogsModule); err != nil {
		log.Fatalf("Failed to register lessonDialogs module: %v", err)
	}

	// GUI module (for main window)
	guiModule := gui.NewGuiModule()
	if err := manager.Register(guiModule); err != nil {
		log.Fatalf("Failed to register gui module: %v", err)
	}

	fmt.Println("✅ Modules registered successfully")

	// Enable all modules
	fmt.Println("Enabling modules...")
	ctx := context.Background()
	if err := manager.EnableAll(ctx); err != nil {
		log.Fatalf("Failed to enable modules: %v", err)
	}
	fmt.Println("✅ Modules enabled successfully")

	// Verify lessonDialogs module is available
	lessonDialogModules := manager.GetModulesByType("lessonDialogs")
	fmt.Printf("Found %d lessonDialogs modules\n", len(lessonDialogModules))

	if len(lessonDialogModules) == 0 {
		log.Fatal("❌ No lessonDialogs modules found!")
	}

	// Test the ShowNewLessonDialog method
	fmt.Println("✅ lessonDialogs module found - testing interface...")

	lessonMod := lessonDialogModules[0]
	dialogMod, hasMethod := lessonMod.(interface{ ShowNewLessonDialog() map[string]interface{} })

	if !hasMethod {
		log.Fatal("❌ ShowNewLessonDialog method not implemented!")
	}

	fmt.Println("✅ ShowNewLessonDialog method is available")

	// Create a test window
	testWindow := widgets.NewQMainWindow(nil, 0)
	testWindow.SetWindowTitle("New Lesson Dialog Test")
	testWindow.SetFixedSize2(400, 300)

	centralWidget := widgets.NewQWidget(nil, 0)
	testWindow.SetCentralWidget(centralWidget)

	layout := widgets.NewQVBoxLayout()
	centralWidget.SetLayout(layout)

	// Instructions
	instructionLabel := widgets.NewQLabel2("New Lesson Dialog Test", nil, 0)
	instructionLabel.SetAlignment(qtcore.Qt__AlignCenter)
	instructionLabel.SetStyleSheet("font-size: 16px; font-weight: bold; margin: 10px;")
	layout.AddWidget(instructionLabel, 0, 0)

	statusLabel := widgets.NewQLabel2("Click 'Test Dialog' to test the new lesson creation dialog", nil, 0)
	statusLabel.SetAlignment(qtcore.Qt__AlignCenter)
	statusLabel.SetWordWrap(true)
	statusLabel.SetStyleSheet("margin: 10px; color: #666;")
	layout.AddWidget(statusLabel, 0, 0)

	// Test button
	testButton := widgets.NewQPushButton2("Test Dialog", nil)
	testButton.SetFixedSize2(200, 50)
	testButton.SetStyleSheet("font-size: 14px; padding: 10px;")
	layout.AddWidget(testButton, 0, qtcore.Qt__AlignCenter)

	// Result display
	resultText := widgets.NewQTextEdit(nil)
	resultText.SetReadOnly(true)
	resultText.SetMaximumHeight(120)
	resultText.SetStyleSheet("background-color: #f5f5f5; margin: 10px;")
	layout.AddWidget(resultText, 0, 0)

	// Connect button
	testButton.ConnectClicked(func(checked bool) {
		fmt.Println("🚀 Testing ShowNewLessonDialog()...")
		statusLabel.SetText("Opening new lesson dialog...")

		// Call the dialog method
		result := dialogMod.ShowNewLessonDialog()

		if result != nil {
			fmt.Printf("✅ Dialog returned data: %v\n", result)
			statusLabel.SetText("✅ Success! Dialog returned data:")

			resultText.Clear()
			resultText.Append("Dialog Results:")
			for key, value := range result {
				resultText.Append(fmt.Sprintf("• %s: %v", key, value))
			}
		} else {
			fmt.Println("ℹ️  Dialog was cancelled or returned nil")
			statusLabel.SetText("ℹ️  Dialog was cancelled")
			resultText.SetPlainText("Dialog was cancelled by user")
		}
	})

	// Quit button
	quitButton := widgets.NewQPushButton2("Quit Test", nil)
	quitButton.SetStyleSheet("background-color: #ff6b6b; color: white; padding: 8px;")
	quitButton.ConnectClicked(func(checked bool) {
		app.Quit()
	})
	layout.AddWidget(quitButton, 0, qtcore.Qt__AlignCenter)

	// Auto-close timer (60 seconds)
	timer := qtcore.NewQTimer(nil)
	timer.ConnectTimeout(func() {
		fmt.Println("⏰ Auto-closing test after 60 seconds")
		app.Quit()
	})
	timer.Start(60000) // 60 seconds

	// Show test window
	testWindow.Show()

	fmt.Println("🎬 Test window shown!")
	fmt.Println("   • Click 'Test Dialog' to open the new lesson dialog")
	fmt.Println("   • Test creating a new lesson with different settings")
	fmt.Println("   • Window will auto-close in 60 seconds")

	// Run the application
	exitCode := app.Exec()

	fmt.Printf("\n✅ Test completed with exit code: %d\n", exitCode)
}
