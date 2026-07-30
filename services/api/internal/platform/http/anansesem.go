package apihttp

import (
	"context"
	"errors"
	"net/http"
	"strings"
	"time"

	"github.com/stanleyHayes/obiara/services/api/internal/games/anansesem/application"
	"github.com/stanleyHayes/obiara/services/api/internal/games/anansesem/domain"
)

type AnansesemStories interface {
	Create(context.Context, application.Command, string) (domain.Story, error)
	Add(context.Context, application.Command, string) (domain.Story, error)
	Edit(context.Context, application.Command, string, string) (domain.Story, error)
	Grant(context.Context, application.Command) (domain.Story, error)
	Publish(context.Context, application.Command) (domain.Edition, error)
	View(context.Context, application.Command) (application.Projection, error)
}

func RegisterAnansesemRoutes(
	mux *http.ServeMux,
	stories AnansesemStories,
	pairs OwarePairResolver,
	auth SessionAuthenticator,
) {
	mux.Handle("POST /v1/circles/{circleId}/stories", createStoryHandler(stories, pairs, auth))
	mux.Handle("GET /v1/circles/{circleId}/stories/{storyId}", viewStoryHandler(stories, pairs, auth))
	mux.Handle("POST /v1/circles/{circleId}/stories/{storyId}/passages", addStoryPassageHandler(stories, pairs, auth))
	mux.Handle("PUT /v1/circles/{circleId}/stories/{storyId}/passages/{passageId}", editStoryPassageHandler(stories, pairs, auth))
	mux.Handle("POST /v1/circles/{circleId}/stories/{storyId}/publication-grants", grantStoryPublicationHandler(stories, pairs, auth))
	mux.Handle("POST /v1/circles/{circleId}/stories/{storyId}/publish", publishStoryHandler(stories, pairs, auth))
}

type storyPassageResponse struct {
	ID        string    `json:"id"`
	Ordinal   int       `json:"ordinal"`
	Content   string    `json:"content"`
	Yours     bool      `json:"yours"`
	CreatedAt time.Time `json:"createdAt"`
	EditedAt  time.Time `json:"editedAt"`
}

type storyEditionResponse struct {
	Version     uint64                  `json:"version"`
	TitleCode   string                  `json:"titleCode"`
	Passages    []domain.EditionPassage `json:"passages"`
	PublishedAt time.Time               `json:"publishedAt"`
}

type storyResponse struct {
	ID          string                 `json:"id"`
	TitleCode   string                 `json:"titleCode"`
	Passages    []storyPassageResponse `json:"passages"`
	YourTurn    bool                   `json:"yourTurn"`
	YourGrant   bool                   `json:"yourGrant"`
	OtherGrant  bool                   `json:"otherGrant"`
	BothGranted bool                   `json:"bothGranted"`
	Editions    []storyEditionResponse `json:"editions"`
	Revision    uint64                 `json:"revision"`
}

func projectStory(story application.Projection) storyResponse {
	result := storyResponse{
		ID: story.ID, TitleCode: story.TitleCode,
		Passages: make([]storyPassageResponse, 0, len(story.Passages)),
		YourTurn: story.YourTurn, YourGrant: story.YourGrant,
		OtherGrant: story.OtherGrant, BothGranted: story.BothGranted, Revision: story.Revision,
		Editions: make([]storyEditionResponse, 0, len(story.Editions)),
	}
	for _, passage := range story.Passages {
		result.Passages = append(result.Passages, storyPassageResponse(passage))
	}
	for _, edition := range story.Editions {
		result.Editions = append(result.Editions, storyEditionResponse{
			Version: edition.Version, TitleCode: edition.TitleCode,
			Passages: edition.Passages, PublishedAt: edition.PublishedAt,
		})
	}
	return result
}

func storyCommand(
	ctx context.Context,
	pairs OwarePairResolver,
	r *http.Request,
	actorID, commandID string,
	expected uint64,
) (application.Command, error) {
	circleID := r.PathValue("circleId")
	second, err := pairs.Pair(ctx, circleID, actorID)
	if err != nil {
		return application.Command{}, err
	}
	return application.Command{
		ID: commandID, StoryID: r.PathValue("storyId"), RoomID: circleID,
		ActorID: actorID, FirstAuthorID: actorID, SecondAuthorID: second,
		ExpectedRevision: expected,
	}, nil
}

