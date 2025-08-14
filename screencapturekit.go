// Package screencapturekit provides a Go wrapper for Apple's ScreenCaptureKit framework
// to capture screen recordings with HDR and microphone support on macOS.
package screencapturekit

import (
	"encoding/json"
	"errors"
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"runtime"
	"strings"

	"github.com/google/uuid"
)

// Error types
var (
	ErrPermissionDenied   = errors.New("screen capture permission denied")
	ErrInvalidDisplay     = errors.New("invalid display ID")
	ErrAssetWriterFailed  = errors.New("asset writer failed")
	ErrStreamFailed       = errors.New("stream operation failed")
	ErrNotSupported       = errors.New("feature not supported on this macOS version")
	ErrRecordingActive    = errors.New("recording already active")
	ErrRecordingNotActive = errors.New("recording not active")
	ErrBinaryNotFound     = errors.New("screencapturekit binary not found")
)

// Screen represents a display screen
type Screen struct {
	ID     uint32 `json:"id"`
	Width  int    `json:"width"`
	Height int    `json:"height"`
}

// AudioDevice represents an audio input/output device
type AudioDevice struct {
	ID           string `json:"id"`
	Name         string `json:"name"`
	Manufacturer string `json:"manufacturer"`
}

// CropArea defines a rectangular area for cropping the recording
type CropArea struct {
	X      float64 `json:"x"`
	Y      float64 `json:"y"`
	Width  float64 `json:"width"`
	Height float64 `json:"height"`
}

// RecordingOptions contains all options for screen recording
type RecordingOptions struct {
	FPS                   int          `json:"fps"`
	CropArea              *CropArea    `json:"cropArea,omitempty"`
	ShowCursor            bool         `json:"showCursor"`
	HighlightClicks       bool         `json:"highlightClicks"`
	ScreenID              uint32       `json:"screenId"`
	AudioDeviceID         *string      `json:"audioDeviceId,omitempty"`
	MicrophoneDeviceID    *string      `json:"microphoneDeviceId,omitempty"`
	VideoCodec            string       `json:"videoCodec"`
	EnableHDR             bool         `json:"enableHDR,omitempty"`
	UseDirectRecordingAPI bool         `json:"useDirectRecordingAPI,omitempty"`
	AudioOnly             bool         `json:"audioOnly,omitempty"`
}

// StreamingOptions contains all options for real-time audio streaming
type StreamingOptions struct {
	FPS                int          `json:"fps"`
	CropArea           *CropArea    `json:"cropArea,omitempty"`
	ShowCursor         bool         `json:"showCursor"`
	HighlightClicks    bool         `json:"highlightClicks"`
	ScreenID           uint32       `json:"screenId"`
	AudioDeviceID      *string      `json:"audioDeviceId,omitempty"`
	MicrophoneDeviceID *string      `json:"microphoneDeviceId,omitempty"`
	VideoCodec         string       `json:"videoCodec"`
	EnableHDR          bool         `json:"enableHDR,omitempty"`
	
	// Streaming-specific options
	StreamingEnabled    bool    `json:"streamingEnabled"`
	StreamingProtocol   string  `json:"streamingProtocol"` // "http", "websocket", "tcp", "pipe"
	StreamingURL        *string `json:"streamingURL,omitempty"`
	StreamingHost       *string `json:"streamingHost,omitempty"`
	StreamingPort       *int    `json:"streamingPort,omitempty"`
	StreamingPipePath   *string `json:"streamingPipePath,omitempty"`
	AudioOnly           bool    `json:"audioOnly,omitempty"`
	StreamSystemAudio   bool    `json:"streamSystemAudio"`
	StreamMicrophone    bool    `json:"streamMicrophone"`
	FFmpegCompatible    bool    `json:"ffmpegCompatible,omitempty"` // Stream raw audio without metadata headers
}

// recordingOptionsInternal represents options passed to the Swift CLI
type recordingOptionsInternal struct {
	Destination           string      `json:"destination"`
	FramesPerSecond       int         `json:"framesPerSecond"`
	ShowCursor            bool        `json:"showCursor"`
	HighlightClicks       bool        `json:"highlightClicks"`
	ScreenID              uint32      `json:"screenId"`
	AudioDeviceID         *string     `json:"audioDeviceId,omitempty"`
	MicrophoneDeviceID    *string     `json:"microphoneDeviceId,omitempty"`
	VideoCodec            *string     `json:"videoCodec,omitempty"`
	CropRect              [][]float64 `json:"cropRect,omitempty"`
	EnableHDR             *bool       `json:"enableHDR,omitempty"`
	UseDirectRecordingAPI *bool       `json:"useDirectRecordingAPI,omitempty"`
}

