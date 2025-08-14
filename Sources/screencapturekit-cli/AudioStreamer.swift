//
//  AudioStreamer.swift
//  ScreenCaptureKit Streaming Extension
//
//  Real-time audio streaming protocols
//

import Foundation
import AVFoundation
import CoreMedia

// MARK: - Streaming Protocols

protocol AudioStreamer {
    var isStreaming: Bool { get }
    func startStreaming() async throws
    func stopStreaming() async throws
    func streamAudioSample(_ sampleBuffer: CMSampleBuffer, from source: AudioSource) async throws
}

enum AudioSource {
    case system
    case microphone
    case mixed
}

enum StreamingError: Error {
    case notConnected
    case encodingFailed
    case networkError(Error)
    case invalidConfiguration
    case streamingNotStarted
}

// MARK: - Audio Encoder

class AudioEncoder {
    private var audioConverter: AudioConverterRef?
    private var outputFormat: AudioStreamBasicDescription
    
    init(outputFormat: AudioStreamBasicDescription) throws {
        self.outputFormat = outputFormat
        try setupEncoder()
    }
    
    deinit {
        if let converter = audioConverter {
            AudioConverterDispose(converter)
        }
    }
    
    private func setupEncoder() throws {
        // Setup audio converter for real-time encoding
        var inputFormat = AudioStreamBasicDescription()
        inputFormat.mSampleRate = 48000
        inputFormat.mFormatID = kAudioFormatLinearPCM
        inputFormat.mFormatFlags = kAudioFormatFlagIsFloat | kAudioFormatFlagsNativeEndian | kAudioFormatFlagIsPacked
        inputFormat.mChannelsPerFrame = 2
        inputFormat.mBitsPerChannel = 32
        inputFormat.mBytesPerFrame = 8
        inputFormat.mBytesPerPacket = 8
        inputFormat.mFramesPerPacket = 1
        
        let status = AudioConverterNew(&inputFormat, &outputFormat, &audioConverter)
        guard status == noErr else {
            throw StreamingError.encodingFailed
        }
    }
    
    func encode(_ sampleBuffer: CMSampleBuffer) throws -> Data {
        guard audioConverter != nil else {
            throw StreamingError.encodingFailed
        }
        
        // Extract audio data from CMSampleBuffer
        guard let blockBuffer = CMSampleBufferGetDataBuffer(sampleBuffer) else {
            throw StreamingError.encodingFailed
        }
        
        var lengthAtOffset: Int = 0
        var totalLength: Int = 0
        var dataPointer: UnsafeMutablePointer<Int8>?
        
        let status = CMBlockBufferGetDataPointer(blockBuffer, atOffset: 0, lengthAtOffsetOut: &lengthAtOffset, totalLengthOut: &totalLength, dataPointerOut: &dataPointer)
        
        guard status == noErr, let audioData = dataPointer else {
            throw StreamingError.encodingFailed
        }
        
        // Convert audio data to target format
        let inputData = Data(bytes: audioData, count: totalLength)
        
        // For simplicity, return raw data (in production, would encode to MP3/AAC)
        return inputData
    }
}

// MARK: - HTTP Streaming

class HTTPAudioStreamer: AudioStreamer {
    private let endpoint: URL
    private let sessionConfig: URLSessionConfiguration
    private var session: URLSession?
    private var encoder: AudioEncoder?
    private(set) var isStreaming = false
    
    init(endpoint: URL) {
        self.endpoint = endpoint
        self.sessionConfig = URLSessionConfiguration.default
        self.sessionConfig.timeoutIntervalForRequest = 10
        self.sessionConfig.timeoutIntervalForResource = 0 // No timeout for streaming
    }
    
    func startStreaming() async throws {
        guard !isStreaming else { return }
        
        session = URLSession(configuration: sessionConfig)
        
        // Setup audio encoder for HTTP streaming (typically AAC or MP3)
        var outputFormat = AudioStreamBasicDescription()
        outputFormat.mSampleRate = 48000
        outputFormat.mFormatID = kAudioFormatMPEG4AAC
        outputFormat.mChannelsPerFrame = 2
        
        encoder = try AudioEncoder(outputFormat: outputFormat)
        isStreaming = true
        
        print("🌊 HTTP audio streaming started to: \(endpoint)")
    }
    
    func stopStreaming() async throws {
        guard isStreaming else { return }
        
        session?.invalidateAndCancel()
        session = nil
        encoder = nil
        isStreaming = false
        
        print("🔌 HTTP audio streaming stopped")
    }
    
