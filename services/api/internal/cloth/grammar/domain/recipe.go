package domain

import (
	"crypto/sha256"
	"encoding/hex"
	"errors"
	"fmt"
	"regexp"
	"slices"
	"strconv"
	"strings"
)

const (
	VersionV1     = "cloth.v1"
	MaxThemes     = 8
	MaxProvenance = 8
)

var (
	ErrUnknownVersion  = errors.New("unknown cloth grammar version")
	ErrInvalidInput    = errors.New("invalid cloth grammar input")
	ErrDuplicateInput  = errors.New("duplicate cloth grammar input")
	ErrCommandMismatch = errors.New("cloth grammar command replay mismatch")
)
var opaque = regexp.MustCompile(`^[a-f0-9]{64}$`)

type Input struct {
	Version                   string
	PairKeys                  [2]string
	ThemeKeys, ProvenanceKeys []string
}
type Token struct{ Name, Value string }
type Recipe struct {
	id, version, renderSeed, commandID, fingerprint string
	pair                                            [2]string
	themes, provenance                              []string
	tokens                                          []Token
	revision                                        uint64
}

func Compile(input Input, commandID string) (Recipe, error) {
	if input.Version != VersionV1 {
		return Recipe{}, ErrUnknownVersion
	}
	if commandID == "" || !validPair(input.PairKeys) || len(input.ThemeKeys) < 1 || len(input.ThemeKeys) > MaxThemes || len(input.ProvenanceKeys) < 1 || len(input.ProvenanceKeys) > MaxProvenance {
		return Recipe{}, ErrInvalidInput
	}
	themes, ok := normalize(input.ThemeKeys)
	if !ok {
		return Recipe{}, ErrDuplicateInput
	}
	provenance, ok := normalize(input.ProvenanceKeys)
	if !ok {
		return Recipe{}, ErrDuplicateInput
	}
	pair := input.PairKeys
	if !opaque.MatchString(pair[0]) || !opaque.MatchString(pair[1]) || pair[0] == pair[1] {
		return Recipe{}, ErrInvalidInput
	}
	if pair[1] < pair[0] {
		pair[0], pair[1] = pair[1], pair[0]
	}
	canonical := strings.Join(append(append([]string{VersionV1, pair[0], pair[1]}, themes...), provenance...), "\x00")
	sum := sha256.Sum256([]byte(canonical))
	seed := hex.EncodeToString(sum[:])
	fingerprintSum := sha256.Sum256([]byte(commandID + "\x00" + canonical))
	tokens := tokensV1(sum)
	return Recipe{id: "recipe:" + seed, version: VersionV1, renderSeed: seed, commandID: commandID, fingerprint: hex.EncodeToString(fingerprintSum[:]), pair: pair, themes: themes, provenance: provenance, tokens: tokens, revision: 1}, nil
}
func Rehydrate(id, version, renderSeed, commandID, fingerprint string, pair [2]string, themes, provenance []string, tokens []Token, revision uint64) (Recipe, error) {
	compiled, err := Compile(Input{version, pair, themes, provenance}, commandID)
	if err != nil {
		return Recipe{}, err
	}
	if revision != 1 || compiled.id != id || compiled.renderSeed != renderSeed || compiled.fingerprint != fingerprint || !slices.Equal(compiled.tokens, tokens) {
		return Recipe{}, ErrInvalidInput
	}
	return compiled, nil
}
func normalize(values []string) ([]string, bool) {
	out := append([]string(nil), values...)
	for _, value := range out {
		if !opaque.MatchString(value) {
			return nil, false
		}
	}
	slices.Sort(out)
	for i := 1; i < len(out); i++ {
		if out[i] == out[i-1] {
			return nil, false
		}
	}
	return out, true
}
func validPair(pair [2]string) bool {
	return opaque.MatchString(pair[0]) && opaque.MatchString(pair[1]) && pair[0] != pair[1]
}
func tokensV1(sum [32]byte) []Token {
	directions := []string{"horizontal", "vertical", "diagonal"}
	motifs := []string{"circle", "diamond", "wave", "knot"}
	return []Token{{"palette.primary", "#" + hex.EncodeToString(sum[0:3])}, {"palette.secondary", "#" + hex.EncodeToString(sum[3:6])}, {"pattern.density", strconv.Itoa(int(sum[6]%5) + 1)}, {"weave.direction", directions[int(sum[7])%len(directions)]}, {"motif.shape", motifs[int(sum[8])%len(motifs)]}, {"border.weight", fmt.Sprintf("%d", int(sum[9]%3)+1)}}
}
func (r Recipe) ID() string           { return r.id }
func (r Recipe) Version() string      { return r.version }
func (r Recipe) RenderSeed() string   { return r.renderSeed }
func (r Recipe) CommandID() string    { return r.commandID }
func (r Recipe) Fingerprint() string  { return r.fingerprint }
func (r Recipe) Pair() [2]string      { return r.pair }
func (r Recipe) Themes() []string     { return append([]string(nil), r.themes...) }
func (r Recipe) Provenance() []string { return append([]string(nil), r.provenance...) }
func (r Recipe) Tokens() []Token      { return append([]Token(nil), r.tokens...) }
func (r Recipe) Revision() uint64     { return r.revision }
