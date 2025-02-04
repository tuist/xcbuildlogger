// swift-tools-version: 6.0
// The swift-tools-version declares the minimum version of Swift required to build this package.

import PackageDescription

let package = Package(
    name: "xcbuildlogger",
    platforms: [.macOS("15.2")],
    products: [
        .executable(name: "xcbuildlogger", targets: ["xcbuildlogger"])
    ],
    dependencies: [
        .package(url: "https://github.com/swift-server/swift-service-lifecycle", .upToNextMajor(from: "2.6.3")),
        .package(url: "https://github.com/apple/swift-log", .upToNextMajor(from: "1.6.2")),
        .package(url: "https://github.com/apple/swift-argument-parser", .upToNextMajor(from: "1.5.0")),
        .package(url: "https://github.com/tuist/path", .upToNextMajor(from: "0.3.8")),
        .package(url: "https://github.com/tuist/filesystem", .upToNextMajor(from: "0.7.2")),
        .package(url: "https://github.com/tuist/command", .upToNextMajor(from: "0.12.1"))
    ],
    targets: [
        .executableTarget(
            name: "xcbuildlogger",
            dependencies: [
                .product(name: "Command", package: "command"),
                .product(name: "Path", package: "path"),
                .product(name: "FileSystem", package: "filesystem"),
                .product(name: "ServiceLifecycle", package: "swift-service-lifecycle"),
                .product(name: "Logging", package: "swift-log"),
                .product(name: "ArgumentParser", package: "swift-argument-parser")
            ])
    ]
)