    func streamAudioSample(_ sampleBuffer: CMSampleBuffer, from source: AudioSource) async throws {
        guard isStreaming, let session = session, let encoder = encoder else {
            throw StreamingError.streamingNotStarted
        }
        
        do {
            let audioData = try encoder.encode(sampleBuffer)
            
            // Create HTTP request
            var request = URLRequest(url: endpoint)
            request.httpMethod = "POST"
            request.setValue("audio/aac", forHTTPHeaderField: "Content-Type")
            request.setValue("streaming", forHTTPHeaderField: "Transfer-Encoding")
            request.setValue("\(source)", forHTTPHeaderField: "X-Audio-Source")
            
            // Send audio data
            let (_, response) = try await session.upload(for: request, from: audioData)
            
            if let httpResponse = response as? HTTPURLResponse, httpResponse.statusCode >= 400 {
                throw StreamingError.networkError(NSError(domain: "HTTP", code: httpResponse.statusCode, userInfo: nil))
            }
            
        } catch {
            throw StreamingError.networkError(error)
        }
    }
}

// MARK: - WebSocket Streaming

@available(macOS 10.15, *)
class WebSocketAudioStreamer: AudioStreamer {
    private let url: URL
    private var webSocketTask: URLSessionWebSocketTask?
    private var session: URLSession?
    private var encoder: AudioEncoder?
    private(set) var isStreaming = false
    
    init(url: URL) {
        self.url = url
    }
    
    func startStreaming() async throws {
        guard !isStreaming else { return }
        
        session = URLSession(configuration: .default)
        webSocketTask = session?.webSocketTask(with: url)
        
        // Setup audio encoder
        var outputFormat = AudioStreamBasicDescription()
        outputFormat.mSampleRate = 48000
        outputFormat.mFormatID = kAudioFormatMPEG4AAC
        outputFormat.mChannelsPerFrame = 2
        
        encoder = try AudioEncoder(outputFormat: outputFormat)
        
        webSocketTask?.resume()
        isStreaming = true
        
        // Start listening for responses
        await listenForMessages()
        
        print("🔗 WebSocket audio streaming started to: \(url)")
    }
    
    func stopStreaming() async throws {
        guard isStreaming else { return }
        
        webSocketTask?.cancel(with: .normalClosure, reason: nil)
        webSocketTask = nil
        session = nil
        encoder = nil
        isStreaming = false
        
        print("🔌 WebSocket audio streaming stopped")
    }
    
    func streamAudioSample(_ sampleBuffer: CMSampleBuffer, from source: AudioSource) async throws {
        guard isStreaming, let webSocketTask = webSocketTask, let encoder = encoder else {
            throw StreamingError.streamingNotStarted
        }
        
        do {
            let audioData = try encoder.encode(sampleBuffer)
            
            // Create message with metadata
            let metadata = [
                "source": "\(source)",
                "timestamp": "\(Date().timeIntervalSince1970)",
                "sampleRate": "48000",
                "channels": "2"
            ]
            
            let message: [String: Any] = [
                "type": "audio",
                "metadata": metadata,
                "data": audioData.base64EncodedString()
            ]
            
            let jsonData = try JSONSerialization.data(withJSONObject: message)
            let wsMessage = URLSessionWebSocketTask.Message.data(jsonData)
            
            try await webSocketTask.send(wsMessage)
            
        } catch {
            throw StreamingError.networkError(error)
        }
    }
    
    private func listenForMessages() async {
        guard let webSocketTask = webSocketTask else { return }
        
        do {
            let message = try await webSocketTask.receive()
            switch message {
            case .string(let text):
                print("📨 WebSocket received: \(text)")
            case .data(let data):
                print("📨 WebSocket received data: \(data.count) bytes")
            @unknown default:
                break
            }
            
            // Continue listening
            await listenForMessages()
        } catch {
            print("⚠️ WebSocket error: \(error)")
        }
    }
}

// MARK: - TCP Streaming

class TCPAudioStreamer: AudioStreamer {
    private let host: String
    private let port: Int
    private var outputStream: OutputStream?
    private var encoder: AudioEncoder?
    private(set) var isStreaming = false
    
    init(host: String, port: Int) {
        self.host = host
        self.port = port
    }
    
