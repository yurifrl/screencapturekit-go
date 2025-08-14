# ScreenCaptureKit Go

A comprehensive Go wrapper for Apple's ScreenCaptureKit framework, providing high-performance screen recording and **real-time audio streaming** capabilities with HDR and microphone support on macOS.

## ✨ Features

### 🎬 **Screen Recording**
- **High-performance screen recording** using Apple's native ScreenCaptureKit framework
- **HDR recording support** on macOS 13.0+ (Ventura and later)
- **Multiple video codecs** (H.264, HEVC/H.265, ProRes)
- **Flexible crop areas** for recording specific screen regions
- **Multiple screen support** with automatic screen discovery
- **Hardware-accelerated encoding** when available
- **Cross-screen recording** support

### 🎵 **Audio Capabilities**
- **System audio capture** with automatic audio mixing
- **Microphone recording** on macOS 15.0+ (Sequoia and later)
- **Audio-only recording** mode with MP3 conversion
- **Mixed audio streams** (system + microphone)

### 🌊 **Real-time Audio Streaming** (New!)
- **HTTP Streaming** - Stream via HTTP POST requests
- **WebSocket Streaming** - Real-time bidirectional streaming
- **TCP Streaming** - Raw binary audio streaming
- **Named Pipe Streaming** - Ultra-low latency local IPC streaming
- **Multi-protocol support** - Stream to multiple destinations
- **Low-latency processing** - Optimized for real-time applications

## Requirements

- **macOS 12.3+** (ScreenCaptureKit framework requirement)
- **Go 1.21+**
- **Xcode Command Line Tools** (for Swift CLI compilation)
- **Screen Recording permissions** (System Preferences > Security & Privacy > Screen Recording)
- **BlackHole or SoundFlower** (for system audio capture)

## Installation

### Quick Start

#### Option 1: Automated Installation (Recommended)
```bash
git clone https://github.com/tfsoares/screencapturekit-go
cd screencapturekit-go
./install.sh
```

#### Option 2: Manual Installation
```bash
# Install the Go module
go get github.com/tfsoares/screencapturekit-go

# The Swift binary will be built automatically on first use
# Or build it manually:
git clone https://github.com/tfsoares/screencapturekit-go
cd screencapturekit-go
swift build --configuration=release
```

#### Option 3: Development Setup
```bash
git clone https://github.com/tfsoares/screencapturekit-go
cd screencapturekit-go
./setup.sh
```

### Prerequisites Setup

1. **Install Xcode Command Line Tools**:
   ```bash
   xcode-select --install
   ```

2. **Verify Swift Installation**:
   ```bash
   swift --version
   # Should show: Swift version 5.9+ 
   ```

3. **Install Audio Routing (for system audio capture)**:
   ```bash
   # Install BlackHole (recommended)
   brew install blackhole-2ch
   
   # Configure: System Preferences > Sound > Output > BlackHole 2ch
   ```

4. **Grant Screen Recording Permissions**:
   - System Preferences → Security & Privacy → Privacy → Screen Recording
   - Add your application or Terminal
   - Restart your application

### Troubleshooting Installation

**Binary Not Found Error**: 
```
Failed to get ScreenCaptureKit audio devices: screencapturekit binary not found
```

**Solutions**:

1. **Automatic Build** (Recommended):
   The binary builds automatically on first use. If this fails:

2. **Manual Build**:
   ```bash
   cd /path/to/screencapturekit-go
   swift build --configuration=release
   ```

3. **Global Installation**:
   ```bash
   git clone https://github.com/tfsoares/screencapturekit-go
   cd screencapturekit-go
   swift build --configuration=release
   sudo cp .build/release/screencapturekit /usr/local/bin/
   ```

4. **Verify Installation**:
   ```bash
   # Check binary exists
   which screencapturekit
   
   # Test functionality
   screencapturekit list screens
   ```

**Common Issues**:

- **"Swift not found"**: Install Xcode Command Line Tools
- **"No screen capture permission"**: Grant permissions in System Preferences
- **"No audio devices"**: Install BlackHole or SoundFlower
- **Build fails**: Check macOS version (requires 12.3+) and Swift version (5.9+)

