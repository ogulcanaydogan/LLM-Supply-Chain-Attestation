package hash

import (
	"bytes"
	"crypto/sha256"
	"encoding/hex"
	"encoding/json"
	"fmt"
	"io"
	"sort"
	"strconv"
)

func CanonicalJSON(v any) ([]byte, error) {
	input, err := json.Marshal(v)
	if err != nil {
		return nil, fmt.Errorf("marshal for canonicalization: %w", err)
	}

	dec := json.NewDecoder(bytes.NewReader(input))
	dec.UseNumber()
	var normalized any
	if err := dec.Decode(&normalized); err != nil {
		return nil, fmt.Errorf("decode for canonicalization: %w", err)
	}

	buf := &bytes.Buffer{}
	if err := writeCanonical(buf, normalized); err != nil {
		return nil, err
	}
	return buf.Bytes(), nil
}

func HashCanonicalJSON(v any) (string, []byte, error) {
	canonical, err := CanonicalJSON(v)
	if err != nil {
		return "", nil, err
	}
	sum := sha256.Sum256(canonical)
	return "sha256:" + hex.EncodeToString(sum[:]), canonical, nil
}

func writeCanonical(w io.Writer, v any) error {
	switch vv := v.(type) {
	case nil:
		_, err := io.WriteString(w, "null")
		return err
	case bool:
		if vv {
			_, err := io.WriteString(w, "true")
			return err
		}
		_, err := io.WriteString(w, "false")
		return err
	case string:
		b, err := json.Marshal(vv)
		if err != nil {
			return err
		}
		_, err = w.Write(b)
		return err
	case json.Number:
		return writeNumber(w, vv.String())
	case float64:
		return writeNumber(w, strconv.FormatFloat(vv, 'f', -1, 64))
	case []any:
		if _, err := io.WriteString(w, "["); err != nil {
			return err
		}
		for i, item := range vv {
			if i > 0 {
				if _, err := io.WriteString(w, ","); err != nil {
					return err
				}
			}
			if err := writeCanonical(w, item); err != nil {
				return err
			}
		}
		_, err := io.WriteString(w, "]")
		return err
	case map[string]any:
		keys := make([]string, 0, len(vv))
		for k := range vv {
			keys = append(keys, k)
		}
		sort.Slice(keys, func(i, j int) bool {
			return keys[i] < keys[j]
		})
		if _, err := io.WriteString(w, "{"); err != nil {
			return err
		}
		for i, k := range keys {
			if i > 0 {
				if _, err := io.WriteString(w, ","); err != nil {
					return err
				}
			}
			kb, err := json.Marshal(k)
			if err != nil {
				return err
			}
			if _, err := w.Write(kb); err != nil {
				return err
			}
			if _, err := io.WriteString(w, ":"); err != nil {
				return err
			}
			if err := writeCanonical(w, vv[k]); err != nil {
				return err
			}
		}
		_, err := io.WriteString(w, "}")
		return err
	default:
		b, err := json.Marshal(v)
		if err != nil {
			return err
		}
		var normalized any
		dec := json.NewDecoder(bytes.NewReader(b))
		dec.UseNumber()
		if err := dec.Decode(&normalized); err != nil {
			return err
		}
		return writeCanonical(w, normalized)
	}
}

func writeNumber(w io.Writer, n string) error {
	if _, err := strconv.ParseFloat(n, 64); err != nil {
		return fmt.Errorf("invalid number %q: %w", n, err)
	}
	_, err := io.WriteString(w, n)
	return err
}
