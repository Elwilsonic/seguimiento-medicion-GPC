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

// CompareVotes compara los votos revelados de la ronda (HU-24).
func CompareVotes(votes []Vote) VoteComparison {
	// Stub de la fase RED.
	return VoteComparison{}
}
