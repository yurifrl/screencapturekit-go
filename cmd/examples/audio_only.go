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

	// Get available screens (still needed even for audio-only)
	screens, err := screencapturekit.GetScreens()
	if err != nil {
		log.Fatalf("Failed to get screens: %v", err)
	}

	// Get available audio devices
	audioDevices, err := screencapturekit.GetAudioDevices()
	if err != nil {
		log.Fatalf("Failed to get audio devices: %v", err)
	}

	// Get available microphone devices
	micDevices, err := screencapturekit.GetMicrophoneDevices()
	if err != nil {
		log.Fatalf("Failed to get microphone devices: %v", err)
	}

	fmt.Printf("Available screens: %d\n", len(screens))
	fmt.Printf("Available audio devices: %d\n", len(audioDevices))
	for i, device := range audioDevices {
		fmt.Printf("  %d: %s (ID: %s)\n", i, device.Name, device.ID)
	}
	
	fmt.Printf("Available microphone devices: %d\n", len(micDevices))
	for i, device := range micDevices {
		fmt.Printf("  %d: %s (ID: %s)\n", i, device.Name, device.ID)
	}

	if len(screens) == 0 {
		log.Fatal("No screens available")
	}

	// Configure audio-only recording options
	options := screencapturekit.RecordingOptions{
		FPS:             30,                 // Still needed but not used for audio-only
		ShowCursor:      false,              // Not relevant for audio-only
		HighlightClicks: false,              // Not relevant for audio-only
		ScreenID:        screens[0].ID,      // Still required by the underlying system
		VideoCodec:      "h264",             // Not used for audio-only
		AudioOnly:       true,               // This enables audio-only mode
	}

	// Add system audio if available
	if len(audioDevices) > 0 {
		options.AudioDeviceID = &audioDevices[0].ID
		fmt.Printf("Using audio device: %s\n", audioDevices[0].Name)
	} else {
		fmt.Println("No audio devices available")
		return
	}

	// Add microphone if available and supported
	if len(micDevices) > 0 && screencapturekit.SupportsMicrophone() {
		options.MicrophoneDeviceID = &micDevices[0].ID
		fmt.Printf("Using microphone: %s\n", micDevices[0].Name)
	} else if len(micDevices) > 0 {
		fmt.Println("Microphone devices available but not supported on this macOS version (requires 15.0+)")
	} else {
		fmt.Println("No microphone devices available")
	}

	fmt.Println("\n🎵 Starting audio-only recording...")
	fmt.Println("This will record system audio (and microphone if supported)")
	fmt.Println("The output will be converted to MP3 format")
	
	err = recorder.StartRecording(options)
	if err != nil {
		log.Fatalf("Failed to start audio recording: %v", err)
	}

	// Record for 15 seconds
	fmt.Println("Recording audio for 15 seconds...")
	for i := 15; i > 0; i-- {
		fmt.Printf("⏱️  %d seconds remaining...\n", i)
		time.Sleep(1 * time.Second)
	}

	fmt.Println("Stopping audio recording...")
	audioPath, err := recorder.StopRecording()
	if err != nil {
		log.Fatalf("Failed to stop recording: %v", err)
	}

	fmt.Printf("🎵 Audio-only recording saved to: %s\n", audioPath)
	fmt.Println("\nNote: The file extension might be .mov but it should contain audio data.")
	fmt.Println("If AudioOnly=true is properly implemented, it should be converted to MP3.")
}