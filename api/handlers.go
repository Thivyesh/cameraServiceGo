package api

import (
	"encoding/json"
	"net/http"

	"github.com/Thivyesh/cameraServiceGo/service"
	"github.com/Thivyesh/cameraServiceGo/types"
	"github.com/gorilla/mux"
	"github.com/gorilla/websocket"
)

// Handler manages HTTP request handling
type Handler struct {
	service *service.CameraService
}

// NewHandler creates a new HTTP handler instance
func NewHandler(svc *service.CameraService) *Handler {
	return &Handler{service: svc}
}

// HandleAddSource handles requests to add a new video source
// @Summary Add new video source
// @Description Add a new camera or video source to the service
// @Tags sources
// @Accept json
// @Produce json
// @Param config body types.SourceConfig true "Source configuration"
// @Success 200 {object} map[string]string
// @Router /sources [post]
func (h *Handler) HandleAddSource(w http.ResponseWriter, r *http.Request) {
	// Parse request body
	var config types.SourceConfig
	if err := json.NewDecoder(r.Body).Decode(&config); err != nil {
		http.Error(w, "Invalid request body", http.StatusBadRequest)
		return
	}

	// Add source to service
	sourceID, err := h.service.AddSource(r.Context(), config)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	// Return success response
	json.NewEncoder(w).Encode(map[string]string{
		"source_id": sourceID,
		"status":    "started",
	})

}

// HandleListSources handles requests to list all sources
// @Summary List all sources
// @Description Get a list of all configured video sources
// @Tags sources
// @Produce json
// @Success 200 {array} types.SourceInfo
// @Router /sources [get]
func (h *Handler) HandleListSources(w http.ResponseWriter, r *http.Request) {
	sources := h.service.ListSources()
	json.NewEncoder(w).Encode(sources)
}

// HandleRemoveSource handles requests to remove a source
// @Summary Remove a video source
// @Description Remove a video source by its ID
// @Tags sources
// @Param id path string true "Source ID"
// @Success 200 "Source removed successfully"
// @Router /sources/{id} [delete]
func (h *Handler) HandleRemoveSource(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)
	sourceID := vars["id"]

	if err := h.service.RemoveSource(sourceID); err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	w.WriteHeader(http.StatusOK)
}

// HandleStreamFrames handles Websocket connections for frame streaming
// @Summary Stream video frames
// @Description Get real-time video frames via WebSocket
// @Tags stream
// @Param id path string true "Source ID"
// @Success 101 "Switching to WebSocket protocol"
// @Router /sources/{id}/stream [get]
func (h *Handler) HandleStreamFrames(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)
	sourceID := vars["id"]

	// Upgrade HTTP connection to Websocket
	upgrader := websocket.Upgrader{
		CheckOrigin: func(r *http.Request) bool {
			return true // Allow all origins for demo
		},
	}

	conn, err := upgrader.Upgrade(w, r, nil)
	if err != nil {
		http.Error(w, "Could not upgrade connection", http.StatusInternalServerError)
		return
	}
	defer conn.Close()

	// Subscribe to source frames
	frames, err := h.service.Subscribe(sourceID)
	if err != nil {
		conn.WriteMessage(websocket.TextMessage, []byte(err.Error()))
		return
	}

	// Stream frames to client
	for frame := range frames {
		if err := conn.WriteMessage(websocket.BinaryMessage, frame.Data); err != nil {
			break
		}
	}
}
