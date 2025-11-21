package strings

import (
	"fmt"
	"math/rand"
	"strings"
)

func RandomName(prefix string, rng *rand.Rand) string {
	length := rng.Intn(maxLength-minLength+1) + minLength

	gen := newNameGenerator(prefix, rng)
	return gen.Rnd(length)
}

const (
	letters   = "abcdefghijklmnopqrstuvwxyzABCDEFGHIJKLMNOPQRSTUVWXYZ"
	minLength = 9
	maxLength = 14
)

var (
	reservedKeywords = map[string]bool{
		"break": true, "default": true, "func": true, "interface": true,
		"select": true, "case": true, "defer": true, "go": true, "map": true,
		"struct": true, "chan": true, "else": true, "goto": true, "package": true,
		"switch": true, "const": true, "fallthrough": true, "if": true, "range": true,
		"type": true, "continue": true, "for": true, "import": true, "return": true,
		"var": true,
	}
)

func newNameGenerator(prefix string, rng *rand.Rand) *randomNameGenerator {
	return &randomNameGenerator{
		//rng: rand.New(rand.NewSource(time.Now().UnixNano())),
		rng:    rng,
		prefix: prefix,
	}
}

type randomNameGenerator struct {
	rng    *rand.Rand
	prefix string
}

func (r *randomNameGenerator) Rnd(length int) string {
	b := make([]byte, length)
	b[0] = letters[rand.Intn(len(letters))]
	for i := 1; i < length; i++ {
		b[i] = letters[rand.Intn(len(letters))]
	}

	name := fmt.Sprintf("%s_%s_%d", r.prefix, string(b), rand.Intn(100000))
	name = strings.ReplaceAll(name, "-", "_")

	if reservedKeywords[strings.ToLower(name)] {
		name += "_X"
	}

	return name
}
