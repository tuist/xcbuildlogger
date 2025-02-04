import ArgumentParser

@main
struct XCBuildLoggerCommand: AsyncParsableCommand {
    static let configuration = CommandConfiguration(
            abstract: "An xcodebuild wrapper that logs the internal messages of Xcode's build system.",
            subcommands: [XcodeBuildCommand.self, LogCommand.self],
            defaultSubcommand: LogCommand.self)
}
