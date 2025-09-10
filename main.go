package main

import (
	"bufio"
	"encoding/json"
	"fmt"
	"os"
	"sort"
	"strings"
)

// Game represents statistics for a single game
type Game struct {
	TotalKills   int            `json:"total_kills"`
	Players      []string       `json:"players"`
	Kills        map[string]int `json:"kills"`
	KillsByMeans map[string]int `json:"kills_by_means"`
}

// GameCollection represents multiple games
type GameCollection map[string]Game

// PlayerRank represents a player's ranking
type PlayerRank struct {
	Name  string
	Kills int
}

func parseLogFile(filePath string) (GameCollection, error) {
	games := make(GameCollection)
	var currentGame *Game
	var currentGameNumber int

	file, err := os.Open(filePath)
	if err != nil {
		return nil, err
	}
	defer file.Close()

	scanner := bufio.NewScanner(file)
	for scanner.Scan() {
		line := scanner.Text()

		// Check for new game initialization
		if strings.Contains(line, "InitGame:") {
			currentGameNumber++
			gameName := fmt.Sprintf("game_%d", currentGameNumber)
			games[gameName] = Game{
				Players:      make([]string, 0),
				Kills:        make(map[string]int),
				KillsByMeans: make(map[string]int),
			}
			temp := games[gameName]
			currentGame = &temp
		}

		if currentGame == nil {
			continue
		}

		// Check for player connections
		if strings.Contains(line, "ClientUserinfoChanged") {
			fields := strings.Split(line, `\`)
			if len(fields) > 1 {
				playerName := fields[1]
				if playerName != "" {
					// Add player if not already in the list
					if !contains(currentGame.Players, playerName) {
						currentGame.Players = append(currentGame.Players, playerName)
						currentGame.Kills[playerName] = 0
					}
				}
			}
		}

		// Check for kills
		if strings.Contains(line, "Kill:") {
			parts := strings.Split(line, ": ")
			if len(parts) < 3 {
				continue
			}

			killMessage := parts[len(parts)-1]
			killParts := strings.Split(killMessage, " ")

			if len(killParts) >= 4 {
				attacker := killParts[0]
				victim := killParts[2]

				// Extract death cause (last part after "by")
				deathCause := killParts[len(killParts)-1]
				currentGame.KillsByMeans[deathCause]++
				currentGame.TotalKills++

				if attacker == "<world>" {
					currentGame.Kills[victim]--
				} else {
					currentGame.Kills[attacker]++
				}
			}
		}
	}

	return games, scanner.Err()
}

func contains(slice []string, item string) bool {
	for _, s := range slice {
		if s == item {
			return true
		}
	}
	return false
}

func generateRanking(games GameCollection) []PlayerRank {
	totalKills := make(map[string]int)

	// Accumulate kills across all games
	for _, game := range games {
		for player, kills := range game.Kills {
			totalKills[player] += kills
		}
	}

	// Convert to slice for sorting
	var ranking []PlayerRank
	for player, kills := range totalKills {
		ranking = append(ranking, PlayerRank{Name: player, Kills: kills})
	}

	// Sort by kills (descending)
	sort.Slice(ranking, func(i, j int) bool {
		return ranking[i].Kills > ranking[j].Kills
	})

	return ranking
}

func main() {
	filePath := "log.txt"

	// Parse the log file
	games, err := parseLogFile(filePath)
	if err != nil {
		fmt.Printf("Error parsing log file: %v\n", err)
		return
	}

	// Generate JSON output for games
	gamesJSON, err := json.MarshalIndent(games, "", "  ")
	if err != nil {
		fmt.Printf("Error generating games JSON: %v\n", err)
		return
	}
	fmt.Println("Games Report:")
	fmt.Println(string(gamesJSON))

	// Generate and print ranking
	fmt.Println("\nRanking Report:")
	ranking := generateRanking(games)
	for i, player := range ranking {
		fmt.Printf("%d. %s - %d kills\n", i+1, player.Name, player.Kills)
	}
}
