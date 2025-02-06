// swift-tools-version: 6.0
// The swift-tools-version declares the minimum version of Swift required to build this package.

import PackageDescription

let package = Package(
    name: "XCBLoggingBuildService",
    platforms: [.macOS("15.2")],
    products: [
        .executable(
            name: "XCBLoggingBuildService",
            targets: ["XCBLoggingBuildService"]),
    ],
    dependencies: [
        .package(url: "https://github.com/tuist/XcodeBuildServiceKit", revision: "734cbdf376a838caaa6f91d1857f439dd1371720"),
    ],
    targets: [
        .executableTarget(
            name: "XCBLoggingBuildService",
            dependencies: [
                .product(name: "MessagePack", package: "XcodeBuildServiceKit")
            ])
    ]
)
