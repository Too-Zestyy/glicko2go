package glicko2go

// TODO: Document

import (
	"errors"
	"math"
)

const (
	GAME_OUTCOME_LOSS float64 = 0
	GAME_OUTCOME_DRAW float64 = 0.5
	GAME_OUTCOME_WIN  float64 = 1

	GLICKO2_DEFAULT_CONVERGENCE_TOLERANCE float64 = 0.000001
	GLICKO2_LOW_SYSTEM_CONSTANT           float64 = 0.3
	GLICKO2_DEFAULT_SYSTEM_CONSTANT       float64 = 0.5
	GLICKO2_HIGH_SYSTEM_CONSTANT          float64 = 1.2
)

var (
	piSquared = math.Pow(math.Pi, 2.0)

	glicko2DefaultSettings Glicko2AlgorithmSettings = Glicko2AlgorithmSettings{
		SystemConstant:       GLICKO2_DEFAULT_SYSTEM_CONSTANT,
		ConvergenceTolerance: GLICKO2_DEFAULT_CONVERGENCE_TOLERANCE,
	}
)

//// Variance calculation (Step 3)

// step3g Calculates `g(φ)`, which is used as a component of game variance within a period
func step3g(deviation float64) float64 {
	deviationSquared := math.Pow(deviation, 2)
	return 1 / math.Sqrt(1+3*deviationSquared/piSquared)
}

// step3E Calculates `E(µ, µj, φj)`, which is used as a component to calculate game variance within a period
func step3E(rating float64, opponentRating float64, opponentDeviation float64) float64 {
	return 1 / (1 + math.Exp(-step3g(opponentDeviation)*(rating-opponentRating)))
}

// calculateVarianceFromGameOutcomes calculates `𝒱`, which is a player's variance within a period solely from game outcomes.
// Equivalent to the entirety of step 3.
func calculateVarianceFromGameOutcomes(playerRating float64, opponentRatings []float64, opponentDeviations []float64) float64 {
	var sum float64

	for i := 0; i < len(opponentRatings); i++ {
		curMatchE := step3E(playerRating, opponentRatings[i], opponentDeviations[i])
		sum += math.Pow(step3g(opponentDeviations[i]), 2) * curMatchE * (1 - curMatchE)
	}

	return 1 / sum
}

//// Rating Improvement (Step 4)

// calculateEstimatedRatingImprovement calculates `∆`, which represents the estimated change in rating compared to the pre-period rating.
// Equivalent to step 4.
func calculateEstimatedRatingImprovement(playerRating float64, opponentRatings []float64, opponentDeviations []float64, gameOutcomes []float64, v float64) float64 {
	var sum float64

	for i := 0; i < len(opponentRatings); i++ {
		sum += step3g(opponentDeviations[i]) * (gameOutcomes[i] - step3E(playerRating, opponentRatings[i], opponentDeviations[i]))
	}

	return v * sum
}

//// Volatility Calculation (Step 5)

// aFromVolatility calculates `a`, which is a constant used within f(x) as defined in step 5.1.
func aFromVolatility(volatility float64) float64 {
	return math.Log(math.Pow(volatility, 2))
}

// fVolatilityFunction calculates f(x) as defined in step 5.1, which is used across various variables to converge on a new volatility value.
func fVolatilityFunction(x float64, systemConstant float64, volatility float64, delta float64, deviation float64, variance float64) float64 {
	a := aFromVolatility(volatility)
	ePowX := math.Exp(x)
	deviationSquared := math.Pow(deviation, 2)

	return (ePowX*(math.Pow(delta, 2)-deviationSquared-variance-ePowX))/(2*math.Pow(deviationSquared+variance+ePowX, 2)) -
		(x-a)/math.Pow(systemConstant, 2)

}

