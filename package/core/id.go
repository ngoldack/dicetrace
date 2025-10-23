package core

import "go.jetify.com/typeid/v2"

type UserID = typeid.TypeID

func NewUserID() UserID {
	return typeid.MustGenerate("user")
}

type EventID = typeid.TypeID

func NewEventID() EventID {
	return typeid.MustGenerate("event")
}

type MatchID = typeid.TypeID

func NewMatchID() MatchID {
	return typeid.MustGenerate("match")
}

type GameID = typeid.TypeID

func NewGameID() GameID {
	return typeid.MustGenerate("game")
}