## Quick Start

### Basic Screen Recording

```go
package main

import (
    "fmt"
    "log"
    "time"
    
    screencapturekit "github.com/tfsoares/screencapturekit-go"
)

func main() {
    // Create recorder
    recorder, err := screencapturekit.NewScreenCaptureKit()
    if err != nil {
        log.Fatal(err)
    }
    defer recorder.Cleanup()

    // Get available screens
    screens, err := screencapturekit.GetScreens()
    if err != nil {
        log.Fatal(err)
    }

    // Configure recording
    options := screencapturekit.RecordingOptions{
        FPS:        30,
        ShowCursor: true,
        ScreenID:   screens[0].ID,
        VideoCodec: "h264",
    }

    // Start recording
    err = recorder.StartRecording(options)
    if err != nil {
        log.Fatal(err)
    }

    // Record for 10 seconds
    time.Sleep(10 * time.Second)

    // Stop and save
    videoPath, err := recorder.StopRecording()
    if err != nil {
        log.Fatal(err)
    }

    fmt.Printf("Recording saved: %s\n", videoPath)
}
```

### Recording with Audio

```go
// Get audio devices
audioDevices, _ := screencapturekit.GetAudioDevices()
micDevices, _ := screencapturekit.GetMicrophoneDevices()

options := screencapturekit.RecordingOptions{
    FPS:      30,
    ScreenID: screens[0].ID,
    AudioDeviceID: &audioDevices[0].ID,        // System audio
    MicrophoneDeviceID: &micDevices[0].ID,     // Microphone (macOS 15.0+)
}
```

### 🌊 Real-time Audio Streaming

#### HTTP Streaming
```go
streamingURL := "http://localhost:8080/audio-stream"

options := screencapturekit.StreamingOptions{
    ScreenID:          screens[0].ID,
    AudioDeviceID:     &audioDevices[0].ID,
    StreamingEnabled:  true,
    StreamingProtocol: "http",
    StreamingURL:      &streamingURL,
    StreamSystemAudio: true,
    StreamMicrophone:  true,
}

recorder.StartStreaming(options)
// Audio now streaming in real-time!
```

#### WebSocket Streaming
```go
wsURL := "ws://localhost:9090/audio-stream"

options := screencapturekit.StreamingOptions{
    StreamingEnabled:  true,
    StreamingProtocol: "websocket",
    StreamingURL:      &wsURL,
    StreamSystemAudio: true,
    StreamMicrophone:  true,
}
```

#### TCP Streaming
```go
host := "localhost"
port := 9999

options := screencapturekit.StreamingOptions{
    StreamingEnabled:  true,
    StreamingProtocol: "tcp",
    StreamingHost:     &host,
    StreamingPort:     &port,
    StreamSystemAudio: true,
}
```

#### Named Pipe Streaming (Ultra-low latency)
```go
pipePath := "/tmp/screencapture_audio.fifo"

options := screencapturekit.StreamingOptions{
    StreamingEnabled:   true,
    StreamingProtocol:  "pipe",
    StreamingPipePath:  &pipePath,
    AudioOnly:          true,
    StreamSystemAudio:  true,
    StreamMicrophone:   true,
}

// Read from pipe in another process
// cat /tmp/screencapture_audio.fifo | ffplay -f f32le -ar 48000 -ac 2 -
```

#### FFmpeg-Compatible Named Pipe Streaming
```go
pipePath := "/tmp/screencapture_ffmpeg.fifo"

options := screencapturekit.StreamingOptions{
    StreamingEnabled:   true,
    StreamingProtocol:  "pipe",
    StreamingPipePath:  &pipePath,
    AudioOnly:          true,
    StreamSystemAudio:  true,
    StreamMicrophone:   true,
    FFmpegCompatible:   true, // Raw f32le stream without metadata
}

// Direct FFmpeg integration
// ffplay -f f32le -ar 48000 -channels 2 /tmp/screencapture_ffmpeg.fifo
// ffmpeg -f f32le -ar 48000 -channels 2 -i /tmp/screencapture_ffmpeg.fifo output.mp3
```

