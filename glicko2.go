package glicko2go

// TODO: Document

import (
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

func step3g(deviation float64) float64 {
	deviationSquared := math.Pow(deviation, 2)
	return 1 / math.Sqrt(1+3*deviationSquared/piSquared)
}

func step3E(rating float64, opponentRating float64, opponentDeviation float64) float64 {
	return 1 / (1 + math.Exp(-step3g(opponentDeviation)*(rating-opponentRating)))
}

// calculateVarianceFromGameOutcomes == 𝒱
func calculateVarianceFromGameOutcomes(playerRating float64, opponentRatings []float64, opponentDeviations []float64) float64 {
	var sum float64

	for i := 0; i < len(opponentRatings); i++ {
		curMatchE := step3E(playerRating, opponentRatings[i], opponentDeviations[i])
		sum += math.Pow(step3g(opponentDeviations[i]), 2) * curMatchE * (1 - curMatchE)
	}

	return 1 / sum
}

// calculateEstimatedRatingImprovement == Step 4 ∆
func calculateEstimatedRatingImprovement(playerRating float64, opponentRatings []float64, opponentDeviations []float64, gameOutcomes []float64, v float64) float64 {
	var sum float64

	for i := 0; i < len(opponentRatings); i++ {
		sum += step3g(opponentDeviations[i]) * (gameOutcomes[i] - step3E(playerRating, opponentRatings[i], opponentDeviations[i]))
	}

	return v * sum
}

func aFromVolatility(volatility float64) float64 {
	return math.Log(math.Pow(volatility, 2))
}

func fThetaFunction(x float64, systemConstant float64, volatility float64, delta float64, deviation float64, variance float64) float64 {
	a := aFromVolatility(volatility)
	ePowX := math.Exp(x)
	deviationSquared := math.Pow(deviation, 2)

	return (ePowX*(math.Pow(delta, 2)-deviationSquared-variance-ePowX))/(2*math.Pow(deviationSquared+variance+ePowX, 2)) -
		(x-a)/math.Pow(systemConstant, 2)

}

// calculateNewVolatility == σ′
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
			if fThetaFunction(a-k*systemConstant, systemConstant, volatility, delta, deviation, variance) < 0 {
				k += 1
			} else {
				break
			}
		}
		B = a - k*systemConstant
	}

	fA := fThetaFunction(A, systemConstant, volatility, delta, deviation, variance)
	fB := fThetaFunction(B, systemConstant, volatility, delta, deviation, variance)

	for {
		if math.Abs(B-A) > convergenceTolerance {

			C := A + (A-B)*fA/(fB-fA)

			fC := fThetaFunction(C, systemConstant, volatility, delta, deviation, variance)

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

// calcVolatilityFromMatches == σ′ convenience function
func calcVolatilityFromMatches(playerRating float64, playerDeviation float64, playerVolatility float64, opponentRatings []float64, opponentDeviations []float64, gameOutcomes []float64,
	convergenceTolerance float64, systemConstant float64, variance float64) float64 {

	delta := calculateEstimatedRatingImprovement(playerRating, opponentRatings, opponentDeviations, gameOutcomes, variance)

	return calculateNewVolatility(convergenceTolerance, systemConstant, playerVolatility, delta, playerDeviation, variance)

}

func calcPreRatingDeviation(curDeviation float64, thetaDash float64) float64 {
	return math.Sqrt(math.Pow(curDeviation, 2) + math.Pow(thetaDash, 2))
}

func RawPlayerUpdaterWithSettings(settings Glicko2AlgorithmSettings) func(playerRating float64, playerDeviation float64, playerVolatility float64,
	opponentRatings []float64, opponentDeviations []float64, gameOutcomes []float64) (float64, float64, float64) {

	return func(playerRating float64, playerDeviation float64, playerVolatility float64, opponentRatings []float64, opponentDeviations []float64, gameOutcomes []float64) (float64, float64, float64) {
		variance := calculateVarianceFromGameOutcomes(playerRating, opponentRatings, opponentDeviations)

		thetaDash := calcVolatilityFromMatches(playerRating, playerDeviation, playerVolatility, opponentRatings, opponentDeviations, gameOutcomes, settings.ConvergenceTolerance, settings.SystemConstant, variance)

		newDeviation := 1 / math.Sqrt((1/math.Pow(calcPreRatingDeviation(playerDeviation, thetaDash), 2))+(1/variance))

		var sum float64

		for i := 0; i < len(opponentRatings); i++ {
			sum += step3g(opponentDeviations[i]) * (gameOutcomes[i] - step3E(playerRating, opponentRatings[i], opponentDeviations[i]))
		}

		newRating := playerRating + (math.Pow(newDeviation, 2) * sum)

		return newRating, newDeviation, thetaDash
	}
}

func RawPlayerUpdaterWithDefaultSettings() func(playerRating float64, playerDeviation float64, playerVolatility float64,
	opponentRatings []float64, opponentDeviations []float64, gameOutcomes []float64) (float64, float64, float64) {
	return RawPlayerUpdaterWithSettings(glicko2DefaultSettings)
}

func PlayerUpdaterWithSettings(settings Glicko2AlgorithmSettings) func(player Glicko2Player,
	opponents []Glicko2Player, gameOutcomes []float64) (float64, float64, float64) {

	return func(player Glicko2Player, opponents []Glicko2Player, gameOutcomes []float64) (float64, float64, float64) {
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

func PlayerUpdaterWithDefaultSettings() func(player Glicko2Player,
	opponents []Glicko2Player, gameOutcomes []float64) (float64, float64, float64) {
	return PlayerUpdaterWithSettings(glicko2DefaultSettings)
}

// TODO: Steps are being skipped - vars like system constant are entirely missed
func UpdatePlayerFromMatches(playerRating float64, playerDeviation float64, playerVolatility float64, opponentRatings []float64, opponentDeviations []float64, gameOutcomes []float64) (float64, float64, float64) {
	variance := calculateVarianceFromGameOutcomes(playerRating, opponentRatings, opponentDeviations)

	thetaDash := calcVolatilityFromMatches(playerRating, playerDeviation, playerVolatility, opponentRatings, opponentDeviations, gameOutcomes, GLICKO2_DEFAULT_CONVERGENCE_TOLERANCE, GLICKO2_DEFAULT_SYSTEM_CONSTANT, variance)

	newDeviation := 1 / math.Sqrt((1/math.Pow(calcPreRatingDeviation(playerDeviation, thetaDash), 2))+(1/variance))

	var sum float64

	for i := 0; i < len(opponentRatings); i++ {
		sum += step3g(opponentDeviations[i]) * (gameOutcomes[i] - step3E(playerRating, opponentRatings[i], opponentDeviations[i]))
	}

	newRating := playerRating + (math.Pow(newDeviation, 2) * sum)

	return newRating, newDeviation, thetaDash
}
