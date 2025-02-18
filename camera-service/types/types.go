// Package types contains all shared data structures for the camera service
package types

import "time"

// FrameData represents a single frame from any video source
type FrameData struct {
	ID        int64     `json:"id"`        // Unique identifier for each frame
	Timestamp time.Time `json:"timestamp"` // When the frame was captured
	Data      []byte    `json:"data"`      // Raw frame data in JPEG format
	Source    string    `json:"source"`    // Identifier of the source
}

// SourceConfig defines the configuration for a video source
type SourceConfig struct {
	Type string `json:"type"` // Type of source: "file", "webcam", "ip_camera"
	URI  string `json:"uri"`  // Location/identifier of the source
	// For files: path to video file
	// For webcam: device ID (e.g., "0" for default camera)
	// For IP camera: RTSP/HTTP URL
}

// SourceInfo provides information about a video source
type SourceInfo struct {
	ID          string `json:"id"`           // Unique identifier for the source
	Type        string `json:"type"`         // Source type
	URI         string `json:"uri"`          // Source location
	IsStreaming bool   `json:"is_streaming"` // Whether source is actively streaming
}
