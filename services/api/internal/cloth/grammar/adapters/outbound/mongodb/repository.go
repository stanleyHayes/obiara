package mongodb

import (
	"context"
	"github.com/stanleyHayes/obiara/internal/platform/mongo"
	"github.com/stanleyHayes/obiara/services/api/internal/cloth/grammar/application"
	"github.com/stanleyHayes/obiara/services/api/internal/cloth/grammar/domain"
	"go.mongodb.org/mongo-driver/v2/bson"
	driver "go.mongodb.org/mongo-driver/v2/mongo"
	"go.mongodb.org/mongo-driver/v2/mongo/options"
)

type Repository struct{ collection *driver.Collection }

func NewRepository(database *driver.Database) *Repository {
	return &Repository{database.Collection("cloth_grammar_recipes")}
}

type tokenDocument struct {
	Name  string `bson:"name"`
	Value string `bson:"value"`
}
type document struct {
	ID          string          `bson:"_id"`
	Version     string          `bson:"version"`
	RenderSeed  string          `bson:"renderSeed"`
	CommandID   string          `bson:"commandId"`
	Fingerprint string          `bson:"fingerprint"`
	Pair        [2]string       `bson:"pair"`
	Themes      []string        `bson:"themes"`
	Provenance  []string        `bson:"provenance"`
	Tokens      []tokenDocument `bson:"tokens"`
	Revision    uint64          `bson:"revision"`
}

func (r *Repository) EnsureIndexes(ctx context.Context) error {
	_, err := r.collection.Indexes().CreateOne(ctx, driver.IndexModel{Keys: bson.D{{Key: "commandId", Value: 1}}, Options: options.Index().SetUnique(true).SetName("cloth_recipe_command_unique")})
	return err
}
func (r *Repository) Store(ctx context.Context, recipe domain.Recipe, expected uint64) (domain.Recipe, bool, error) {
	if expected != 0 {
		return domain.Recipe{}, false, application.ErrConcurrentChange
	}
	if _, err := r.collection.InsertOne(ctx, toDocument(recipe)); err == nil {
		return recipe, false, nil
	} else if !mongo.IsDuplicateKey(err) {
		return domain.Recipe{}, false, err
	}
	var byCommand document
	if err := r.collection.FindOne(ctx, bson.M{"commandId": recipe.CommandID()}).Decode(&byCommand); err == nil {
		existing, decodeErr := fromDocument(byCommand)
		if decodeErr != nil {
			return domain.Recipe{}, false, decodeErr
		}
		if existing.Fingerprint() != recipe.Fingerprint() {
			return domain.Recipe{}, false, domain.ErrCommandMismatch
		}
		return existing, true, nil
	}
	existing, err := r.Find(ctx, recipe.ID())
	if err != nil {
		return domain.Recipe{}, false, application.ErrConcurrentChange
	}
	if existing.RenderSeed() != recipe.RenderSeed() {
		return domain.Recipe{}, false, application.ErrConcurrentChange
	}
	return existing, true, nil
}
func (r *Repository) Find(ctx context.Context, id string) (domain.Recipe, error) {
	var doc document
	if err := r.collection.FindOne(ctx, bson.M{"_id": id}).Decode(&doc); err != nil {
		return domain.Recipe{}, err
	}
	return fromDocument(doc)
}
func toDocument(r domain.Recipe) document {
	tokens := make([]tokenDocument, 0, len(r.Tokens()))
	for _, token := range r.Tokens() {
		tokens = append(tokens, tokenDocument{token.Name, token.Value})
	}
	return document{r.ID(), r.Version(), r.RenderSeed(), r.CommandID(), r.Fingerprint(), r.Pair(), r.Themes(), r.Provenance(), tokens, r.Revision()}
}
func fromDocument(d document) (domain.Recipe, error) {
	tokens := make([]domain.Token, 0, len(d.Tokens))
	for _, token := range d.Tokens {
		tokens = append(tokens, domain.Token{Name: token.Name, Value: token.Value})
	}
	return domain.Rehydrate(d.ID, d.Version, d.RenderSeed, d.CommandID, d.Fingerprint, d.Pair, d.Themes, d.Provenance, tokens, d.Revision)
}
