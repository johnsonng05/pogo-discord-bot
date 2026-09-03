package api

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"log"
	"net/http"
	"pogo-discord-bot/internal/cache"
	"pogo-discord-bot/internal/models"
	"strings"
	"time"

	"github.com/redis/go-redis/v9"
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
	Cache   *cache.Cache
}

// New constructs a Client with a 10 second timeout.
func New(rdb *cache.Cache) *Client {
	return &Client{
		Client:  &http.Client{Timeout: 10 * time.Second},
		Timeout: 10 * time.Second,
		Cache:   rdb,
	}
}

// fetchBytes downloads and returns the bytes from the given URL.
func (c *Client) fetchBytes(url string) ([]byte, error) {
	ctx, cancel := context.WithTimeout(context.Background(), c.Timeout)
	defer cancel()

	req, err := http.NewRequestWithContext(ctx, http.MethodGet, url, nil)
	if err != nil {
		return nil, fmt.Errorf("failed to create request: %s: %w", url, err)
	}
	resp, err := c.Client.Do(req)
	if err != nil {
		return nil, fmt.Errorf("failed to execute request: %s: %w", url, err)
	}

	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("GET %s: unexpected status %s", url, resp.Status)
	}

	return io.ReadAll(resp.Body)
}

// fetchJSON retrieves JSON data from the cache or the given URL, caches successful
// responses, and unmarshals the response into dest.
func (c *Client) fetchJSON(cacheKey string, url string, dest any) error {
	ctx := context.Background()
	if c.Cache != nil {
		cached, err := c.Cache.GetCached(ctx, cacheKey)
		if err == nil {
			log.Printf("cache hit %s", cacheKey)
			return json.Unmarshal(cached, dest)
		}
		if !errors.Is(err, redis.Nil) {
			log.Printf("cache get %s: %v", cacheKey, err)
		}
	}
	body, err := c.fetchBytes(url)
	if err != nil {
		return err
	}
	if c.Cache != nil {
		if err := c.Cache.SetCached(ctx, cacheKey, body, cache.DefaultTTL); err != nil {
			log.Printf("cache set %s: %v", cacheKey, err)
		}
	}
	return json.Unmarshal(body, dest)
}

// FetchEvents fetches events.json and decodes it into []models.Event.
func (c *Client) FetchEvents() ([]models.Event, error) {

	events := []models.Event{}

	if err := c.fetchJSON(cache.KeyEvents, EventsURL, &events); err != nil {
		return nil, err
	}

	return events, nil
}

// FetchRaidBosses downloads and decodes ScrapedDuck raids.json.
func (c *Client) FetchRaidBosses() ([]models.RaidBoss, error) {

	raidBosses := []models.RaidBoss{}

	if err := c.fetchJSON(cache.KeyRaids, RaidBossesURL, &raidBosses); err != nil {
		return nil, err
	}

	return raidBosses, nil
}

// FetchPokemonStats downloads and decodes pokemon_stats.json.
func (c *Client) FetchPokemonStats() ([]models.PokemonStats, error) {

	pokemonStats := []models.PokemonStats{}

	if err := c.fetchJSON(cache.KeyPokemonStats, PokemonStatsURL, &pokemonStats); err != nil {
		return nil, err
	}

	return pokemonStats, nil
}

// FetchPokemonMoves downloads and decodes current_pokemon_moves.json.
func (c *Client) FetchPokemonMoves() ([]models.PokemonMoves, error) {

	pokemonMoves := []models.PokemonMoves{}

	if err := c.fetchJSON(cache.KeyPokemonMoves, PokemonMovesURL, &pokemonMoves); err != nil {
		return nil, err
	}

	return pokemonMoves, nil
}

// FetchPokemonTypes downloads and decodes pokemon_types.json.
func (c *Client) FetchPokemonTypes() ([]models.PokemonTypes, error) {
	pokemonTypes := []models.PokemonTypes{}
	if err := c.fetchJSON(cache.KeyPokemonTypes, PokemonTypesURL, &pokemonTypes); err != nil {
		return nil, err
	}
	return pokemonTypes, nil
}

// FetchTypeEffectiveness downloads and decodes type_effectiveness.json.
func (c *Client) FetchTypeEffectiveness() (*models.TypeEffectiveness, error) {

	typeEffectiveness := &models.TypeEffectiveness{}

	if err := c.fetchJSON(cache.KeyTypeEffectiveness, TypeEffectivenessURL, typeEffectiveness); err != nil {
		return nil, err
	}

	return typeEffectiveness, nil
}

// FetchPokedexAPI downloads and decodes the gamemaster pokedex (includes GO icon URLs).
func (c *Client) FetchPokedexAPI() ([]models.PokedexAPIEntry, error) {
	entries := []models.PokedexAPIEntry{}
	if err := c.fetchJSON(cache.KeyPokedex, PokedexAPIURL, &entries); err != nil {
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
func grabGOImage(entries []models.PokedexAPIEntry, name string, form string) string {
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
