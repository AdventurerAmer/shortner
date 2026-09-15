package config

import (
	"encoding/json"
	"fmt"
)

type Env string

const (
	EnvLocal      Env = "local"
	EnvStaging    Env = "staging"
	EnvProduction Env = "production"
)

var EnvValues = []Env{EnvLocal, EnvStaging, EnvProduction}

func (e Env) String() string {
	return string(e)
}

func (e Env) MarshalJSON() ([]byte, error) {
	return json.Marshal(e.String())
}

func (e *Env) UnmarshalJSON(data []byte) error {
	var s string
	if err := json.Unmarshal(data, &s); err != nil {
		return err
	}

	switch s {
	case "local":
		*e = EnvLocal
	case "staging":
		*e = EnvStaging
	case "production":
		*e = EnvProduction
	default:
		return fmt.Errorf("invalid env: %s", s)
	}

	return nil
}
