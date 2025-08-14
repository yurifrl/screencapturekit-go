#!/usr/bin/env python3
"""
Simple WebSocket server to receive audio streams from ScreenCaptureKit Go wrapper.
Requires: pip3 install websockets

Usage: python3 websocket_audio_server.py
Then run: make run-websocket-stream
"""

import asyncio
import websockets
import json
import base64
from datetime import datetime

async def handle_audio_stream(websocket, path):
    """Handle incoming WebSocket audio stream"""
    client_address = f"{websocket.remote_address[0]}:{websocket.remote_address[1]}"
    print(f"🔗 New WebSocket connection from {client_address}")
    
    try:
        async for message in websocket:
            timestamp = datetime.now().strftime('%H:%M:%S.%f')[:-3]
            
            try:
                # Try to parse as JSON
                if isinstance(message, str):
                    data = json.loads(message)
                    
                    if data.get('type') == 'audio':
                        metadata = data.get('metadata', {})
                        audio_data_b64 = data.get('data', '')
                        
                        # Decode base64 audio data
                        audio_data = base64.b64decode(audio_data_b64)
                        
                        print(f"[{timestamp}] 🎵 Received {len(audio_data)} bytes of audio")
                        print(f"    Source: {metadata.get('source', 'unknown')}")
                        print(f"    Sample Rate: {metadata.get('sampleRate', 'unknown')}")
                        print(f"    Channels: {metadata.get('channels', 'unknown')}")
                        
                        # Send acknowledgment back to client
                        response = {
                            'type': 'ack',
                            'received_bytes': len(audio_data),
                            'timestamp': timestamp
                        }
                        await websocket.send(json.dumps(response))
                    
                else:
                    # Binary data
                    print(f"[{timestamp}] 📦 Received {len(message)} bytes of binary data")
                    
                    # Send binary acknowledgment
                    ack_data = json.dumps({
                        'type': 'binary_ack',
                        'received_bytes': len(message),
                        'timestamp': timestamp
                    }).encode()
                    await websocket.send(ack_data)
                    
            except json.JSONDecodeError:
                print(f"[{timestamp}] ⚠️ Received non-JSON message: {len(message)} bytes")
            except Exception as e:
                print(f"[{timestamp}] ❌ Error processing message: {e}")
                
    except websockets.exceptions.ConnectionClosed:
        print(f"🔌 WebSocket connection from {client_address} closed")
    except Exception as e:
        print(f"❌ WebSocket error: {e}")

async def main():
    """Start the WebSocket server"""
    host = "localhost"
    port = 9090
    
    print(f"🔗 WebSocket Audio Stream Server")
    print(f"📡 Listening on ws://{host}:{port}")
    print(f"🎵 Ready to receive audio streams")
    print(f"⏹️  Press Ctrl+C to stop")
    print()
    
    # Start the WebSocket server
    server = await websockets.serve(
        handle_audio_stream, 
        host, 
        port,
        ping_interval=None,  # Disable ping for audio streaming
        max_size=10**7,      # 10MB max message size for audio data
    )
    
    try:
        await server.wait_closed()
    except KeyboardInterrupt:
        print("\n🔌 Server stopped")
        server.close()
        await server.wait_closed()

if __name__ == '__main__':
    try:
        asyncio.run(main())
    except ImportError:
        print("❌ Error: websockets library not installed")
        print("💡 Install it with: pip3 install websockets")
        exit(1)