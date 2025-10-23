package core

import (
	"encoding/json"
	"io"
)

// User encoding/decoding
func EncodeUser(w io.Writer, user *User) error {
	encoder := json.NewEncoder(w)
	return encoder.Encode(user)
}

func DecodeUser(r io.Reader) (*User, error) {
	decoder := json.NewDecoder(r)
	var user User
	if err := decoder.Decode(&user); err != nil {
		return nil, err
	}
	return &user, nil
}

// Attendee encoding/decoding
func EncodeAttendee(w io.Writer, attendee *Attendee) error {
	encoder := json.NewEncoder(w)
	return encoder.Encode(attendee)
}

func DecodeAttendee(r io.Reader) (*Attendee, error) {
	decoder := json.NewDecoder(r)
	var attendee Attendee
	if err := decoder.Decode(&attendee); err != nil {
		return nil, err
	}
	return &attendee, nil
}

// Event encoding/decoding
func EncodeEvent(w io.Writer, event *Event) error {
	encoder := json.NewEncoder(w)
	return encoder.Encode(event)
}

func DecodeEvent(r io.Reader) (*Event, error) {
	decoder := json.NewDecoder(r)
	var event Event
	if err := decoder.Decode(&event); err != nil {
		return nil, err
	}
	return &event, nil
}

// Match encoding/decoding
func EncodeMatch(w io.Writer, match *Match) error {
	encoder := json.NewEncoder(w)
	return encoder.Encode(match)
}

func DecodeMatch(r io.Reader) (*Match, error) {
	decoder := json.NewDecoder(r)
	var match Match
	if err := decoder.Decode(&match); err != nil {
		return nil, err
	}
	return &match, nil
}

// Score encoding/decoding
func EncodeScore(w io.Writer, score *Score) error {
	encoder := json.NewEncoder(w)
	return encoder.Encode(score)
}

func DecodeScore(r io.Reader) (*Score, error) {
	decoder := json.NewDecoder(r)
	var score Score
	if err := decoder.Decode(&score); err != nil {
		return nil, err
	}
	return &score, nil
}

// Scoreboard encoding/decoding
func EncodeScoreboard(w io.Writer, scoreboard *Scoreboard) error {
	encoder := json.NewEncoder(w)
	return encoder.Encode(scoreboard)
}

func DecodeScoreboard(r io.Reader) (*Scoreboard, error) {
	decoder := json.NewDecoder(r)
	var scoreboard Scoreboard
	if err := decoder.Decode(&scoreboard); err != nil {
		return nil, err
	}
	return &scoreboard, nil
}

// Game encoding/decoding
func EncodeGame(w io.Writer, game *Game) error {
	encoder := json.NewEncoder(w)
	return encoder.Encode(game)
}

func DecodeGame(r io.Reader) (*Game, error) {
	decoder := json.NewDecoder(r)
	var game Game
	if err := decoder.Decode(&game); err != nil {
		return nil, err
	}
	return &game, nil
}
