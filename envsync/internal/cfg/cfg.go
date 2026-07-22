package cfg

import (
	"fmt"
	"reflect"
	"strconv"

	"github.com/jackc/pgx/v5"
)

type Cfg struct {
	DB_URL          string `env:"DB_URL"`
	MAX_CONNECTIONS uint16 `env:"MAX_CONNECTIONS"`
	CACHE_ENABLED   bool   `env:"CACHE_ENABLED"`
}

func Parse(raw map[string]string) (Cfg, error) {
	var result Cfg

	// reflect for dynamic mapping by inspecting struct at runtime
	// using this we can loop through any struct, read their env tags
	// and map raw data to the struct without hardcoding every field name
	v := reflect.ValueOf(&result).Elem()
	t := v.Type()

	for i := range t.NumField() {
		field := t.Field(i)
		key := field.Tag.Get("env")
		if key == "" {
			continue
		}

		valStr, exists := raw[key]
		if !exists {
			return result, fmt.Errorf("missing required config key: %s", key)
		}

		fieldVal := v.Field(i)
		switch fieldVal.Kind() {
		case reflect.String:
			fieldVal.SetString(valStr)

		case reflect.Bool:
			b, err := strconv.ParseBool(valStr)
			if err != nil {
				return result, fmt.Errorf("invalid %s: %s", key, valStr)
			}
			fieldVal.SetBool(b)

		case reflect.Uint16:
			n, err := strconv.ParseUint(valStr, 10, 16)
			if err != nil {
				return result, fmt.Errorf("invalid %s: %s", key, valStr)
			}
			fieldVal.SetUint(n)

		default:
			return result, fmt.Errorf("unsupported struct type %s for key %s",
				fieldVal.Kind(), key)
		}
	}

	// ensure DB_URL is a valid postgres string
	if _, err := pgx.ParseConfig(result.DB_URL); err != nil {
		return result, fmt.Errorf("invalid DB_URL format: %s", result.DB_URL)
	}

	return result, nil
}
