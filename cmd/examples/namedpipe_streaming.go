package main

import (
	"fmt"
	"log"
	"os"
	"path/filepath"
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

	// Use microphone if available and supported
	var micID *string
	if len(micDevices) > 0 && screencapturekit.SupportsMicrophone() {
		micID = &micDevices[0].ID
		fmt.Printf("🎤 Using microphone: %s\n", micDevices[0].Name)
	}

	// Create named pipe path
	tempDir := os.TempDir()
	pipePath := filepath.Join(tempDir, "screencapture_audio.fifo")
	
	// Clean up any existing pipe
	os.Remove(pipePath)

	// Configure named pipe streaming options
	options := screencapturekit.StreamingOptions{
		FPS:                30,
		ShowCursor:         false,
		HighlightClicks:    false,
		ScreenID:           screens[0].ID,
		AudioDeviceID:      audioID,
		MicrophoneDeviceID: micID,
		VideoCodec:         "h264", // Not used for audio-only
		
		// Named Pipe Streaming configuration
		StreamingEnabled:   true,
		StreamingProtocol:  "pipe",
		StreamingPipePath:  &pipePath,
		AudioOnly:          true,
		StreamSystemAudio:  true,
		StreamMicrophone:   micID != nil,
	}

	fmt.Println("\n📁 Named Pipe Audio Streaming Example")
	fmt.Printf("📡 Streaming to named pipe: %s\n", pipePath)
	fmt.Println("📋 What Named Pipes Provide:")
	fmt.Println("   ✅ Ultra-low latency (local IPC)")
	fmt.Println("   ✅ High throughput (no network overhead)")
	fmt.Println("   ✅ OS-level buffering and flow control")
	fmt.Println("   ✅ Perfect for local audio processing")
	fmt.Println("   ✅ Compatible with any program that can read files")
	fmt.Println()
	fmt.Println("💡 Consumer Examples:")
	fmt.Println("   # Read raw audio with dd")
	fmt.Printf("   dd if=%s of=audio.raw bs=4096\n", pipePath)
	fmt.Println()
	fmt.Println("   # Play directly with ffplay")
	fmt.Printf("   ffplay -f f32le -ar 48000 -ac 2 %s\n", pipePath)
	fmt.Println()
	fmt.Println("   # Process with Python")
	fmt.Printf("   python3 test-servers/namedpipe_reader.py %s\n", pipePath)
	fmt.Println()

	fmt.Print("Press Enter to start named pipe streaming...")
	fmt.Scanln()

	fmt.Println("🚀 Starting named pipe audio streaming...")
	err = recorder.StartStreaming(options)
	if err != nil {
		log.Fatalf("Failed to start named pipe streaming: %v", err)
	}

	// Stream for 30 seconds
	fmt.Println("📁 Streaming audio to named pipe for 30 seconds...")
	if micID != nil {
		fmt.Println("🎤 Streaming both system audio AND microphone")
	} else {
		fmt.Println("🎵 Streaming system audio only")
	}
	fmt.Printf("📡 Data format: 32-bit float PCM, 48kHz, stereo\n")
	fmt.Printf("📦 Packet format: [source_len][source][timestamp][audio_data]\n")
	fmt.Printf("💡 Pipe will block until a reader connects\n")
	
	fmt.Println("\n🔄 Open another terminal and try:")
	fmt.Printf("   cat %s > /dev/null   # Basic read test\n", pipePath)
	fmt.Printf("   hexdump -C %s | head # Inspect data format\n", pipePath)
	fmt.Println()
	
	for i := 30; i > 0; i-- {
		fmt.Printf("⏱️  Streaming... %d seconds remaining\n", i)
		time.Sleep(1 * time.Second)
	}

	fmt.Println("🔌 Stopping named pipe streaming...")
	err = recorder.StopStreaming()
	if err != nil {
		log.Fatalf("Failed to stop streaming: %v", err)
	}

	// Clean up the named pipe
	os.Remove(pipePath)

	fmt.Println("✅ Named pipe audio streaming example completed!")
	fmt.Println()
	fmt.Println("📊 What happened:")
	fmt.Println("   • Real-time audio captured from system/microphone")
	fmt.Println("   • Audio encoded as 32-bit float PCM (native format)")
	fmt.Println("   • Data written to named pipe with metadata headers")
	fmt.Println("   • OS handled buffering and synchronization")
	fmt.Println("   • Zero network overhead for maximum performance")
	fmt.Println()
	fmt.Println("🎯 Perfect Use Cases:")
	fmt.Println("   • Local audio processing pipelines")
	fmt.Println("   • Real-time audio analysis (ML/DSP)")
	fmt.Println("   • Integration with audio software (DAWs, etc.)")
	fmt.Println("   • High-performance audio streaming within the system")
	fmt.Println("   • Debugging audio capture without network complexity")
}