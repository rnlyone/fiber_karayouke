package websocket

import (
	"encoding/json"
	"sort"
	"sync"
	"time"
)

// Video represents a video/song in the playlist
type Video struct {
	ID         string  `json:"id"`
	Title      string  `json:"title"`
	Artist     string  `json:"artist"`
	Song       string  `json:"song"`
	CoverURL   string  `json:"coverUrl"`
	Duration   string  `json:"duration"`
	SingerName string  `json:"singerName"`
	CreatedAt  string  `json:"createdAt"`
	PlayedAt   *string `json:"playedAt"`
}

// RoomMeta contains room metadata
type RoomMeta struct {
	Name      string `json:"name,omitempty"`
	CreatedAt string `json:"createdAt,omitempty"`
}

// RoomSettings contains room settings
type RoomSettings struct {
	OrderByFairness bool `json:"orderByFairness"`
}

// RoomState represents the state of a karaoke room
type RoomState struct {
	Playlist []Video      `json:"playlist"`
	Settings RoomSettings `json:"settings"`
	Meta     *RoomMeta    `json:"meta"`
}

// Room represents a karaoke room with WebSocket connections
type Room struct {
	Key         string
	State       RoomState
	Connections map[*Connection]bool
	mu          sync.RWMutex
	lastAccess  time.Time
}

// RoomManager manages all karaoke rooms
type RoomManager struct {
	rooms map[string]*Room
	mu    sync.RWMutex
}

// NewRoomManager creates a new room manager
func NewRoomManager() *RoomManager {
	rm := &RoomManager{
		rooms: make(map[string]*Room),
	}
	// Start cleanup goroutine for expired rooms (30 days)
	go rm.cleanupExpiredRooms()
	return rm
}

// GetOrCreateRoom gets an existing room or creates a new one
func (rm *RoomManager) GetOrCreateRoom(roomKey string) *Room {
	rm.mu.Lock()
	defer rm.mu.Unlock()

	if room, exists := rm.rooms[roomKey]; exists {
		room.lastAccess = time.Now()
		return room
	}

	room := &Room{
		Key: roomKey,
		State: RoomState{
			Playlist: []Video{},
			Settings: RoomSettings{OrderByFairness: true},
			Meta:     nil,
		},
		Connections: make(map[*Connection]bool),
		lastAccess:  time.Now(),
	}
	rm.rooms[roomKey] = room
	return room
}

// cleanupExpiredRooms removes rooms that haven't been accessed in 30 days
func (rm *RoomManager) cleanupExpiredRooms() {
	ticker := time.NewTicker(24 * time.Hour)
	for range ticker.C {
		rm.mu.Lock()
		threshold := time.Now().Add(-30 * 24 * time.Hour)
		for key, room := range rm.rooms {
			if room.lastAccess.Before(threshold) && len(room.Connections) == 0 {
				delete(rm.rooms, key)
			}
		}
		rm.mu.Unlock()
	}
}

// AddConnection adds a connection to the room
func (r *Room) AddConnection(conn *Connection) {
	r.mu.Lock()
	defer r.mu.Unlock()
	r.Connections[conn] = true
	r.lastAccess = time.Now()
}

// RemoveConnection removes a connection from the room
func (r *Room) RemoveConnection(conn *Connection) {
	r.mu.Lock()
	defer r.mu.Unlock()
	delete(r.Connections, conn)
}

// Broadcast sends a message to all connections in the room
func (r *Room) Broadcast(message []byte) {
	r.mu.RLock()
	defer r.mu.RUnlock()
	for conn := range r.Connections {
		conn.Send(message)
	}
}

// SendState sends the current state to a specific connection
func (r *Room) SendState(conn *Connection) {
	r.mu.RLock()
	defer r.mu.RUnlock()
	msg := map[string]interface{}{
		"type":  "state",
		"state": r.State,
	}
	data, _ := json.Marshal(msg)
	conn.Send(data)
}

// BroadcastState sends the current state to all connections
func (r *Room) BroadcastState() {
	r.mu.RLock()
	msg := map[string]interface{}{
		"type":  "state",
		"state": r.State,
	}
	data, _ := json.Marshal(msg)
	r.mu.RUnlock()
	r.Broadcast(data)
}

