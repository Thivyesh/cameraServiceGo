// Package source handles video capture from different types of sources
package source

import (
	"context"
	"fmt"
	"log"
	"sync"
	"time"

	"github.com/Thivyesh/cameraServiceGo/types"
	"gocv.io/x/gocv"
)

// VideoSource manages video capture from a single source
type VideoSource struct {
	config    types.SourceConfig   // Source configuration
	capture   *gocv.VideoCapture   // OpenCV video capture
	isActive  bool                 // Whether source is streaming
	frames    chan types.FrameData // Channel for frame distribution
	closeOnce sync.Once            // Ensures cleanup happens only once
	mu        sync.RWMutex         // Protects shared state
}

// NewVideoSource creates a new video source instance
func NewVideoSource(config types.SourceConfig) *VideoSource {
	return &VideoSource{
		config:   config,
		frames:   make(chan types.FrameData, 100), // Buffer 100 frames
		isActive: false,
	}
}

// Start initializes video capture and begins frame streaming
func (s *VideoSource) Start(ctx context.Context) error {
	s.mu.Lock()
	defer s.mu.Unlock()

	if s.isActive {
		return fmt.Errorf("source already active")
	}

	// Initialize capture based on source type
	var err error
	switch s.config.Type {
	case "file":
		s.capture, err = gocv.OpenVideoCapture(s.config.URI)
	case "webcam":
		deviceID := 0
		fmt.Sscanf(s.config.URI, "%d", &deviceID)
		s.capture, err = gocv.OpenVideoCapture(deviceID)
	case "ip_camera":
		s.capture, err = gocv.OpenVideoCapture(s.config.URI)
	default:
		return fmt.Errorf("unsupported source type: %s", s.config.Type)
	}

	if err != nil {
		return fmt.Errorf("failed to open video source: %v", err)
	}

	s.isActive = true

	// Start frame capture in background
	go s.captureFrames(ctx)

	return nil
}

// captureFrames continuously captures frames from the source
func (s *VideoSource) captureFrames(ctx context.Context) {
	// Ensure cleanup on exit
	defer s.closeOnce.Do(func() {
		log.Printf("Cleaning up video source: %s", s.config.URI)
		close(s.frames)
		if s.capture != nil {
			s.capture.Close()
		}
		s.isActive = false
	})

	// Create reusable matrix for frame capture
	img := gocv.NewMat()
	defer img.Close()

	frameID := int64(0)
	log.Printf("Starting frame capture for source: %s", s.config.URI)

	for s.isActive {
		select {
		case <-ctx.Done():
			log.Printf("Context cancelled for source: %s", s.config.URI)
			return
		default:
			// Read next frame
			if ok := s.capture.Read(&img); !ok {
				log.Printf("Failed to read frame from source: %s", s.config.URI)
				if s.config.Type == "file" {
					s.capture.Set(gocv.VideoCapturePosFrames, 0)
					continue
				}
				return
			}

			if img.Empty() {
				log.Printf("Received empty frame from source: %s", s.config.URI)
				continue
			}

			// Encode frame to JPEG
			buf, err := gocv.IMEncode(".jpg", img)
			if err != nil {
				log.Printf("Error encoding frame: %v", err)
				continue
			}

			// Convert NativeByteBuffer to []byte
			frameBytes := buf.GetBytes()
			buf.Close()

			// Create frame data
			frame := types.FrameData{
				ID:        frameID,
				Timestamp: time.Now(),
				Data:      frameBytes,
				Source:    s.config.URI,
			}
			frameID++

			// Send frame to channel, skip if buffer full
			select {
			case s.frames <- frame:
				if frameID%30 == 0 { // Log every 30 frames
					log.Printf("Sent frame %d from source: %s (size: %d bytes)",
						frameID, s.config.URI, len(frameBytes))
				}
			default:
				log.Printf("Frame buffer full, dropping frame %d from source: %s",
					frameID, s.config.URI)
			}
		}
	}
}

// Stop gracefully stops the video capture
func (s *VideoSource) Stop() {
	s.mu.Lock()
	defer s.mu.Unlock()
	s.isActive = false
}

// GetFrames returns the channel for receiving frames
func (s *VideoSource) GetFrames() <-chan types.FrameData {
	return s.frames
}

// GetInfo returns current source information
func (s *VideoSource) GetInfo() types.SourceInfo {
	s.mu.RLock()
	defer s.mu.RUnlock()

	return types.SourceInfo{
		ID:          fmt.Sprintf("%s_%s", s.config.Type, s.config.URI),
		Type:        s.config.Type,
		URI:         s.config.URI,
		IsStreaming: s.isActive,
	}
}
