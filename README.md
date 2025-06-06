# glicko2go

`glicko2go` is an implementation of the glicko 2 algorithm in Go. For more information on the algorithm and the base for this implementation, [see the paper detailing an example of its usage](https://www.glicko.net/glicko/glicko2.pdf).

---
# Usage

## Players

Players within this package are represented by `GlickoPlayer` and `Glicko2Player`, and can be converted between.

To create a Glicko 2 player that has a Glicko Rating of 2000 and Glicko Rating Deviation of 100:
```go
package main

import (
	"fmt"
	"github.com/Too-Zestyy/glicko2go"
)

func main() {
	glickoPlayer := glicko2go.GlickoPlayer{
		Rating:          2000,
		RatingDeviation: 100,
	}

	glicko2Player := glicko2go.ConvertToGlicko2(
		glickoPlayer,
		glicko2go.GLICKO2_DEFAULT_PLAYER_VOLATILITY, // Or whatever volatility value is best for your system
	)

	fmt.Printf("Glicko Player within glicko 2 scale:\n"+
		"	- Rating: %v\n"+
		"	- Rating Deviation: %v\n"+
		"	- Volatility: %v\n",
		glicko2Player.Rating, glicko2Player.RatingDeviation, glicko2Player.RatingVolatility,
	)
}
```
```
Glicko Player within glicko 2 scale:
	- Rating: 2.878231246308668
	- Rating Deviation: 0.5756462492617337
	- Volatility: 0.06
```

Functions that provide reasonable shorthands and defaults for a variety of values are provided, in addition to conversions between Glicko and Glicko 2:
```go
package main

import (
	"fmt"
	"github.com/Too-Zestyy/glicko2go"
)

func main() {
	glicko2Player := glicko2go.NewDefaultGlicko2Player()

	fmt.Printf(
		"Default Glicko 2 Player\n"+
			"	- Rating: %v\n"+
			"	- Rating Deviation: %v\n"+
			"	- Volatility: %v\n",
		glicko2Player.Rating, glicko2Player.RatingDeviation, glicko2Player.RatingVolatility)

	fmt.Println("--------------------")

	glickoPlayer := glicko2go.ConvertToGlicko(glicko2Player)

	fmt.Printf("Default Glicko 2 Player on Glicko scale: \n"+
		"	- Rating: %v\n"+
		"	- Rating Deviation: %v\n",
		glickoPlayer.Rating, glickoPlayer.RatingDeviation)
}
```
```
Default Glicko 2 Player
	- Rating: 0
	- Rating Deviation: 2.014761872416068
	- Volatility: 0.06
--------------------
Default Glicko 2 Player on Glicko scale: 
	- Rating: 1500
	- Rating Deviation: 350
```

## Period Updaters

For common use-cases such as a full-stack application with a database storing player stats, period updaters can abstract calculating updates for players over a period, including those who have not played within a period. Either `DefaultPeriodCalculator` or `PeriodCalculatorWithSettings` will be suitable, depending on if you want to use the same constants as within the paper's example.
These functions take a map of `Glicko2Player`s with `int` IDs, and a slice of `Glicko2MatchByID`s - this should be suitable for databases with a player table and match table.

For example, to calculate the output of the example matches within the paper:

```go
package main

import (
	"fmt"
	"github.com/Too-Zestyy/glicko2go"
)

func main() {

	players := map[int]glicko2go.Glicko2Player{
		1: glicko2go.ConvertToGlicko2WithDefaultVolatility(glicko2go.GlickoPlayer{
			Rating:          1500,
			RatingDeviation: 200,
		}),
		2: glicko2go.ConvertToGlicko2WithDefaultVolatility(glicko2go.GlickoPlayer{
			Rating:          1400,
			RatingDeviation: 30,
		}),
		3: glicko2go.ConvertToGlicko2WithDefaultVolatility(glicko2go.GlickoPlayer{
			Rating:          1550,
			RatingDeviation: 100,
		}),
		4: glicko2go.ConvertToGlicko2WithDefaultVolatility(glicko2go.GlickoPlayer{
			Rating:          1700,
			RatingDeviation: 300,
		}),
		// Another player that is not within the example, and so plays no matches within the period
		5: glicko2go.ConvertToGlicko2WithDefaultVolatility(glicko2go.GlickoPlayer{
			Rating:          1500,
			RatingDeviation: 300,
		}),
	}

	matches := []glicko2go.Glicko2MatchByID{
		{
			Player1ID: 1,
			Player2ID: 2,
			Result:    glicko2go.GAME_OUTCOME_WIN,
		},
		{
			Player1ID: 1,
			Player2ID: 3,
			Result:    glicko2go.GAME_OUTCOME_LOSS,
		},
		{
			Player1ID: 1,
			Player2ID: 4,
			Result:    glicko2go.GAME_OUTCOME_LOSS,
		},
	}

	periodUpdater := glicko2go.DefaultPeriodCalculator()

	playersAfterPeriod, err := periodUpdater(players, matches)

	if err != nil {
		panic(err)
	}

	for id, player := range playersAfterPeriod {
		glicko1PlayerStats := glicko2go.ConvertToGlicko(player)
		fmt.Printf(
			"Player %v post-period: \n"+
				"	- Rating: %v\n"+
				"	- Rating Deviation: %v\n"+
				"	- Volatility: %v\n"+
				"--------------------\n",
			id, glicko1PlayerStats.Rating, glicko1PlayerStats.RatingDeviation, player.RatingVolatility)
	}

}
```
```
Player 1 post-period: 
	- Rating: 1464.0506705393013
	- Rating Deviation: 151.51652412385727
	- Volatility: 0.059995984286488495
--------------------
Player 2 post-period: 
	- Rating: 1398.1435582337338
	- Rating Deviation: 31.67021528115062
	- Volatility: 0.05999912372888531
--------------------
Player 3 post-period: 
	- Rating: 1570.394740240854
	- Rating Deviation: 97.70916852200307
	- Volatility: 0.05999941947199381
--------------------
Player 4 post-period: 
	- Rating: 1784.4217901320874
	- Rating Deviation: 251.56556453224735
	- Volatility: 0.059999011763670944
--------------------
Player 5 post-period: 
	- Rating: 1500
	- Rating Deviation: 300.18101263493105
	- Volatility: 0.06
--------------------
```
## Advanced usage

For those that need a more specific interface, there are public functions at various levels of abstraction, with the lowest being `UpdatePlayerFromMatches`. This function is the base for all abstracted functions provided (such as the period updaters) and will allow anyone to create a custom interface for their needs.

To update Player 1 from the aforementioned example:

```go
package main

import (
	"fmt"
	"github.com/Too-Zestyy/glicko2go"
)

func main() {

	opponentRatings := []float64{
		glicko2go.GlickoRatingToGlicko2(1400),
		glicko2go.GlickoRatingToGlicko2(1550),
		glicko2go.GlickoRatingToGlicko2(1700),
	}
	opponentDeviations := []float64{
		glicko2go.GlickoDeviationToGlicko2(30),
		glicko2go.GlickoDeviationToGlicko2(100),
		glicko2go.GlickoDeviationToGlicko2(300),
	}
	gameOutcomes := []float64{glicko2go.GAME_OUTCOME_WIN, glicko2go.GAME_OUTCOME_LOSS, glicko2go.GAME_OUTCOME_LOSS}

	newRating, newDeviation, newVolatility, err := glicko2go.UpdatePlayerFromMatches(
		glicko2go.GlickoRatingToGlicko2(1500),       // The Rating of the player to update
		glicko2go.GlickoDeviationToGlicko2(200),     // The Deviation of the player to update
		glicko2go.GLICKO2_DEFAULT_PLAYER_VOLATILITY, // The Volatility of the player to update
		opponentRatings,
		opponentDeviations,
		gameOutcomes,
		glicko2go.Glicko2AlgorithmSettings{
			SystemConstant:       glicko2go.GLICKO2_DEFAULT_SYSTEM_CONSTANT,
			ConvergenceTolerance: glicko2go.GLICKO2_DEFAULT_CONVERGENCE_TOLERANCE,
		},
	)

	if err != nil {
		panic(err)
	}

	fmt.Printf(
		"Player after period: \n"+
			"	- Rating: %v\n"+
			"	- Deviation: %v\n"+
			"	- Volatility: %v\n",
		glicko2go.Glicko2RatingToGlicko(newRating), glicko2go.Glicko2DeviationToGlicko(newDeviation), newVolatility)

}
```
```
Player after period: 
	- Rating: 1464.0506705393013
	- Deviation: 151.51652412385727
	- Volatility: 0.059995984286488495
```
