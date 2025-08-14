//
//  main.swift
//  Unified ScreenCaptureKit CLI with Streaming Support
//

import Foundation

// Choose which CLI to run based on command line arguments
if CommandLine.arguments.contains("stream") {
    // Use streaming-enabled CLI
    await StreamingScreenCaptureKitCLI.main()
} else {
    // Use original CLI for backward compatibility  
    await ScreenCaptureKitCLI.main()
}