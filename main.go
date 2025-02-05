package main

import (
	"encoding/binary"
	"encoding/json"
	"fmt"
	"io"
	"log"
	"os"
	"os/exec"

	"github.com/vmihailenco/msgpack/v5"
)

// RPCPacket represents a structured decoded message.
type RPCPacket struct {
	Channel uint64      `json:"channel"`
	Body    string      `json:"body"`
	Payload interface{} `json:"payload"`
}

func main() {
	xcbuildServicePath := "/Applications/Xcode.app/Contents/SharedFrameworks/XCBuild.framework/PlugIns/XCBBuildService.bundle/Contents/MacOS/XCBBuildService"

	// Open log file (only for decoded messages)
	logFilePath := "/tmp/xcodebuildlogging_decoded.txt"
	logFile, err := os.OpenFile(logFilePath, os.O_CREATE|os.O_WRONLY|os.O_APPEND, 0644)
	if err != nil {
		log.Fatalf("Failed to open log file: %v", err)
	}
	defer logFile.Close()

	// Create a new command for XCBBuildService
	cmd := exec.Command(xcbuildServicePath)

	// Setup pipes
	stdoutPipe, _ := cmd.StdoutPipe()
	stderrPipe, _ := cmd.StderrPipe()

	// Keep raw forwarding intact
	cmd.Stdin = os.Stdin
	cmd.Stdout = os.Stdout
	cmd.Stderr = os.Stderr

	// Start XCBBuildService
	if err := cmd.Start(); err != nil {
		log.Fatalf("Failed to start XCBBuildService: %v", err)
	}

	// **Non-Intrusive Decoding for Logging**
	go logDecodedMessages(os.Stdin, "Xcode → XCBBuildService", logFile)
	go logDecodedMessages(stdoutPipe, "XCBBuildService → Xcode", logFile)
	go logDecodedMessages(stderrPipe, "XCBBuildService STDERR", logFile)

	// Wait for process completion
	if err := cmd.Wait(); err != nil {
		log.Printf("XCBBuildService exited with error: %v", err)
	}
}

// **Decodes messages for logging but DOES NOT MODIFY original forwarding**
func logDecodedMessages(input io.Reader, direction string, logFile *os.File) {
	for {
		packet, _, err := readRPCPacket(input)
		if err == io.EOF {
			break
		}
		if err != nil {
			logToFile(logFile, fmt.Sprintf("[%s] Error decoding: %v", direction, err))
			continue
		}
		logDecodedPacket(logFile, packet, direction)
	}
}

// **Reads and decodes an RPC message without modifying the stream**
func readRPCPacket(reader io.Reader) (*RPCPacket, []byte, error) {
	header := make([]byte, 12)
	_, err := io.ReadFull(reader, header)
	if err != nil {
		return nil, nil, err
	}

	// Parse header
	channel := binary.LittleEndian.Uint64(header[:8])
	payloadSize := binary.LittleEndian.Uint32(header[8:12])

	// Read the payload
	payload := make([]byte, payloadSize)
	_, err = io.ReadFull(reader, payload)
	if err != nil {
		return nil, nil, err
	}

	// Decode MessagePack
	var decodedData []interface{} // Expect an array like Swift’s logic
	err = msgpack.Unmarshal(payload, &decodedData)
	if err != nil {
		return nil, nil, fmt.Errorf("MsgPack decode error: %v", err)
	}

	// Extract details
	commandName, extractedPayload := extractCommandAndPayload(decodedData)

	// Return decoded structure **(but do NOT modify original data)**
	return &RPCPacket{
		Channel: channel,
		Body:    commandName,
		Payload: extractedPayload,
	}, append(header, payload...), nil
}

// **Extracts command name and payload from MessagePack array**
func extractCommandAndPayload(data []interface{}) (string, interface{}) {
	if len(data) == 0 {
		return "UNKNOWN_COMMAND", nil
	}
	commandName, ok := data[0].(string)
	if !ok {
		commandName = "UNKNOWN_COMMAND"
	}
	if len(data) > 1 {
		return commandName, data[1]
	}
	return commandName, nil
}

// **Logs structured decoded messages**
func logDecodedPacket(logFile *os.File, packet *RPCPacket, direction string) {
	packetJSON, _ := json.MarshalIndent(packet, "", "  ")
	logToFile(logFile, fmt.Sprintf("[%s] Decoded RPCPacket: %s", direction, packetJSON))
}

// **Writes log messages without interfering with execution**
func logToFile(logFile *os.File, message string) {
	logFile.WriteString(message + "\n")
	logFile.Sync()
}
