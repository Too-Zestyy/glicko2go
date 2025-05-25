package glicko2go

import (
	"testing"
)

// Values used for the example at https://www.glicko.net/glicko/glicko2.pdf
var (
	playerRating    float64 = 1500
	playerDeviation float64 = 200

	opponentRatings    = []float64{1400, 1550, 1700}
	opponentDeviations = []float64{30, 100, 300}
	gameOutcomes       = []float64{GAME_OUTCOME_WIN, GAME_OUTCOME_LOSS, GAME_OUTCOME_LOSS}
)

func calculateExampleWithoutPlayerStructs() (float64, float64, float64, error) {
	var g2Ratings []float64
	var g2Deviations []float64

	for i := 0; i < len(opponentDeviations); i++ {
		g2Ratings = append(g2Ratings, GlickoRatingToGlicko2(opponentRatings[i]))
		g2Deviations = append(g2Deviations, GlickoDeviationToGlicko2(opponentDeviations[i]))
	}

	return UpdatePlayerFromMatches(
		GlickoRatingToGlicko2(playerRating),
		GlickoDeviationToGlicko2(playerDeviation),
		GLICKO2_DEFAULT_PLAYER_VOLATILITY,
		g2Ratings, g2Deviations, gameOutcomes,
		Glicko2AlgorithmSettings{
			SystemConstant:       GLICKO2_DEFAULT_SYSTEM_CONSTANT,
			ConvergenceTolerance: GLICKO2_DEFAULT_CONVERGENCE_TOLERANCE,
		},
	)
}

func calculateExampleWithPlayerStructs() (Glicko2Player, error) {
	playerToUpdate := ConvertToDefaultGlicko2(GlickoPlayer{
		Rating:          playerRating,
		RatingDeviation: playerDeviation,
	})

	var opponents []Glicko2Player

	for i := 0; i < len(opponentRatings); i++ {
		opponents = append(opponents, ConvertToDefaultGlicko2(GlickoPlayer{
			Rating:          opponentRatings[i],
			RatingDeviation: opponentDeviations[i],
		}))
	}

	playerUpdater := PlayerUpdaterWithDefaultSettings()

	return playerUpdater(playerToUpdate, opponents, gameOutcomes)
}

// TODO: Add tests for conversions and basic rating calculations

func TestGlicko2CalculationMethodsMatch(t *testing.T) {
	dRating, dDeviation, dVolatility, noStructError := calculateExampleWithoutPlayerStructs()
	updatedPlayer, structError := calculateExampleWithPlayerStructs()

	if noStructError != nil {
		t.Fatal(noStructError)
	}
	if structError != nil {
		t.Fatal(structError)
	}

	if dRating != updatedPlayer.Rating {
		t.Errorf("Ratings do not match when using structs vs without: Struct rating: %v, No-struct rating: %v", updatedPlayer.Rating, dRating)
	}
	if dDeviation != updatedPlayer.RatingDeviation {
		t.Errorf("Deviations do not match when using structs vs without: Struct deviation: %v, No-struct deviation: %v", updatedPlayer.RatingDeviation, dDeviation)
	}
	if dVolatility != updatedPlayer.RatingVolatility {
		t.Errorf("Volatilities do not match when using structs vs without: Struct volatility: %v, No-struct volatility: %v", updatedPlayer.RatingVolatility, dVolatility)
	}

}
