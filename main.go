package main

import (
	"bytes"
	"encoding/binary"
	"encoding/json"
	"fmt"
	"io"
	"log"
	"os"
	"os/exec"
	"os/signal"
	"syscall"

	"github.com/vmihailenco/msgpack/v5"
)

// RPCPacket represents a decoded message
type RPCPacket struct {
	Channel uint64      `json:"channel"` // The RPC channel being communicated on
	Body    interface{} `json:"body"`    // The content of the message
}

// Add a method to implement the Stringer interface for debugging
func (p RPCPacket) String() string {
	return fmt.Sprintf("Channel: %d - Body: %v", p.Channel, p.Body)
}

func main() {
	xcbuildServicePath := "/Applications/Xcode.app/Contents/SharedFrameworks/XCBuild.framework/PlugIns/XCBBuildService.bundle/Contents/MacOS/XCBBuildService"

	// Open log file
	logFilePath := "/tmp/xcode_xcbbuildservice.log"
	logFile, err := os.OpenFile(logFilePath, os.O_CREATE|os.O_WRONLY|os.O_APPEND, 0644)
	if err != nil {
		log.Fatalf("Failed to open log file: %v", err)
	}
	defer logFile.Close()

	cmd := exec.Command(xcbuildServicePath)

	// Create pipes
	stdinPipe, err := cmd.StdinPipe()
	if err != nil {
		log.Fatalf("Failed to create stdin pipe: %v", err)
	}
	stdoutPipe, err := cmd.StdoutPipe()
	if err != nil {
		log.Fatalf("Failed to create stdout pipe: %v", err)
	}
	stderrPipe, err := cmd.StderrPipe()
	if err != nil {
		log.Fatalf("Failed to create stderr pipe: %v", err)
	}

	// Start the process
	if err := cmd.Start(); err != nil {
		log.Fatalf("Failed to start XCBBuildService: %v", err)
	}

	// Handle signals and forward them to the child process
	signalChan := make(chan os.Signal, 1)
	signal.Notify(signalChan, syscall.SIGINT, syscall.SIGTERM)
	go func() {
		for sig := range signalChan {
			cmd.Process.Signal(sig)
		}
	}()

	// Handle Xcode -> XCBBuildService messages (stdin)
	go func() {
		decodeAndForwardStdin(os.Stdin, stdinPipe, logFile)
	}()

	// Handle XCBBuildService -> Xcode messages (stdout & stderr)
	go forwardOutput(stdoutPipe, os.Stdout, logFile, "XCBBuildService STDOUT")
	go forwardOutput(stderrPipe, os.Stderr, logFile, "XCBBuildService STDERR")

	// Wait for XCBBuildService to exit
	if err := cmd.Wait(); err != nil {
		logToFile(logFile, fmt.Sprintf("XCBBuildService exited with error: %v", err))
	}
}

// decodeAndForwardStdin intercepts, decodes, logs, and forwards stdin messages
func decodeAndForwardStdin(input io.Reader, output io.Writer, logFile *os.File) {
	for {
		packet, rawData, err := readRPCPacket(input)
		if err == io.EOF {
			break
		}
		if err != nil {
			logToFile(logFile, fmt.Sprintf("Error reading RPCPacket: %v", err))
			continue
		}

		// Log decoded packet with full metadata
		logDecodedPacket(logFile, packet)

		// Forward raw data to XCBBuildService
		_, err = output.Write(rawData)
		if err != nil {
			logToFile(logFile, fmt.Sprintf("Error forwarding data to XCBBuildService: %v", err))
			break
		}
	}
}
func readRPCPacket(reader io.Reader) (*RPCPacket, []byte, error) {
	// Read the 12-byte header (8-byte channel + 4-byte payload size)
	header := make([]byte, 12)
	_, err := io.ReadFull(reader, header)
	if err != nil {
		return nil, nil, err
	}

	// Parse header fields
	channel := binary.LittleEndian.Uint64(header[:8])
	payloadSize := binary.LittleEndian.Uint32(header[8:12])

	// Read the payload
	payload := make([]byte, payloadSize)
	_, err = io.ReadFull(reader, payload)
	if err != nil {
		return nil, nil, err
	}

	// Decode body (MessagePack)
	var body interface{}
	err = msgpack.Unmarshal(payload, &body)
	if err != nil {
		return nil, nil, fmt.Errorf("MsgPack decode error: %v", err)
	}

	// Decode raw_message (if possible)
	var rawDecoded interface{}
	err = msgpack.Unmarshal(payload, &rawDecoded)
	if err != nil {
		rawDecoded = string(payload) // Fallback to string if decoding fails
	}

	// Return the fully decoded structure
	return &RPCPacket{
		Channel: channel,
		Body:    body,
	}, append(header, payload...), nil
}

// forwardOutput forwards output from XCBBuildService and logs it
func forwardOutput(input io.Reader, output io.Writer, logFile *os.File, prefix string) {
	tee := io.TeeReader(input, output)
	buf := new(bytes.Buffer)
	_, _ = io.Copy(buf, tee)

	// Log raw output
	logToFile(logFile, fmt.Sprintf("[%s] %s", prefix, buf.String()))
}

// logDecodedPacket writes a decoded RPCPacket with metadata
func logDecodedPacket(logFile *os.File, packet *RPCPacket) {
	packetJSON, _ := json.MarshalIndent(packet, "", "  ") // Pretty-print JSON
	logToFile(logFile, fmt.Sprintf("Decoded RPCPacket: %s", packetJSON))
}

// logToFile writes log messages safely
func logToFile(logFile *os.File, message string) {
	logFile.WriteString(message + "\n")
	logFile.Sync() // Ensure it's written to disk
}
