package main

import (
	"fmt"
	"math/rand"
	"net"
	"sort"
	"time"
)

// function to generate random light hexa-decimal color
func GenerateLightRandomColor() string {
	r := GetRandNumber(0, 80) + 150
	g := GetRandNumber(0, 80) + 150
	b := GetRandNumber(0, 80) + 150

	return fmt.Sprintf("#%02X%02X%02X", r, g, b)
}

// GenerateVibrantRandomColor returns a vibrant, visually strong color (not too dark/light)
func GenerateVibrantRandomColor() string {
	r := GetRandNumber(50, 205)
	g := GetRandNumber(50, 205)
	b := GetRandNumber(50, 205)

	// Randomly boost one of the channels for vibrance
	switch GetRandNumber(0, 3) {
	case 0:
		r = GetRandNumber(150, 255)
	case 1:
		g = GetRandNumber(150, 255)
	case 2:
		b = GetRandNumber(150, 255)
	}

	return fmt.Sprintf("#%02X%02X%02X", r, g, b)
}

func GetCurrentPlayers(game *Game) []Player {
	players := make([]Player, 0)
	for _, p := range game.Players {
		players = append(players, p)
	}

	sort.Slice(players, func(i, j int) bool {
		return players[i].Rank < players[j].Rank
	})
	return players
}

func ParseClientIP(ip string) string {
	client_ip := net.ParseIP(ip)
	if client_ip.IsLoopback() {
		return "localhost"
	}
	return ip
}

func GetRandNumber(L, R int) int {
	rand.New(rand.NewSource(time.Now().Unix()))
	randVal := rand.Intn(R - L)
	return L + randVal
}
