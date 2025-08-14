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
	pipePath := filepath.Join(tempDir, "screencapture_ffmpeg.fifo")
	
	// Clean up any existing pipe
	os.Remove(pipePath)

	// Configure FFmpeg-compatible named pipe streaming options
	options := screencapturekit.StreamingOptions{
		FPS:                30,
		ShowCursor:         false,
		HighlightClicks:    false,
		ScreenID:           screens[0].ID,
		AudioDeviceID:      audioID,
		MicrophoneDeviceID: micID,
		VideoCodec:         "h264", // Not used for audio-only
		
		// FFmpeg-Compatible Named Pipe Streaming configuration
		StreamingEnabled:   true,
		StreamingProtocol:  "pipe",
		StreamingPipePath:  &pipePath,
		AudioOnly:          true,
		StreamSystemAudio:  true,
		StreamMicrophone:   micID != nil,
		FFmpegCompatible:   true, // Enable FFmpeg compatibility mode
	}

	fmt.Println("\n📁 FFmpeg-Compatible Named Pipe Audio Streaming Example")
	fmt.Printf("📡 Streaming to named pipe: %s\n", pipePath)
	fmt.Println("🎵 Audio Format: f32le (32-bit float PCM), 48kHz, stereo")
	fmt.Println("🔧 FFmpeg Compatible: Raw audio stream (no metadata headers)")
	fmt.Println()
	fmt.Println("📋 What FFmpeg Compatibility Provides:")
	fmt.Println("   ✅ Direct compatibility with FFmpeg tools")
	fmt.Println("   ✅ Raw f32le audio stream format")
	fmt.Println("   ✅ No custom headers or metadata")
	fmt.Println("   ✅ Standard Unix named pipe (FIFO)")
	fmt.Println("   ✅ Perfect for FFmpeg processing pipelines")
	fmt.Println()
	fmt.Println("💡 FFmpeg Usage Examples:")
	fmt.Println("   # Play directly with ffplay")
	fmt.Printf("   ffplay -f f32le -ar 48000 -channels 2 %s\n", pipePath)
	fmt.Println()
	fmt.Println("   # Convert to MP3")
	fmt.Printf("   ffmpeg -f f32le -ar 48000 -channels 2 -i %s output.mp3\n", pipePath)
	fmt.Println()
	fmt.Println("   # Stream to network")
	fmt.Printf("   ffmpeg -f f32le -ar 48000 -channels 2 -i %s -f mp3 icecast://server:8000/stream\n", pipePath)
	fmt.Println()
	fmt.Println("   # Real-time analysis with custom tool")
	fmt.Printf("   your-audio-tool < %s\n", pipePath)
	fmt.Println()

	fmt.Print("Press Enter to start FFmpeg-compatible streaming...")
	fmt.Scanln()

	fmt.Println("🚀 Starting FFmpeg-compatible named pipe audio streaming...")
	err = recorder.StartStreaming(options)
	if err != nil {
		log.Fatalf("Failed to start FFmpeg-compatible streaming: %v", err)
	}

	// Stream for 60 seconds
	fmt.Println("📁 Streaming raw audio to named pipe for 60 seconds...")
	if micID != nil {
		fmt.Println("🎤 Streaming both system audio AND microphone")
	} else {
		fmt.Println("🎵 Streaming system audio only")
	}
	fmt.Printf("🎵 Format: f32le, 48kHz, stereo (FFmpeg native format)\n")
	fmt.Printf("📦 Data: Pure audio stream (no headers)\n")
	fmt.Printf("💡 Pipe blocks until FFmpeg/reader connects\n")
	
	fmt.Println("\n🔄 Open another terminal and try these commands:")
	fmt.Printf("   # Basic playback test\n")
	fmt.Printf("   ffplay -f f32le -ar 48000 -channels 2 %s\n", pipePath)
	fmt.Println()
	fmt.Printf("   # Save to WAV file\n")
	fmt.Printf("   ffmpeg -f f32le -ar 48000 -channels 2 -i %s -t 10 output.wav\n", pipePath)
	fmt.Println()
	fmt.Printf("   # Monitor with ffprobe\n")
	fmt.Printf("   ffprobe -f f32le -ar 48000 -channels 2 %s\n", pipePath)
	fmt.Println()
	
	// Show countdown
	for i := 60; i > 0; i-- {
		fmt.Printf("⏱️  Streaming... %d seconds remaining\n", i)
		time.Sleep(1 * time.Second)
	}

	fmt.Println("🔌 Stopping FFmpeg-compatible streaming...")
	err = recorder.StopStreaming()
	if err != nil {
		log.Fatalf("Failed to stop streaming: %v", err)
	}

	// Clean up the named pipe
	os.Remove(pipePath)

	fmt.Println("✅ FFmpeg-compatible streaming example completed!")
	fmt.Println()
	fmt.Println("📊 What happened:")
	fmt.Println("   • Real-time audio captured from system/microphone")
	fmt.Println("   • Audio streamed as raw f32le PCM (32-bit float)")
	fmt.Println("   • Direct compatibility with FFmpeg tools")
	fmt.Println("   • No custom headers or metadata overhead")
	fmt.Println("   • Standard Unix named pipe for maximum compatibility")
	fmt.Println()
	fmt.Println("🎯 Perfect for:")
	fmt.Println("   • FFmpeg processing pipelines")
	fmt.Println("   • Audio format conversion workflows")
	fmt.Println("   • Live streaming to network services")
	fmt.Println("   • Custom audio analysis tools")
	fmt.Println("   • Integration with existing Unix audio tools")
	fmt.Println()
	fmt.Println("🔧 Technical Details:")
	fmt.Println("   • Format: IEEE 754 32-bit float, little-endian")
	fmt.Println("   • Sample Rate: 48000 Hz")
	fmt.Println("   • Channels: 2 (stereo)")
	fmt.Println("   • Interleaved: Left, Right, Left, Right...")
	fmt.Println("   • No headers, pure audio stream")
}