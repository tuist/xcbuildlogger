import ArgumentParser
import FileSystem
import Path
import Command
import Foundation

struct XcodeBuildCommand: AsyncParsableCommand {
    
    static let configuration = CommandConfiguration(commandName: "xcodebuild", abstract: "Shells out to xcodebuild forwarding the internal logs.")
    @Argument(parsing: .allUnrecognized) var xcodebuildArgs: [String] = []
    
    func run() async throws {
        let xcbuildloggerPath = try AbsolutePath(validating: Bundle.main.executablePath!)
        let commandRunner = CommandRunner()
        
        // Arguments
        var xcodebuildArgs = ["/usr/bin/xcrun", "xcodebuild"]
        xcodebuildArgs.append(contentsOf: self.xcodebuildArgs)

        // Environment
        var environment = ProcessInfo.processInfo.environment
        environment["XCBBUILDSERVICE_PATH"] = xcbuildloggerPath.pathString
 
        try await commandRunner.run(arguments: xcodebuildArgs, environment: environment).pipedStream().awaitCompletion()
    }
}
