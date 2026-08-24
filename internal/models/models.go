package models

// Event is the parsing target for ScrapedDuck event timelines.

type Event struct {
	Name string `json:"name"`
	Heading string `json:"heading"`
	Start string `json:"start"`
	End string `json:"end"`
	Image string `json:"image"`
	Link string `json:"link"`
}

// RaidBosses is the parsing target for live raid tier bosses.

type RaidBosses struct {
	Current map[string][]RaidBoss `json:"current"`
	Previous map[string][]RaidBoss `json:"previous"`
}

type RaidBoss struct {
	Name string `json:"name"`
	Type []string `json:"type"`
	PossibleShiny bool `json:"possible_shiny"`
	BoostedWeather []string `json:"boosted_weather"`
	Form string `json:"form"`
}

// PokemonStats is one entry from the stats index.

type PokemonStats struct {
	PokemonName string `json:"pokemon_name"`
	PokemonID int `json:"pokemon_id"`
	BaseAttack int `json:"base_attack"`
	BaseDefense int `json:"base_defense"`
	BaseStamina int `json:"base_stamina"`
	Form string `json:"form"`
}

// PokemonMoves is one entry from the current moveset index.

type PokemonMoves struct {
	PokemonName string `json:"pokemon_name"`
	PokemonID int `json:"pokemon_id"`
	Form string `json:"form"`
	ChargedMoves []string `json:"charged_moves"`
	FastMoves []string `json:"fast_moves"`
	EliteChargedMoves []string `json:"elite_charged_moves"`
	EliteFastMoves []string `json:"elite_fast_moves"`
}

// TypeEffectiveness is the type-matchup matrix.

type TypeEffectiveness map[string]map[string]float64