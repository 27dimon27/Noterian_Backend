package websocket

import (
	"log"
	"net/http"
	"time"

	"github.com/go-park-mail-ru/2026_1_WHITECROWSOFT/internal/types"
	"github.com/google/uuid"
	"github.com/gorilla/websocket"
)

var upgrader = websocket.Upgrader{
	CheckOrigin: func(r *http.Request) bool {
		return true
	},
	ReadBufferSize:  1024,
	WriteBufferSize: 1024,
}

type WebSocketHandler struct {
	hub            *Hub
	noteUsecase    NoteUsecaseInterface
	profileUsecase ProfileUsecaseInterface
}

func NewWebSocketHandler(
	hub *Hub,
	noteUsecase NoteUsecaseInterface,
	profileUsecase ProfileUsecaseInterface,
) *WebSocketHandler {
	return &WebSocketHandler{
		hub:            hub,
		noteUsecase:    noteUsecase,
		profileUsecase: profileUsecase,
	}
}

func (h *WebSocketHandler) ServeWS(w http.ResponseWriter, r *http.Request) {
	userID, ok := r.Context().Value(types.UserIDKey).(uuid.UUID)
	if !ok {
		http.Error(w, "Unauthorized", http.StatusUnauthorized)
		return
	}

	noteIDStr := r.PathValue("noteId")
	if noteIDStr == "" {
		http.Error(w, "Note ID required", http.StatusBadRequest)
		return
	}

	noteID, err := uuid.Parse(noteIDStr)
	if err != nil {
		http.Error(w, "Invalid note ID", http.StatusBadRequest)
		return
	}

	note, err := h.noteUsecase.GetNote(r.Context(), noteID, userID)
	if err != nil {
		http.Error(w, "Note not found", http.StatusNotFound)
		return
	}

	if !note.IsPublic {
		http.Error(w, "Forbidden", http.StatusForbidden)
		return
	}

	profile, err := h.profileUsecase.GetProfile(r.Context(), userID)
	if err != nil {
		http.Error(w, "Failed to get user profile", http.StatusInternalServerError)
		return
	}

	conn, err := upgrader.Upgrade(w, r, nil)
	if err != nil {
		log.Printf("Failed to upgrade connection: %v", err)
		return
	}

	client := &ClientInfo{
		UserID:     userID.String(),
		UserName:   profile.Username,
		NoteID:     noteID.String(),
		LastCursor: CursorPosition{},
		Send:       make(chan WebSocketMessage, 256),
		LastPing:   time.Now().Unix(),
	}

	h.hub.register <- client

	go h.writePump(client, conn)
	go h.readPump(client, conn)
}

func (h *WebSocketHandler) writePump(client *ClientInfo, conn *websocket.Conn) {
	ticker := time.NewTicker(30 * time.Second)
	defer func() {
		ticker.Stop()
		conn.Close()
	}()

	for {
		select {
		case message, ok := <-client.Send:
			conn.SetWriteDeadline(time.Now().Add(10 * time.Second))
			if !ok {
				conn.WriteMessage(websocket.CloseMessage, []byte{})
				return
			}

			if err := conn.WriteJSON(message); err != nil {
				log.Printf("Failed to write JSON: %v", err)
				return
			}

		case <-ticker.C:
			conn.SetWriteDeadline(time.Now().Add(10 * time.Second))
			if err := conn.WriteMessage(websocket.PingMessage, nil); err != nil {
				return
			}
		}
	}
}

func (h *WebSocketHandler) readPump(client *ClientInfo, conn *websocket.Conn) {
	defer func() {
		h.hub.unregister <- client
		conn.Close()
	}()

	conn.SetReadLimit(4096)
	conn.SetReadDeadline(time.Now().Add(60 * time.Second))
	conn.SetPongHandler(func(string) error {
		conn.SetReadDeadline(time.Now().Add(60 * time.Second))
		client.LastPing = time.Now().Unix()
		return nil
	})

	for {
		var msg WebSocketMessage
		if err := conn.ReadJSON(&msg); err != nil {
			if websocket.IsUnexpectedCloseError(err, websocket.CloseGoingAway, websocket.CloseAbnormalClosure) {
				log.Printf("WebSocket error: %v", err)
			}
			break
		}

		msg.Timestamp = time.Now().UnixNano()
		msg.NoteID = client.NoteID

		if msg.IsLocal {
			msg.UserID = client.UserID
			msg.UserName = client.UserName
		}

		h.hub.HandleOperation(client.NoteID, client.UserID, msg)
	}
}
