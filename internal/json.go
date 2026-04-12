package internal

import (
	"encoding/json"
	"fmt"
	"io"
)

func decodeJSON(r io.Reader, v any) error {
	if err := json.NewDecoder(r).Decode(v); err != nil {
		return fmt.Errorf("json decode: %w", err)
	}
	return nil
}
