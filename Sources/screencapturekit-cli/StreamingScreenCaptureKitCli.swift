//
//  StreamingScreenCaptureKitCli.swift
//  ScreenCaptureKit with Real-time Audio Streaming
//
//  Enhanced version with streaming capabilities
//

import ArgumentParser
import AVFoundation
import Foundation
import CoreGraphics
import ScreenCaptureKit

struct StreamingOptions: Decodable {
    let destination: URL?
    let framesPerSecond: Int
    let cropRect: CGRect?
    let showCursor: Bool
    let highlightClicks: Bool
    let screenId: CGDirectDisplayID
    let audioDeviceId: String?
    let microphoneDeviceId: String?
    let videoCodec: String?
    let enableHDR: Bool?
    let useDirectRecordingAPI: Bool?
    
    // Streaming options
    let streamingEnabled: Bool?
    let streamingProtocol: String? // "http", "websocket", "tcp", "pipe"
    let streamingURL: String?
    let streamingHost: String?
    let streamingPort: Int?
    let streamingPipePath: String?
    let audioOnly: Bool?
    let streamSystemAudio: Bool?
    let streamMicrophone: Bool?
}

struct StreamingScreenCaptureKitCLI: AsyncParsableCommand {
    static var configuration = CommandConfiguration(
        abstract: "ScreenCaptureKit wrapper with real-time streaming capabilities",
        subcommands: [List.self, Record.self, Stream.self],
        defaultSubcommand: Record.self
    )
}

extension StreamingScreenCaptureKitCLI {
    struct List: AsyncParsableCommand {
        static let configuration = CommandConfiguration(
            abstract: "List available screens and audio devices",
            subcommands: [ScreenCaptureKitCLI.List.Screens.self, ScreenCaptureKitCLI.List.AudioDevices.self, ScreenCaptureKitCLI.List.MicrophoneDevices.self]
        )
    }

    struct Stream: AsyncParsableCommand {
        static let configuration = CommandConfiguration(abstract: "Start real-time audio streaming")

        @Argument(help: "Stringified JSON object with streaming options")
        var options: String

        mutating func run() async throws {
            var keepRunning = true
            let options: StreamingOptions = try options.jsonDecoded()

            print("🌊 Starting real-time audio streaming...")
            print("Options: \(options)")

            // Check for screen recording permission
            guard CGPreflightScreenCaptureAccess() else {
                throw RecordingError("No screen capture permission")
            }

            let audioStreamer = try await AudioStreamingRecorder(options: options)
            
            print("🎵 Starting audio streaming...")
            try await audioStreamer.start()

            // Signal handling for graceful shutdown
            signal(SIGKILL, SIG_IGN)
            signal(SIGINT, SIG_IGN)
            signal(SIGTERM, SIG_IGN)
            
            let sigintSrc = DispatchSource.makeSignalSource(signal: SIGINT, queue: .main)
            sigintSrc.setEventHandler {
                print("🔌 Got SIGINT - stopping stream")
                keepRunning = false
            }
            sigintSrc.resume()
            
            let sigKillSrc = DispatchSource.makeSignalSource(signal: SIGKILL, queue: .main)
            sigKillSrc.setEventHandler {
                print("🔌 Got SIGKILL - stopping stream")
                keepRunning = false
            }
            sigKillSrc.resume()
            
            let sigTermSrc = DispatchSource.makeSignalSource(signal: SIGTERM, queue: .main)
            sigTermSrc.setEventHandler {
                print("🔌 Got SIGTERM - stopping stream")
                keepRunning = false
            }
            sigTermSrc.resume()

            // Keep streaming until interrupted
            while keepRunning {
                sleep(1)
            }

            try await audioStreamer.stop()
            print("✅ Audio streaming stopped")
        }
    }
    
    struct Record: AsyncParsableCommand {
        static let configuration = CommandConfiguration(abstract: "Start a recording with the given options (original functionality)")

        @Argument(help: "Stringified JSON object with options")
        var options: String

        mutating func run() async throws {
            // Original recording functionality
            var keepRunning = true
            let options: Options = try options.jsonDecoded()

            print("🎬 Starting file recording...")
            print(options)
            
            // Check for screen recording permission
            guard CGPreflightScreenCaptureAccess() else {
                throw RecordingError("No screen capture permission")
            }

            let screenRecorder = try await ScreenRecorder(
                url: options.destination, 
                displayID: options.screenId, 
                showCursor: options.showCursor, 
                cropRect: options.cropRect,
                audioDeviceId: options.audioDeviceId,
                microphoneDeviceId: options.microphoneDeviceId,
                enableHDR: options.enableHDR ?? false,
                useDirectRecordingAPI: options.useDirectRecordingAPI ?? false
            )
            
            print("📹 Starting screen recording of display \(options.screenId)")
            try await screenRecorder.start()

            // Signal handling (same as original)
            signal(SIGKILL, SIG_IGN)
            signal(SIGINT, SIG_IGN)
            signal(SIGTERM, SIG_IGN)
            
            let sigintSrc = DispatchSource.makeSignalSource(signal: SIGINT, queue: .main)
            sigintSrc.setEventHandler {
                print("Got SIGINT")
                keepRunning = false
            }
            sigintSrc.resume()
            
            let sigKillSrc = DispatchSource.makeSignalSource(signal: SIGKILL, queue: .main)
            sigKillSrc.setEventHandler {
                print("Got SIGKILL")
                keepRunning = false
            }
            sigKillSrc.resume()
            
            let sigTermSrc = DispatchSource.makeSignalSource(signal: SIGTERM, queue: .main)
            sigTermSrc.setEventHandler {
                print("Got SIGTERM")
                keepRunning = false
            }
            sigTermSrc.resume()

            while keepRunning {
                sleep(1)
            }

            try await screenRecorder.stop()
            print("✅ Recording complete")
        }
    }
}

