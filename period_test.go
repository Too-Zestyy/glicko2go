package glicko2go

import (
	"fmt"
	"testing"
)

// getAllBinaryPermutationsOfLength returns all binary permutations of length `len`.
// Another way of phrasing this is returning all binary values with `len` bits.
func getAllBinaryPermutationsOfLength(len int) ([][]bool, error) {

	if len < 1 {
		return nil, fmt.Errorf("length of each permutation cannot be less than 1. Got: %v", len)
	}

	permutationCount := 2 << (len - 1)

	permutations := make([][]bool, permutationCount)

	for i := 0; i < permutationCount; i++ {
		remainder := i
		for binaryDigitVal := permutationCount >> 1; binaryDigitVal >= 1; binaryDigitVal = binaryDigitVal >> 1 {
			if remainder >= binaryDigitVal {
				remainder -= binaryDigitVal
				permutations[i] = append(permutations[i], true)
			} else {
				permutations[i] = append(permutations[i], false)
			}
		}
	}

	return permutations, nil
}

// getExamplePlayers returns the players defined within `Example of the Glicko-2 system`.
func getExamplePlayers() map[int]Glicko2Player {
	return map[int]Glicko2Player{
		1: ConvertToGlicko2WithDefaultVolatility(GlickoPlayer{
			Rating:          1500,
			RatingDeviation: 200,
		}),
		2: ConvertToGlicko2WithDefaultVolatility(GlickoPlayer{
			Rating:          1400,
			RatingDeviation: 30,
		}),
		3: ConvertToGlicko2WithDefaultVolatility(GlickoPlayer{
			Rating:          1550,
			RatingDeviation: 100,
		}),
		4: ConvertToGlicko2WithDefaultVolatility(GlickoPlayer{
			Rating:          1700,
			RatingDeviation: 300,
		}),
	}
}

// getExampleMatchList returns the example set of matches defined alongside getExamplePlayers.
func getExampleMatchList() []Glicko2MatchByID {
	return []Glicko2MatchByID{
		{
			Player1ID: 1,
			Player2ID: 2,
			Result:    GAME_OUTCOME_WIN,
		},
		{
			Player1ID: 1,
			Player2ID: 3,
			Result:    GAME_OUTCOME_LOSS,
		},
		{
			Player1ID: 1,
			Player2ID: 4,
			Result:    GAME_OUTCOME_LOSS,
		},
	}
}

// getExamplePlayersAndMatches is a convenience function to get the players and matches defined in
// `Example of the Glicko-2 system`'s example. See: https://www.glicko.net/glicko/glicko2.pdf
func getExamplePlayersAndMatches() (map[int]Glicko2Player, []Glicko2MatchByID) {
	return getExamplePlayers(), getExampleMatchList()
}

// TestPeriodWinAndLoss is a smoke test to ensure thet winning and losing correctly puts 1 of 2 identical players above the other.
func TestPeriodWinAndLoss(t *testing.T) {

	// Use 2 previously unrated players
	players := map[int]Glicko2Player{
		1: NewDefaultGlicko2Player(),
		2: NewDefaultGlicko2Player(),
	}

	periodCalculator := DefaultPeriodCalculator()

	postP1WinPlayers, err := periodCalculator(players, []Glicko2MatchByID{
		{
			Player1ID: 1,
			Player2ID: 2,
			Result:    GAME_OUTCOME_WIN,
		},
	})
	if err != nil {
		t.Errorf("Error in period calculator for P1 winning a match vs P2: %v", err)
	}
	if postP1WinPlayers[1].Rating <= postP1WinPlayers[2].Rating {
		t.Errorf("P1 winning vs P2 does not place P1 above P2. \nP1: %v\nP2: %v", postP1WinPlayers[1], postP1WinPlayers[2])
	}
	if postP1WinPlayers[1].Rating <= 0 {
		t.Errorf("P1 does not gain rating after winning vs an identical player. \nP1: %v", postP1WinPlayers[1])
	}
	if postP1WinPlayers[2].Rating >= 0 {
		t.Errorf("P2 does not lose rating after losing vs an identical player. \nP2: %v", postP1WinPlayers[2])
	}

	postP1LossPlayers, err := periodCalculator(players, []Glicko2MatchByID{
		{
			Player1ID: 1,
			Player2ID: 2,
			Result:    GAME_OUTCOME_LOSS,
		},
	})
	if err != nil {
		t.Errorf("Error in period calculator for P1 losing a match vs P2: %v", err)
	}
	if postP1LossPlayers[1].Rating >= postP1LossPlayers[2].Rating {
		t.Errorf("P1 losing vs P2 does not place P1 below P2. \nP1: %v\nP2: %v", postP1LossPlayers[1], postP1LossPlayers[2])
	}
	if postP1LossPlayers[1].Rating >= 0 {
		t.Errorf("P1 does not lose rating after lose vs an identical player. \nP1: %v", postP1LossPlayers[1])
	}
	if postP1LossPlayers[2].Rating <= 0 {
		t.Errorf("P2 does not gain rating after winning vs an identical player. \nP2: %v", postP1LossPlayers[2])
	}
}

