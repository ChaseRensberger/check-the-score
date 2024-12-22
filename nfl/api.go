package nfl

import (
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"os"

	"github.com/charmbracelet/bubbles/table"
	"github.com/charmbracelet/lipgloss"
)

var baseStyle = lipgloss.NewStyle().
	BorderStyle(lipgloss.NormalBorder()).
	BorderForeground(lipgloss.Color("240"))

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

	columns := []table.Column{
		{Title: "Time/Status"},
		{Title: "Away Team"},
		{Title: "Score"},
		{Title: "Home Team"},
	}

	rows := []table.Row{}
	for _, game := range games {
		rows = append(rows, table.Row{
			game.Status,
			game.AwayTeam,
			game.Score,
			game.HomeTeam,
		})
	}

	maxWidths := make([]int, 4)
	maxWidths[0] = len("Time/Status")
	maxWidths[1] = len("Away Team")
	maxWidths[2] = len("Score")
	maxWidths[3] = len("Home Team")

	for _, game := range games {
		if len(game.Status) > maxWidths[0] {
			maxWidths[0] = len(game.Status)
		}
		if len(game.AwayTeam) > maxWidths[1] {
			maxWidths[1] = len(game.AwayTeam)
		}
		if len(game.Score) > maxWidths[2] {
			maxWidths[2] = len(game.Score)
		}
		if len(game.HomeTeam) > maxWidths[3] {
			maxWidths[3] = len(game.HomeTeam)
		}
	}

	totalWidth := 0
	for i, width := range maxWidths {
		columns[i].Width = width + 2
		totalWidth += width + 2
	}

	totalWidth += 5

	t := table.New(
		table.WithColumns(columns),
		table.WithRows(rows),
		table.WithHeight(len(rows)),
		table.WithWidth(totalWidth),
	)

	s := table.DefaultStyles()
	s.Header = s.Header.
		BorderStyle(lipgloss.NormalBorder()).
		BorderForeground(lipgloss.Color("240")).
		BorderBottom(true).
		Bold(true)

	t.SetStyles(s)

	fmt.Println(baseStyle.Render(t.View()))
}
