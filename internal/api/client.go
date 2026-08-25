package api

import (
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"pogo-bot/internal/models"
	"time"
)

// Community JSON endpoints. Niantic has no public API; these are maintained
// by the Pokémon GO community (see README.md → "Data Sources").
const (
	EventsURL            = "https://raw.githubusercontent.com/bigfoott/ScrapedDuck/data/events.json"
	RaidBossesURL        = "https://pogoapi.net/api/v1/raid_bosses.json"
	PokemonStatsURL      = "https://pogoapi.net/api/v1/pokemon_stats.json"
	PokemonMovesURL      = "https://pogoapi.net/api/v1/current_pokemon_moves.json"
	TypeEffectivenessURL = "https://pogoapi.net/api/v1/type_effectiveness.json"
)

// Client talks to the community datasets over HTTP.
type Client struct {
	Client  *http.Client
	Timeout time.Duration
}


// New constructs a Client with a 10 second timeout.
func New() *Client {
	return &Client{
		Client:  &http.Client{Timeout: 10 * time.Second},
		Timeout: 10 * time.Second,
	}
}

// fetchData downloads and decodes the data from the URL into the struct.
func (c *Client) fetchData(url string, dest any) error {

	ctx, cancel := context.WithTimeout(context.Background(), c.Timeout)
	defer cancel()

	req, err := http.NewRequestWithContext(ctx, http.MethodGet, url, nil)
	if err != nil {
		return fmt.Errorf("failed to create request: %s: %w", url, err)
	}

	resp, err := c.Client.Do(req)
	if err != nil {
		return fmt.Errorf("failed to execute request: %s: %w", url, err)
	}

	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return fmt.Errorf("GET %s: unexpected status %s", url, resp.Status)
	}

	if err := json.NewDecoder(resp.Body).Decode(dest); err != nil {
		return fmt.Errorf("failed to decode response body: %s: %w", url, err)
	}
	return nil
}

// FetchEvents fetches events.json and decodes it into []models.Event.

func (c *Client) FetchEvents() ([]models.Event, error) {

	events := []models.Event{}

	if err := c.fetchData(EventsURL, &events); err != nil {
		return nil, err
	}

	return events, nil
}
