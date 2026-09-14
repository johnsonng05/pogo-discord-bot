package api

import (
	"context"
	"encoding/json"
	"os"
	"testing"

	"pogo-discord-bot/internal/cache"
	"pogo-discord-bot/internal/models"
)

func TestFetchEventsLive(t *testing.T) {
	client := New(nil) // nil cache — live HTTP tests skip Redis
	events, err := client.FetchEvents()
	if err != nil {
		t.Fatal(err)
	}
	if len(events) == 0 {
		t.Fatal("got 0 events")
	}
	if events[0].Name == "" || events[0].Heading == "" || events[0].Start == "" || events[0].End == "" {
		t.Fatalf("fields empty — check json tags: %+v", events[0])
	}
}

func TestFetchRaidBosses(t *testing.T) {
	client := New(nil) // nil cache — live HTTP tests skip Redis
	raidBosses, err := client.FetchRaidBosses()
	if err != nil {
		t.Fatal(err)
	}
	if len(raidBosses) == 0 {
		t.Fatal("got 0 raid bosses")
	}

	first := raidBosses[0]
	if first.Name == "" || first.Tier == "" {
		t.Fatalf("identity fields empty — check json tags: %+v", first)
	}
	if len(first.Types) == 0 || first.Types[0].Name == "" {
		t.Fatalf("types empty — check json tags: %+v", first)
	}
	if first.CombatPower.Normal.Max == 0 {
		t.Fatalf("combatPower empty — check json tags: %+v", first)
	}
}

func TestFetchPokemonStats(t *testing.T) {
	client := New(nil) // nil cache — live HTTP tests skip Redis
	pokemonStats, err := client.FetchPokemonStats()
	if err != nil {
		t.Fatal(err)
	}
	if len(pokemonStats) == 0 {
		t.Fatal("got 0 pokemon stats")
	}
	if pokemonStats[0].PokemonName == "" || pokemonStats[0].PokemonID == 0 || pokemonStats[0].BaseAttack == 0 || pokemonStats[0].BaseDefense == 0 || pokemonStats[0].BaseStamina == 0 || pokemonStats[0].Form == "" {
		t.Fatalf("fields empty — check json tags: %+v", pokemonStats[0])
	}
}

func TestFetchPokemonMoves(t *testing.T) {
	client := New(nil) // nil cache — live HTTP tests skip Redis
	pokemonMoves, err := client.FetchPokemonMoves()
	if err != nil {
		t.Fatal(err)
	}
	if len(pokemonMoves) == 0 {
		t.Fatal("got 0 pokemon moves")
	}

	first := pokemonMoves[0]
	if first.PokemonName == "" || first.PokemonID == 0 || first.Form == "" {
		t.Fatalf("identity fields empty — check json tags: %+v", first)
	}

	// Elite lists are often []. Fast/charged are usually filled, but do not
	// require that on index 0 — scan for one row that has a moveset.
	hasMoveset := false
	for _, row := range pokemonMoves {
		if len(row.FastMoves) > 0 || len(row.ChargedMoves) > 0 {
			hasMoveset = true
			break
		}
	}
	if !hasMoveset {
		t.Fatal("no fast/charged moves on any row — check json tags")
	}
}

func TestFetchTypeEffectiveness(t *testing.T) {
	client := New(nil) // nil cache — live HTTP tests skip Redis
	typeEffectiveness, err := client.FetchTypeEffectiveness()
	if err != nil {
		t.Fatal(err)
	}
	if typeEffectiveness == nil {
		t.Fatal("got nil type effectiveness")
	}
}

func TestLookupPokemon(t *testing.T) {
	client := New(nil) // nil cache — live HTTP tests skip Redis
	pokemonProfile, err := client.LookupPokemon("Pikachu", "")
	if err != nil {
		t.Fatal(err)
	}
	if pokemonProfile == nil {
		t.Fatal("got nil pokemonProfile")
	}
	if pokemonProfile.Stats.PokemonName != "Pikachu" || pokemonProfile.Stats.PokemonID != 25 || pokemonProfile.Stats.Form != "Normal" {
		t.Fatalf("pokemonProfile fields empty — check json tags: %+v", pokemonProfile)
	}
	if pokemonProfile.Moves.PokemonName != "Pikachu" || pokemonProfile.Moves.PokemonID != 25 || pokemonProfile.Moves.Form != "Normal" {
		t.Fatalf("pokemonProfile fields empty — check json tags: %+v", pokemonProfile)
	}
	if pokemonProfile.Types.PokemonName != "Pikachu" || pokemonProfile.Types.PokemonID != 25 || pokemonProfile.Types.Form != "Normal" {
		t.Fatalf("pokemonProfile fields empty — check json tags: %+v", pokemonProfile)
	}
	if len(pokemonProfile.Moves.FastMoves) == 0 || len(pokemonProfile.Moves.ChargedMoves) == 0 {
		t.Fatalf("pokemonProfile fields empty — check json tags: %+v", pokemonProfile)
	}
	if pokemonProfile.Stats.BaseAttack == 0 || pokemonProfile.Stats.BaseDefense == 0 || pokemonProfile.Stats.BaseStamina == 0 {
		t.Fatalf("pokemonProfile fields empty — check json tags: %+v", pokemonProfile)
	}
	if len(pokemonProfile.Types.Type) == 0 {
		t.Fatalf("pokemonProfile fields empty — check json tags: %+v", pokemonProfile)
	}
	if pokemonProfile.GOImage == "" {
		t.Fatalf("pokemonProfile fields empty — check json tags: %+v", pokemonProfile)
	}
}

