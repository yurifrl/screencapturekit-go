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
	fmt.Printf("Available microphone devices: %d\n", len(micDevices))

	if len(screens) == 0 {
		log.Fatal("No screens available")
	}

	// Configure recording options with audio
	options := screencapturekit.RecordingOptions{
		FPS:             30,
		ShowCursor:      true,
		HighlightClicks: false,
		ScreenID:        screens[0].ID,
		VideoCodec:      "h264",
		EnableHDR:       false,
	}

	// Add system audio if available
	if len(audioDevices) > 0 {
		options.AudioDeviceID = &audioDevices[0].ID
		fmt.Printf("Using audio device: %s\n", audioDevices[0].Name)
	}

	// Add microphone if available and supported
	if len(micDevices) > 0 && screencapturekit.SupportsMicrophone() {
		options.MicrophoneDeviceID = &micDevices[0].ID
		fmt.Printf("Using microphone: %s\n", micDevices[0].Name)
	} else if len(micDevices) > 0 {
		fmt.Println("Microphone devices available but not supported on this macOS version (requires 15.0+)")
	}

	fmt.Println("Starting recording with audio...")
	err = recorder.StartRecording(options)
	if err != nil {
		log.Fatalf("Failed to start recording: %v", err)
	}

	// Record for 15 seconds
	time.Sleep(15 * time.Second)

	fmt.Println("Stopping recording...")
	videoPath, err := recorder.StopRecording()
	if err != nil {
		log.Fatalf("Failed to stop recording: %v", err)
	}

	fmt.Printf("Recording with audio saved to: %s\n", videoPath)
}