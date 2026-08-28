package models

// Event is the parsing target for ScrapedDuck event timelines.

type Event struct {
	Name    string `json:"name"`
	Heading string `json:"heading"`
	Start   string `json:"start"`
	End     string `json:"end"`
	Image   string `json:"image"`
	Link    string `json:"link"`
}

// RaidBoss is one entry from ScrapedDuck raids.json (Leek Duck).

type RaidBoss struct {
	Name           string       `json:"name"`
	Tier           string       `json:"tier"`
	CanBeShiny     bool         `json:"canBeShiny"`
	Types          []NamedImage `json:"types"`
	CombatPower    CombatPower  `json:"combatPower"`
	BoostedWeather []NamedImage `json:"boostedWeather"`
	Image          string       `json:"image"`
}

// NamedImage is a labeled asset (type or weather) with an icon URL.

type NamedImage struct {
	Name  string `json:"name"`
	Image string `json:"image"`
}

// CombatPower is the catch CP range for normal and weather-boosted.

type CombatPower struct {
	Normal  CPRange `json:"normal"`
	Boosted CPRange `json:"boosted"`
}

type CPRange struct {
	Min int `json:"min"`
	Max int `json:"max"`
}

// PokemonStats is one entry from the stats index.

type PokemonStats struct {
	PokemonName string `json:"pokemon_name"`
	PokemonID   int    `json:"pokemon_id"`
	BaseAttack  int    `json:"base_attack"`
	BaseDefense int    `json:"base_defense"`
	BaseStamina int    `json:"base_stamina"`
	Form        string `json:"form"`
}

// PokemonMoves is one entry from the current moveset index.

type PokemonMoves struct {
	PokemonName       string   `json:"pokemon_name"`
	PokemonID         int      `json:"pokemon_id"`
	Form              string   `json:"form"`
	ChargedMoves      []string `json:"charged_moves"`
	FastMoves         []string `json:"fast_moves"`
	EliteChargedMoves []string `json:"elite_charged_moves"`
	EliteFastMoves    []string `json:"elite_fast_moves"`
}

type PokemonTypes struct {
	PokemonName string   `json:"pokemon_name"`
	PokemonID   int      `json:"pokemon_id"`
	Form        string   `json:"form"`
	Type        []string `json:"type"`
}

type PokemonProfile struct {
	Stats   PokemonStats
	Moves   PokemonMoves
	Types   PokemonTypes
	GOImage string // GO icon from pokemon-go-api pokedex.json
}

// PokedexAPIEntry is one species from pokemon-go-api (gamemaster + asset URLs).
type PokedexAPIEntry struct {
	Names struct {
		English string `json:"English"`
	} `json:"names"`
	Assets struct {
		Image string `json:"image"`
	} `json:"assets"`
	AssetForms []AssetForm `json:"assetForms"`
}

type AssetForm struct {
	Form  *string `json:"form"`
	Image string  `json:"image"`
}

// TypeEffectiveness is the type-matchup matrix.

type TypeEffectiveness map[string]map[string]float64
