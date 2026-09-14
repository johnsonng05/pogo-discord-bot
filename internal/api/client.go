package api

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"log"
	"net/http"
	"pogo-discord-bot/internal/cache"
	"pogo-discord-bot/internal/models"
	"slices"
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

// decodeJSON GETs url and streams the body into dest with json.Decoder.
func (c *Client) decodeJSON(url string, dest any) error {
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
		return fmt.Errorf("decode %s: %w", url, err)
	}
	return nil
}

// fetchJSON retrieves JSON from Redis or the given URL, streaming the HTTP
// response into dest on a miss, then caching a marshaled copy of dest.
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
	if err := c.decodeJSON(url, dest); err != nil {
		return err
	}
	if c.Cache != nil {
		payload, err := json.Marshal(dest)
		if err != nil {
			log.Printf("cache marshal %s: %v", cacheKey, err)
		} else if err := c.Cache.SetCached(ctx, cacheKey, payload, cache.DefaultTTL); err != nil {
			log.Printf("cache set %s: %v", cacheKey, err)
		}
	}
	return nil
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

// FetchGOImages returns a slim nested map: pokemon name → form → GO icon URL.
func (c *Client) FetchGOImages() (map[string]map[string]string, error) {
	if images, ok := c.getCachedGOImages(); ok {
		return images, nil
	}
	images, _, err := c.warmPokedexCaches()
	return images, err
}

// FetchMegaForms returns a slim nested map: pokemon name → mega form → MegaForm.
func (c *Client) FetchMegaForms() (map[string]map[string]models.MegaForm, error) {
	if megas, ok := c.getCachedMegaForms(); ok {
		return megas, nil
	}
	_, megas, err := c.warmPokedexCaches()
	return megas, err
}

func (c *Client) getCachedGOImages() (map[string]map[string]string, bool) {
	if c.Cache == nil {
		return nil, false
	}
	ctx := context.Background()
	cached, err := c.Cache.GetCached(ctx, cache.KeyGOImages)
	if err != nil {
		if !errors.Is(err, redis.Nil) {
			log.Printf("cache get %s: %v", cache.KeyGOImages, err)
		}
		return nil, false
	}
	var images map[string]map[string]string
	if err := json.Unmarshal(cached, &images); err != nil {
		log.Printf("cache get %s: corrupt payload: %v", cache.KeyGOImages, err)
		if delErr := c.Cache.Delete(ctx, cache.KeyGOImages); delErr != nil {
			log.Printf("corrupt cache delete %s: %v", cache.KeyGOImages, delErr)
		}
		return nil, false
	}
	log.Printf("cache hit %s", cache.KeyGOImages)
	return images, true
}

func (c *Client) getCachedMegaForms() (map[string]map[string]models.MegaForm, bool) {
	if c.Cache == nil {
		return nil, false
	}
	ctx := context.Background()
	cached, err := c.Cache.GetCached(ctx, cache.KeyMegaForms)
	if err != nil {
		if !errors.Is(err, redis.Nil) {
			log.Printf("cache get %s: %v", cache.KeyMegaForms, err)
		}
		return nil, false
	}
	var megas map[string]map[string]models.MegaForm
	if err := json.Unmarshal(cached, &megas); err != nil {
		log.Printf("cache get %s: corrupt payload: %v", cache.KeyMegaForms, err)
		if delErr := c.Cache.Delete(ctx, cache.KeyMegaForms); delErr != nil {
			log.Printf("corrupt cache delete %s: %v", cache.KeyMegaForms, delErr)
		}
		return nil, false
	}
	log.Printf("cache hit %s", cache.KeyMegaForms)
	return megas, true
}

// warmPokedexCaches downloads pokedex.json once and caches both image and mega maps.
func (c *Client) warmPokedexCaches() (map[string]map[string]string, map[string]map[string]models.MegaForm, error) {
	var entries []models.PokedexEntry
	if err := c.decodeJSON(PokedexAPIURL, &entries); err != nil {
		return nil, nil, err
	}
	images := buildGOImageMap(entries)
	megas := buildMegaFormMap(entries)

	if c.Cache != nil {
		ctx := context.Background()
		if payload, err := json.Marshal(images); err != nil {
			log.Printf("cache marshal %s: %v", cache.KeyGOImages, err)
		} else if err := c.Cache.SetCached(ctx, cache.KeyGOImages, payload, cache.GOImagesTTL); err != nil {
			log.Printf("cache set %s: %v", cache.KeyGOImages, err)
		}
		if payload, err := json.Marshal(megas); err != nil {
			log.Printf("cache marshal %s: %v", cache.KeyMegaForms, err)
		} else if err := c.Cache.SetCached(ctx, cache.KeyMegaForms, payload, cache.MegaFormsTTL); err != nil {
			log.Printf("cache set %s: %v", cache.KeyMegaForms, err)
		}
	}
	return images, megas, nil
}