// TestPeriodDraw is a smoke test to ensure that identical players are not updated erratically after a draw
func TestPeriodDraw(t *testing.T) {

	// Use 2 previously unrated players
	players := map[int]Glicko2Player{
		1: NewDefaultGlicko2Player(),
		2: NewDefaultGlicko2Player(),
	}
	periodCalculator := DefaultPeriodCalculator()
	postPeriodPlayers, err := periodCalculator(players, []Glicko2MatchByID{
		{
			Player1ID: 1,
			Player2ID: 2,
			Result:    GAME_OUTCOME_DRAW,
		},
	})

	if err != nil {
		t.Errorf("Error when calculating a match draw: %v", err)
	}
	if postPeriodPlayers[1].Rating != 0 || postPeriodPlayers[2].Rating != 0 {
		t.Errorf("Player rating(s) have changed after a draw when both are identical. \nP1: %v\nP2: %v", postPeriodPlayers[1].Rating, postPeriodPlayers[2].Rating)
	}

	glicko2DefaultDeviation := GlickoDeviationToGlicko2(GLICKO_DEFAULT_PLAYER_DEVIATION)
	if postPeriodPlayers[1].RatingDeviation >= glicko2DefaultDeviation || postPeriodPlayers[2].RatingDeviation >= glicko2DefaultDeviation {
		t.Errorf("Player deviation(s) have not decreased after a draw when both are identical. \nP1: %v\nP2: %v", postPeriodPlayers[1].RatingDeviation, postPeriodPlayers[2].RatingDeviation)
	}

	if postPeriodPlayers[1].RatingVolatility >= GLICKO2_DEFAULT_PLAYER_VOLATILITY || postPeriodPlayers[2].RatingVolatility >= GLICKO2_DEFAULT_PLAYER_VOLATILITY {
		t.Errorf("Player volatility(s) have not decreased after a draw when both are identical. \nP1: %v\nP2: %v", postPeriodPlayers[1].RatingVolatility, postPeriodPlayers[2].RatingVolatility)
	}

}

// TestPeriodCalculatorMatchInversion ensures that period calculators correctly fill opposing player's matches
// by inverting matches from getExamplePlayersAndMatches in every combination.
// If inverting a match causes a different result after the period, then the test fails.
func TestPeriodCalculatorMatchInversion(t *testing.T) {
	players, matchList := getExamplePlayersAndMatches()

	periodCalculator := DefaultPeriodCalculator()
	referencePostPeriodPlayers, err := periodCalculator(players, matchList)

	// Implies that there is an error in calculation, as opposed to inversions causing issues
	if err != nil {
		t.Fatalf("Error calculating reference results: %v", err)
	}

	// Used to test every combination of inversions
	inversionPermutations, err := getAllBinaryPermutationsOfLength(len(matchList))
	if err != nil {
		t.Fatalf("Error getting inversion inversionPermutations: %v", err)
	}

	for _, permutation := range inversionPermutations {
		matchListWithInversionPermutation := getExampleMatchList()
		for matchIdx, isInverted := range permutation {
			if isInverted {
				curMatch := matchListWithInversionPermutation[matchIdx]
				invertedMatch := Glicko2MatchByID{
					Player1ID: curMatch.Player2ID,
					Player2ID: curMatch.Player1ID,
					Result:    1 - curMatch.Result,
				}
				matchListWithInversionPermutation[matchIdx] = invertedMatch
			}
		}

		postPeriodPlayers, err := periodCalculator(players, matchListWithInversionPermutation)
		if err != nil {
			t.Fatalf("Error calculating results: %v\nInversion configuration: %v", err, permutation)
		}

		for id := range postPeriodPlayers {
			if referencePostPeriodPlayers[id] != postPeriodPlayers[id] {
				t.Errorf("Player %v changes when a match is inverted \nPre-Inversion:  %v\nPost-Inversion: %v\nInversion configuration: %v",
					id, referencePostPeriodPlayers[id], postPeriodPlayers[id], permutation)
			}
		}
	}

}