// HandleMessage processes an incoming WebSocket message
func (r *Room) HandleMessage(conn *Connection, message []byte) {
	var payload map[string]interface{}
	if err := json.Unmarshal(message, &payload); err != nil {
		return
	}

	msgType, ok := payload["type"].(string)
	if !ok {
		return
	}

	r.lastAccess = time.Now()

	switch msgType {
	case "getState":
		r.SendState(conn)
		return

	case "setRoomMeta":
		r.mu.Lock()
		name, _ := payload["name"].(string)
		createdAt, _ := payload["createdAt"].(string)
		if r.State.Meta == nil {
			r.State.Meta = &RoomMeta{}
		}
		r.State.Meta.Name = name
		r.State.Meta.CreatedAt = createdAt
		r.mu.Unlock()

	case "updateRoom":
		r.mu.Lock()
		if patch, ok := payload["patch"].(map[string]interface{}); ok {
			if meta, ok := patch["meta"].(map[string]interface{}); ok {
				if r.State.Meta == nil {
					r.State.Meta = &RoomMeta{}
				}
				if name, ok := meta["name"].(string); ok {
					r.State.Meta.Name = name
				}
				if createdAt, ok := meta["createdAt"].(string); ok {
					r.State.Meta.CreatedAt = createdAt
				}
			}
		}
		r.mu.Unlock()

	case "add-video":
		r.mu.Lock()
		id, _ := payload["id"].(string)

		// Check if already queued
		alreadyQueued := false
		for _, v := range r.State.Playlist {
			if v.ID == id && v.PlayedAt == nil {
				alreadyQueued = true
				break
			}
		}

		if !alreadyQueued {
			title, _ := payload["title"].(string)
			artist, _ := payload["artist"].(string)
			if artist == "" {
				artist = "Unknown"
			}
			song, _ := payload["song"].(string)
			if song == "" {
				song = title
			}
			coverURL, _ := payload["coverUrl"].(string)
			duration, _ := payload["duration"].(string)
			singerName, _ := payload["singerName"].(string)
			if singerName == "" {
				singerName = "Guest"
			}

			newVideo := Video{
				ID:         id,
				Title:      title,
				Artist:     artist,
				Song:       song,
				CoverURL:   coverURL,
				Duration:   duration,
				SingerName: singerName,
				CreatedAt:  time.Now().UTC().Format(time.RFC3339),
				PlayedAt:   nil,
			}

			insertPos, _ := payload["insertPosition"].(string)
			currentIndex := -1
			for i, v := range r.State.Playlist {
				if v.PlayedAt == nil {
					currentIndex = i
					break
				}
			}

			// Only apply special positioning for "next" - anything else goes to end
			if insertPos == "next" && currentIndex != -1 {
				offset := 1
				targetIndex := currentIndex + offset
				if targetIndex > len(r.State.Playlist) {
					targetIndex = len(r.State.Playlist)
				}
				// Insert at position (without fairness reordering to maintain exact positioning)
				r.State.Playlist = append(r.State.Playlist[:targetIndex], append([]Video{newVideo}, r.State.Playlist[targetIndex:]...)...)
			} else {
				// Add to end of playlist (without fairness reordering to maintain insertion order)
				r.State.Playlist = append(r.State.Playlist, newVideo)
			}
		}
		r.mu.Unlock()

	case "reorder-upcoming":
		r.mu.Lock()
		idsRaw, ok := payload["ids"].([]interface{})
		if ok {
			ids := make([]string, len(idsRaw))
			for i, id := range idsRaw {
				ids[i], _ = id.(string)
			}

			played := []Video{}
			unplayed := []Video{}
			for _, v := range r.State.Playlist {
				if v.PlayedAt != nil {
					played = append(played, v)
				} else {
					unplayed = append(unplayed, v)
				}
			}

			if len(unplayed) > 0 {
				lookup := make(map[string]Video)
				for _, v := range unplayed {
					lookup[v.ID] = v
				}

				ordered := []Video{}
				for _, id := range ids {
					if v, exists := lookup[id]; exists {
						ordered = append(ordered, v)
						delete(lookup, id)
					}
				}
				for _, v := range lookup {
					ordered = append(ordered, v)
				}

				r.State.Playlist = append(played, ordered...)
			}
		}
		r.mu.Unlock()

	case "remove-video":
		r.mu.Lock()
		id, _ := payload["id"].(string)
		for i, v := range r.State.Playlist {
			if v.ID == id {
				r.State.Playlist = append(r.State.Playlist[:i], r.State.Playlist[i+1:]...)
				break
			}
		}
		r.mu.Unlock()

	case "mark-as-played":
		r.mu.Lock()
		id, _ := payload["id"].(string)
		changed := false
		for i, v := range r.State.Playlist {
			if v.ID == id && v.PlayedAt == nil {
				now := time.Now().UTC().Format(time.RFC3339)
				r.State.Playlist[i].PlayedAt = &now
				changed = true
				break
			}
		}
		if !changed {
			// Song was already played (duplicate request) - skip broadcast
			r.mu.Unlock()
			return
		}
		// Clean up old played songs to prevent memory accumulation (keep last 10 played)
		r.cleanupOldPlayedSongs()
		r.mu.Unlock()

	case "horn":
		msg := map[string]string{"type": "horn"}
		data, _ := json.Marshal(msg)
		r.Broadcast(data)
		return

	case "emoji":
		emoji, _ := payload["emoji"].(string)
		if emoji != "" {
			msg := map[string]interface{}{"type": "emoji", "emoji": emoji}
			data, _ := json.Marshal(msg)
			r.Broadcast(data)
		}
		return

	default:
		return
	}

	r.BroadcastState()
}

