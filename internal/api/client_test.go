package api

import "testing"

func TestFetchEventsLive(t *testing.T) {
	client := New()
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
	client := New()
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
	client := New()
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
	client := New()
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
	client := New()
	typeEffectiveness, err := client.FetchTypeEffectiveness()
	if err != nil {
		t.Fatal(err)
	}
	if typeEffectiveness == nil {
		t.Fatal("got nil type effectiveness")
	}
}

func TestLookupPokemon(t *testing.T) {
	client := New()
	pokemonProfile, err := client.LookupPokemon("Pikachu")
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
