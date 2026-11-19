package models

import (
	"fmt"
	"log"
	"strconv"
	"strings"
)

// Model represents a Gemini backend routing configuration.
// Mode corresponds to the Gemini upstream mode ID, Think specifies the default reasoning depth,
// and Extra contains index-based overrides for Gemini's sparse RPC request array.
type Model struct {
	Mode   int         `json:"mode"`
	Think  int         `json:"think"`
	Desc   string      `json:"desc"`
	Header string      `json:"header,omitempty"`
	Extra  map[int]any `json:"extra,omitempty"`
}

var MODELS = map[string]Model{
	"gemini-4-carbon": {
		Mode:   8,
		Think:  4,
		Header: "95a4c0abcba24364",
		Desc:   "Frontier model for complex reasoning and agentic tasks (Gemini 4 Carbon, requires Google AI Pro cookie)",
	},
	"gemini-4-carbon-thinking": {
		Mode:   8,
		Think:  0,
		Header: "95a4c0abcba24364",
		Desc:   "Deep thinking mode for Gemini 4 Carbon (~20k chars output)",
	},
	"carbon": {
		Mode:   8,
		Think:  4,
		Header: "95a4c0abcba24364",
		Desc:   "Alias for gemini-4-carbon",
	},
	"gemini-4-argon": {
		Mode:   8,
		Think:  4,
		Header: "95a4c0abcba24364",
		Desc:   "Frontier model for complex tasks (Gemini 4 Argon, requires Google AI Pro cookie)",
	},
	"gemini-4-argon-thinking": {
		Mode:   8,
		Think:  0,
		Header: "95a4c0abcba24364",
		Desc:   "Deep thinking mode for Gemini 4 Argon (~20k chars output)",
	},
	"argon": {
		Mode:   8,
		Think:  4,
		Header: "95a4c0abcba24364",
		Desc:   "Alias for gemini-4-argon",
	},
	"gemini-4": {
		Mode:   8,
		Think:  4,
		Header: "95a4c0abcba24364",
		Desc:   "Alias for gemini-4-argon",
	},
	"gemini-4-thinking": {
		Mode:   8,
		Think:  0,
		Header: "95a4c0abcba24364",
		Desc:   "Alias for frontier thinking model (gemini-4-carbon-thinking)",
	},
	"gemini-3.8-flash": {
		Mode:  1,
		Think: 4,
		Desc:  "Latest all-around model (Gemini 3.8 Flash)",
	},
	"gemini-3.8-flash-thinking": {
		Mode:  2,
		Think: 0,
		Desc:  "Deep thinking mode for Gemini 3.8 Flash (~20k chars output)",
	},
	"gemini-3.7-flash": {
		Mode:  1,
		Think: 4,
		Desc:  "All-around model (Gemini 3.7 Flash)",
	},
	"gemini-3.7-flash-thinking": {
		Mode:  2,
		Think: 0,
		Desc:  "Deep thinking mode for Gemini 3.7 Flash (~20k chars output)",
	},
	"gemini-3.6-flash": {
		Mode:  1,
		Think: 4,
		Desc:  "All-around model (Gemini 3.6 Flash)",
	},
	"gemini-3.5-flash": {
		Mode:  1,
		Think: 4,
		Desc:  "Alias for gemini-3.6-flash (backend upgraded)",
	},
	"gemini-3.5-flash-thinking": {
		Mode:  2,
		Think: 0,
		Desc:  "Deep thinking mode, longest output (~20k chars)",
	},
	"gemini-flash-thinking": {
		Mode:  2,
		Think: 0,
		Desc:  "Alias for latest flash thinking mode (gemini-3.8-flash-thinking)",
	},
	"gemini-flash": {
		Mode:  1,
		Think: 4,
		Desc:  "Alias for latest flash model (gemini-3.8-flash)",
	},
	"gemini-3.1-pro": {
		Mode:  3,
		Think: 4,
		Desc:  "Pro model (requires cookie for real routing)",
	},
	"gemini-3.1-pro-thinking": {
		Mode:  3,
		Think: 0,
		Desc:  "Deep thinking mode for Gemini 3.1 Pro (~20k chars output)",
	},
	"gemini-3.1-pro-enhanced": {
		Mode:  3,
		Think: 4,
		Desc:  "Pro with enhanced output (experimental)",
		Extra: map[int]any{31: 2, 80: 3},
	},
	"gemini-pro": {
		Mode:  3,
		Think: 4,
		Desc:  "Alias for latest pro model (gemini-3.1-pro)",
	},
	"gemini-pro-thinking": {
		Mode:  3,
		Think: 0,
		Desc:  "Alias for gemini-3.1-pro-thinking",
	},
	"gemini-auto": {
		Mode:  4,
		Think: 4,
		Desc:  "Auto model selection",
	},
	"gemini-3.5-flash-thinking-lite": {
		Mode:  5,
		Think: 0,
		Desc:  "Dynamic thinking with adaptive depth",
	},
	"gemini-flash-lite": {
		Mode:  6,
		Think: 4,
		Desc:  "Lightweight fast model",
	},
	// Backward-compatibility and standard aliases
	"gemini-2.5-pro": {
		Mode:  3,
		Think: 4,
		Desc:  "Alias for gemini-3.1-pro",
	},
	"gemini-2.5-flash": {
		Mode:  1,
		Think: 4,
		Desc:  "Alias for gemini-3.8-flash",
	},
	"gemini-2.0-flash": {
		Mode:  1,
		Think: 4,
		Desc:  "Alias for gemini-3.8-flash",
	},
	"gemini-2.0-flash-thinking": {
		Mode:  2,
		Think: 0,
		Desc:  "Alias for gemini-3.8-flash-thinking",
	},
	"gemini-2.0-flash-thinking-exp": {
		Mode:  2,
		Think: 0,
		Desc:  "Alias for gemini-3.8-flash-thinking",
	},
	"gemini-1.5-pro": {
		Mode:  3,
		Think: 4,
		Desc:  "Alias for gemini-3.1-pro",
	},
	"gemini-1.5-flash": {
		Mode:  1,
		Think: 4,
		Desc:  "Alias for gemini-3.8-flash",
	},
}

