package internal

import (
	"bytes"
	"log/slog"
	"strconv"

	"github.com/kkjdaniel/gogeek/thing"
	"github.com/nats-io/nats.go/micro"
	"github.com/ngoldack/dicetrace/package/core"
	"go.jetify.com/typeid/v2"
)

const ErrorBGGGameNotFound = "bgg_game_not_found"
const ErrorBGGGameIDMissing = "bgg_game_id_missing"

func HandlerGetGameByID() micro.Handler {
	return micro.HandlerFunc(func(r micro.Request) {
		slog.Info("HandlerGetGameByID called", "headers", r.Headers())
		bggId := r.Headers().Get("bgg_id")
		if bggId == "" {
			r.Error(ErrorBGGGameIDMissing, "BGG game ID is missing", nil)
			return
		}

		//to int
		bggIdInt, err := strconv.Atoi(bggId)
		if err != nil {
			r.Error(ErrorBGGGameIDMissing, "BGG game ID is not a valid number", nil)
			return
		}

		items, err := thing.Query([]int{bggIdInt})
		if err != nil {
			r.Error(ErrorBGGGameNotFound, "BGG game not found", nil)
			return
		}

		if len(items.Items) == 0 {
			r.Error(ErrorBGGGameNotFound, "BGG game not found", nil)
			return
		}

		games := make([]*core.Game, 0, len(items.Items))
		for _, item := range items.Items {
			// only process boardgames
			if item.Type != "boardgame" {
				continue
			}

			game := &core.Game{
				GameID: typeid.MustGenerate("game"),
				BGGID:  item.ID,
				Name:   item.Name[0].Value,
				Categories: func() []string {
					cats := make([]string, 0)
					for _, link := range item.Links {
						if link.Type == "boardgamecategory" {
							cats = append(cats, link.Value)
						}
					}
					return cats
				}(),
			}
			games = append(games, game)
		}

		if len(games) == 0 {
			r.Error(ErrorBGGGameNotFound, "BGG game not found", nil)
			return
		}

		// For simplicity, return only the first game found
		game := games[0]

		buf := bytes.NewBuffer(nil)
		if err := core.EncodeGame(buf, game); err != nil {
			r.Error("internal_error", "failed to encode game", nil)
			return
		}

		slog.Info("response", "game", game)

		r.Respond(buf.Bytes())
	})
}
