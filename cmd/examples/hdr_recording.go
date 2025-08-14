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

	// Check if HDR is supported
	if !screencapturekit.SupportsHDR() {
		log.Fatal("HDR recording is not supported on this macOS version (requires 13.0+)")
	}

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

	// Configure recording options with HDR
	options := screencapturekit.RecordingOptions{
		FPS:             60, // Higher FPS for HDR content
		ShowCursor:      true,
		HighlightClicks: false,
		ScreenID:        screens[0].ID, // Use first screen
		VideoCodec:      "hevc",        // HEVC is better for HDR
		EnableHDR:       true,
	}

	// Check if HEVC is supported
	if screencapturekit.SupportsHEVC() {
		fmt.Println("Using HEVC codec for HDR recording")
	} else {
		fmt.Println("HEVC not supported, falling back to H.264")
		options.VideoCodec = "h264"
	}

	fmt.Println("Starting HDR recording...")
	err = recorder.StartRecording(options)
	if err != nil {
		log.Fatalf("Failed to start HDR recording: %v", err)
	}

	// Record for 10 seconds
	fmt.Println("Recording HDR content for 10 seconds...")
	time.Sleep(10 * time.Second)

	fmt.Println("Stopping HDR recording...")
	videoPath, err := recorder.StopRecording()
	if err != nil {
		log.Fatalf("Failed to stop recording: %v", err)
	}

	fmt.Printf("HDR recording saved to: %s\n", videoPath)
}