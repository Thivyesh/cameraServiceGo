// Package types contains all shared data structures for the camera service
package types

import "time"

// FrameData represents a single frame from any video source
// @Description Video frame data structure
type FrameData struct {
	ID        int64     `json:"id"`        // Unique identifier for each frame
	Timestamp time.Time `json:"timestamp"` // When the frame was captured
	Data      []byte    `json:"data"`      // Raw frame data in JPEG format
	Source    string    `json:"source"`    // Identifier of the source
}

// SourceConfig defines the configuration for a video source
// @Description Configuration for a video source
type SourceConfig struct {
	// @Description Type of video source (webcam, file, ip_camera)
	Type string `json:"type"`
	// @Description URI or identifier for the video source
	URI string `json:"uri"`
	// For files: path to video file
	// For webcam: device ID (e.g., "0" for default camera)
	// For IP camera: RTSP/HTTP URL
}

// SourceInfo provides information about a video source
// @Description Information about a video source
type SourceInfo struct {
	// @Description Unique identifier for the source
	ID string `json:"id"` // Unique identifier for the source
	// @Description Type of video source
	Type string `json:"type"` // Source type
	// @Description URI of the video source
	URI string `json:"uri"` // Source location
	// @Description Whether the source is currently streaming
	IsStreaming bool `json:"is_streaming"` // Whether source is actively streaming
}
