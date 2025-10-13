package main

import (
	"fmt"
	"log"
	"os"
	"strconv"
	"sync"
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

type Player struct {
	ID       string   `json:"id"`
	Name     string   `json:"name"`
	Position Position `json:"postion"`
	Rank     int      `json:"rank"`
}

type Event struct {
	EventType string
}

type Game struct {
	Players map[string]Player
	Board   [][]Cell
	Size    int
	Mu      sync.Mutex
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

	N := boardDim * boardDim

	// Building Candidates for portals
	validCellIDs := []int{}
	for cellID := range N {
		// Excluding starting and ending cells
		if cellID == 0 || cellID == (boardDim-1)*boardDim {
			continue
		}
		validCellIDs = append(validCellIDs, cellID)
	}

	// Creating Portals
	for range maxPortals {
		idx := GetRandNumber(0, len(validCellIDs))
		cellID := validCellIDs[idx]

		srrow, srcol := cellID/boardDim, cellID%boardDim

		validCellIDs = append(validCellIDs[:idx], validCellIDs[idx+1:]...)

		idx = GetRandNumber(0, len(validCellIDs))
		cellID = validCellIDs[idx]
		destRow, destCol := cellID/boardDim, cellID%boardDim
		validCellIDs = append(validCellIDs[:idx], validCellIDs[idx+1:]...)

		color := GenerateLightRandomColor()
		grid[srrow][srcol] = Cell{
			IsPortal: true,
			Dest: Position{
				Row: destRow,
				Col: destCol,
			},
			Color: color,
		}
		grid[destRow][destCol].IsPortal = true
		grid[destRow][destCol].Color = color
	}

	// Assigning Values to the Cells
	dir := 1
	defaultCellColor := os.Getenv("DEFAULT_CELL_COLOR")
	cellVal := N
	for row := range boardDim {
		for col := range boardDim {
			if dir == 1 {
				grid[row][col].Value = cellVal
			} else {
				grid[row][boardDim-col-1].Value = cellVal
			}

			if !grid[row][col].IsPortal {
				grid[row][col].Dest = Position{
					Row: row,
					Col: col,
				}
				grid[row][col].Color = defaultCellColor
			}
			cellVal--
		}

		dir ^= 1

	}

	game.Board = grid
	game.Size = boardDim
	game.Players = make(map[string]Player, maxPlayers)
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
	game.Players[playerID] = Player{
		ID:   playerID,
		Name: playerName,
		Position: Position{
			Row: startRow,
			Col: startCol,
		},
		Rank: 0,
	}

	// Adding player to the cell
	game.Board[startRow][startCol].Players = append(game.Board[startRow][startCol].Players, game.Players[playerID])

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

	if _, exists := game.Players[playerID]; !exists {
		return "", fmt.Errorf("Player doesn't exists")
	}

	playerName := game.Players[playerID].Name

	// removing player from the cell
	game.removePlayerFromCell(playerID)

	delete(game.Players, playerID)
	return playerName, nil
}

// Updates the player position in the board
// based on the dice roll
func (game *Game) MovePlayer(steps int, playerID string) (Position, bool, Position, error) {
	game.Mu.Lock()
	defer game.Mu.Unlock()

	// Getting player position
	playerState, exists := game.Players[playerID]
	if !exists {
		return Position{}, false, Position{}, fmt.Errorf("Player doesn't exists")
	}

	row, col := playerState.Position.Row, playerState.Position.Col
	size := game.Size

	// removing player from the game board
	game.removePlayerFromCell(playerID)

	for steps > 0 {
		left2Right := ((size-1-row)%2 == 0)

		if left2Right {
			rem := (size - 1) - col
			if steps <= rem {
				col += steps
				steps = 0
			} else {
				col = size - 1
				steps -= rem
				if row > 0 {
					row--
				} else {
					steps = 0
				}
			}
		} else {
			rem := col
			if steps <= rem {
				col -= steps
				steps = 0
			} else {
				col = 0
				steps -= rem
				if row > 0 {
					row--
				} else {
					steps = 0
				}
			}
		}
	}

	// Checking teleportation happening or not
	teleported := false
	dest := Position{}
	cell := game.Board[row][col]

	if cell.IsPortal {
		dest = cell.Dest
		row, col = dest.Row, dest.Col
		teleported = true
	}

	playerState.Position = Position{
		Row: row,
		Col: col,
	}

	game.Players[playerID] = playerState
	game.Board[row][col].Players = append(game.Board[row][col].Players, game.Players[playerID])

	return playerState.Position, teleported, dest, nil
}