    func startStreaming() async throws {
        guard !isStreaming else { return }
        
        // Setup TCP connection
        var readStream: Unmanaged<CFReadStream>?
        var writeStream: Unmanaged<CFWriteStream>?
        
        CFStreamCreatePairWithSocketToHost(nil, host as CFString, UInt32(port), &readStream, &writeStream)
        
        guard let outputStream = writeStream?.takeRetainedValue() else {
            throw StreamingError.networkError(NSError(domain: "TCP", code: -1, userInfo: [NSLocalizedDescriptionKey: "Failed to create output stream"]))
        }
        
        if let outputStream = outputStream as? OutputStream {
            self.outputStream = outputStream
            outputStream.open()
        } else {
            throw StreamingError.networkError(NSError(domain: "TCP", code: -1, userInfo: [NSLocalizedDescriptionKey: "Failed to cast output stream"]))
        }
        
        // Setup audio encoder
        var outputFormat = AudioStreamBasicDescription()
        outputFormat.mSampleRate = 48000
        outputFormat.mFormatID = kAudioFormatLinearPCM
        outputFormat.mFormatFlags = kAudioFormatFlagIsFloat | kAudioFormatFlagsNativeEndian
        outputFormat.mChannelsPerFrame = 2
        outputFormat.mBitsPerChannel = 32
        outputFormat.mBytesPerFrame = 8
        outputFormat.mBytesPerPacket = 8
        outputFormat.mFramesPerPacket = 1
        
        encoder = try AudioEncoder(outputFormat: outputFormat)
        isStreaming = true
        
        print("🔌 TCP audio streaming started to: \(host):\(port)")
    }
    
    func stopStreaming() async throws {
        guard isStreaming else { return }
        
        outputStream?.close()
        outputStream = nil
        encoder = nil
        isStreaming = false
        
        print("🔌 TCP audio streaming stopped")
    }
    
    func streamAudioSample(_ sampleBuffer: CMSampleBuffer, from source: AudioSource) async throws {
        guard isStreaming, let outputStream = outputStream, let encoder = encoder else {
            throw StreamingError.streamingNotStarted
        }
        
        do {
            let audioData = try encoder.encode(sampleBuffer)
            
            // Create packet with header
            let sourceData = "\(source)".data(using: .utf8)!
            let header = Data([UInt8(sourceData.count)]) + sourceData
            let packet = header + audioData
            
            // Send data
            let bytesWritten = packet.withUnsafeBytes { bytes in
                outputStream.write(bytes.bindMemory(to: UInt8.self).baseAddress!, maxLength: packet.count)
            }
            
            if bytesWritten < 0 {
                throw StreamingError.networkError(outputStream.streamError ?? NSError(domain: "TCP", code: -1, userInfo: nil))
            }
            
        } catch {
            throw StreamingError.networkError(error)
        }
    }
}

// MARK: - Named Pipe Streaming

class NamedPipeAudioStreamer: AudioStreamer {
    private let pipePath: String
    private var outputStream: OutputStream?
    private var encoder: AudioEncoder?
    private(set) var isStreaming = false
    private let writeQueue = DispatchQueue(label: "NamedPipeAudioStreamer.WriteQueue")
    
    init(pipePath: String) {
        self.pipePath = pipePath
    }
    
    func startStreaming() async throws {
        guard !isStreaming else { return }
        
        // Create named pipe if it doesn't exist
        let fileManager = FileManager.default
        if !fileManager.fileExists(atPath: pipePath) {
            // Use mkfifo to create the named pipe
            let result = mkfifo(pipePath, S_IRUSR | S_IWUSR | S_IRGRP | S_IWGRP)
            guard result == 0 else {
                throw StreamingError.networkError(NSError(domain: "NamedPipe", code: Int(errno), userInfo: [NSLocalizedDescriptionKey: "Failed to create named pipe: \(String(cString: strerror(errno)))"]))
            }
        }
        
        // Open the named pipe for writing
        guard let outputStream = OutputStream(toFileAtPath: pipePath, append: false) else {
            throw StreamingError.networkError(NSError(domain: "NamedPipe", code: -1, userInfo: [NSLocalizedDescriptionKey: "Failed to create output stream for named pipe"]))
        }
        
        self.outputStream = outputStream
        outputStream.open()
        
        // Wait for the stream to be ready
        var attempts = 0
        while outputStream.streamStatus != .open && attempts < 100 {
            usleep(10000) // 10ms
            attempts += 1
        }
        
        guard outputStream.streamStatus == .open else {
            throw StreamingError.networkError(NSError(domain: "NamedPipe", code: -2, userInfo: [NSLocalizedDescriptionKey: "Named pipe failed to open after timeout"]))
        }
        
        // Setup audio encoder for raw PCM output (best for named pipes)
        var outputFormat = AudioStreamBasicDescription()
        outputFormat.mSampleRate = 48000
        outputFormat.mFormatID = kAudioFormatLinearPCM
        outputFormat.mFormatFlags = kAudioFormatFlagIsFloat | kAudioFormatFlagsNativeEndian | kAudioFormatFlagIsPacked
        outputFormat.mChannelsPerFrame = 2
        outputFormat.mBitsPerChannel = 32
        outputFormat.mBytesPerFrame = 8
        outputFormat.mBytesPerPacket = 8
        outputFormat.mFramesPerPacket = 1
        
        encoder = try AudioEncoder(outputFormat: outputFormat)
        isStreaming = true
        
        print("📁 Named pipe audio streaming started: \(pipePath)")
        print("💡 Consumers can read from: \(pipePath)")
    }
    
