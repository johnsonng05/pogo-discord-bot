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
    if raidBosses == nil {
        t.Fatal("got nil raid bosses")
    }
    if len(raidBosses.Current) == 0 || len(raidBosses.Previous) == 0 {
        t.Fatal("current/previous maps empty — check json tags")
    }

    // Individual tiers are often [] during a rotation. Only require that
    // at least one current tier has bosses with mapped fields.
    found := false
    for _, bosses := range raidBosses.Current {
        if len(bosses) == 0 {
            continue
        }
        found = true
        if bosses[0].Name == "" {
            t.Fatalf("boss name empty — check json tags: %+v", bosses[0])
        }
        break
    }
    if !found {
        t.Fatal("no current raid bosses in any tier")
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