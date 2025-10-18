package main

import (
	"fmt"
	"log"
	"os"
	"strconv"
	"sync"
	"time"
)

type Position struct {
	Row int
	Col int
}

type Cell struct {
	IsPortal bool
	Dest     Position
	Value    int
	Color    string
	Players  []Player
}

type TimerState struct {
	StartedAt time.Time
	EndedAt   time.Time
	Active    bool
	Elasped   time.Duration
}

func (t *TimerState) StartNow() {
	t.StartedAt = time.Now().UTC()
	t.EndedAt = time.Time{}
	t.Active = true
}

func (t *TimerState) StopNow() {
	if !t.Active {
		return
	}
	t.EndedAt = time.Now().UTC()
	t.Active = false
	t.Elasped = t.EndedAt.Sub(t.StartedAt)
}

type Player struct {
	ID       string   `json:"id"`
	Name     string   `json:"name"`
	Position Position `json:"postion"`
	Rank     int      `json:"rank"`
	Timer    TimerState
}

type Event struct {
	EventType string
}

type Game struct {
	Players     map[string]Player
	Board       [][]Cell
	Size        int
	Mu          sync.Mutex
	Finder      map[int]Position
	LastCellVal int
}

// Initializes the Game board and Players
func (game *Game) InitGame() {
	// Collecting Game Features
	maxPlayers, err := strconv.Atoi(os.Getenv("MAX_PLAYERS"))
	if err != nil {
		log.Fatalf("error while parsing MAX_PLAYERS env | error: %v\n", err)
	}
	boardDim, err := strconv.Atoi(os.Getenv("BOARD_DIM"))
	if err != nil {
		log.Fatalf("error while parsing BOARD_DIM env | error: %v\n", err)
	}
	maxPortals, err := strconv.Atoi(os.Getenv("MAX_PORTALS"))
	if err != nil {
		log.Fatalf("error while parsing MAX_PORTALS env | error: %v\n", err)
	}

	// Creating the Grid
	grid := make([][]Cell, boardDim)
	for row := range boardDim {
		grid[row] = make([]Cell, boardDim)
	}

	// Assigning last cell value
	game.LastCellVal = boardDim * boardDim

	// Assigning Values to the Cells
	dir := 1
	defaultCellColor := os.Getenv("DEFAULT_CELL_COLOR")
	cellVal := game.LastCellVal
	finder := make(map[int]Position)
	for row := range boardDim {
		for col := range boardDim {
			if dir == 1 {
				grid[row][col].Value = cellVal
				finder[cellVal] = Position{
					Row: row,
					Col: col,
				}
			} else {
				grid[row][boardDim-col-1].Value = cellVal
				finder[cellVal] = Position{
					Row: row,
					Col: boardDim - col - 1,
				}
			}
			grid[row][col].Color = defaultCellColor
			cellVal--
		}

		dir ^= 1

	}

	// Building Candidates for portals
	validCellVals := []int{}
	for cellID := range game.LastCellVal {
		cellID++
		// Excluding starting and ending cells
		if cellID == 1 || cellID == game.LastCellVal {
			continue
		}
		validCellVals = append(validCellVals, cellID)
	}

	// Creating Portals
	for range maxPortals {
		idx := GetRandNumber(0, len(validCellVals))
		cellVal := validCellVals[idx]

		srrow, srcol := finder[cellVal].Row, finder[cellVal].Col

		// Removing the cell Value
		validCellVals = append(validCellVals[:idx], validCellVals[idx+1:]...)

		idx = GetRandNumber(0, len(validCellVals))
		cellVal = validCellVals[idx]
		destRow, destCol := finder[cellVal].Row, finder[cellVal].Col
		validCellVals = append(validCellVals[:idx], validCellVals[idx+1:]...)

		color := GenerateLightRandomColor()
		grid[srrow][srcol].IsPortal = true
		grid[srrow][srcol].Dest = Position{
			Row: destRow,
			Col: destCol,
		}
		grid[srrow][srcol].Color = color
		grid[destRow][destCol].Color = color
	}

	game.Board = grid
	game.Size = boardDim
	game.Players = make(map[string]Player, maxPlayers)
	game.Finder = finder
}