// findStatsByName finds the first Normal form of the given Pokémon name.
func findStatsByName(stats []models.PokemonStats, name string, form string) (*models.PokemonStats, bool) {
	var firstMatch *models.PokemonStats
	for i := range stats {
		if !strings.EqualFold(stats[i].PokemonName, name) {
			continue
		}
		if form != "" {
			if strings.EqualFold(stats[i].Form, form) {
				return &stats[i], true
			}
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

// formsByName returns a list of all forms for the given Pokémon name.
func formsByName(stats []models.PokemonStats, name string) []string {
	forms := []string{}
	for i := range stats {
		if !strings.EqualFold(stats[i].PokemonName, name) {
			continue
		}
		if !slices.Contains(forms, stats[i].Form) {
			forms = append(forms, stats[i].Form)
		}
	}
	slices.Sort(forms)
	return forms
}

// isMegaForm reports whether form is a mega evolution request.
func isMegaForm(form string) bool {
	key := normalizeGOImageForm(form)
	return key == "mega" || strings.HasPrefix(key, "mega_")
}

// megaFormKeyFromID maps pokedex mega ids like CHARIZARD_MEGA_X → mega_x.
func megaFormKeyFromID(id string) string {
	upper := strings.ToUpper(id)
	switch {
	case strings.HasSuffix(upper, "_MEGA_X"):
		return "mega_x"
	case strings.HasSuffix(upper, "_MEGA_Y"):
		return "mega_y"
	case strings.HasSuffix(upper, "_MEGA"):
		return "mega"
	default:
		return ""
	}
}

// displayMegaForm turns a normalized key into the pogoapi-style label.
func displayMegaForm(key string) string {
	switch key {
	case "mega_x":
		return "Mega_X"
	case "mega_y":
		return "Mega_Y"
	case "mega":
		return "Mega"
	default:
		return key
	}
}

// buildMegaFormMap reduces megaEvolutions to name → form → MegaForm.
func buildMegaFormMap(entries []models.PokedexEntry) map[string]map[string]models.MegaForm {
	megas := make(map[string]map[string]models.MegaForm)
	for i := range entries {
		if len(entries[i].MegaEvolutions) == 0 {
			continue
		}
		name := normalizeGOImageName(entries[i].Names.English)
		if name == "" {
			continue
		}
		forms, ok := megas[name]
		if !ok {
			forms = make(map[string]models.MegaForm)
			megas[name] = forms
		}
		for id, mega := range entries[i].MegaEvolutions {
			key := megaFormKeyFromID(id)
			if key == "" {
				key = megaFormKeyFromID(mega.ID)
			}
			if key == "" {
				continue
			}
			types := []string{}
			if mega.PrimaryType.Names.English != "" {
				types = append(types, mega.PrimaryType.Names.English)
			}
			if mega.SecondaryType != nil && mega.SecondaryType.Names.English != "" {
				types = append(types, mega.SecondaryType.Names.English)
			}
			forms[key] = models.MegaForm{
				Form:        displayMegaForm(key),
				BaseAttack:  mega.Stats.Attack,
				BaseDefense: mega.Stats.Defense,
				BaseStamina: mega.Stats.Stamina,
				Types:       types,
				Image:       mega.Assets.Image,
			}
		}
	}
	return megas
}

func findMegaForm(megas map[string]map[string]models.MegaForm, name, form string) (models.MegaForm, bool) {
	forms := megas[normalizeGOImageName(name)]
	if forms == nil {
		return models.MegaForm{}, false
	}
	mega, ok := forms[normalizeGOImageForm(form)]
	return mega, ok
}

func megaFormsByName(megas map[string]map[string]models.MegaForm, name string) []string {
	forms := megas[normalizeGOImageName(name)]
	if len(forms) == 0 {
		return nil
	}
	out := make([]string, 0, len(forms))
	for _, mega := range forms {
		out = append(out, mega.Form)
	}
	slices.Sort(out)
	return out
}

func availableForms(stats []models.PokemonStats, megas map[string]map[string]models.MegaForm, name string) []string {
	forms := formsByName(stats, name)
	for _, mega := range megaFormsByName(megas, name) {
		if !slices.Contains(forms, mega) {
			forms = append(forms, mega)
		}
	}
	slices.Sort(forms)
	return forms
}

// formatFormsForDisplay replaces underscores so form lists read naturally for users.
func formatFormsForDisplay(forms []string) string {
	labels := make([]string, len(forms))
	for i, form := range forms {
		labels[i] = strings.ReplaceAll(form, "_", " ")
	}
	return strings.Join(labels, ", ")
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

// normalizeGOImageName lowercases a Pokémon name for map keys.
func normalizeGOImageName(name string) string {
	return strings.ToLower(strings.TrimSpace(name))
}

// normalizeGOImageForm lowercases a form key; empty / Normal become "normal".
func normalizeGOImageForm(form string) string {
	if form == "" || strings.EqualFold(form, "Normal") {
		return "normal"
	}
	form = strings.ToLower(strings.TrimSpace(form))
	form = strings.ReplaceAll(form, " ", "_")
	form = strings.ReplaceAll(form, "-", "_")
	return form
}

// buildGOImageMap reduces the full pokedex to name → form → icon URL.
func buildGOImageMap(entries []models.PokedexEntry) map[string]map[string]string {
	images := make(map[string]map[string]string, len(entries))
	for i := range entries {
		name := normalizeGOImageName(entries[i].Names.English)
		if name == "" {
			continue
		}
		forms, ok := images[name]
		if !ok {
			forms = make(map[string]string)
			images[name] = forms
		}
		if entries[i].Assets.Image != "" {
			forms["normal"] = entries[i].Assets.Image
		}
		for _, formSprite := range entries[i].AssetForms {
			if formSprite.Form == nil || *formSprite.Form == "" || formSprite.Image == "" {
				continue
			}
			forms[normalizeGOImageForm(*formSprite.Form)] = formSprite.Image
		}
	}
	return images
}

// grabGOImage finds a GO icon URL for the given pokemon name and form.
// Unknown forms fall back to the default (normal) icon when present.
func grabGOImage(images map[string]map[string]string, name string, form string) string {
	forms := images[normalizeGOImageName(name)]
	if forms == nil {
		return ""
	}
	if img, ok := forms[normalizeGOImageForm(form)]; ok {
		return img
	}
	return forms["normal"]
}

func (c *Client) LookupPokemon(name string, form string) (*models.PokemonProfile, error) {
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
	goImages, err := c.FetchGOImages()
	if err != nil {
		return nil, err
	}
	megaForms, err := c.FetchMegaForms()
	if err != nil {
		return nil, err
	}

	if isMegaForm(form) {
		return c.lookupMegaPokemon(name, form, pokemonStats, pokemonMoves, megaForms)
	}

	pokemon, ok := findStatsByName(pokemonStats, name, form)
	if !ok {
		forms := availableForms(pokemonStats, megaForms, name)
		if form != "" {
			return nil, fmt.Errorf(
				"Form %q was not found for %s. \nAvailable forms: %s",
				form, name, formatFormsForDisplay(forms),
			)
		}
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
		GOImage: grabGOImage(goImages, pokemon.PokemonName, pokemon.Form),
	}, nil
}

func (c *Client) lookupMegaPokemon(name, form string, pokemonStats []models.PokemonStats, pokemonMoves []models.PokemonMoves, megaForms map[string]map[string]models.MegaForm) (*models.PokemonProfile, error) {
	mega, ok := findMegaForm(megaForms, name, form)
	if !ok {
		forms := availableForms(pokemonStats, megaForms, name)
		return nil, fmt.Errorf(
			"Form %q was not found for %s. \nAvailable forms: %s",
			form, name, formatFormsForDisplay(forms),
		)
	}

	base, ok := findStatsByName(pokemonStats, name, "")
	if !ok {
		return nil, fmt.Errorf("stats not found for %s", name)
	}

	// Megas share the base species moveset in GO; use Normal moves.
	move, ok := c.findMoves(pokemonMoves, base.PokemonName, base.PokemonID, "Normal")
	if !ok {
		move, ok = c.findMoves(pokemonMoves, base.PokemonName, base.PokemonID, base.Form)
		if !ok {
			return nil, fmt.Errorf("moves not found for %s", name)
		}
	}

	return &models.PokemonProfile{
		Stats: models.PokemonStats{
			PokemonName: base.PokemonName,
			PokemonID:   base.PokemonID,
			BaseAttack:  mega.BaseAttack,
			BaseDefense: mega.BaseDefense,
			BaseStamina: mega.BaseStamina,
			Form:        mega.Form,
		},
		Moves: models.PokemonMoves{
			PokemonName:       move.PokemonName,
			PokemonID:         move.PokemonID,
			Form:              mega.Form,
			ChargedMoves:      move.ChargedMoves,
			FastMoves:         move.FastMoves,
			EliteChargedMoves: move.EliteChargedMoves,
			EliteFastMoves:    move.EliteFastMoves,
		},
		Types: models.PokemonTypes{
			PokemonName: base.PokemonName,
			PokemonID:   base.PokemonID,
			Form:        mega.Form,
			Type:        mega.Types,
		},
		GOImage: mega.Image,
	}, nil
}
