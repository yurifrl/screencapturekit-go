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

	// Configure recording options
	options := screencapturekit.RecordingOptions{
		FPS:             30,
		ShowCursor:      true,
		HighlightClicks: false,
		ScreenID:        screens[0].ID, // Use first screen
		VideoCodec:      "h264",
		EnableHDR:       false,
	}

	fmt.Println("Starting recording...")
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

	fmt.Printf("Recording saved to: %s\n", videoPath)
}