### HDR Recording

```go
if screencapturekit.SupportsHDR() {
    options := screencapturekit.RecordingOptions{
        FPS:        60,
        ScreenID:   screens[0].ID,
        VideoCodec: "hevc",  // HEVC recommended for HDR
        EnableHDR:  true,
    }
}
```

### Audio-Only Recording

```go
options := screencapturekit.RecordingOptions{
    FPS:               30,  // Ignored for audio-only
    ScreenID:          screens[0].ID,
    AudioDeviceID:     &audioDevices[0].ID,
    MicrophoneDeviceID: &micDevices[0].ID,
    AudioOnly:         true,  // Outputs MP3 file
}
```

## API Reference

### Core Types

#### `RecordingOptions`
```go
type RecordingOptions struct {
    FPS                   int          // Frames per second (1-60)
    CropArea              *CropArea    // Optional crop area
    ShowCursor            bool         // Show cursor in recording
    HighlightClicks       bool         // Highlight mouse clicks
    ScreenID              uint32       // Target screen ID
    AudioDeviceID         *string      // System audio device ID
    MicrophoneDeviceID    *string      // Microphone device ID
    VideoCodec            string       // Video codec ("h264", "hevc", "proRes422", "proRes4444")
    EnableHDR             bool         // Enable HDR recording (macOS 13.0+)
    UseDirectRecordingAPI bool         // Use direct recording API (macOS 15.0+)
    AudioOnly             bool         // Audio-only recording
}
```

#### `StreamingOptions`
```go
type StreamingOptions struct {
    // Basic screen capture settings
    FPS                int
    ScreenID           uint32
    AudioDeviceID      *string
    MicrophoneDeviceID *string
    
    // Streaming configuration
    StreamingEnabled   bool
    StreamingProtocol  string  // "http", "websocket", "tcp", "pipe"
    StreamingURL       *string // For HTTP/WebSocket
    StreamingHost      *string // For TCP
    StreamingPort      *int    // For TCP
    StreamingPipePath  *string // For Named Pipes
    
    // Named pipe options
    FFmpegCompatible   bool    // Stream raw audio without metadata (FFmpeg compatible)
    
    // Audio sources
    AudioOnly          bool    // Audio-only mode
    StreamSystemAudio  bool    // Capture system/desktop audio
    StreamMicrophone   bool    // Capture microphone (macOS 15.0+)
}
```

### Global Functions

```go
// Device Discovery
func GetScreens() ([]Screen, error)
func GetAudioDevices() ([]AudioDevice, error)
func GetMicrophoneDevices() ([]AudioDevice, error)

// Feature Detection
func SupportsHDR() bool
func SupportsHEVC() bool
func SupportsMicrophone() bool

// Create Recorder
func NewScreenCaptureKit() (*ScreenCaptureKit, error)
```

### Streaming Methods

```go
// Start real-time audio streaming
func (sck *ScreenCaptureKit) StartStreaming(options StreamingOptions) error

// Stop streaming
func (sck *ScreenCaptureKit) StopStreaming() error

// Check streaming status
func (sck *ScreenCaptureKit) IsStreaming() bool

// Get current streaming options
func (sck *ScreenCaptureKit) GetStreamingOptions() *StreamingOptions
```

## Examples

The `cmd/examples/` directory contains complete working examples:

### Recording Examples
- `basic_recording.go` - Simple screen recording
- `audio_recording.go` - Recording with system audio & microphone
- `hdr_recording.go` - HDR recording with HEVC codec
- `cropped_recording.go` - Recording specific screen areas
- `audio_only.go` - Audio-only recording with MP3 output

### Streaming Examples (New!)
- `http_streaming.go` - Real-time HTTP audio streaming
- `websocket_streaming.go` - WebSocket audio streaming
- `tcp_streaming.go` - Raw TCP audio streaming
- `namedpipe_streaming.go` - Ultra-low latency named pipe streaming
- `namedpipe_ffmpeg_streaming.go` - FFmpeg-compatible named pipe streaming

