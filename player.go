package glicko2go

const (
	// GLICKO_DEFAULT_PLAYER_RATING is the default player rating for someone that has not been previously rated,
	// as described in step 1.
	GLICKO_DEFAULT_PLAYER_RATING = 1500
	// GLICKO_DEFAULT_PLAYER_DEVIATION is the default player rating deviation for someone that has not been previously rated,
	// as described in step 1. Represents the width of the ratings that the system is 99% the player's skill is within.
	GLICKO_DEFAULT_PLAYER_DEVIATION = 350
	// GLICKO2_DEFAULT_PLAYER_VOLATILITY is the default player volatility for someone that has not been previously rated,
	// as described in step 1. Represents the consistency of the player's performance.
	GLICKO2_DEFAULT_PLAYER_VOLATILITY = 0.06
)

// NewDefaultGlickoPlayer creates a GlickoPlayer using identical values to NewDefaultGlicko2Player.
func NewDefaultGlickoPlayer() GlickoPlayer {
	return GlickoPlayer{
		Rating:          GLICKO_DEFAULT_PLAYER_RATING,
		RatingDeviation: GLICKO_DEFAULT_PLAYER_DEVIATION,
	}
}

// NewDefaultGlicko2Player creates a Glicko2Player representing a previously unrated player.
// Uses the values defined in step 1. These values are exposed via:
//
//   - GLICKO_DEFAULT_PLAYER_RATING
//   - GLICKO_DEFAULT_PLAYER_DEVIATION
//   - GLICKO2_DEFAULT_PLAYER_VOLATILITY
func NewDefaultGlicko2Player() Glicko2Player {
	return ConvertToGlicko2WithDefaultVolatility(NewDefaultGlickoPlayer())
}
