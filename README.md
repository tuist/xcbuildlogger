# xcbuildlogger

`xcbuildlogger` is a CLI that invokes `xcodebuild` and logs the messages sent between `xcodebuild` and `XCBBuildService`.

## Usage

You can run it with Mise:

```bash
mise run x spm:tuist/xcbuildlogger -- xcbuildlogger xcodebuild -scheme Test -workspace Test/Test.xcodeproj
```

## Development

1. Clone the repository: `git clone https://github.com/MobileNativeFoundation/XCBBuildServiceProxyKit`.
2. Run: `swift run xcbuildlogger xcodebuild -scheme Test -project Test/Test.xcodeproj`

## Credit

- [XCBBuildServiceProxyKit](https://github.com/MobileNativeFoundation/XCBBuildServiceProxyKit) for having pioneered and built the tools to understand and work with the `XCBBuildService` protocol.