func createStoryHandler(stories AnansesemStories, pairs OwarePairResolver, auth SessionAuthenticator) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		actor, ok := subanSubject(w, r, auth)
		if !ok {
			return
		}
		commandID, ok := circleCommandID(w, r)
		if !ok {
			return
		}
		var body struct {
			TitleCode string `json:"titleCode"`
		}
		if !decodeCircleJSON(w, r, &body) {
			return
		}
		command, err := storyCommand(r.Context(), pairs, r, actor, commandID, 0)
		if err == nil {
			var story domain.Story
			story, err = stories.Create(r.Context(), command, strings.TrimSpace(body.TitleCode))
			command.StoryID = story.ID()
		}
		if err == nil {
			var view application.Projection
			view, err = stories.View(r.Context(), command)
			if err == nil {
				writeSuccess(w, r, http.StatusCreated, projectStory(view))
				return
			}
		}
		writeAnansesemError(w, r, err)
	})
}

func viewStoryHandler(stories AnansesemStories, pairs OwarePairResolver, auth SessionAuthenticator) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		actor, ok := subanSubject(w, r, auth)
		if !ok {
			return
		}
		command, err := storyCommand(r.Context(), pairs, r, actor, "", 0)
		if err == nil {
			var view application.Projection
			view, err = stories.View(r.Context(), command)
			if err == nil {
				writeSuccess(w, r, http.StatusOK, projectStory(view))
				return
			}
		}
		writeAnansesemError(w, r, err)
	})
}

type storyMutationRequest struct {
	Content          string `json:"content,omitempty"`
	ExpectedRevision uint64 `json:"expectedRevision"`
}

func addStoryPassageHandler(stories AnansesemStories, pairs OwarePairResolver, auth SessionAuthenticator) http.Handler {
	return mutateStoryHandler(stories, pairs, auth, func(ctx context.Context, command application.Command, body storyMutationRequest, _ *http.Request) error {
		_, err := stories.Add(ctx, command, strings.TrimSpace(body.Content))
		return err
	})
}

func editStoryPassageHandler(stories AnansesemStories, pairs OwarePairResolver, auth SessionAuthenticator) http.Handler {
	return mutateStoryHandler(stories, pairs, auth, func(ctx context.Context, command application.Command, body storyMutationRequest, r *http.Request) error {
		_, err := stories.Edit(ctx, command, r.PathValue("passageId"), strings.TrimSpace(body.Content))
		return err
	})
}

func grantStoryPublicationHandler(stories AnansesemStories, pairs OwarePairResolver, auth SessionAuthenticator) http.Handler {
	return mutateStoryHandler(stories, pairs, auth, func(ctx context.Context, command application.Command, _ storyMutationRequest, _ *http.Request) error {
		_, err := stories.Grant(ctx, command)
		return err
	})
}

func publishStoryHandler(stories AnansesemStories, pairs OwarePairResolver, auth SessionAuthenticator) http.Handler {
	return mutateStoryHandler(stories, pairs, auth, func(ctx context.Context, command application.Command, _ storyMutationRequest, _ *http.Request) error {
		_, err := stories.Publish(ctx, command)
		return err
	})
}

func mutateStoryHandler(
	stories AnansesemStories,
	pairs OwarePairResolver,
	auth SessionAuthenticator,
	mutate func(context.Context, application.Command, storyMutationRequest, *http.Request) error,
) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		actor, ok := subanSubject(w, r, auth)
		if !ok {
			return
		}
		commandID, ok := circleCommandID(w, r)
		if !ok {
			return
		}
		var body storyMutationRequest
		if !decodeCircleJSON(w, r, &body) {
			return
		}
		command, err := storyCommand(r.Context(), pairs, r, actor, commandID, body.ExpectedRevision)
		if err == nil {
			err = mutate(r.Context(), command, body, r)
		}
		if err == nil {
			var view application.Projection
			view, err = stories.View(r.Context(), command)
			if err == nil {
				writeSuccess(w, r, http.StatusOK, projectStory(view))
				return
			}
		}
		writeAnansesemError(w, r, err)
	})
}

func writeAnansesemError(w http.ResponseWriter, r *http.Request, err error) {
	switch {
	case errors.Is(err, domain.ErrInvalid):
		writeError(w, r, http.StatusUnprocessableEntity, APIError{
			Code: "validation_failed", Message: "The story action is invalid.",
		})
	case errors.Is(err, application.ErrConflict), errors.Is(err, application.ErrApplied):
		writeError(w, r, http.StatusConflict, APIError{
			Code: "story_conflict", Message: "The story changed. Refresh and try again.",
		})
	default:
		writeError(w, r, http.StatusNotFound, APIError{
			Code:    "story_not_available",
			Message: "That private story is not available in this room.",
		})
	}
}
