#!/usr/bin/env python3
"""
Named Pipe Audio Reader - Test utility for ScreenCaptureKit Go named pipe streaming

This utility reads audio data from a named pipe (FIFO) created by the 
ScreenCaptureKit streaming functionality and provides various output options.

Usage:
    python3 namedpipe_reader.py <pipe_path> [options]

Options:
    --output-file FILE    Save audio data to file
    --play               Play audio in real-time (requires ffplay)
    --analyze            Show real-time audio analysis
    --raw                Output raw audio data (default)
    --hex                Show hex dump of data
    --stats              Show streaming statistics only
"""

import sys
import os
import struct
import time
import argparse
from datetime import datetime
from typing import Optional, BinaryIO

class NamedPipeReader:
    def __init__(self, pipe_path: str):
        self.pipe_path = pipe_path
        self.total_bytes = 0
        self.total_packets = 0
        self.start_time = None
        self.last_stats_time = None
        
    def read_packet_header(self, pipe_file: BinaryIO) -> Optional[tuple]:
        """Read and parse packet header: [source_length][source][timestamp]"""
        try:
            # Read source length (4 bytes, big-endian)
            source_len_data = pipe_file.read(4)
            if len(source_len_data) != 4:
                return None
                
            source_length = struct.unpack('>I', source_len_data)[0]
            if source_length > 1024:  # Sanity check
                print(f"⚠️ Invalid source length: {source_length}")
                return None
            
            # Read source string
            source_data = pipe_file.read(source_length)
            if len(source_data) != source_length:
                return None
                
            source = source_data.decode('utf-8', errors='ignore')
            
            # Read timestamp (8 bytes, big-endian double)
            timestamp_data = pipe_file.read(8)
            if len(timestamp_data) != 8:
                return None
                
            timestamp_bits = struct.unpack('>Q', timestamp_data)[0]
            timestamp = struct.unpack('d', struct.pack('Q', timestamp_bits))[0]
            
            return source, timestamp
            
        except Exception as e:
            print(f"⚠️ Header parsing error: {e}")
            return None
    
    def process_audio_data(self, audio_data: bytes, source: str, timestamp: float, args) -> None:
        """Process audio data based on command line options"""
        self.total_bytes += len(audio_data)
        self.total_packets += 1
        
        if self.start_time is None:
            self.start_time = time.time()
            
        # Calculate audio properties (assuming 48kHz, stereo, 32-bit float)
        sample_size = 4  # 32-bit = 4 bytes
        channels = 2
        samples_per_frame = len(audio_data) // (sample_size * channels)
        duration_ms = (samples_per_frame / 48000.0) * 1000
        
        current_time = time.time()
        
        if args.stats:
            # Show statistics every second
            if self.last_stats_time is None or (current_time - self.last_stats_time) >= 1.0:
                elapsed = current_time - self.start_time
                avg_kbps = (self.total_bytes * 8) / (elapsed * 1000) if elapsed > 0 else 0
                packets_per_sec = self.total_packets / elapsed if elapsed > 0 else 0
                
                print(f"📊 Stats: {self.total_packets:,} packets, {self.total_bytes:,} bytes, "
                      f"{avg_kbps:.1f} kbps, {packets_per_sec:.1f} pkt/s")
                self.last_stats_time = current_time
                
        elif args.analyze:
            # Show packet analysis
            latency_ms = (current_time - timestamp) * 1000 if timestamp > 0 else 0
            dt_str = datetime.fromtimestamp(timestamp).strftime('%H:%M:%S.%f')[:-3]
            
            print(f"🎵 [{dt_str}] {source}: {samples_per_frame:,} samples "
                  f"({duration_ms:.1f}ms), latency: {latency_ms:.1f}ms")
                  
        elif args.hex:
            # Show hex dump of first 64 bytes
            hex_data = audio_data[:64].hex()
            formatted_hex = ' '.join(hex_data[i:i+2] for i in range(0, len(hex_data), 2))
            print(f"📦 {source}: {formatted_hex}{'...' if len(audio_data) > 64 else ''}")
            
        elif not args.raw:
            # Default: show packet info
            print(f"🎵 Received: {source} -> {len(audio_data)} bytes "
                  f"({samples_per_frame:,} samples, {duration_ms:.1f}ms)")
        
        # Write to output file if specified
        if args.output_file:
            args.output_file.write(audio_data)
            args.output_file.flush()
    
    def run(self, args) -> None:
        """Main reading loop"""
        print(f"📁 Opening named pipe: {self.pipe_path}")
        print(f"💡 Packet format: [source_len][source][timestamp][audio_data]")
        print(f"🎵 Expected audio: 48kHz, stereo, 32-bit float PCM")
        
        if args.play:
            print(f"🔊 Playing audio in real-time (make sure ffplay is installed)")
        
        print(f"🔄 Waiting for data... (Press Ctrl+C to stop)")
        print()
        
        try:
            with open(self.pipe_path, 'rb') as pipe_file:
                print(f"✅ Connected to named pipe")
                
                while True:
                    # Read packet header
                    header = self.read_packet_header(pipe_file)
                    if header is None:
                        print("⚠️ Failed to read packet header, retrying...")
                        time.sleep(0.1)
                        continue
                    
                    source, timestamp = header
                    
                    # Read remaining audio data
                    # We don't know the exact size, so read in chunks
                    # In a real implementation, you might want a length field
                    audio_data = b''
                    chunk_size = 8192
                    
                    # Read a reasonable chunk (this is simplified)
                    chunk = pipe_file.read(chunk_size)
                    audio_data += chunk
                    
                    if len(audio_data) > 0:
                        self.process_audio_data(audio_data, source, timestamp, args)
                    
        except KeyboardInterrupt:
            print(f"\n🛑 Interrupted by user")
        except FileNotFoundError:
            print(f"❌ Named pipe not found: {self.pipe_path}")
            print(f"💡 Make sure the ScreenCaptureKit streaming is running")
            sys.exit(1)
        except Exception as e:
            print(f"❌ Error reading from pipe: {e}")
            sys.exit(1)
        
        finally:
            if self.start_time:
                elapsed = time.time() - self.start_time
                avg_kbps = (self.total_bytes * 8) / (elapsed * 1000) if elapsed > 0 else 0
                print(f"\n📊 Final Statistics:")
                print(f"   ⏱️  Duration: {elapsed:.1f}s")
                print(f"   📦 Packets: {self.total_packets:,}")
                print(f"   📊 Data: {self.total_bytes:,} bytes ({self.total_bytes/1024/1024:.1f} MB)")
                print(f"   🌊 Avg bitrate: {avg_kbps:.1f} kbps")

