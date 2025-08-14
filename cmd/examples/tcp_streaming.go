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

	// Find the best audio device for desktop audio
	var audioID *string
	var audioDeviceName string

	// Priority: BlackHole -> SoundFlower -> Built-in
	for _, device := range audioDevices {
		if device.Name == "BlackHole 2ch" {
			audioID = &device.ID
			audioDeviceName = device.Name
			break
		}
	}

	if audioID == nil {
		for _, device := range audioDevices {
			if device.Name == "Soundflower (2ch)" {
				audioID = &device.ID
				audioDeviceName = device.Name
				break
			}
		}
	}

	// Fallback to first available device
	if audioID == nil {
		audioID = &audioDevices[0].ID
		audioDeviceName = audioDevices[0].Name
	}

	fmt.Printf("🎵 Using audio device: %s\n", audioDeviceName)

	// TCP streaming configuration
	host := "localhost"
	port := 9999

	// Configure TCP streaming options
	options := screencapturekit.StreamingOptions{
		FPS:                30,
		ShowCursor:         false,
		HighlightClicks:    false,
		ScreenID:           screens[0].ID,
		AudioDeviceID:      audioID,
		VideoCodec:         "h264", // Not used for audio-only
		
		// TCP Streaming configuration
		StreamingEnabled:   true,
		StreamingProtocol:  "tcp",
		StreamingHost:      &host,
		StreamingPort:      &port,
		AudioOnly:          true,
		StreamSystemAudio:  true,
		StreamMicrophone:   false, // TCP example uses system audio only
	}

	fmt.Println("\n🔌 TCP Audio Streaming Example")
	fmt.Printf("📡 Streaming to: %s:%d\n", host, port)
	fmt.Println("📋 Setup Instructions:")
	fmt.Println("   1. Start a TCP server to receive the stream:")
	fmt.Println("      nc -l 9999")
	fmt.Println("   2. Or use telnet to listen:")
	fmt.Println("      telnet localhost 9999")
	fmt.Println("   3. Set your system audio output to BlackHole/Soundflower if available")
	fmt.Println("   4. Play some music or audio to test the stream")
	fmt.Println()

	fmt.Print("Press Enter to start TCP streaming (make sure your TCP server is ready)...")
	fmt.Scanln()

	fmt.Println("🚀 Starting TCP audio streaming...")
	err = recorder.StartStreaming(options)
	if err != nil {
		log.Fatalf("Failed to start TCP streaming: %v", err)
	}

	// Stream for 20 seconds
	fmt.Println("📡 Streaming raw audio data via TCP for 20 seconds...")
	fmt.Println("💡 You should see binary audio data flowing to your TCP server")
	fmt.Println("🎵 The data includes both audio samples and source metadata")
	
	for i := 20; i > 0; i-- {
		fmt.Printf("⏱️  Streaming... %d seconds remaining\n", i)
		time.Sleep(1 * time.Second)
	}

	fmt.Println("🔌 Stopping TCP streaming...")
	err = recorder.StopStreaming()
	if err != nil {
		log.Fatalf("Failed to stop streaming: %v", err)
	}

	fmt.Println("✅ TCP audio streaming example completed!")
	fmt.Println()
	fmt.Println("📊 What happened:")
	fmt.Println("   • Raw audio data captured from system audio device")
	fmt.Println("   • Audio samples sent as binary data over TCP connection")
	fmt.Println("   • Low-latency streaming with minimal protocol overhead")
	fmt.Println("   • Simple packet format with source identification")
	fmt.Println("   • Suitable for high-performance audio streaming applications")
	fmt.Println()
	fmt.Println("💡 Use Cases:")
	fmt.Println("   • Real-time audio transmission to remote servers")
	fmt.Println("   • Integration with audio processing pipelines")
	fmt.Println("   • Custom audio streaming protocols")
	fmt.Println("   • Low-latency audio distribution systems")
}