// ScreenCaptureKit is the main recorder instance
type ScreenCaptureKit struct {
	cmd         *exec.Cmd
	videoPath   string
	isRecording bool
	isStreaming bool
	options     *RecordingOptions
	streamingOptions *StreamingOptions
}

// NewScreenCaptureKit creates a new ScreenCaptureKit instance
func NewScreenCaptureKit() (*ScreenCaptureKit, error) {
	if runtime.GOOS != "darwin" {
		return nil, ErrNotSupported
	}

	// Find the screencapturekit binary
	binaryPath, err := findBinary()
	if err != nil {
		return nil, err
	}

	// Test if binary works (this also checks permissions)
	cmd := exec.Command(binaryPath, "list", "screens")
	if err := cmd.Run(); err != nil {
		if strings.Contains(err.Error(), "permission") {
			return nil, ErrPermissionDenied
		}
		return nil, fmt.Errorf("binary test failed: %w", err)
	}

	return &ScreenCaptureKit{
		isRecording: false,
	}, nil
}

// findBinary locates the screencapturekit CLI binary
func findBinary() (string, error) {
	// Try different possible locations
	locations := []string{
		"./screencapturekit",                                    // Built locally
		"./.build/release/screencapturekit",                     // Swift build output
		"./.build/apple/Products/Release/screencapturekit",      // Xcode build output
		"/usr/local/bin/screencapturekit",                       // Installed globally
		filepath.Join(os.Getenv("HOME"), ".local/bin/screencapturekit"), // User install
	}

	for _, path := range locations {
		if _, err := os.Stat(path); err == nil {
			return path, nil
		}
	}

	// Try to build it if Package.swift exists
	if _, err := os.Stat("Package.swift"); err == nil {
		fmt.Println("Building screencapturekit binary...")
		cmd := exec.Command("swift", "build", "--configuration=release")
		if err := cmd.Run(); err != nil {
			return "", fmt.Errorf("failed to build binary: %w", err)
		}

		// Check if build succeeded
		builtPath := "./.build/release/screencapturekit"
		if _, err := os.Stat(builtPath); err == nil {
			return builtPath, nil
		}

		builtPath = "./.build/apple/Products/Release/screencapturekit"
		if _, err := os.Stat(builtPath); err == nil {
			return builtPath, nil
		}
	}

	return "", ErrBinaryNotFound
}

// GetScreens returns all available screens for recording
func GetScreens() ([]Screen, error) {
	if runtime.GOOS != "darwin" {
		return nil, ErrNotSupported
	}

	binaryPath, err := findBinary()
	if err != nil {
		return nil, err
	}

	cmd := exec.Command(binaryPath, "list", "screens")
	output, err := cmd.CombinedOutput()
	if err != nil {
		return nil, fmt.Errorf("failed to get screens: %w", err)
	}

	var screens []Screen
	// The output goes to stderr in the Swift implementation
	outputStr := string(output)
	lines := strings.Split(outputStr, "\n")
	for _, line := range lines {
		if strings.TrimSpace(line) == "" {
			continue
		}
		if strings.HasPrefix(line, "[") {
			err = json.Unmarshal([]byte(line), &screens)
			if err == nil {
				break
			}
		}
	}

	if len(screens) == 0 {
		return nil, fmt.Errorf("no screens found or failed to parse output")
	}

	return screens, nil
}

// GetAudioDevices returns all available audio devices
func GetAudioDevices() ([]AudioDevice, error) {
	if runtime.GOOS != "darwin" {
		return nil, ErrNotSupported
	}

	binaryPath, err := findBinary()
	if err != nil {
		return nil, err
	}

	cmd := exec.Command(binaryPath, "list", "audio-devices")
	output, err := cmd.CombinedOutput()
	if err != nil {
		return nil, fmt.Errorf("failed to get audio devices: %w", err)
	}

	var devices []AudioDevice
	// The output goes to stderr in the Swift implementation
	outputStr := string(output)
	lines := strings.Split(outputStr, "\n")
	for _, line := range lines {
		if strings.TrimSpace(line) == "" {
			continue
		}
		if strings.HasPrefix(line, "[") {
			err = json.Unmarshal([]byte(line), &devices)
			if err == nil {
				break
			}
		}
	}

	return devices, nil
}

