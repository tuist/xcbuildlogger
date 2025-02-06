// swift-tools-version: 6.0
// The swift-tools-version declares the minimum version of Swift required to build this package.

import PackageDescription

let package = Package(
    name: "xcodebuildlogging",
    platforms: [.macOS("15.2")],
    products: [
        .executable(
            name: "xcodebuildlogging",
            targets: ["xcodebuildlogging"]),
    ],
    dependencies: [
        .package(url: "https://github.com/tuist/XcodeBuildServiceKit", revision: "734cbdf376a838caaa6f91d1857f439dd1371720"),
    ],
    targets: [
        // Targets are the basic building blocks of a package, defining a module or a test suite.
        // Targets can depend on other targets in this package and products from dependencies.
        .executableTarget(
            name: "xcodebuildlogging",
            dependencies: [
                .product(name: "MessagePack", package: "XcodeBuildServiceKit")
            ])
    ]
)
