package core_test

import (
	"bytes"
	"testing"

	"github.com/ngoldack/dicetrace/package/core"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestEncodeDecodeUser(t *testing.T) {
	t.Parallel()
	original := &core.User{
		UserID:      core.NewUserID(),
		Username:    "testuser",
		BGGUsername: "bgguser",
	}

	var buf bytes.Buffer
	err := core.EncodeUser(&buf, original)
	require.NoError(t, err)

	decoded, err := core.DecodeUser(&buf)
	require.NoError(t, err)
	assert.Equal(t, original, decoded)
}

func TestEncodeDecodeAttendee(t *testing.T) {
	t.Parallel()
	original := &core.Attendee{
		User: core.User{
			UserID:      core.NewUserID(),
			Username:    "testuser",
			BGGUsername: "bgguser",
		},
		Status: core.AttendeeStatusConfirmed,
	}

	var buf bytes.Buffer
	err := core.EncodeAttendee(&buf, original)
	require.NoError(t, err)

	decoded, err := core.DecodeAttendee(&buf)
	require.NoError(t, err)
	assert.Equal(t, original, decoded)
}

func TestEncodeDecodeEvent(t *testing.T) {
	t.Parallel()
	original := &core.Event{
		EventID: core.NewEventID(),
		Status:  core.EventStatusOngoing,
		Attendees: []core.Attendee{
			{
				User: core.User{
					UserID:      core.NewUserID(),
					Username:    "user1",
					BGGUsername: "bgguser1",
				},
				Status: core.AttendeeStatusConfirmed,
			},
			{
				User: core.User{
					UserID:      core.NewUserID(),
					Username:    "user2",
					BGGUsername: "bgguser2",
				},
				Status: core.AttendeeStatusPending,
			},
		},
		Matches: []core.Match{},
	}

	var buf bytes.Buffer
	err := core.EncodeEvent(&buf, original)
	require.NoError(t, err)

	decoded, err := core.DecodeEvent(&buf)
	require.NoError(t, err)
	assert.Equal(t, original, decoded)
}

func TestEncodeDecodeMatch(t *testing.T) {
	t.Parallel()
	original := &core.Match{
		MatchID: core.NewMatchID(),
		Game: core.Game{
			GameID:     core.NewGameID(),
			BGGID:      174430,
			Rating:     8.5,
			Name:       "Gloomhaven",
			Categories: []string{"Adventure", "Fantasy"},
		},
		Players: []core.User{
			{
				UserID:      core.NewUserID(),
				Username:    "player1",
				BGGUsername: "bggplayer1",
			},
			{
				UserID:      core.NewUserID(),
				Username:    "player2",
				BGGUsername: "bggplayer2",
			},
		},
		Scoreboard: core.Scoreboard{
			Scores: []core.Score{
				{UserID: core.NewUserID(), Value: 100},
				{UserID: core.NewUserID(), Value: 95},
			},
			ScoreUnit: core.ScoreUnitPoints,
		},
	}

	var buf bytes.Buffer
	err := core.EncodeMatch(&buf, original)
	require.NoError(t, err)

	decoded, err := core.DecodeMatch(&buf)
	require.NoError(t, err)
	assert.Equal(t, original, decoded)
}

func TestEncodeDecodeScore(t *testing.T) {
	t.Parallel()
	original := &core.Score{
		UserID: core.NewUserID(),
		Value:  42.5,
	}

	var buf bytes.Buffer
	err := core.EncodeScore(&buf, original)
	require.NoError(t, err)

	decoded, err := core.DecodeScore(&buf)
	require.NoError(t, err)
	assert.Equal(t, original, decoded)
}

func TestEncodeDecodeScoreboard(t *testing.T) {
	t.Parallel()
	original := &core.Scoreboard{
		Scores: []core.Score{
			{UserID: core.NewUserID(), Value: 100},
			{UserID: core.NewUserID(), Value: 95},
			{UserID: core.NewUserID(), Value: 80},
		},
		ScoreUnit: core.ScoreUnitPoints,
	}

	var buf bytes.Buffer
	err := core.EncodeScoreboard(&buf, original)
	require.NoError(t, err)

	decoded, err := core.DecodeScoreboard(&buf)
	require.NoError(t, err)
	assert.Equal(t, original, decoded)
}

func TestEncodeDecodeGame(t *testing.T) {
	t.Parallel()
	original := &core.Game{
		GameID:     core.NewGameID(),
		BGGID:      174430,
		Rating:     8.5,
		Name:       "Gloomhaven",
		Categories: []string{"Adventure", "Fantasy", "Strategy"},
	}

	var buf bytes.Buffer
	err := core.EncodeGame(&buf, original)
	require.NoError(t, err)

	decoded, err := core.DecodeGame(&buf)
	require.NoError(t, err)
	assert.Equal(t, original, decoded)
}

func TestEncodeDecodeAttendeeStatus(t *testing.T) {
	t.Parallel()
	testCases := []struct {
		name   string
		status core.AttendeeStatus
	}{
		{"confirmed", core.AttendeeStatusConfirmed},
		{"pending", core.AttendeeStatusPending},
		{"declined", core.AttendeeStatusDeclined},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			attendee := &core.Attendee{
				User: core.User{
					UserID:   core.NewUserID(),
					Username: "testuser",
				},
				Status: tc.status,
			}

			var buf bytes.Buffer
			err := core.EncodeAttendee(&buf, attendee)
			require.NoError(t, err)

			decoded, err := core.DecodeAttendee(&buf)
			require.NoError(t, err)
			assert.Equal(t, tc.status, decoded.Status)
		})
	}
}

func TestEncodeDecodeEventStatus(t *testing.T) {
	t.Parallel()
	testCases := []struct {
		name   string
		status core.EventStatus
	}{
		{"scheduled", core.EventStatusScheduled},
		{"ongoing", core.EventStatusOngoing},
		{"completed", core.EventStatusCompleted},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			event := &core.Event{
				EventID:   core.NewEventID(),
				Status:    tc.status,
				Attendees: []core.Attendee{},
				Matches:   []core.Match{},
			}

			var buf bytes.Buffer
			err := core.EncodeEvent(&buf, event)
			require.NoError(t, err)

			decoded, err := core.DecodeEvent(&buf)
			require.NoError(t, err)
			assert.Equal(t, tc.status, decoded.Status)
		})
	}
}

func TestEncodeDecodeScoreUnit(t *testing.T) {
	t.Parallel()
	testCases := []struct {
		name      string
		scoreUnit core.ScoreUnit
	}{
		{"points", core.ScoreUnitPoints},
		{"custom", core.ScoreUnitCustom},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			scoreboard := &core.Scoreboard{
				Scores: []core.Score{
					{UserID: core.NewUserID(), Value: 100},
				},
				ScoreUnit: tc.scoreUnit,
			}

			var buf bytes.Buffer
			err := core.EncodeScoreboard(&buf, scoreboard)
			require.NoError(t, err)

			decoded, err := core.DecodeScoreboard(&buf)
			require.NoError(t, err)
			assert.Equal(t, tc.scoreUnit, decoded.ScoreUnit)
		})
	}
}