// MARK: - Audio Streaming Recorder

@available(macOS 12.3, *)
class AudioStreamingRecorder {
    private let options: StreamingOptions
    private var stream: SCStream?
    private var streamOutput: StreamingOutput?
    private var audioStreamer: AudioStreamer?
    
    init(options: StreamingOptions) throws {
        self.options = options
        try setupStreamer()
    }
    
    private func setupStreamer() throws {
        guard let streamingEnabled = options.streamingEnabled, streamingEnabled else {
            throw StreamingError.invalidConfiguration
        }
        
        guard let protocolType = options.streamingProtocol else {
            throw StreamingError.invalidConfiguration
        }
        
        switch protocolType.lowercased() {
        case "http":
            guard let urlString = options.streamingURL, let url = URL(string: urlString) else {
                throw StreamingError.invalidConfiguration
            }
            audioStreamer = HTTPAudioStreamer(endpoint: url)
            
        case "websocket", "ws":
            guard let urlString = options.streamingURL, let url = URL(string: urlString) else {
                throw StreamingError.invalidConfiguration
            }
            if #available(macOS 10.15, *) {
                audioStreamer = WebSocketAudioStreamer(url: url)
            } else {
                throw StreamingError.invalidConfiguration
            }
            
        case "tcp":
            guard let host = options.streamingHost, let port = options.streamingPort else {
                throw StreamingError.invalidConfiguration
            }
            audioStreamer = TCPAudioStreamer(host: host, port: port)
            
        case "pipe", "namedpipe":
            guard let pipePath = options.streamingPipePath else {
                throw StreamingError.invalidConfiguration
            }
            audioStreamer = NamedPipeAudioStreamer(pipePath: pipePath)
            
        default:
            throw StreamingError.invalidConfiguration
        }
    }
    
    func start() async throws {
        guard let audioStreamer = audioStreamer else {
            throw StreamingError.invalidConfiguration
        }
        
        // Start the audio streamer
        try await audioStreamer.startStreaming()
        
        // Setup ScreenCaptureKit stream
        let sharableContent = try await SCShareableContent.current
        print("Displays: \(sharableContent.displays.count), Windows: \(sharableContent.windows.count), Apps: \(sharableContent.applications.count)")
        
        guard let display = sharableContent.displays.first(where: { $0.displayID == options.screenId }) else {
            throw RecordingError("No display with ID \(options.screenId) found")
        }
        
        let filter = SCContentFilter(display: display, excludingWindows: [])
        let config = SCStreamConfiguration()
        
        // Configure audio capture
        if options.streamSystemAudio == true {
            config.capturesAudio = true
            config.excludesCurrentProcessAudio = true
            print("✅ System audio capture enabled")
        }
        
        if options.streamMicrophone == true {
            if #available(macOS 15.0, *) {
                config.captureMicrophone = true
                if let microphoneDeviceId = options.microphoneDeviceId {
                    config.microphoneCaptureDeviceID = microphoneDeviceId
                }
                print("✅ Microphone capture enabled")
            } else {
                print("⚠️ Microphone capture requires macOS 15.0+")
            }
        }
        
        // Create stream
        stream = SCStream(filter: filter, configuration: config, delegate: nil)
        
        // Create streaming output
        streamOutput = StreamingOutput(audioStreamer: audioStreamer, options: options)
        
        // Add stream outputs
        if let stream = stream, let streamOutput = streamOutput {
            if options.streamSystemAudio == true {
                try stream.addStreamOutput(streamOutput, type: .audio, sampleHandlerQueue: DispatchQueue(label: "AudioStreamingQueue"))
            }
            
            if options.streamMicrophone == true {
                if #available(macOS 15.0, *) {
                    try stream.addStreamOutput(streamOutput, type: .microphone, sampleHandlerQueue: DispatchQueue(label: "MicrophoneStreamingQueue"))
                }
            }
            
            // Start capture
            try await stream.startCapture()
            streamOutput.sessionStarted = true
        }
        
        print("🌊 Audio streaming active")
    }
    
    func stop() async throws {
        streamOutput?.sessionStarted = false
        
        if let stream = stream {
            try await stream.stopCapture()
        }
        
        if let audioStreamer = audioStreamer {
            try await audioStreamer.stopStreaming()
        }
        
        stream = nil
        streamOutput = nil
    }
}

// MARK: - Streaming Output Handler

private class StreamingOutput: NSObject, SCStreamOutput {
    let audioStreamer: AudioStreamer
    let options: StreamingOptions
    var sessionStarted = false
    
    init(audioStreamer: AudioStreamer, options: StreamingOptions) {
        self.audioStreamer = audioStreamer
        self.options = options
    }
    
    func stream(_: SCStream, didOutputSampleBuffer sampleBuffer: CMSampleBuffer, of type: SCStreamOutputType) {
        guard sessionStarted else { return }
        guard sampleBuffer.isValid else { return }
        
        Task {
            do {
                switch type {
                case .audio:
                    if options.streamSystemAudio == true {
                        try await audioStreamer.streamAudioSample(sampleBuffer, from: .system)
                    }
                case .microphone:
                    if options.streamMicrophone == true {
                        try await audioStreamer.streamAudioSample(sampleBuffer, from: .microphone)
                    }
                default:
                    break
                }
            } catch {
                print("⚠️ Streaming error: \(error)")
            }
        }
    }
}