## Building and Running

```bash
# Build everything
make all

# Build all examples
make examples

# Run recording examples
make run-basic
make run-audio
make run-hdr
make run-cropped
make run-audio-only

# Run streaming examples
make run-http-stream
make run-websocket-stream
make run-tcp-stream
make run-namedpipe-stream
make run-namedpipe-ffmpeg-stream

# Development
make dev-setup
make format
make lint
```

## Audio Setup

### System Audio Capture
For capturing desktop audio, install a virtual audio driver:

```bash
# Install BlackHole (recommended)
brew install blackhole-2ch

# Configure in System Preferences:
# Sound > Output > BlackHole 2ch
```

### Test Streaming Servers
Start test servers to receive audio streams:

```bash
# HTTP Server
python3 test-servers/http_audio_server.py

# WebSocket Server (requires: pip3 install websockets)
python3 test-servers/websocket_audio_server.py

# TCP Server
python3 test-servers/tcp_audio_server.py

# Named Pipe Reader (for metadata-rich format)
python3 test-servers/namedpipe_reader.py /tmp/screencapture_audio.fifo --analyze
```

### FFmpeg Integration
For direct FFmpeg compatibility, use the FFmpeg-compatible mode:

```bash
# Direct playback (FFmpeg 7.0+ syntax)
ffplay -f f32le -ar 48000 -channels 2 /tmp/screencapture_ffmpeg.fifo

# Convert to MP3
ffmpeg -f f32le -ar 48000 -channels 2 -i /tmp/screencapture_ffmpeg.fifo output.mp3

# Stream to network
ffmpeg -f f32le -ar 48000 -channels 2 -i /tmp/screencapture_ffmpeg.fifo -f mp3 icecast://server:8000/stream

# Real-time processing
ffmpeg -f f32le -ar 48000 -channels 2 -i /tmp/screencapture_ffmpeg.fifo -af volume=0.5 -f pulse default
```

## Performance & Latency

| Protocol | Latency | Use Case |
|----------|---------|----------|
| **Named Pipes** | ~1-10ms | Ultra-low latency local processing |
| **TCP** | ~10-50ms | High-performance network streaming |
| **WebSocket** | ~20-100ms | Real-time web applications |
| **HTTP** | ~50-200ms | Simple integration, logging |

## Use Cases

### 🎮 **Live Streaming Applications**
- Stream desktop audio to Twitch/YouTube
- Discord bots with computer audio
- Podcast platforms with real-time distribution

### 🔬 **Audio Analysis & Processing**
- Real-time audio processing pipelines
- Machine learning audio classification
- Audio effects and DSP applications

### 🌐 **Remote Monitoring**
- System audio monitoring
- Remote desktop audio component
- Surveillance and security systems

### ⚡ **Local Audio Processing**
- Integration with DAWs and audio software
- High-performance audio streaming within system
- Debugging audio capture without network complexity

## Permissions

Your application needs Screen Recording permission:

1. **System Preferences > Security & Privacy > Privacy > Screen Recording**
2. Add your application or terminal
3. Restart your application

For microphone streaming (macOS 15.0+), also enable:
- **System Preferences > Security & Privacy > Privacy > Microphone**

## Platform Support

| macOS Version | Basic Recording | HDR | Microphone | Direct API | Streaming |
|---------------|----------------|-----|------------|------------|-----------|
| 12.3+ | ✅ | ❌ | ❌ | ❌ | ✅ |
| 13.0+ | ✅ | ✅ | ❌ | ❌ | ✅ |
| 14.0+ | ✅ | ✅ | ❌ | ❌ | ✅ |
| 15.0+ | ✅ | ✅ | ✅ | ✅ | ✅ |

## Documentation

- **[Streaming Guide](STREAMING.md)** - Comprehensive real-time streaming documentation
- **[Examples](cmd/examples/)** - Complete working examples
- **[Test Servers](test-servers/)** - Audio streaming test utilities

## Contributing

Contributions are welcome! Please feel free to submit a Pull Request.

## License

MIT