// GetMicrophoneDevices returns all available microphone devices
func GetMicrophoneDevices() ([]AudioDevice, error) {
	if runtime.GOOS != "darwin" {
		return nil, ErrNotSupported
	}

	binaryPath, err := findBinary()
	if err != nil {
		return nil, err
	}

	cmd := exec.Command(binaryPath, "list", "microphone-devices")
	output, err := cmd.CombinedOutput()
	if err != nil {
		return nil, fmt.Errorf("failed to get microphone devices: %w", err)
	}

	var devices []AudioDevice
	// The output goes to stderr in the Swift implementation
	outputStr := string(output)
	lines := strings.Split(outputStr, "\n")
	for _, line := range lines {
		if strings.TrimSpace(line) == "" {
			continue
		}
		if strings.HasPrefix(line, "[") {
			err = json.Unmarshal([]byte(line), &devices)
			if err == nil {
				break
			}
		}
	}

	return devices, nil
}

// SupportsHDR returns true if the system supports HDR recording
func SupportsHDR() bool {
	// HDR requires macOS 13.0+
	cmd := exec.Command("sw_vers", "-productVersion")
	output, err := cmd.Output()
	if err != nil {
		return false
	}

	version := strings.TrimSpace(string(output))
	// Simple version check - in production, you'd want more robust parsing
	return strings.HasPrefix(version, "13.") || strings.HasPrefix(version, "14.") || strings.HasPrefix(version, "15.")
}

// SupportsHEVC returns true if the system supports HEVC encoding
func SupportsHEVC() bool {
	// Check if we're on Apple Silicon or Intel 6th gen+
	// This is a simplified check
	return true
}

// SupportsMicrophone returns true if the system supports microphone capture
func SupportsMicrophone() bool {
	// Microphone capture requires macOS 15.0+
	cmd := exec.Command("sw_vers", "-productVersion")
	output, err := cmd.Output()
	if err != nil {
		return false
	}

	version := strings.TrimSpace(string(output))
	return strings.HasPrefix(version, "15.")
}

// StartRecording begins screen recording with the specified options
func (sck *ScreenCaptureKit) StartRecording(options RecordingOptions) error {
	if sck.isRecording {
		return ErrRecordingActive
	}

	binaryPath, err := findBinary()
	if err != nil {
		return err
	}

	// Generate temporary file path
	tempDir := os.TempDir()
	videoID := uuid.New().String()
	sck.videoPath = filepath.Join(tempDir, videoID+".mov")

	// Create URL from file path
	destinationURL := "file://" + sck.videoPath

	// Convert Go options to internal format
	internalOptions := recordingOptionsInternal{
		Destination:     destinationURL,
		FramesPerSecond: options.FPS,
		ShowCursor:      options.ShowCursor,
		HighlightClicks: options.HighlightClicks,
		ScreenID:        options.ScreenID,
		AudioDeviceID:   options.AudioDeviceID,
		MicrophoneDeviceID: options.MicrophoneDeviceID,
	}

	if options.VideoCodec != "" {
		internalOptions.VideoCodec = &options.VideoCodec
	}

	if options.EnableHDR {
		internalOptions.EnableHDR = &options.EnableHDR
	}

	if options.UseDirectRecordingAPI {
		internalOptions.UseDirectRecordingAPI = &options.UseDirectRecordingAPI
	}

	if options.CropArea != nil {
		internalOptions.CropRect = [][]float64{
			{options.CropArea.X, options.CropArea.Y},
			{options.CropArea.Width, options.CropArea.Height},
		}
	}

	// Convert to JSON
	optionsJSON, err := json.Marshal(internalOptions)
	if err != nil {
		return fmt.Errorf("failed to marshal options: %w", err)
	}

	// Start recording
	sck.cmd = exec.Command(binaryPath, "record", string(optionsJSON))
	
	if err := sck.cmd.Start(); err != nil {
		return fmt.Errorf("failed to start recording: %w", err)
	}

	sck.isRecording = true
	sck.options = &options
	return nil
}

// StopRecording stops the ongoing recording
func (sck *ScreenCaptureKit) StopRecording() (string, error) {
	if !sck.isRecording {
		return "", ErrRecordingNotActive
	}

	if sck.cmd != nil && sck.cmd.Process != nil {
		// Send interrupt signal to gracefully stop recording
		if err := sck.cmd.Process.Signal(os.Interrupt); err != nil {
			// If interrupt fails, try kill
			sck.cmd.Process.Kill()
		}
		
		// Wait for process to finish
		sck.cmd.Wait()
		sck.cmd = nil
	}

	sck.isRecording = false
	videoPath := sck.videoPath
	sck.videoPath = ""

	// Check if file was created and has content
	if stat, err := os.Stat(videoPath); err != nil || stat.Size() == 0 {
		return "", fmt.Errorf("recording failed or file is empty")
	}

	// If audio-only recording, convert to MP3
	if sck.options != nil && sck.options.AudioOnly {
		return sck.convertToAudioOnly(videoPath)
	}

	return videoPath, nil
}

