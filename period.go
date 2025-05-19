package glicko2go

// TODO: Add function that allows for matches to be direct assignment of players instead of IDs

func Glicko2PeriodCalculatorWithSettings(settings Glicko2AlgorithmSettings) func(
	players map[int]Glicko2Player,
	matches []Glicko2MatchByID) (map[int]Glicko2Player, error) {

	playerUpdater := PlayerUpdaterWithSettings(settings)

	return func(players map[int]Glicko2Player, matches []Glicko2MatchByID) (map[int]Glicko2Player, error) {

		updatedPlayers := make(map[int]Glicko2Player)

		// Should be faster than looping through all matches per player
		matchLists := make(map[int]Glicko2MatchSet)
		for _, match := range matches {

			// Add the original match to player 1's match list
			if p1List, ok := matchLists[match.Player1ID]; ok {
				p1List.Opponents = append(p1List.Opponents, players[match.Player2ID])
				p1List.Results = append(p1List.Results, match.Result)

				matchLists[match.Player1ID] = p1List
			} else {
				matchLists[match.Player1ID] = Glicko2MatchSet{
					Opponents: []Glicko2Player{players[match.Player2ID]},
					Results:   []float64{match.Result},
				}
			}

			// The same match is implicitly the opposite for player 2 - P1 winning against P2 == P2 losing against P1
			if p2List, ok := matchLists[match.Player2ID]; ok {
				p2List.Opponents = append(p2List.Opponents, players[match.Player1ID])
				p2List.Results = append(p2List.Results, 1-match.Result)

				matchLists[match.Player2ID] = p2List
			} else {
				matchLists[match.Player1ID] = Glicko2MatchSet{
					Opponents: []Glicko2Player{players[match.Player2ID]},
					Results:   []float64{1 - match.Result},
				}
			}
		}

		for playerID, player := range players {
			playerMatchList := Glicko2MatchSet{}
			if filledMatchList, ok := matchLists[playerID]; ok {
				playerMatchList = filledMatchList
			}

			updatedPlayer, err := playerUpdater(player, playerMatchList.Opponents, playerMatchList.Results)

			if err != nil {
				return nil, err
			}

			updatedPlayers[playerID] = updatedPlayer

		}

		return updatedPlayers, nil
	}
}

func DefaultGlicko2PeriodCalculator() func(
	players map[int]Glicko2Player,
	matches []Glicko2MatchByID) (map[int]Glicko2Player, error) {
	return Glicko2PeriodCalculatorWithSettings(glicko2DefaultSettings)
}