    func stopStreaming() async throws {
        guard isStreaming else { return }
        
        // Close the output stream
        outputStream?.close()
        outputStream = nil
        encoder = nil
        isStreaming = false
        
        print("📁 Named pipe audio streaming stopped")
    }
    
    func streamAudioSample(_ sampleBuffer: CMSampleBuffer, from source: AudioSource) async throws {
        guard isStreaming, let outputStream = outputStream, let encoder = encoder else {
            throw StreamingError.streamingNotStarted
        }
        
        // Use write queue to prevent blocking the audio thread
        writeQueue.async {
            do {
                let audioData = try encoder.encode(sampleBuffer)
                
                // Create packet with header (source + timestamp + data)
                let sourceString = "\(source)"
                let sourceData = sourceString.data(using: .utf8)!
                let timestamp = Date().timeIntervalSince1970
                
                // Packet format: [4 bytes: source_length][source_data][8 bytes: timestamp][audio_data]
                var packet = Data()
                
                // Add source length (4 bytes)
                var sourceLength = UInt32(sourceData.count).bigEndian
                packet.append(Data(bytes: &sourceLength, count: 4))
                
                // Add source data
                packet.append(sourceData)
                
                // Add timestamp (8 bytes)
                var timestampBytes = timestamp.bitPattern.bigEndian
                packet.append(Data(bytes: &timestampBytes, count: 8))
                
                // Add audio data
                packet.append(audioData)
                
                // Write to named pipe
                let bytesWritten = packet.withUnsafeBytes { bytes in
                    outputStream.write(bytes.bindMemory(to: UInt8.self).baseAddress!, maxLength: packet.count)
                }
                
                if bytesWritten < 0 {
                    print("⚠️ Named pipe write error: \(outputStream.streamError?.localizedDescription ?? "unknown")")
                } else if bytesWritten < packet.count {
                    print("⚠️ Named pipe partial write: \(bytesWritten)/\(packet.count) bytes")
                }
                
            } catch {
                print("⚠️ Named pipe encoding error: \(error)")
            }
        }
    }
}

// MARK: - Multi-target Streaming

class MultiTargetAudioStreamer: AudioStreamer {
    private var streamers: [AudioStreamer] = []
    private(set) var isStreaming = false
    
    func addStreamer(_ streamer: AudioStreamer) {
        streamers.append(streamer)
    }
    
    func startStreaming() async throws {
        guard !isStreaming else { return }
        
        for streamer in streamers {
            do {
                try await streamer.startStreaming()
            } catch {
                print("⚠️ Failed to start streamer: \(error)")
            }
        }
        
        isStreaming = streamers.contains { $0.isStreaming }
        print("🌊 Multi-target streaming started with \(streamers.filter(\.isStreaming).count) active streams")
    }
    
    func stopStreaming() async throws {
        guard isStreaming else { return }
        
        for streamer in streamers {
            do {
                try await streamer.stopStreaming()
            } catch {
                print("⚠️ Failed to stop streamer: \(error)")
            }
        }
        
        isStreaming = false
        print("🔌 Multi-target streaming stopped")
    }
    
    func streamAudioSample(_ sampleBuffer: CMSampleBuffer, from source: AudioSource) async throws {
        guard isStreaming else {
            throw StreamingError.streamingNotStarted
        }
        
        // Stream to all active streamers concurrently
        await withTaskGroup(of: Void.self) { group in
            for streamer in streamers where streamer.isStreaming {
                group.addTask {
                    do {
                        try await streamer.streamAudioSample(sampleBuffer, from: source)
                    } catch {
                        print("⚠️ Streaming error: \(error)")
                    }
                }
            }
        }
    }
}