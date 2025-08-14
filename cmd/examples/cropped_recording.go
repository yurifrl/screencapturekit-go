package main

import (
	"fmt"
	"log"
	"time"

	screencapturekit "github.com/tfsoares/screencapturekit-go"
)

func main() {
	// Create a new ScreenCaptureKit instance
	recorder, err := screencapturekit.NewScreenCaptureKit()
	if err != nil {
		log.Fatalf("Failed to create recorder: %v", err)
	}
	defer recorder.Cleanup()

	// Get available screens
	screens, err := screencapturekit.GetScreens()
	if err != nil {
		log.Fatalf("Failed to get screens: %v", err)
	}

	fmt.Printf("Available screens: %d\n", len(screens))
	for i, screen := range screens {
		fmt.Printf("Screen %d: ID=%d, Size=%dx%d\n", i, screen.ID, screen.Width, screen.Height)
	}

	if len(screens) == 0 {
		log.Fatal("No screens available")
	}

	selectedScreen := screens[0]

	// Define a crop area (center quarter of the screen)
	cropArea := &screencapturekit.CropArea{
		X:      float64(selectedScreen.Width) * 0.25,  // Start at 25% from left
		Y:      float64(selectedScreen.Height) * 0.25, // Start at 25% from top
		Width:  float64(selectedScreen.Width) * 0.5,   // Width is 50% of screen
		Height: float64(selectedScreen.Height) * 0.5,  // Height is 50% of screen
	}

	fmt.Printf("Crop area: x=%.0f, y=%.0f, width=%.0f, height=%.0f\n",
		cropArea.X, cropArea.Y, cropArea.Width, cropArea.Height)

	// Configure recording options with crop area
	options := screencapturekit.RecordingOptions{
		FPS:             30,
		ShowCursor:      true,
		HighlightClicks: true, // Highlight clicks to see interaction
		ScreenID:        selectedScreen.ID,
		VideoCodec:      "h264",
		EnableHDR:       false,
		CropArea:        cropArea,
	}

	fmt.Println("Starting cropped recording...")
	fmt.Println("Try moving your mouse and clicking in the center area of the screen")
	
	err = recorder.StartRecording(options)
	if err != nil {
		log.Fatalf("Failed to start recording: %v", err)
	}

	// Record for 10 seconds
	time.Sleep(10 * time.Second)

	fmt.Println("Stopping recording...")
	videoPath, err := recorder.StopRecording()
	if err != nil {
		log.Fatalf("Failed to stop recording: %v", err)
	}

	fmt.Printf("Cropped recording saved to: %s\n", videoPath)
}