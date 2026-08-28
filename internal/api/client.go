package api

import (
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"pogo-bot/internal/models"
	"strings"
	"time"
)

// Community JSON endpoints. Niantic has no public API; these are maintained
// by the Pokémon GO community (see README.md → "Data Sources").
const (
	EventsURL            = "https://raw.githubusercontent.com/bigfoott/ScrapedDuck/data/events.json"
	RaidBossesURL        = "https://raw.githubusercontent.com/bigfoott/ScrapedDuck/data/raids.json"
	PokemonStatsURL      = "https://pogoapi.net/api/v1/pokemon_stats.json"
	PokemonMovesURL      = "https://pogoapi.net/api/v1/current_pokemon_moves.json"
	PokemonTypesURL      = "https://pogoapi.net/api/v1/pokemon_types.json"
	TypeEffectivenessURL = "https://pogoapi.net/api/v1/type_effectiveness.json"
	PokedexAPIURL        = "https://pokemon-go-api.github.io/pokemon-go-api/api/pokedex.json"
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

// FetchRaidBosses downloads and decodes ScrapedDuck raids.json.
func (c *Client) FetchRaidBosses() ([]models.RaidBoss, error) {

	raidBosses := []models.RaidBoss{}

	if err := c.fetchData(RaidBossesURL, &raidBosses); err != nil {
		return nil, err
	}

	return raidBosses, nil
}

// FetchPokemonStats downloads and decodes pokemon_stats.json.
func (c *Client) FetchPokemonStats() ([]models.PokemonStats, error) {

	pokemonStats := []models.PokemonStats{}

	if err := c.fetchData(PokemonStatsURL, &pokemonStats); err != nil {
		return nil, err
	}

	return pokemonStats, nil
}

// FetchPokemonMoves downloads and decodes current_pokemon_moves.json.
func (c *Client) FetchPokemonMoves() ([]models.PokemonMoves, error) {

	pokemonMoves := []models.PokemonMoves{}

	if err := c.fetchData(PokemonMovesURL, &pokemonMoves); err != nil {
		return nil, err
	}

	return pokemonMoves, nil
}

// FetchPokemonTypes downloads and decodes pokemon_types.json.
func (c *Client) FetchPokemonTypes() ([]models.PokemonTypes, error) {
	pokemonTypes := []models.PokemonTypes{}
	if err := c.fetchData(PokemonTypesURL, &pokemonTypes); err != nil {
		return nil, err
	}
	return pokemonTypes, nil
}

// FetchTypeEffectiveness downloads and decodes type_effectiveness.json.
func (c *Client) FetchTypeEffectiveness() (*models.TypeEffectiveness, error) {

	typeEffectiveness := &models.TypeEffectiveness{}

	if err := c.fetchData(TypeEffectivenessURL, typeEffectiveness); err != nil {
		return nil, err
	}

	return typeEffectiveness, nil
}

// FetchPokedexAPI downloads and decodes the gamemaster pokedex (includes GO icon URLs).
func (c *Client) FetchPokedexAPI() ([]models.PokedexAPIEntry, error) {
	entries := []models.PokedexAPIEntry{}
	if err := c.fetchData(PokedexAPIURL, &entries); err != nil {
		return nil, err
	}
	return entries, nil
}

// findStatsByName finds the first Normal form of the given Pokémon name.
func findStatsByName(stats []models.PokemonStats, name string) (*models.PokemonStats, bool) {
	var firstMatch *models.PokemonStats
	for i := range stats {
		if !strings.EqualFold(stats[i].PokemonName, name) {
			continue
		}
		if stats[i].Form == "Normal" {
			return &stats[i], true
		}
		if firstMatch == nil {
			firstMatch = &stats[i]
		}
	}
	if firstMatch != nil {
		return firstMatch, true
	}
	return nil, false
}

// findMoves finds the first match of the given Pokémon name, ID, and form.
func (c *Client) findMoves(moves []models.PokemonMoves, name string, id int, form string) (*models.PokemonMoves, bool) {
	for i, move := range moves {
		if move.PokemonName == name && move.PokemonID == id && move.Form == form {
			return &moves[i], true
		}
	}
	return nil, false
}

// findTypes finds the first match of the given Pokémon name, ID, and form.
func (c *Client) findTypes(types []models.PokemonTypes, name string, id int, form string) (*models.PokemonTypes, bool) {
	for i, t := range types {
		if t.PokemonName == name && t.PokemonID == id && t.Form == form {
			return &types[i], true
		}
	}
	return nil, false
}

// grabGOImage finds a GO icon URL from pokemon-go-api for the given pokemon name and form.
func grabGOImage(entries []models.PokedexAPIEntry, name, form string) string {
	for i := range entries {
		if !strings.EqualFold(entries[i].Names.English, name) {
			continue
		}
		if form == "" || strings.EqualFold(form, "Normal") {
			return entries[i].Assets.Image
		}
		for _, formSprite := range entries[i].AssetForms {
			if formSprite.Form != nil && strings.EqualFold(*formSprite.Form, form) {
				return formSprite.Image
			}
		}
		return entries[i].Assets.Image
	}
	return ""
}

func (c *Client) LookupPokemon(name string) (*models.PokemonProfile, error) {
	pokemonStats, err := c.FetchPokemonStats()
	if err != nil {
		return nil, err
	}
	pokemonMoves, err := c.FetchPokemonMoves()
	if err != nil {
		return nil, err
	}
	pokemonTypes, err := c.FetchPokemonTypes()
	if err != nil {
		return nil, err
	}
	pokedexAPI, err := c.FetchPokedexAPI()
	if err != nil {
		return nil, err
	}

	pokemon, ok := findStatsByName(pokemonStats, name)
	if !ok {
		return nil, fmt.Errorf("stats not found for %s", name)
	}
	move, ok := c.findMoves(pokemonMoves, pokemon.PokemonName, pokemon.PokemonID, pokemon.Form)
	if !ok {
		return nil, fmt.Errorf("moves not found for %s", name)
	}
	types, ok := c.findTypes(pokemonTypes, pokemon.PokemonName, pokemon.PokemonID, pokemon.Form)
	if !ok {
		return nil, fmt.Errorf("types not found for %s", name)
	}

	return &models.PokemonProfile{
		Stats:   *pokemon,
		Moves:   *move,
		Types:   *types,
		GOImage: grabGOImage(pokedexAPI, pokemon.PokemonName, pokemon.Form),
	}, nil
}
