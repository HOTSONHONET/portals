package main

import (
	"net/http"
	"strconv"
	"time"

	"github.com/gin-gonic/gin"
)

type GameHandler struct {
	Game   *Game
	Broker *Broker
	Render func(name string, data any) string
}

func NewGameHander(game *Game, broker *Broker, render func(string, any) string) *GameHandler {
	return &GameHandler{
		Game:   game,
		Broker: broker,
		Render: render,
	}
}

func (h *GameHandler) currentPlayerIDFromCookie(c *gin.Context) (string, error) {
	return c.Cookie("portals_player_id")
}

func (h *GameHandler) SetPortalsCookie(c *gin.Context) {
	if _, err := c.Cookie("portals_player_id"); err != nil {
		c.SetCookie("portals_player_id", strconv.FormatInt(time.Now().UnixNano(), 36), 24*60*60*7, "/", "", false, true)
	}
	me, _ := h.currentPlayerIDFromCookie(c)
	c.HTML(http.StatusOK, "index.html", gin.H{
		"Game": h.Game,
		"Me":   me,
	})
}

func (h *GameHandler) BroadCastEvents(c *gin.Context) {
	c.Writer.Header().Set("Content-Type", "text/event-stream")
	c.Writer.Header().Set("Cache-Control", "no-cache")
	c.Writer.Header().Set("Connection", "keep-alive")
	c.Writer.Header().Set("X-Accel-Buffering", "no")

	flusher, ok := c.Writer.(http.Flusher)
	if !ok {
		c.Status(http.StatusInternalServerError)
		return
	}

	ch := make(chan string, 8)
	h.Broker.Add(ch)
	defer h.Broker.Remove(ch)

	// Broadcasting initial state
	player_id, err := h.currentPlayerIDFromCookie(c)
	if err != nil {
		c.Status(http.StatusBadRequest)
		return
	}

	// Sending initial events
	board := h.Render("_board.html", gin.H{"Game": h.Game})
	players := h.Render("_players.html", gin.H{"Game": h.Game})
	dice := h.Render("_dice.html", gin.H{"Game": h.Game, "Me": player_id})
	tokens := h.Render("_tokens.html", gin.H{"Game": h.Game})

	c.Writer.Write([]byte(sseEvent("board", board)))
	c.Writer.Write([]byte(sseEvent("players", players)))
	c.Writer.Write([]byte(sseEvent("dice", dice)))
	c.Writer.Write([]byte(sseEvent("tokens", tokens)))
	flusher.Flush()

	// Pump
	for {
		select {
		case <-c.Request.Context().Done():
			return
		case msg := <-ch:
			_, _ = c.Writer.Write([]byte(msg))
			flusher.Flush()
		}
	}
}

func (h *GameHandler) JoinGame(c *gin.Context) {
	name := c.PostForm("player_name")
	if name == "" {
		c.String(http.StatusBadRequest, "Name required")
		return
	}

	player_id, err := h.currentPlayerIDFromCookie(c)
	if err != nil {
		c.String(http.StatusBadRequest, err.Error())
		return
	}
	if err := h.Game.AddPlayer(player_id, name); err != nil {
		c.String(http.StatusBadRequest, err.Error())
		return
	}

	// Boardcasting players + board
	h.Broker.Broadcast("players", h.Render("_players.html", gin.H{"Game": h.Game}))
	h.Broker.Broadcast("board", h.Render("_board.html", gin.H{"Game": h.Game}))
	h.Broker.Broadcast("dice", h.Render("_dice.html", gin.H{"Game": h.Game}))
	h.Broker.Broadcast("tokens", h.Render("_tokens.html", gin.H{"Game": h.Game}))

	// Swaping join section
	c.HTML(http.StatusOK, "_joined_header.html", gin.H{"PlayerName": name})
}

func (h *GameHandler) RemovePlayer(c *gin.Context) {
	player_id, err := h.currentPlayerIDFromCookie(c)
	if err != nil {
		c.String(http.StatusBadRequest, err.Error())
		return
	}

	if err = h.Game.RemovePlayer(player_id); err != nil {
		c.String(http.StatusBadRequest, err.Error())
		return
	}

	// BoardCasting Events
	h.Broker.Broadcast("players", h.Render("_players.html", gin.H{"Game": h.Game}))
	h.Broker.Broadcast("board", h.Render("_board.html", gin.H{"Game": h.Game}))
	h.Broker.Broadcast("dice", h.Render("_dice.html", gin.H{"Game": h.Game}))
	h.Broker.Broadcast("tokens", h.Render("_tokens.html", gin.H{"Game": h.Game}))

	c.HTML(http.StatusOK, "_join_form.html", nil)

}

func (h *GameHandler) RollDice(c *gin.Context) {
	player_id, err := h.currentPlayerIDFromCookie(c)
	if err != nil {
		c.String(http.StatusBadRequest, err.Error())
		return
	}

	roll := GetRandNumber(1, 7)
	_, _, _, moveErr := h.Game.MovePlayer(roll, player_id)
	if moveErr != nil {
		c.String(http.StatusBadRequest, moveErr.Error())
		return
	}

	// BoardCasting Events
	h.Broker.Broadcast("players", h.Render("_players.html", gin.H{"Game": h.Game}))
	h.Broker.Broadcast("board", h.Render("_board.html", gin.H{"Game": h.Game}))
	h.Broker.Broadcast("dice", h.Render("_dice.html", gin.H{"Game": h.Game, "Me": player_id}))
	h.Broker.Broadcast("tokens", h.Render("_tokens.html", gin.H{"Game": h.Game}))

	c.HTML(
		http.StatusOK,
		"_dice.html",
		gin.H{
			"Game":       h.Game,
			"Me":         player_id,
			"JustRolled": roll,
		},
	)
}