func TestBuildGOImageMapAndGrab(t *testing.T) {
	fall := "FALL_2019"
	entries := []models.PokedexEntry{
		{
			Names: struct {
				English string `json:"English"`
			}{English: "Pikachu"},
			Assets: struct {
				Image string `json:"image"`
			}{Image: "https://example.com/pikachu.png"},
			AssetForms: []models.AssetForm{
				{Form: &fall, Image: "https://example.com/pikachu-fall.png"},
			},
		},
	}

	images := buildGOImageMap(entries)
	if got := grabGOImage(images, "pikachu", "Normal"); got != "https://example.com/pikachu.png" {
		t.Fatalf("normal form: got %q", got)
	}
	if got := grabGOImage(images, "Pikachu", "Fall_2019"); got != "https://example.com/pikachu-fall.png" {
		t.Fatalf("named form: got %q", got)
	}
	if forms := images["pikachu"]; forms["normal"] == "" || forms["fall_2019"] == "" {
		t.Fatalf("expected nested form keys, got %#v", forms)
	}
	if got := grabGOImage(images, "Pikachu", "Unknown_Form"); got != "https://example.com/pikachu.png" {
		t.Fatalf("fallback: got %q", got)
	}
	if got := grabGOImage(images, "MissingNo", "Normal"); got != "" {
		t.Fatalf("missing: got %q", got)
	}
}

func TestBuildMegaFormMap(t *testing.T) {
	const raw = `[
		{
			"names": {"English": "Charizard"},
			"megaEvolutions": {
				"CHARIZARD_MEGA_X": {
					"id": "CHARIZARD_MEGA_X",
					"stats": {"attack": 273, "defense": 213, "stamina": 186},
					"primaryType": {"names": {"English": "Fire"}},
					"secondaryType": {"names": {"English": "Dragon"}},
					"assets": {"image": "https://example.com/charizard-megax.png"}
				},
				"CHARIZARD_MEGA_Y": {
					"id": "CHARIZARD_MEGA_Y",
					"stats": {"attack": 319, "defense": 212, "stamina": 186},
					"primaryType": {"names": {"English": "Fire"}},
					"secondaryType": {"names": {"English": "Flying"}},
					"assets": {"image": "https://example.com/charizard-megay.png"}
				}
			}
		},
		{
			"names": {"English": "Venusaur"},
			"megaEvolutions": {
				"VENUSAUR_MEGA": {
					"id": "VENUSAUR_MEGA",
					"stats": {"attack": 241, "defense": 246, "stamina": 190},
					"primaryType": {"names": {"English": "Grass"}},
					"secondaryType": {"names": {"English": "Poison"}},
					"assets": {"image": "https://example.com/venusaur-mega.png"}
				}
			}
		}
	]`
	var entries []models.PokedexEntry
	if err := json.Unmarshal([]byte(raw), &entries); err != nil {
		t.Fatal(err)
	}

	megas := buildMegaFormMap(entries)
	x, ok := findMegaForm(megas, "Charizard", "Mega_X")
	if !ok {
		t.Fatal("expected Mega_X for Charizard")
	}
	if x.Form != "Mega_X" || x.BaseAttack != 273 || x.BaseDefense != 213 || x.BaseStamina != 186 {
		t.Fatalf("Mega_X stats: %+v", x)
	}
	if len(x.Types) != 2 || x.Types[0] != "Fire" || x.Types[1] != "Dragon" {
		t.Fatalf("Mega_X types: %#v", x.Types)
	}
	if x.Image == "" {
		t.Fatal("expected Mega_X image")
	}

	if _, ok := findMegaForm(megas, "Charizard", "mega x"); !ok {
		t.Fatal("expected spaced mega x to normalize")
	}
	if _, ok := findMegaForm(megas, "Venusaur", "Mega"); !ok {
		t.Fatal("expected Mega for Venusaur")
	}
	if _, ok := findMegaForm(megas, "Pikachu", "Mega"); ok {
		t.Fatal("unexpected mega for Pikachu")
	}
}

func TestFormatFormsForDisplay(t *testing.T) {
	got := formatFormsForDisplay([]string{"Normal", "Fall_2019", "Mega_X", "Copy_2019"})
	want := "Normal, Fall 2019, Mega X, Copy 2019"
	if got != want {
		t.Fatalf("got %q, want %q", got, want)
	}
}

