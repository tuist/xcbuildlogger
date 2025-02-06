# XCBLoggingBuildService

`XCBLoggingBuildService` is an Xcode's `XCBBuildService` executable that logs the messages sent between Xcode and the service.
This is useful to debug issues with the build service.

## Usage

1. Clone the repository: `git clone https://github.com/tuist/XCBLoggingBuildService`.
2. Build the project: `swift buil`.
3. Use the service:
  - With **xcodebuild**: `XCBBUILDSERVICE_PATH=$(pwd)/.build/debug/XCBLoggingBuildService xcodebuild ...`.
  - In **Xcode**: `env XCBBUILDSERVICE_PATH=$(pwd)/.build/debug/XCBLoggingBuildService /Applications/Xcode.app/Contents/MacOS/Xcode`.
4. The logs are persisted in `/tmp/XCBLoggingBuildService.log`. You can use `tail -f /tmp/XCBLoggingBuildService.log` in a different terminal session to see the logs in real-time.