// Add the player with the given player ID in the Game
func (game *Game) AddPlayer(playerID, playerName string) error {
	game.Mu.Lock()
	defer game.Mu.Unlock()

	// Checking if the player already exists
	_, exists := game.Players[playerID]
	if exists {
		return fmt.Errorf("Player already exists")
	}

	startRow, startCol := game.Size-1, 0

	player := Player{
		ID:   playerID,
		Name: playerName,
		Position: Position{
			Row: startRow,
			Col: startCol,
		},
		Rank:  0,
		Timer: TimerState{},
	}
	player.Timer.StartNow()
	game.Players[playerID] = player

	// Adding player to the cell
	game.Board[startRow][startCol].Players = append(game.Board[startRow][startCol].Players, player)

	return nil
}

// Remove player from the cell
func (game *Game) removePlayerFromCell(playerID string) {
	idx := -1
	pos := game.Players[playerID].Position
	row, col := pos.Row, pos.Col

	for i, player := range game.Board[row][col].Players {
		if player.ID == playerID {
			idx = i
			break
		}
	}

	if idx != -1 {
		game.Board[row][col].Players = append(game.Board[row][col].Players[:idx], game.Board[row][col].Players[idx+1:]...)
	}
}

// Remove the player with the given player ID from the Game
func (game *Game) RemovePlayer(playerID string) (string, error) {
	game.Mu.Lock()
	defer game.Mu.Unlock()

	player, exists := game.Players[playerID]
	if !exists {
		return "", fmt.Errorf("Player doesn't exists")
	}

	playerName := game.Players[playerID].Name

	// removing player from the cell
	game.removePlayerFromCell(playerID)

	if player.Timer.Active {
		player.Timer.Active = false
		player.Timer.EndedAt = time.Time{}
	}

	delete(game.Players, playerID)
	return playerName, nil
}

// Updates the player position in the board
// based on the dice roll
// returns PlayerState, hasTeleported, hasMoved, hasCompleted, CurrentValue of the cell where player is, error
func (game *Game) MovePlayer(steps int, playerID string) (Player, bool, bool, bool, int, error) {
	game.Mu.Lock()
	defer game.Mu.Unlock()

	// Getting player position
	playerState, exists := game.Players[playerID]
	if !exists {
		return Player{}, false, false, false, -1, fmt.Errorf("Player doesn't exists")
	}

	row, col := playerState.Position.Row, playerState.Position.Col

	newVal := game.Board[row][col].Value + steps
	if newVal > game.LastCellVal {
		return playerState, false, false, false, game.Board[row][col].Value, nil
	}

	log.Printf("old: %v | roll: %v | new: %v\n | newValPos: %v\n", game.Board[row][col].Value, steps, newVal, game.Finder[newVal])

	// removing player from the game board
	game.removePlayerFromCell(playerID)

	// Moving the player
	row, col = game.Finder[newVal].Row, game.Finder[newVal].Col

	// Checking teleportation happening or not
	teleported := false
	dest := Position{}
	cell := game.Board[row][col]

	if cell.IsPortal {
		dest = cell.Dest
		row, col = dest.Row, dest.Col
		teleported = true
	}

	// Checking if player has completed the game
	hasCompleted := false
	if row == 0 && col == 0 {
		playerState.Timer.StopNow()
		hasCompleted = true
	}

	playerState.Position = Position{
		Row: row,
		Col: col,
	}

	game.Players[playerID] = playerState
	game.Board[row][col].Players = append(
		game.Board[row][col].Players,
		game.Players[playerID],
	)

	return playerState, teleported, true, hasCompleted, game.Board[row][col].Value, nil
}