type Resolved struct {
	Name     string
	Mode     int
	Think    int
	Header   string
	Fallback bool
	Extra    map[int]any
}

const DefaultModelName = "gemini-3.8-flash"

// Resolve maps a requested model string to its corresponding backend configuration.
// It supports model name suffixes like "@think=N" (e.g. "gemini-3.8-flash@think=0")
// to dynamically override the model's default thinking depth without modifying the model mapping.
// It automatically normalizes whitespace, strips "models/" SDK prefixes, and supports case-insensitive lookups.
func Resolve(modelName, defaultName string) (Resolved, error) {
	defaultName = strings.TrimSpace(defaultName)
	defaultName = strings.TrimPrefix(defaultName, "models/")
	if defaultName == "" {
		defaultName = DefaultModelName
	}
	defaultCfg, defaultExists := MODELS[defaultName]
	if !defaultExists {
		if lowerCfg, ok := MODELS[strings.ToLower(defaultName)]; ok {
			defaultName = strings.ToLower(defaultName)
			defaultCfg = lowerCfg
			defaultExists = true
		}
	}
	if !defaultExists {
		defaultName = DefaultModelName
		defaultCfg = MODELS[DefaultModelName]
	}

	modelName = strings.TrimSpace(modelName)
	var thinkOverride *int
	if idx := strings.LastIndex(strings.ToLower(modelName), "@think="); idx != -1 {
		thinkStr := strings.TrimSpace(modelName[idx+len("@think="):])
		modelName = strings.TrimSpace(modelName[:idx])
		val, err := strconv.Atoi(thinkStr)
		if err != nil {
			return Resolved{}, fmt.Errorf("invalid think level: %s", thinkStr)
		}
		thinkOverride = &val
	}

	modelName = strings.TrimPrefix(modelName, "models/")

	cfg, exists := MODELS[modelName]
	if !exists {
		lower := strings.ToLower(modelName)
		if lowerCfg, lowerExists := MODELS[lower]; lowerExists {
			modelName = lower
			cfg = lowerCfg
			exists = true
		}
	}

	fallback := false
	if !exists {
		log.Printf("Unknown model '%s', falling back to '%s'", modelName, defaultName)
		modelName = defaultName
		cfg = defaultCfg
		fallback = true
	}

	thinkMode := cfg.Think
	if thinkOverride != nil {
		thinkMode = *thinkOverride
	}

	return Resolved{
		Name:     modelName,
		Mode:     cfg.Mode,
		Think:    thinkMode,
		Header:   cfg.Header,
		Fallback: fallback,
		Extra:    cfg.Extra,
	}, nil
}
