# FFmpeg Usage Guide

## Common FFmpeg/FFplay Issues and Solutions

### Issue: "Failed to set value '2' for option 'ac': Option not found"

This error occurs with newer versions of FFmpeg (7.0+) where the option syntax has changed.

## Corrected FFmpeg Commands

### ❌ Old Syntax (doesn't work with FFmpeg 7.0+)
```bash
ffplay -f f32le -ar 48000 -ac 2 /tmp/pipe.fifo
```

### ✅ New Syntax (works with all FFmpeg versions)
```bash
ffplay -f f32le -ar 48000 -channels 2 /tmp/pipe.fifo
```

## Complete Command Reference

### Direct Playback
```bash
# New syntax (recommended)
ffplay -f f32le -ar 48000 -channels 2 /tmp/screencapture_ffmpeg.fifo

# Alternative syntax
ffplay -f f32le -sample_rate 48000 -channels 2 /tmp/screencapture_ffmpeg.fifo

# With additional options for better playback
ffplay -f f32le -ar 48000 -channels 2 -showmode 1 /tmp/screencapture_ffmpeg.fifo
```

### Format Conversion
```bash
# Convert to WAV
ffmpeg -f f32le -ar 48000 -channels 2 -i /tmp/screencapture_ffmpeg.fifo output.wav

# Convert to MP3
ffmpeg -f f32le -ar 48000 -channels 2 -i /tmp/screencapture_ffmpeg.fifo -b:a 192k output.mp3

# Convert to AAC
ffmpeg -f f32le -ar 48000 -channels 2 -i /tmp/screencapture_ffmpeg.fifo -c:a aac -b:a 128k output.m4a
```

### Live Streaming
```bash
# Stream to Icecast
ffmpeg -f f32le -ar 48000 -channels 2 -i /tmp/screencapture_ffmpeg.fifo \
       -c:a mp3 -b:a 128k -f mp3 icecast://username:password@server:8000/mountpoint

# Stream to RTMP (Twitch/YouTube)  
ffmpeg -f f32le -ar 48000 -channels 2 -i /tmp/screencapture_ffmpeg.fifo \
       -c:a aac -b:a 128k -f flv rtmp://live.twitch.tv/live/YOUR_STREAM_KEY
```

### Real-time Processing
```bash
# Apply volume adjustment
ffmpeg -f f32le -ar 48000 -channels 2 -i /tmp/screencapture_ffmpeg.fifo \
       -af volume=0.8 -f pulse default

# Apply noise reduction
ffmpeg -f f32le -ar 48000 -channels 2 -i /tmp/screencapture_ffmpeg.fifo \
       -af afftdn=nr=20 -f alsa hw:0,0

# Real-time echo effect
ffmpeg -f f32le -ar 48000 -channels 2 -i /tmp/screencapture_ffmpeg.fifo \
       -af aecho=0.8:0.88:60:0.4 -f pulse default
```

### Audio Analysis
```bash
# Show audio waveform while playing
ffplay -f f32le -ar 48000 -channels 2 -showmode 1 -vf showwaves /tmp/screencapture_ffmpeg.fifo

# Show frequency spectrum
ffplay -f f32le -ar 48000 -channels 2 -showmode 2 /tmp/screencapture_ffmpeg.fifo

# Analyze audio properties
ffprobe -f f32le -ar 48000 -channels 2 -show_streams /tmp/screencapture_ffmpeg.fifo
```

## Version Compatibility

### FFmpeg 7.0+ (Current)
- Use `channels` instead of `ac`
- Use `ar` or `sample_rate` for sample rate
- Use `f32le` for 32-bit float format

### FFmpeg 6.x and Earlier (Legacy)
- Use `ac` for audio channels (deprecated)
- Use `ar` for sample rate
- Use `f32le` for 32-bit float format

## Troubleshooting

### Common Issues and Solutions

#### Issue: Pipe blocks/hangs
**Solution**: Make sure the named pipe exists and the ScreenCaptureKit streaming is running first.

```bash
# Check if pipe exists
ls -la /tmp/screencapture_ffmpeg.fifo

# Should show: prw-r--r-- (pipe file)
```

#### Issue: "No such file or directory"
**Solution**: Start the ScreenCaptureKit streaming first, it creates the pipe automatically.

#### Issue: Audio distortion/noise
**Solution**: Verify audio device configuration and try different volume levels:

```bash
# Test with lower volume
ffplay -f f32le -ar 48000 -channels 2 -af volume=0.5 /tmp/screencapture_ffmpeg.fifo
```

#### Issue: "Invalid data found when processing input"
**Solution**: Make sure both producer and consumer agree on the format:

```bash
# Verify format with ffprobe first
ffprobe -f f32le -ar 48000 -channels 2 /tmp/screencapture_ffmpeg.fifo
```

## Testing Your Setup

### Quick Test Commands

1. **Start ScreenCaptureKit streaming** (in terminal 1):
```bash
make run-namedpipe-ffmpeg-stream
```

2. **Test playback** (in terminal 2):
```bash
# Basic playback test
ffplay -f f32le -ar 48000 -channels 2 /tmp/screencapture_ffmpeg.fifo

# With visual waveform
ffplay -f f32le -ar 48000 -channels 2 -showmode 1 /tmp/screencapture_ffmpeg.fifo

# Save 10 seconds to file for testing
ffmpeg -f f32le -ar 48000 -channels 2 -i /tmp/screencapture_ffmpeg.fifo -t 10 test_audio.wav
```

## Advanced Usage

### Chaining Multiple Effects
```bash
ffmpeg -f f32le -ar 48000 -channels 2 -i /tmp/screencapture_ffmpeg.fifo \
       -af volume=0.8,highpass=f=100,lowpass=f=8000 \
       -c:a aac -b:a 128k output.m4a
```

### Real-time Monitoring
```bash
# Monitor levels while streaming
ffplay -f f32le -ar 48000 -channels 2 -af volumedetect -autoexit /tmp/screencapture_ffmpeg.fifo
```

### Network Streaming with SRT
```bash
ffmpeg -f f32le -ar 48000 -channels 2 -i /tmp/screencapture_ffmpeg.fifo \
       -c:a aac -b:a 128k -f mpegts srt://192.168.1.100:9999
```

This guide should resolve the FFmpeg compatibility issues and provide comprehensive usage examples.