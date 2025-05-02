package jqutil

import (
	"context"
	"encoding/json"
	"fmt"
	"strconv"
	"strings"

	"github.com/itchyny/gojq"
)

type EvalOptions struct {
	Query   string
	Unquote bool
	Data    any
}

func Eval(ctx context.Context, opts EvalOptions) (string, error) {
	enc := newEncoder(false, 0)

	query, err := gojq.Parse(opts.Query)
	if err != nil {
		return "", fmt.Errorf("invalid jq query %q: %w", opts.Query, err)
	}

	iter := query.RunWithContext(ctx, opts.Data)
	for {
		v, ok := iter.Next()
		if !ok {
			break
		}
		if err, ok := v.(error); ok {
			return "", err
		}
		if err := enc.encode(v); err != nil {
			return "", err
		}
	}

	res := enc.w.String()
	if opts.Unquote {
		unq, err := strconv.Unquote(res)
		if err == nil {
			res = unq
		}
	}

	return res, nil
}

func ForEach(ctx context.Context, opts EvalOptions, action func(any) error) error {
	res, err := Eval(ctx, opts)
	if err != nil {
		return err
	}

	var tmp any
	if err := json.Unmarshal([]byte(res), &tmp); err != nil {
		return err
	}

	items, ok := tmp.([]any)
	if !ok {
		return fmt.Errorf("query %q must return a JSON array", opts.Query)
	}

	for _, el := range items {
		if err := action(el); err != nil {
			return err
		}
	}

	return nil
}

/*
func MaybeQuery(s string) (string, bool) {
	start := strings.Index(s, "${")
	if start == -1 {
		return s, false
	}

	start += len("${")
	end := strings.LastIndexByte(s[start:], '}')
	if end == -1 {
		return s, false
	}

	return strings.TrimSpace(s[start : start+end]), true
}
*/

func MaybeQuery(s string) (string, bool) {
	start := strings.Index(s, "${")
	if start == -1 {
		return s, false
	}
	start += len("${")

	bracketCount := 1
	end := start

	for end < len(s) {
		switch s[end] {
		case '{':
			bracketCount++
		case '}':
			bracketCount--
			if bracketCount == 0 {
				return strings.TrimSpace(s[start:end]), true
			}
		}
		end++
	}

	return s, false
}

func Extract(ctx context.Context, opts EvalOptions) (map[string]any, error) {
	res, err := Eval(ctx, opts)
	if err != nil {
		return map[string]any{}, err
	}

	var tmp any
	if err := json.Unmarshal([]byte(res), &tmp); err != nil {
		return map[string]any{}, err
	}

	if arr, ok := tmp.([]any); ok {
		return map[string]any{
			"items": arr,
		}, nil
	}

	return map[string]any{}, fmt.Errorf("query %q must return a JSON array", opts.Query)
}
