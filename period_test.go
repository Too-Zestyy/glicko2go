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
		1: ConvertToDefaultGlicko2(GlickoPlayer{
			Rating:          1500,
			RatingDeviation: 200,
		}),
		2: ConvertToDefaultGlicko2(GlickoPlayer{
			Rating:          1400,
			RatingDeviation: 30,
		}),
		3: ConvertToDefaultGlicko2(GlickoPlayer{
			Rating:          1550,
			RatingDeviation: 100,
		}),
		4: ConvertToDefaultGlicko2(GlickoPlayer{
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
