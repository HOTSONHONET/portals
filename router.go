package main

import (
	"bytes"
	"html/template"
	"log"

	"github.com/gin-gonic/gin"
)

func Arise() *gin.Engine {
	router := gin.Default()

	// Loading all the templates
	sub := func(x, y int) int {
		return x - y
	}

	seq := func(start, end int) []int {
		n := end - start + 1
		if n <= 0 {
			return []int{}
		}

		out := make([]int, n)
		for i := range n {
			out[i] = start + i
		}
		return out
	}
	templ := template.Must(
		template.New("all").
			Funcs(
				template.FuncMap{
					"seq": seq,
					"sub": sub,
				}).ParseGlob("templates/*.html"),
	)
	router.SetHTMLTemplate(templ)

	// Initializing the game
	game := &Game{}
	game.InitGame()

	// Initializing the stream
	streamer := NewStreamer()

	// Creating broker
	broker := NewBroker()

	// Func to render templates for Broadcasting
	Render := func(name string, data any) string {
		var buf bytes.Buffer
		if err := templ.ExecuteTemplate(&buf, name, data); err != nil {
			log.Printf("error while rendering | err: %v\n", err)
			return ""
		}

		return buf.String()
	}
	h := NewGameHander(game, broker, streamer, Render)
	router.GET("/", h.SetPortalsCookie)
	router.GET("/events", h.BroadCastEvents)
	router.GET("/dice-roll", h.RollDice)
	router.POST("/join", h.JoinGame)
	router.POST("/leave", h.RemovePlayer)

	return router
}
