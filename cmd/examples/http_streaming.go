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

	if len(screens) == 0 {
		log.Fatal("No screens available")
	}

	if len(audioDevices) == 0 {
		log.Fatal("No audio devices available")
	}

	fmt.Printf("Available screens: %d\n", len(screens))
	fmt.Printf("Available audio devices: %d\n", len(audioDevices))

	// Find BlackHole for desktop audio capture
	var blackHoleID *string
	for _, device := range audioDevices {
		if device.Name == "BlackHole 2ch" {
			blackHoleID = &device.ID
			fmt.Printf("🎵 Found BlackHole for desktop audio: %s\n", device.ID)
			break
		}
	}

	// Fallback to first audio device if BlackHole not found
	if blackHoleID == nil {
		blackHoleID = &audioDevices[0].ID
		fmt.Printf("🎵 Using audio device: %s\n", audioDevices[0].Name)
	}

	// HTTP streaming endpoint
	streamingURL := "http://localhost:8080/audio-stream"

	// Configure HTTP streaming options
	options := screencapturekit.StreamingOptions{
		FPS:                30,
		ShowCursor:         false, // Not relevant for audio streaming
		HighlightClicks:    false,
		ScreenID:           screens[0].ID,
		AudioDeviceID:      blackHoleID,
		VideoCodec:         "h264", // Not used for audio-only
		
		// Streaming configuration
		StreamingEnabled:   true,
		StreamingProtocol:  "http",
		StreamingURL:       &streamingURL,
		AudioOnly:          true,
		StreamSystemAudio:  true,
		StreamMicrophone:   false, // System audio only for this example
	}

	fmt.Println("\n🌊 HTTP Audio Streaming Example")
	fmt.Printf("📡 Streaming to: %s\n", streamingURL)
	fmt.Println("📋 Setup Instructions:")
	fmt.Println("   1. Start a server to receive the stream:")
	fmt.Println("      python3 -m http.server 8080")
	fmt.Println("   2. Or use netcat to listen:")
	fmt.Println("      nc -l 8080")
	fmt.Println("   3. Set your system audio output to BlackHole if available")
	fmt.Println("   4. Play some audio to test the stream")
	fmt.Println()

	fmt.Print("Press Enter to start streaming (make sure your server is ready)...")
	fmt.Scanln()

	fmt.Println("🚀 Starting HTTP audio streaming...")
	err = recorder.StartStreaming(options)
	if err != nil {
		log.Fatalf("Failed to start HTTP streaming: %v", err)
	}

	// Stream for 30 seconds
	fmt.Println("📡 Streaming audio via HTTP for 30 seconds...")
	fmt.Println("💡 You should see HTTP POST requests on your server")
	
	for i := 30; i > 0; i-- {
		fmt.Printf("⏱️  Streaming... %d seconds remaining\n", i)
		time.Sleep(1 * time.Second)
	}

	fmt.Println("🔌 Stopping HTTP streaming...")
	err = recorder.StopStreaming()
	if err != nil {
		log.Fatalf("Failed to stop streaming: %v", err)
	}

	fmt.Println("✅ HTTP audio streaming example completed!")
	fmt.Println()
	fmt.Println("📊 What happened:")
	fmt.Println("   • Real-time audio data was captured from your system")
	fmt.Println("   • Audio was encoded and sent via HTTP POST requests")
	fmt.Println("   • Each audio frame was transmitted separately")
	fmt.Println("   • The receiving server got continuous audio stream data")
}