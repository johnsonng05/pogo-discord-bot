package models

import (
	"bytes"
	"encoding/json"
)

type AssetForm struct {
	Form  *string `json:"form"`
	Image string  `json:"image"`
}

// PokedexEntry is one species from pokemon-go-api (gamemaster + asset URLs).
type PokedexEntry struct {
	Names struct {
		English string `json:"English"`
	} `json:"names"`
	Assets struct {
		Image string `json:"image"`
	} `json:"assets"`
	AssetForms     []AssetForm      `json:"assetForms"`
	MegaEvolutions MegaEvolutionMap `json:"megaEvolutions"`
}

// MegaEvolutionMap accepts either a mega id → mega object map, or [] when empty.
type MegaEvolutionMap map[string]MegaEvolution

func (megaMap *MegaEvolutionMap) UnmarshalJSON(data []byte) error {
	trimmed := bytes.TrimSpace(data)
	if len(trimmed) == 0 || string(trimmed) == "null" || string(trimmed) == "[]" {
		*megaMap = nil
		return nil
	}
	if trimmed[0] == '[' {
		*megaMap = nil
		return nil
	}
	var output map[string]MegaEvolution
	if err := json.Unmarshal(data, &output); err != nil {
		return err
	}
	*megaMap = output
	return nil
}

// MegaEvolution is one mega from pokemon-go-api megaEvolutions.
type MegaEvolution struct {
	ID    string `json:"id"`
	Names struct {
		English string `json:"English"`
	} `json:"names"`
	Stats struct {
		Attack  int `json:"attack"`
		Defense int `json:"defense"`
		Stamina int `json:"stamina"`
	} `json:"stats"`
	PrimaryType struct {
		Names struct {
			English string `json:"English"`
		} `json:"names"`
	} `json:"primaryType"`
	SecondaryType *struct {
		Names struct {
			English string `json:"English"`
		} `json:"names"`
	} `json:"secondaryType"`
	Assets struct {
		Image string `json:"image"`
	} `json:"assets"`
}

// MegaForm is a slim cached mega row (name → form → MegaForm).
type MegaForm struct {
	Form        string   `json:"form"` // Mega, Mega_X, Mega_Y
	BaseAttack  int      `json:"base_attack"`
	BaseDefense int      `json:"base_defense"`
	BaseStamina int      `json:"base_stamina"`
	Types       []string `json:"types"`
	Image       string   `json:"image"`
}
