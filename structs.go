package glicko2go

// GlickoPlayer Represents a player within the original Glicko System.
// Used to allow conversions from Glicko 2 to the old scale, if preferred.
type GlickoPlayer struct {
	Rating          float64
	RatingDeviation float64
}

// Glicko2Player Represents a player within the Glicko 2 rating system.
type Glicko2Player struct {
	GlickoPlayer
	RatingVolatility float64
}

type Glicko2AlgorithmSettings struct {
	SystemConstant       float64
	ConvergenceTolerance float64
}

type Glicko2PlayerPeriodMatches struct {
	Opponents []Glicko2Player
	Results   []float64
}

type Glicko2MatchByID struct {
	Player1ID int
	Player2ID int
	Result    float64
}