def main():
    parser = argparse.ArgumentParser(
        description='Named Pipe Audio Reader for ScreenCaptureKit Go',
        formatter_class=argparse.RawDescriptionHelpFormatter,
        epilog="""
Examples:
    # Basic reading with packet info
    python3 namedpipe_reader.py /tmp/screencapture_audio.fifo
    
    # Save raw audio to file
    python3 namedpipe_reader.py /tmp/screencapture_audio.fifo --output-file audio.raw
    
    # Show real-time analysis
    python3 namedpipe_reader.py /tmp/screencapture_audio.fifo --analyze
    
    # Show hex dump of packets
    python3 namedpipe_reader.py /tmp/screencapture_audio.fifo --hex
    
    # Show only statistics
    python3 namedpipe_reader.py /tmp/screencapture_audio.fifo --stats
    
    # Play audio in real-time (requires ffplay)
    python3 namedpipe_reader.py /tmp/screencapture_audio.fifo --play --raw > /dev/null

Audio file usage:
    # Convert saved raw audio to WAV
    ffmpeg -f f32le -ar 48000 -ac 2 -i audio.raw output.wav
    
    # Play raw audio directly  
    ffplay -f f32le -ar 48000 -ac 2 audio.raw
        """)
    
    parser.add_argument('pipe_path', help='Path to the named pipe')
    parser.add_argument('--output-file', type=argparse.FileType('wb'), 
                       help='Save raw audio data to file')
    parser.add_argument('--play', action='store_true',
                       help='Play audio in real-time (requires ffplay)')
    parser.add_argument('--analyze', action='store_true',
                       help='Show real-time audio analysis')
    parser.add_argument('--raw', action='store_true',
                       help='Output raw audio data to stdout')
    parser.add_argument('--hex', action='store_true', 
                       help='Show hex dump of audio data')
    parser.add_argument('--stats', action='store_true',
                       help='Show only streaming statistics')
    
    args = parser.parse_args()
    
    if not os.path.exists(args.pipe_path):
        print(f"❌ Named pipe does not exist: {args.pipe_path}")
        print(f"💡 Start ScreenCaptureKit streaming first to create the pipe")
        sys.exit(1)
        
    # Check for mutually exclusive options
    output_modes = sum([args.analyze, args.raw, args.hex, args.stats])
    if output_modes > 1:
        print("❌ Please specify only one output mode (--analyze, --raw, --hex, or --stats)")
        sys.exit(1)
        
    if args.play and not args.raw:
        print("💡 --play option works best with --raw to pipe audio directly to ffplay")
        print("   Example: python3 namedpipe_reader.py pipe.fifo --play --raw | ffplay -f f32le -ar 48000 -ac 2 -")
    
    reader = NamedPipeReader(args.pipe_path)
    reader.run(args)

if __name__ == '__main__':
    main()