// calculateNewVolatility Calculates the post-period volatility of a player (AKA `σ′`).
func calculateNewVolatility(convergenceTolerance float64, systemConstant float64, volatility float64, delta float64, deviation float64, variance float64) float64 {
	a := aFromVolatility(volatility)

	// Get bracketing values to speed up convergence
	A := a
	var B float64
	if math.Pow(delta, 2) > math.Pow(deviation, 2)+variance {
		B = math.Log(math.Pow(delta, 2) - math.Pow(deviation, 2) - variance)
	} else {
		var k float64 = 1
		for {
			if fVolatilityFunction(a-k*systemConstant, systemConstant, volatility, delta, deviation, variance) < 0 {
				k += 1
			} else {
				break
			}
		}
		B = a - k*systemConstant
	}

	fA := fVolatilityFunction(A, systemConstant, volatility, delta, deviation, variance)
	fB := fVolatilityFunction(B, systemConstant, volatility, delta, deviation, variance)

	for {
		if math.Abs(B-A) > convergenceTolerance {

			C := A + (A-B)*fA/(fB-fA)

			fC := fVolatilityFunction(C, systemConstant, volatility, delta, deviation, variance)

			if fC*fB <= 0 {
				A = B
				fA = fB
			} else {
				fA /= 2
			}

			B = C
			fB = fC

		} else {
			return math.Exp(A / 2)
		}
	}

}

// calcVolatilityFromMatches is a convenience function to calculate volatility from a list of matches, without needing intermediate values.
func calcVolatilityFromMatches(playerRating float64, playerDeviation float64, playerVolatility float64, periodVariance float64, opponentRatings []float64, opponentDeviations []float64, gameOutcomes []float64,
	convergenceTolerance float64, systemConstant float64) float64 {

	delta := calculateEstimatedRatingImprovement(playerRating, opponentRatings, opponentDeviations, gameOutcomes, periodVariance)

	return calculateNewVolatility(convergenceTolerance, systemConstant, playerVolatility, delta, playerDeviation, periodVariance)

}

//// Pre-rating deviation (Step 6)

// calcPreRatingDeviation calculates the deviation value that is both used as a component for post-period volatility,
// and to update volatility when players have not played a game within a period.
func calcPreRatingDeviation(deviation float64, volatility float64) float64 {
	return math.Sqrt(math.Pow(deviation, 2) + math.Pow(volatility, 2))
}

//// Update player stats (Step 7)

// calcPlayedPeriodDeviation Calculates the post-period volatility for a player when matches have been played within the period.
func calcPlayedPeriodDeviation(playerDeviation float64, variance float64, newVolatility float64) float64 {
	return 1 / math.Sqrt((1/math.Pow(calcPreRatingDeviation(playerDeviation, newVolatility), 2))+(1/variance))
}

func calcPlayedPeriodRating(playerRating float64, postPeriodDeviation float64, opponentRatings []float64, opponentDeviations []float64, gameOutcomes []float64) float64 {
	var sum float64

	for i := 0; i < len(opponentRatings); i++ {
		sum += step3g(opponentDeviations[i]) * (gameOutcomes[i] - step3E(playerRating, opponentRatings[i], opponentDeviations[i]))
	}

	return playerRating + (math.Pow(postPeriodDeviation, 2) * sum)
}

// UpdatePlayerFromMatches Calculates a players new rating, deviation and volatility after a single period.
//
// For more details, see https://www.glicko.net/glicko/glicko2.pdf
func UpdatePlayerFromMatches(playerRating float64, playerDeviation float64, playerVolatility float64, opponentRatings []float64, opponentDeviations []float64, gameOutcomes []float64, settings Glicko2AlgorithmSettings) (float64, float64, float64, error) {

	// Argument Validation
	if len(opponentRatings) != len(opponentDeviations) || len(opponentRatings) != len(gameOutcomes) || len(opponentDeviations) != len(gameOutcomes) {
		return -1, -1, -1, errors.New("the lengths of opponent ratings, deviations and game outcomes must be the same length")
	}

	if len(gameOutcomes) == 0 {
		return playerRating, calcPreRatingDeviation(playerDeviation, playerVolatility), playerVolatility, nil
	} else {
		variance := calculateVarianceFromGameOutcomes(playerRating, opponentRatings, opponentDeviations)

		newVolatility := calcVolatilityFromMatches(
			playerRating, playerDeviation, playerVolatility, variance,
			opponentRatings, opponentDeviations, gameOutcomes,
			settings.ConvergenceTolerance, settings.SystemConstant)

		newDeviation := calcPlayedPeriodDeviation(playerDeviation, variance, newVolatility)

		newRating := calcPlayedPeriodRating(playerRating, newDeviation, opponentRatings, opponentDeviations, gameOutcomes)

		return newRating, newDeviation, newVolatility, nil
	}
}

//// Convenience functions

