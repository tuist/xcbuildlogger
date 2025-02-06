import Foundation
import Dispatch
import Darwin
import MessagePack
import SwiftyJSON

struct RPCPacket {
    let channel: UInt64
    let payload: [MessagePackValue]
}

let xcbuildServicePath = "/Applications/Xcode.app/Contents/SharedFrameworks/XCBuild.framework/PlugIns/XCBBuildService.bundle/Contents/MacOS/XCBBuildService"
let logFilePath = "/tmp/xcode_xcbbuildservice.log"

@main
struct xcodebuildlogging {
    static func main() async throws {
        let process = Process()
        process.executableURL = URL(fileURLWithPath: xcbuildServicePath)

        let stdinPipe = Pipe()
        let stdoutPipe = Pipe()
        let stderrPipe = Pipe()
        process.standardInput = stdinPipe
        process.standardOutput = stdoutPipe
        process.standardError = stderrPipe
        
        do {
            try process.run()
        } catch {
            logToFile("Failed to start XCBBuildService: \(error)")
            exit(1)
        }

        DispatchQueue.global().async {
            handleStream(input: FileHandle.standardInput, output: stdinPipe.fileHandleForWriting)
        }

        DispatchQueue.global().async {
            handleStream(input: stdoutPipe.fileHandleForReading, output: FileHandle.standardOutput)
        }

        DispatchQueue.global().async {
            handleStream(input: stderrPipe.fileHandleForReading, output: FileHandle.standardError)
        }

        process.waitUntilExit()
        logToFile("XCBBuildService terminated")
    }
    
    static func logToFile(_ message: String) {
        if let handle = FileHandle(forWritingAtPath: logFilePath) {
            handle.seekToEndOfFile()
            if let data = (message + "\n").data(using: .utf8) {
                handle.write(data)
            }
            handle.closeFile()
        } else {
            try? (message + "\n").write(toFile: logFilePath, atomically: true, encoding: .utf8)
        }
    }
    
    static func readRPCPacket(from handle: FileHandle) -> (RPCPacket?, Data?)? {
        let headerSize = 12
        guard let header = try? handle.read(upToCount: headerSize), header.count == headerSize else {
            return nil
        }
        
        let channel = header.withUnsafeBytes { $0.load(as: UInt64.self) }
        let payloadSize = header.withUnsafeBytes { $0.load(fromByteOffset: 8, as: UInt32.self) }
        
        guard let payloadData = try? handle.read(upToCount: Int(payloadSize)), payloadData.count == payloadSize else {
            return nil
        }
        let unpackedData = try! MessagePackValue.unpackAll(payloadData)
        
        let packet = RPCPacket(channel: channel, payload: unpackedData)
        return (packet, header + payloadData)
    }

    static func handleStream(input: FileHandle, output: FileHandle) {
        while true {
            if let (packet, rawData) = readRPCPacket(from: input) {
                if let packet = packet {
                    let jsonData = try! JSONSerialization.data(withJSONObject: ["channel": packet.channel, "payload": packet.payload.map(\.description)])
                    logToFile(String(data: jsonData, encoding: .utf8)!)
                }
                output.write(Data(rawData!))
            } else {
                break
            }
        }
    }
}
