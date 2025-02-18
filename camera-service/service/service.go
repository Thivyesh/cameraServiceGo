// Package service provides the main camera service functionality
package service

import (
	"context"
	"fmt"
	"sync"

	"github.com/yourusername/accident-prediction-api/go-services/camera-service/source"
	"github.com/yourusername/accident-prediction-api/go-services/camera-service/types"
)

// CameraService manages multiple video sources and their subscribers
type CameraService struct {
	mu          sync.RWMutex                      // Protects shared state
	sources     map[string]*source.VideoSource    // Active video sources
	subscribers map[string][]chan types.FrameData // Subscribers per source
}

// NewCameraService creates a new camera service instance
func NewCameraService() *CameraService {
	return &CameraService{
		sources:     make(map[string]*source.VideoSource),
		subscribers: make(map[string][]chan types.FrameData),
	}
}

// AddSource adds a new video source to the service
func (s *CameraService) AddSource(ctx context.Context, config types.SourceConfig) (string, error) {
	s.mu.Lock()
	defer s.mu.Unlock()

	// Generate source ID
	sourceID := fmt.Sprintf("%s_%s", config.Type, config.URI)

	// Check if source already exists
	if _, exists := s.sources[sourceID]; exists {
		return "", fmt.Errorf("source already exists: %s", sourceID)
	}

	// Create and initialize new source
	videoSource := source.NewVideoSource(config)
	if err := videoSource.Start(ctx); err != nil {
		return "", fmt.Errorf("failed to start source: %v", err)
	}

	// Store source
	s.sources[sourceID] = videoSource
	s.subscribers[sourceID] = make([]chan types.FrameData, 0)

	// Start frame distribution
	go s.distributeFrames(ctx, sourceID)

	return sourceID, nil
}

// Subscribe creates a new subscription to a source's frames
func (s *CameraService) Subscribe(sourceID string) (<-chan types.FrameData, error) {
	s.mu.Lock()
	defer s.mu.Unlock()

	source, exists := s.sources[sourceID]
	if !exists {
		return nil, fmt.Errorf("source not found: %s", sourceID)
	}

	// Create subscriber channel
	subscriber := make(chan types.FrameData, 100)
	s.subscribers[sourceID] = append(s.subscribers[sourceID], subscriber)

	return subscriber, nil
}

// distributeFrames handles frame distribution to subscribers
func (s *CameraService) distributeFrames(ctx context.Context, sourceID string) {
	source := s.sources[sourceID]
	frames := source.GetFrames()

	for {
		select {
		case <-ctx.Done():
			return
		case frame, ok := <-frames:
			if !ok {
				return
			}

			// Get current subscribers
			s.mu.RLock()
			subs := s.subscribers[sourceID]
			s.mu.RUnlock()

			// Distribute frame to all subscribers
			for _, sub := range subs {
				select {
				case sub <- frame:
					// Frame sent successfully
				default:
					// Skip if subscriber is not keeping up
				}
			}
		}
	}
}

// RemoveSource stops and removes a video source
func (s *CameraService) RemoveSource(sourceID string) error {
	s.mu.Lock()
	defer s.mu.Unlock()

	source, exists := s.sources[sourceID]
	if !exists {
		return fmt.Errorf("source not found: %s", sourceID)
	}

	// Stop the source
	source.Stop()

	// Close all subscriber channels
	for _, sub := range s.subscribers[sourceID] {
		close(sub)
	}

	// Remove source and subscribers
	delete(s.sources, sourceID)
	delete(s.subscribers, sourceID)

	return nil
}

// ListSources returns information about all active sources
func (s *CameraService) ListSources() []types.SourceInfo {
	s.mu.RLock()
	defer s.mu.RUnlock()

	sources := make([]types.SourceInfo, 0, len(s.sources))
	for _, src := range s.sources {
		sources = append(sources, src.GetInfo())
	}
	return sources
}