// cleanupOldPlayedSongs removes old played songs, keeping only the last 10 played
// This prevents memory accumulation on TV browsers and constrained devices
func (r *Room) cleanupOldPlayedSongs() {
	played := []Video{}
	unplayed := []Video{}

	for _, v := range r.State.Playlist {
		if v.PlayedAt != nil {
			played = append(played, v)
		} else {
			unplayed = append(unplayed, v)
		}
	}

	// Keep only the last 10 played songs
	if len(played) > 10 {
		played = played[len(played)-10:]
	}

	// Reconstruct playlist with limited played songs + all unplayed
	r.State.Playlist = append(played, unplayed...)
}

// reorderPlaylistByFairness reorders the playlist to be fair across singers
func (r *Room) reorderPlaylistByFairness() {
	if len(r.State.Playlist) == 0 {
		return
	}

	// Group by singer
	grouped := make(map[string][]Video)
	for _, v := range r.State.Playlist {
		grouped[v.SingerName] = append(grouped[v.SingerName], v)
	}

	// Sort each group by createdAt
	for singer := range grouped {
		sort.Slice(grouped[singer], func(i, j int) bool {
			return grouped[singer][i].CreatedAt < grouped[singer][j].CreatedAt
		})
	}

	// Round-robin merge
	result := []Video{}
	for len(result) < len(r.State.Playlist) {
		round := []Video{}
		for singer := range grouped {
			if len(grouped[singer]) > 0 {
				round = append(round, grouped[singer][0])
				grouped[singer] = grouped[singer][1:]
			}
		}
		// Sort round by createdAt for consistent ordering
		sort.Slice(round, func(i, j int) bool {
			return round[i].CreatedAt < round[j].CreatedAt
		})
		result = append(result, round...)
	}

	// Separate played, current, and next videos
	playedVideos := []Video{}
	currentFound := false
	var currentVideo Video
	nextVideos := []Video{}

	for _, v := range result {
		if v.PlayedAt != nil {
			playedVideos = append(playedVideos, v)
		} else if !currentFound {
			currentFound = true
			currentVideo = v
		} else {
			nextVideos = append(nextVideos, v)
		}
	}

	// Reconstruct playlist
	r.State.Playlist = playedVideos
	if currentFound {
		r.State.Playlist = append(r.State.Playlist, currentVideo)
	}
	r.State.Playlist = append(r.State.Playlist, nextVideos...)
}