func TestLookupPokemonMega(t *testing.T) {
	client := New(nil)
	profile, err := client.LookupPokemon("Charizard", "Mega_X")
	if err != nil {
		t.Fatal(err)
	}
	if profile.Stats.Form != "Mega_X" {
		t.Fatalf("form: got %q", profile.Stats.Form)
	}
	if profile.Stats.PokemonID != 6 || profile.Stats.PokemonName != "Charizard" {
		t.Fatalf("identity: %+v", profile.Stats)
	}
	if profile.Stats.BaseAttack < 250 {
		t.Fatalf("expected mega attack boost, got %d", profile.Stats.BaseAttack)
	}
	if len(profile.Types.Type) == 0 {
		t.Fatal("expected mega types")
	}
	if len(profile.Moves.FastMoves) == 0 || len(profile.Moves.ChargedMoves) == 0 {
		t.Fatal("expected base moves on mega profile")
	}
	if profile.Moves.Form != "Mega_X" {
		t.Fatalf("moves form: got %q", profile.Moves.Form)
	}
	if profile.GOImage == "" {
		t.Fatal("expected mega image")
	}

	_, err = client.LookupPokemon("Pikachu", "Mega")
	if err == nil {
		t.Fatal("expected error for Pikachu Mega")
	}
}

func TestFetchGOImagesCachesSlimPayload(t *testing.T) {
	redisURL := os.Getenv("REDIS_URL")
	if redisURL == "" {
		t.Skip("REDIS_URL not set")
	}

	rdb, err := cache.New(redisURL)
	if err != nil {
		t.Fatal(err)
	}
	defer rdb.Close()

	ctx := context.Background()
	if err := rdb.Delete(ctx, cache.KeyGOImages, cache.KeyMegaForms); err != nil {
		t.Fatal(err)
	}
	t.Cleanup(func() { _ = rdb.Delete(ctx, cache.KeyGOImages, cache.KeyMegaForms) })

	client := New(rdb)
	images, err := client.FetchGOImages()
	if err != nil {
		t.Fatal(err)
	}
	if len(images) == 0 {
		t.Fatal("got 0 go images")
	}
	if img := grabGOImage(images, "Pikachu", "Normal"); img == "" {
		t.Fatal("expected Pikachu icon URL")
	}

	cached, err := rdb.GetCached(ctx, cache.KeyGOImages)
	if err != nil {
		t.Fatal(err)
	}
	// Full pokedex is ~14MB; slim map must stay under Upstash's 10MB request limit.
	const maxBytes = 2 * 1024 * 1024
	if len(cached) > maxBytes {
		t.Fatalf("cached go_images too large: %d bytes (limit %d)", len(cached), maxBytes)
	}

	megaCached, err := rdb.GetCached(ctx, cache.KeyMegaForms)
	if err != nil {
		t.Fatal(err)
	}
	if len(megaCached) == 0 {
		t.Fatal("expected mega_forms to be warmed with go_images")
	}
	if len(megaCached) > maxBytes {
		t.Fatalf("cached mega_forms too large: %d bytes (limit %d)", len(megaCached), maxBytes)
	}

	again, err := client.FetchGOImages()
	if err != nil {
		t.Fatal(err)
	}
	if len(again) != len(images) {
		t.Fatalf("cache hit size %d != first fetch %d", len(again), len(images))
	}
}

func TestFetchEventsCachesWithTTL(t *testing.T) {
	redisURL := os.Getenv("REDIS_URL")
	if redisURL == "" {
		t.Skip("REDIS_URL not set")
	}

	rdb, err := cache.New(redisURL)
	if err != nil {
		t.Fatal(err)
	}
	defer rdb.Close()

	ctx := context.Background()
	if err := rdb.Delete(ctx, cache.KeyEvents); err != nil {
		t.Fatal(err)
	}
	t.Cleanup(func() { _ = rdb.Delete(ctx, cache.KeyEvents) })

	client := New(rdb)

	events, err := client.FetchEvents()
	if err != nil {
		t.Fatal(err)
	}
	if len(events) == 0 {
		t.Fatal("got 0 events")
	}

	remaining, err := rdb.TTL(ctx, cache.KeyEvents)
	if err != nil {
		t.Fatal(err)
	}
	if remaining <= 0 {
		t.Fatalf("expected positive TTL on %s, got %v", cache.KeyEvents, remaining)
	}
	if remaining > cache.DefaultTTL {
		t.Fatalf("TTL %v exceeds DefaultTTL %v", remaining, cache.DefaultTTL)
	}

	// Second fetch should be a cache hit (key still present with TTL counting down).
	eventsAgain, err := client.FetchEvents()
	if err != nil {
		t.Fatal(err)
	}
	if len(eventsAgain) != len(events) {
		t.Fatalf("cached fetch length %d != first fetch %d", len(eventsAgain), len(events))
	}

	cached, err := rdb.GetCached(ctx, cache.KeyEvents)
	if err != nil {
		t.Fatal(err)
	}
	if len(cached) == 0 {
		t.Fatal("expected cached JSON bytes")
	}

	after, err := rdb.TTL(ctx, cache.KeyEvents)
	if err != nil {
		t.Fatal(err)
	}
	if after <= 0 {
		t.Fatalf("expected key to remain cached after second fetch, TTL=%v", after)
	}
}
