package poker

// VoteGroup son los puntos de un extremo de la comparación y los integrantes
// que los votaron (HU-24).
type VoteGroup struct {
	Points int
	Voters []string
}

// VoteComparison indica si los votos revelados tienen diferencias y cuáles
// son el voto menor y el mayor (RN-24.1 y RN-24.2).
type VoteComparison struct {
	HasDifferences bool
	Lowest         VoteGroup
	Highest        VoteGroup
}

// CompareVotes compara los votos revelados de la ronda (HU-24). Hay
// diferencia si no todos tienen los mismos puntos; el menor y el mayor llevan
// a sus integrantes en el orden de los votos. Es una función pura y devuelve
// listas nuevas (RN-24.10).
func CompareVotes(votes []Vote) VoteComparison {
	comparison := VoteComparison{
		Lowest:  VoteGroup{Voters: []string{}},
		Highest: VoteGroup{Voters: []string{}},
	}
	if len(votes) == 0 {
		return comparison
	}

	comparison.Lowest.Points = votes[0].Points
	comparison.Highest.Points = votes[0].Points
	for _, vote := range votes[1:] {
		comparison.Lowest.Points = min(comparison.Lowest.Points, vote.Points)
		comparison.Highest.Points = max(comparison.Highest.Points, vote.Points)
	}
	for _, vote := range votes {
		if vote.Points == comparison.Lowest.Points {
			comparison.Lowest.Voters = append(comparison.Lowest.Voters, vote.Voter)
		}
		if vote.Points == comparison.Highest.Points {
			comparison.Highest.Voters = append(comparison.Highest.Voters, vote.Voter)
		}
	}
	comparison.HasDifferences = comparison.Lowest.Points != comparison.Highest.Points
	return comparison
}