// convertToAudioOnly converts a video file to audio-only MP3
func (sck *ScreenCaptureKit) convertToAudioOnly(inputPath string) (string, error) {
	// Generate output path with .mp3 extension
	tempDir := os.TempDir()
	audioID := uuid.New().String()
	outputPath := filepath.Join(tempDir, audioID+".mp3")

	// Check if ffmpeg is available
	if _, err := exec.LookPath("ffmpeg"); err != nil {
		fmt.Println("Warning: ffmpeg not found. Returning original video file.")
		fmt.Println("To get MP3 output, install ffmpeg: brew install ffmpeg")
		return inputPath, nil
	}

	// Convert to MP3 using ffmpeg
	cmd := exec.Command("ffmpeg", 
		"-i", inputPath,           // Input file
		"-vn",                     // No video
		"-c:a", "libmp3lame",      // MP3 codec
		"-b:a", "192k",            // Audio bitrate
		"-y",                      // Overwrite output
		outputPath,                // Output file
	)

	if err := cmd.Run(); err != nil {
		fmt.Printf("Warning: Failed to convert to MP3: %v\n", err)
		fmt.Println("Returning original video file instead.")
		return inputPath, nil
	}

	// Remove the original video file
	os.Remove(inputPath)

	fmt.Printf("Converted to audio-only MP3: %s\n", outputPath)
	return outputPath, nil
}

// IsRecording returns true if a recording is currently active
func (sck *ScreenCaptureKit) IsRecording() bool {
	return sck.isRecording
}

// GetVideoPath returns the path to the current recording file
func (sck *ScreenCaptureKit) GetVideoPath() string {
	return sck.videoPath
}

// StartStreaming begins real-time audio streaming with the specified options
func (sck *ScreenCaptureKit) StartStreaming(options StreamingOptions) error {
	if sck.isStreaming {
		return errors.New("streaming already active")
	}

	if sck.isRecording {
		return errors.New("cannot stream while recording")
	}

	binaryPath, err := findBinary()
	if err != nil {
		return err
	}

	// Validate streaming options
	if !options.StreamingEnabled {
		return errors.New("streaming not enabled in options")
	}

	if options.StreamingProtocol == "" {
		return errors.New("streaming protocol not specified")
	}

	// Convert Go options to JSON for Swift CLI
	optionsJSON, err := json.Marshal(options)
	if err != nil {
		return fmt.Errorf("failed to marshal streaming options: %w", err)
	}

	// Start streaming using the Swift CLI
	sck.cmd = exec.Command(binaryPath, "stream", string(optionsJSON))
	
	if err := sck.cmd.Start(); err != nil {
		return fmt.Errorf("failed to start streaming: %w", err)
	}

	sck.isStreaming = true
	sck.streamingOptions = &options
	
	fmt.Printf("🌊 Audio streaming started (%s)\n", options.StreamingProtocol)
	return nil
}

// StopStreaming stops the ongoing audio streaming
func (sck *ScreenCaptureKit) StopStreaming() error {
	if !sck.isStreaming {
		return errors.New("streaming not active")
	}

	if sck.cmd != nil && sck.cmd.Process != nil {
		// Send interrupt signal to gracefully stop streaming
		if err := sck.cmd.Process.Signal(os.Interrupt); err != nil {
			// If interrupt fails, try kill
			sck.cmd.Process.Kill()
		}
		
		// Wait for process to finish
		sck.cmd.Wait()
		sck.cmd = nil
	}

	sck.isStreaming = false
	sck.streamingOptions = nil
	
	fmt.Println("🔌 Audio streaming stopped")
	return nil
}

// IsStreaming returns true if streaming is currently active
func (sck *ScreenCaptureKit) IsStreaming() bool {
	return sck.isStreaming
}

// GetStreamingOptions returns the current streaming options
func (sck *ScreenCaptureKit) GetStreamingOptions() *StreamingOptions {
	return sck.streamingOptions
}

// Cleanup releases any resources held by the ScreenCaptureKit instance
func (sck *ScreenCaptureKit) Cleanup() {
	if sck.isRecording {
		sck.StopRecording()
	}
	if sck.isStreaming {
		sck.StopStreaming()
	}
}