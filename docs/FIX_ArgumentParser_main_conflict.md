# Fix: ArgumentParser @main Conflict

## Problem
Binary shows help text instead of running commands when using `AsyncParsableCommand` with Swift 6.

## Root Cause
Cannot use `@main` attribute when you have a `main.swift` file.

**GitHub Issue:** https://github.com/apple/swift-argument-parser/issues/393

## Solution

**Option 1: Use @main (Recommended)**
1. Delete `main.swift`
2. Add `@main` to your command struct:
```swift
@main
struct YourCommandCLI: AsyncParsableCommand {
    // ...
}
```

**Option 2: Keep main.swift**
1. Remove `@main` from command struct
2. In `main.swift`, call directly without async context manipulation:
```swift
YourCommandCLI.main()
```

## What We Did
- Deleted `Sources/screencapturekit-cli/main.swift`
- Added `@main` to `ScreenCaptureKitCLI` struct
- Rebuilt: `swift build -c release`

## Result
✅ Binary now parses arguments correctly
✅ `make run-basic` works
