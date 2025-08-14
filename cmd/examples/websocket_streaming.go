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

	// Get microphone devices
	micDevices, err := screencapturekit.GetMicrophoneDevices()
	if err != nil {
		log.Fatalf("Failed to get microphone devices: %v", err)
	}

	if len(screens) == 0 {
		log.Fatal("No screens available")
	}

	if len(audioDevices) == 0 {
		log.Fatal("No audio devices available")
	}

	fmt.Printf("Available screens: %d\n", len(screens))
	fmt.Printf("Available audio devices: %d\n", len(audioDevices))
	fmt.Printf("Available microphone devices: %d\n", len(micDevices))

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

	// Use microphone if available and supported
	var micID *string
	if len(micDevices) > 0 && screencapturekit.SupportsMicrophone() {
		micID = &micDevices[0].ID
		fmt.Printf("🎤 Using microphone: %s\n", micDevices[0].Name)
	}

	// WebSocket streaming endpoint
	streamingURL := "ws://localhost:9090/audio-stream"

	// Configure WebSocket streaming options
	options := screencapturekit.StreamingOptions{
		FPS:                30,
		ShowCursor:         false,
		HighlightClicks:    false,
		ScreenID:           screens[0].ID,
		AudioDeviceID:      blackHoleID,
		MicrophoneDeviceID: micID,
		VideoCodec:         "h264", // Not used for audio-only
		
		// Streaming configuration
		StreamingEnabled:   true,
		StreamingProtocol:  "websocket",
		StreamingURL:       &streamingURL,
		AudioOnly:          true,
		StreamSystemAudio:  true,
		StreamMicrophone:   micID != nil,
	}

	fmt.Println("\n🔗 WebSocket Audio Streaming Example")
	fmt.Printf("📡 Streaming to: %s\n", streamingURL)
	fmt.Println("📋 Setup Instructions:")
	fmt.Println("   1. Install and run a WebSocket server:")
	fmt.Println("      npm install -g wscat")
	fmt.Println("      wscat -l 9090")
	fmt.Println("   2. Or use a simple WebSocket echo server")
	fmt.Println("   3. Set system audio output to BlackHole")
	fmt.Println("   4. Enable microphone if you want both audio sources")
	fmt.Println()

	fmt.Print("Press Enter to start WebSocket streaming (make sure your WebSocket server is ready)...")
	fmt.Scanln()

	fmt.Println("🚀 Starting WebSocket audio streaming...")
	err = recorder.StartStreaming(options)
	if err != nil {
		log.Fatalf("Failed to start WebSocket streaming: %v", err)
	}

	// Stream for 45 seconds
	fmt.Println("📡 Streaming audio via WebSocket for 45 seconds...")
	if micID != nil {
		fmt.Println("🎤 Streaming both system audio AND microphone")
	} else {
		fmt.Println("🎵 Streaming system audio only")
	}
	fmt.Println("💡 You should see real-time JSON messages on your WebSocket server")
	
	for i := 45; i > 0; i-- {
		fmt.Printf("⏱️  Streaming... %d seconds remaining\n", i)
		time.Sleep(1 * time.Second)
	}

	fmt.Println("🔌 Stopping WebSocket streaming...")
	err = recorder.StopStreaming()
	if err != nil {
		log.Fatalf("Failed to stop streaming: %v", err)
	}

	fmt.Println("✅ WebSocket audio streaming example completed!")
	fmt.Println()
	fmt.Println("📊 What happened:")
	fmt.Println("   • Real-time audio captured from system and/or microphone")
	fmt.Println("   • Audio encoded and sent as WebSocket messages")
	fmt.Println("   • Bidirectional communication possible")
	fmt.Println("   • JSON metadata included with each audio frame")
	fmt.Println("   • Efficient real-time streaming protocol")
}