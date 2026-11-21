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

// DynamicModel represents an upstream model discovered at runtime from the
// account roster (otAQ7b). Ticket is the routing hash sent in the
// x-goog-ext-525001261-jspb header and Mode is the upstream mode category.
type DynamicModel struct {
	Name   string
	Ticket string
	Desc   string
	Mode   int
}

var MODELS = map[string]Model{
	"gemini-4-carbon": {
		Mode:   8,
		Think:  4,
		Header: "f44d2958d1f0189a",
		Desc:   "Long-horizon agentic reasoning (Gemini 4 Carbon, Ultra accounts)",
	},
	"gemini-4-carbon-thinking": {
		Mode:   8,
		Think:  0,
		Header: "f44d2958d1f0189a",
		Desc:   "Deep thinking mode for Gemini 4 Carbon (~20k chars output)",
	},
	"carbon": {
		Mode:   8,
		Think:  4,
		Header: "f44d2958d1f0189a",
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
	"gemini-4-neon": {
		Mode:   3,
		Think:  4,
		Header: "e6fa609c3fa255c0",
		Desc:   "Deeper analysis and coding (Gemini 4 Neon, Pro/Ultra accounts)",
	},
	"gemini-4-neon-thinking": {
		Mode:   3,
		Think:  0,
		Header: "e6fa609c3fa255c0",
		Desc:   "Deep thinking mode for Gemini 4 Neon (~20k chars output)",
	},
	"neon": {
		Mode:   3,
		Think:  4,
		Header: "e6fa609c3fa255c0",
		Desc:   "Alias for gemini-4-neon",
	},
	"gemini-4-hydrogen": {
		Mode:   1,
		Think:  4,
		Header: "56fdd199312815e2",
		Desc:   "Everyday coding and agentic workflows (Gemini 4 Hydrogen)",
	},
	"gemini-4-hydrogen-thinking": {
		Mode:   1,
		Think:  0,
		Header: "56fdd199312815e2",
		Desc:   "Thinking mode for Gemini 4 Hydrogen",
	},
	"hydrogen": {
		Mode:   1,
		Think:  4,
		Header: "56fdd199312815e2",
		Desc:   "Alias for gemini-4-hydrogen",
	},
	"gemini-4-thinking": {
		Mode:   8,
		Think:  0,
		Header: "f44d2958d1f0189a",
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
// Optional dynamic models discovered from the account roster are consulted when the name is not
// present in the static registry; static entries always take precedence.
func Resolve(modelName, defaultName string, dynamic ...DynamicModel) (Resolved, error) {
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
		if dm, ok := lookupDynamic(modelName, dynamic); ok {
			thinkMode := 4
			if thinkOverride != nil {
				thinkMode = *thinkOverride
			}
			return Resolved{
				Name:   dm.Name,
				Mode:   dm.Mode,
				Think:  thinkMode,
				Header: dm.Ticket,
				Extra:  nil,
			}, nil
		}
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

func lookupDynamic(modelName string, dynamic []DynamicModel) (DynamicModel, bool) {
	name := strings.ToLower(modelName)
	for _, dm := range dynamic {
		candidate := strings.ToLower(dm.Name)
		if candidate == name || strings.TrimPrefix(candidate, "gemini-") == name {
			return dm, true
		}
	}
	return DynamicModel{}, false
}

// DynamicNameFromDisplay converts an upstream display label such as "4 Hydrogen"
// or "3.8 Flash" into an API model name like "gemini-4-hydrogen".
func DynamicNameFromDisplay(display string) (string, bool) {
	fields := strings.Fields(strings.TrimSpace(display))
	if len(fields) != 2 {
		return "", false
	}
	version := fields[0]
	if version[0] < '0' || version[0] > '9' {
		return "", false
	}
	label := strings.ToLower(fields[1])
	label = strings.ReplaceAll(label, " ", "-")
	return "gemini-" + version + "-" + label, true
}
