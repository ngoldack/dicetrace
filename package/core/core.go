package core

type EventStatus string

const (
	EventStatusScheduled EventStatus = "scheduled"
	EventStatusOngoing   EventStatus = "ongoing"
	EventStatusCompleted EventStatus = "completed"
)

type User struct {
	UserID      UserID `json:"user_id"`
	Username    string `json:"username"`
	BGGUsername string `json:"bgg_username"`
}

type AttendeeStatus string

const (
	AttendeeStatusConfirmed AttendeeStatus = "confirmed"
	AttendeeStatusPending   AttendeeStatus = "pending"
	AttendeeStatusDeclined  AttendeeStatus = "declined"
)

type Attendee struct {
	User   User           `json:"user"`
	Status AttendeeStatus `json:"status"`
}

type Event struct {
	EventID EventID `json:"event_id"`

	Status EventStatus `json:"status"`

	Attendees []Attendee `json:"attendees"`
	Matches   []Match    `json:"matches"`
}

type Match struct {
	MatchID MatchID `json:"match_id"`

	Game    Game   `json:"game"`
	Players []User `json:"players"`

	Scoreboard Scoreboard `json:"scoreboard"`
}

type ScoreUnit string

const (
	ScoreUnitPoints ScoreUnit = "points"
	ScoreUnitCustom ScoreUnit = "custom"
)

type Score struct {
	UserID UserID  `json:"user_id"`
	Value  float64 `json:"value"`
}

type Scoreboard struct {
	// Scores listed in the order of players in the Match.Players slice
	Scores    []Score   `json:"scores"`
	ScoreUnit ScoreUnit `json:"score_unit"`
}

type Game struct {
	GameID GameID `json:"game_id"`
	BGGID  int    `json:"bgg_id"`

	Rating float64 `json:"rating"`

	Name       string   `json:"name"`
	Categories []string `json:"categories"`
}
