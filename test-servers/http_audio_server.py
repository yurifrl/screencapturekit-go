#!/usr/bin/env python3
"""
Simple HTTP server to receive audio streams from ScreenCaptureKit Go wrapper.
Usage: python3 http_audio_server.py
Then run: make run-http-stream
"""

from http.server import HTTPServer, BaseHTTPRequestHandler
import json
import time
from datetime import datetime

class AudioStreamHandler(BaseHTTPRequestHandler):
    
    def do_POST(self):
        """Handle incoming audio stream data"""
        if self.path == '/audio-stream':
            # Get content length
            content_length = int(self.headers.get('Content-Length', 0))
            
            # Read the audio data
            audio_data = self.rfile.read(content_length)
            
            # Get headers
            content_type = self.headers.get('Content-Type', 'unknown')
            audio_source = self.headers.get('X-Audio-Source', 'unknown')
            
            # Log the received data
            timestamp = datetime.now().strftime('%H:%M:%S.%f')[:-3]
            print(f"[{timestamp}] 🎵 Received {len(audio_data)} bytes of audio data")
            print(f"    Source: {audio_source}")
            print(f"    Content-Type: {content_type}")
            
            # Send successful response
            self.send_response(200)
            self.send_header('Content-Type', 'application/json')
            self.end_headers()
            
            response = {
                'status': 'success',
                'received_bytes': len(audio_data),
                'timestamp': timestamp
            }
            self.wfile.write(json.dumps(response).encode())
            
        else:
            # Handle other paths
            self.send_response(404)
            self.end_headers()
    
    def do_GET(self):
        """Handle GET requests (for testing)"""
        if self.path == '/':
            self.send_response(200)
            self.send_header('Content-Type', 'text/html')
            self.end_headers()
            
            html = """
            <!DOCTYPE html>
            <html>
            <head>
                <title>ScreenCaptureKit Audio Stream Server</title>
                <style>
                    body { font-family: Arial, sans-serif; margin: 40px; }
                    .status { color: green; font-weight: bold; }
                    .info { background: #f0f0f0; padding: 10px; margin: 10px 0; }
                </style>
            </head>
            <body>
                <h1>🎵 ScreenCaptureKit Audio Stream Server</h1>
                <div class="status">✅ Server is running and ready to receive audio streams</div>
                
                <div class="info">
                    <h3>📡 Listening for streams at:</h3>
                    <code>POST http://localhost:8080/audio-stream</code>
                </div>
                
                <div class="info">
                    <h3>🚀 To start streaming:</h3>
                    <pre>make run-http-stream</pre>
                </div>
                
                <div class="info">
                    <h3>📊 Expected data:</h3>
                    <ul>
                        <li>Audio samples in real-time</li>
                        <li>Content-Type: audio/aac (or similar)</li>
                        <li>X-Audio-Source header indicating source</li>
                    </ul>
                </div>
            </body>
            </html>
            """
            self.wfile.write(html.encode())
        else:
            self.send_response(404)
            self.end_headers()
    
    def log_message(self, format, *args):
        """Override to provide cleaner logging"""
        pass

def run_server(port=8080):
    """Run the HTTP audio stream server"""
    server_address = ('localhost', port)
    httpd = HTTPServer(server_address, AudioStreamHandler)
    
    print(f"🌊 HTTP Audio Stream Server")
    print(f"📡 Listening on http://localhost:{port}")
    print(f"🎵 Ready to receive audio streams at /audio-stream")
    print(f"🌐 Visit http://localhost:{port} for status")
    print(f"⏹️  Press Ctrl+C to stop")
    print()
    
    try:
        httpd.serve_forever()
    except KeyboardInterrupt:
        print("\n🔌 Server stopped")
        httpd.server_close()

if __name__ == '__main__':
    run_server()