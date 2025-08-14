#!/usr/bin/env python3
"""
Simple TCP server to receive raw audio streams from ScreenCaptureKit Go wrapper.
Usage: python3 tcp_audio_server.py
Then run: make run-tcp-stream
"""

import socket
import threading
import struct
from datetime import datetime

def handle_client(client_socket, client_address):
    """Handle incoming TCP audio stream from a client"""
    print(f"🔌 New TCP connection from {client_address[0]}:{client_address[1]}")
    
    try:
        while True:
            # Read data from client
            data = client_socket.recv(4096)
            
            if not data:
                break
                
            timestamp = datetime.now().strftime('%H:%M:%S.%f')[:-3]
            
            try:
                # Try to parse the packet format
                if len(data) > 1:
                    # First byte should be source name length
                    source_name_length = data[0]
                    
                    if len(data) > source_name_length + 1:
                        # Extract source name
                        source_name = data[1:1+source_name_length].decode('utf-8')
                        
                        # Remaining data is audio
                        audio_data = data[1+source_name_length:]
                        
                        print(f"[{timestamp}] 🎵 Received {len(audio_data)} bytes of audio")
                        print(f"    Source: {source_name}")
                        print(f"    Raw audio data length: {len(audio_data)}")
                    else:
                        print(f"[{timestamp}] 📦 Received {len(data)} bytes (incomplete packet)")
                else:
                    print(f"[{timestamp}] 📦 Received {len(data)} bytes (too small)")
                    
            except Exception as e:
                print(f"[{timestamp}] ⚠️ Error parsing packet: {e}")
                print(f"    Raw data length: {len(data)}")
                
    except ConnectionResetError:
        print(f"🔌 Connection from {client_address[0]}:{client_address[1]} reset by peer")
    except Exception as e:
        print(f"❌ Error handling client {client_address[0]}:{client_address[1]}: {e}")
    finally:
        client_socket.close()
        print(f"🔌 Closed connection to {client_address[0]}:{client_address[1]}")

def start_server(host='localhost', port=9999):
    """Start the TCP audio stream server"""
    server_socket = socket.socket(socket.AF_INET, socket.SOCK_STREAM)
    server_socket.setsockopt(socket.SOL_SOCKET, socket.SO_REUSEADDR, 1)
    
    try:
        server_socket.bind((host, port))
        server_socket.listen(5)
        
        print(f"🔌 TCP Audio Stream Server")
        print(f"📡 Listening on {host}:{port}")
        print(f"🎵 Ready to receive raw audio streams")
        print(f"⏹️  Press Ctrl+C to stop")
        print()
        
        while True:
            client_socket, client_address = server_socket.accept()
            
            # Handle each client in a separate thread
            client_thread = threading.Thread(
                target=handle_client,
                args=(client_socket, client_address)
            )
            client_thread.daemon = True
            client_thread.start()
            
    except KeyboardInterrupt:
        print("\n🔌 Server stopped")
    except Exception as e:
        print(f"❌ Server error: {e}")
    finally:
        server_socket.close()

if __name__ == '__main__':
    start_server()