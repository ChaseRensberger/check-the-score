package nfl

import (
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"os"
	"strings"

	"github.com/charmbracelet/lipgloss"
	"github.com/charmbracelet/lipgloss/table"
)

type GameData struct {
	Status   string `json:"status"`
	AwayTeam string `json:"away_team"`
	HomeTeam string `json:"home_team"`
	Score    string `json:"score"`
}

type ESPNResponse struct {
	Events []struct {
		Status struct {
			Type struct {
				State       string `json:"state"`
				ShortDetail string `json:"shortDetail"`
			} `json:"type"`
		} `json:"status"`
		Competitions []struct {
			Competitors []struct {
				Team struct {
					DisplayName string `json:"displayName"`
				} `json:"team"`
				Score string `json:"score"`
			} `json:"competitors"`
		} `json:"competitions"`
	} `json:"events"`
}

func getCurrentWeekGames() ([]GameData, error) {
	url := "https://site.api.espn.com/apis/site/v2/sports/football/nfl/scoreboard"

	resp, err := http.Get(url)
	if err != nil {
		return nil, fmt.Errorf("failed to fetch NFL data: %v", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != 200 {
		return nil, fmt.Errorf("failed to fetch NFL data: status code %d", resp.StatusCode)
	}

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, fmt.Errorf("failed to read response body: %v", err)
	}

	var data ESPNResponse
	if err := json.Unmarshal(body, &data); err != nil {
		return nil, fmt.Errorf("failed to parse JSON: %v", err)
	}

	var games []GameData
	for _, event := range data.Events {
		game := GameData{
			Status:   getGameStatus(event),
			AwayTeam: event.Competitions[0].Competitors[1].Team.DisplayName,
			HomeTeam: event.Competitions[0].Competitors[0].Team.DisplayName,
			Score:    getGameScore(event),
		}
		games = append(games, game)
	}

	return games, nil
}

func getGameStatus(event struct {
	Status struct {
		Type struct {
			State       string `json:"state"`
			ShortDetail string `json:"shortDetail"`
		} `json:"type"`
	} `json:"status"`
	Competitions []struct {
		Competitors []struct {
			Team struct {
				DisplayName string `json:"displayName"`
			} `json:"team"`
			Score string `json:"score"`
		} `json:"competitors"`
	} `json:"competitions"`
},
) string {
	status := event.Status.Type.State
	if status == "pre" {
		return event.Status.Type.ShortDetail
	} else if status == "in" {
		return fmt.Sprintf("In Progress - %s", event.Status.Type.ShortDetail)
	}
	return "Final"
}

func getGameScore(event struct {
	Status struct {
		Type struct {
			State       string `json:"state"`
			ShortDetail string `json:"shortDetail"`
		} `json:"type"`
	} `json:"status"`
	Competitions []struct {
		Competitors []struct {
			Team struct {
				DisplayName string `json:"displayName"`
			} `json:"team"`
			Score string `json:"score"`
		} `json:"competitors"`
	} `json:"competitions"`
},
) string {
	if event.Status.Type.State == "pre" {
		return "vs"
	}

	awayScore := event.Competitions[0].Competitors[1].Score
	homeScore := event.Competitions[0].Competitors[0].Score
	return fmt.Sprintf("%s - %s", awayScore, homeScore)
}

func DisplayNFLGames() {
	games, err := getCurrentWeekGames()
	if err != nil {
		fmt.Printf("Error: %v\n", err)
		os.Exit(1)
	}
	re := lipgloss.NewRenderer(os.Stdout)
	baseStyle := re.NewStyle().Padding(0, 2)
	headerStyle := baseStyle.Foreground(lipgloss.Color("252")).Bold(true)
	selectedStyle := baseStyle.Foreground(lipgloss.Color("#01BE85")).Background(lipgloss.Color("#00432F"))

	headers := []string{"Time/Status", "Away Team", "Score", "Home Team"}
	rowsInProgress := make(map[int]bool)

	rows := [][]string{}
	for gameIdx, game := range games {
		rows = append(rows, []string{
			game.Status,
			game.AwayTeam,
			game.Score,
			game.HomeTeam,
		})

		if strings.HasPrefix(game.Status, "In Progress") {
			rowsInProgress[gameIdx] = true
		}
	}

	t := table.New().
		Headers(headers...).
		Rows(rows...).
		Border(lipgloss.NormalBorder()).
		BorderStyle(re.NewStyle().Foreground(lipgloss.Color("238"))).
		StyleFunc(func(row, col int) lipgloss.Style {
			if row == table.HeaderRow {
				return headerStyle
			}

			if rowsInProgress[row] {
				return selectedStyle
			}

			return baseStyle.Foreground(lipgloss.Color("252"))
		})

	fmt.Println(t)
}
