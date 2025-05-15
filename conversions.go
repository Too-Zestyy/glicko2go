package glicko2go

const (
	GLICKO2_DEFAULT_PLAYER_VOLATILITY float64 = 0.06
)

func GlickoRatingToGlicko2(gRating float64) float64 {
	return (gRating - 1500) / 173.7178
}

func Glicko2RatingtoGlicko(g2Rating float64) float64 {
	return g2Rating*173.7178 + 1500
}

func GlickoDeviationToGlicko2(gDeviation float64) float64 {
	return gDeviation / 173.7178
}

func Glicko2DeviationToGlicko(g2Deviation float64) float64 {
	return g2Deviation * 173.7178
}

func ConvertToGlicko2(gp GlickoPlayer, volatility float64) Glicko2Player {
	return Glicko2Player{
		GlickoPlayer: GlickoPlayer{
			Rating:          GlickoRatingToGlicko2(gp.Rating),
			RatingDeviation: GlickoDeviationToGlicko2(gp.RatingDeviation),
		},
		RatingVolatility: volatility,
	}
}

func ConvertToDefaultGlicko2(gp GlickoPlayer) Glicko2Player {
	return ConvertToGlicko2(gp, GLICKO2_DEFAULT_PLAYER_VOLATILITY)
}

func ConvertToGlicko1(g2p Glicko2Player) GlickoPlayer {
	return GlickoPlayer{
		Rating:          Glicko2RatingtoGlicko(g2p.Rating),
		RatingDeviation: Glicko2DeviationToGlicko(g2p.RatingDeviation),
	}
}
