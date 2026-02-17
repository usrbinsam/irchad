package live

import (
	"sync"
)

func NewStreamRegistry() *StreamRegistry {
	return &StreamRegistry{
		mu:      sync.RWMutex{},
		streams: make(map[string]map[string]StreamHandler),
	}
}

type StreamRegistry struct {
	mu      sync.RWMutex
	streams map[string]map[string]StreamHandler // ParticipantID -> trackID -> StreamHandler
}

func (r *StreamRegistry) Add(participantID, trackID string, handler StreamHandler) {
	r.mu.Lock()
	defer r.mu.Unlock()
	if _, ok := r.streams[participantID]; !ok {
		r.streams[participantID] = make(map[string]StreamHandler)
	}
	r.streams[participantID][trackID] = handler
}

func (r *StreamRegistry) Remove(participantID, trackID string) {
	r.mu.Lock()
	defer r.mu.Unlock()
	if tracks, ok := r.streams[participantID]; ok {
		if handler, exists := tracks[trackID]; exists {
			handler.Close()
			delete(tracks, trackID)
		}

		if len(tracks) == 0 {
			delete(r.streams, participantID)
		}
	}
}