// RawPlayerUpdaterWithSettings Returns a function used to return an updated player after a period.
//
// A player is denoted by `playerRating`, `playerDeviation` and `playerVolatility`.
//
// Matches are derived from `opponentRatings`, `opponentDeviations` and `gameOutcomes` using the order of their contents.
//
// `settings` can be used to denote the constants used for the application's Glicko environment.
func RawPlayerUpdaterWithSettings(settings Glicko2AlgorithmSettings) func(playerRating float64, playerDeviation float64, playerVolatility float64,
	opponentRatings []float64, opponentDeviations []float64, gameOutcomes []float64) (Glicko2Player, error) {

	return func(playerRating float64, playerDeviation float64, playerVolatility float64, opponentRatings []float64, opponentDeviations []float64, gameOutcomes []float64) (Glicko2Player, error) {

		newRating, newDeviation, newVolatility, err := UpdatePlayerFromMatches(playerRating, playerDeviation, playerVolatility, opponentRatings, opponentDeviations, gameOutcomes, settings)

		if err != nil {
			return Glicko2Player{}, err
		}

		return Glicko2Player{
			GlickoPlayer: GlickoPlayer{
				Rating:          newRating,
				RatingDeviation: newDeviation,
			},
			RatingVolatility: newVolatility,
		}, nil
	}
}

// RawPlayerUpdaterWithDefaultSettings returns a RawPlayerUpdaterWithSettings function with the following settings:
//   - System constant: GLICKO2_DEFAULT_SYSTEM_CONSTANT
//   - Convergence tolerance: GLICKO2_DEFAULT_CONVERGENCE_TOLERANCE
func RawPlayerUpdaterWithDefaultSettings() func(playerRating float64, playerDeviation float64, playerVolatility float64,
	opponentRatings []float64, opponentDeviations []float64, gameOutcomes []float64) (Glicko2Player, error) {
	return RawPlayerUpdaterWithSettings(glicko2DefaultSettings)
}

// PlayerUpdaterWithSettings returns a wrapper of RawPlayerUpdaterWithSettings that allows for a player's period to be calculated
// via Glicko2Player structures over arrays of each value.
func PlayerUpdaterWithSettings(settings Glicko2AlgorithmSettings) func(player Glicko2Player,
	opponents []Glicko2Player, gameOutcomes []float64) (Glicko2Player, error) {

	return func(player Glicko2Player, opponents []Glicko2Player, gameOutcomes []float64) (Glicko2Player, error) {
		playerUpdater := RawPlayerUpdaterWithSettings(settings)

		var opponentRatings []float64
		var opponentDeviations []float64

		for _, opponent := range opponents {
			opponentRatings = append(opponentRatings, opponent.Rating)
			opponentDeviations = append(opponentDeviations, opponent.RatingDeviation)
		}

		return playerUpdater(player.Rating, player.RatingDeviation, player.RatingVolatility, opponentRatings, opponentDeviations, gameOutcomes)
	}

}

// PlayerUpdaterWithDefaultSettings provides default settings for PlayerUpdaterWithSettings,
// in an identical fashion to RawPlayerUpdaterWithDefaultSettings.
func PlayerUpdaterWithDefaultSettings() func(player Glicko2Player,
	opponents []Glicko2Player, gameOutcomes []float64) (Glicko2Player, error) {
	return PlayerUpdaterWithSettings(glicko2DefaultSettings)
}

func NewPlayerUpdaterWithSettings(settings Glicko2AlgorithmSettings) func(player Glicko2Player, periodGames []Glicko2MatchForPlayer) (Glicko2Player, error) {

	return func(player Glicko2Player, periodGames []Glicko2MatchForPlayer) (Glicko2Player, error) {
		playerUpdater := RawPlayerUpdaterWithSettings(settings)

		var opponentRatings []float64
		var opponentDeviations []float64
		// TODO: Be more consistent with usage of game result vs outcome
		var gameResults []float64

		for _, game := range periodGames {
			opponentRatings = append(opponentRatings, game.Opponent.Rating)
			opponentDeviations = append(opponentDeviations, game.Opponent.RatingDeviation)
			gameResults = append(gameResults, game.Result)
		}

		return playerUpdater(player.Rating, player.RatingDeviation, player.RatingVolatility, opponentRatings, opponentDeviations, gameResults)
	}
}
