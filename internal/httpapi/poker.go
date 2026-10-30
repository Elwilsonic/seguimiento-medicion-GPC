package httpapi

import (
	"net/http"

	"seguimiento-medicion-gpc/internal/poker"
)

type pokerVoteRequest struct {
	Integrante string `json:"integrante"`
	Puntos     int    `json:"puntos"`
}

type pokerSessionResponse struct {
	ID     int                 `json:"id"`
	ItemID int                 `json:"item_id"`
	Estado string              `json:"estado"`
	Ronda  int                 `json:"ronda"`
	Votos  []pokerVoteResponse `json:"votos"`
	Cartas []int               `json:"cartas"`
	// Comparacion solo la completa la revelación (HU-24, D2); las demás rutas
	// la omiten y su JSON no cambia.
	Comparacion *pokerComparisonResponse `json:"comparacion,omitempty"`
}

type pokerVoteResponse struct {
	Integrante string `json:"integrante"`
	Puntos     int    `json:"puntos"`
}

type pokerComparisonResponse struct {
	HayDiferencias bool                   `json:"hay_diferencias"`
	Menor          pokerVoteGroupResponse `json:"menor"`
	Mayor          pokerVoteGroupResponse `json:"mayor"`
}

type pokerVoteGroupResponse struct {
	Puntos      int      `json:"puntos"`
	Integrantes []string `json:"integrantes"`
}

// methodNotAllowed responde 405 a los métodos distintos de POST en la ruta de
// sesiones de Planning Poker (CA-21.7).
func methodNotAllowed(w http.ResponseWriter, _ *http.Request) {
	w.Header().Set("Allow", http.MethodPost)
	http.Error(w, http.StatusText(http.StatusMethodNotAllowed), http.StatusMethodNotAllowed)
}

// openPokerSession atiende POST /api/items/{id}/poker-sessions: abre una
// sesión de Planning Poker para el ítem (HU-21, CA-21.7).
func (s *server) openPokerSession(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		w.Header().Set("Allow", http.MethodPost)
		http.Error(w, http.StatusText(http.StatusMethodNotAllowed), http.StatusMethodNotAllowed)
		return
	}

	id, err := pathID(r)
	if err != nil {
		writeError(w, err)
		return
	}

	session, err := s.poker.OpenSession(id)
	if err != nil {
		writeError(w, err)
		return
	}

	votos := make([]pokerVoteResponse, len(session.Votes))
	for i, voto := range session.Votes {
		votos[i] = pokerVoteResponse{Integrante: voto.Voter, Puntos: voto.Points}
	}
	writeJSON(w, http.StatusCreated, pokerSessionResponse{
		ID:     session.ID,
		ItemID: session.ItemID,
		Estado: string(session.Status),
		Ronda:  session.Round,
		Votos:  votos,
		Cartas: poker.Cards(),
	})
}

func (s *server) votePokerSession(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		w.Header().Set("Allow", http.MethodPost)
		http.Error(w, http.StatusText(http.StatusMethodNotAllowed), http.StatusMethodNotAllowed)
		return
	}
	id, err := pathID(r)
	if err != nil {
		writeError(w, err)
		return
	}
	var request pokerVoteRequest
	if err := decodeJSON(w, r, &request); err != nil {
		writeError(w, err)
		return
	}
	session, err := s.poker.Vote(id, request.Integrante, request.Puntos)
	if err != nil {
		writeError(w, err)
		return
	}
	writeJSON(w, http.StatusOK, pokerSessionResponse{
		ID: session.ID, ItemID: session.ItemID, Estado: string(session.Status),
		Ronda: session.Round, Votos: []pokerVoteResponse{}, Cartas: poker.Cards(),
	})
}

// revealPokerSession atiende POST /api/poker-sessions/{id}/reveal: revela
// todos los votos de la sesión (HU-23, CA-23.8). La ruta no recibe cuerpo; si
// llega uno, se ignora.
func (s *server) revealPokerSession(w http.ResponseWriter, r *http.Request) {
	id, err := pathID(r)
	if err != nil {
		writeError(w, err)
		return
	}
	session, err := s.poker.Reveal(id)
	if err != nil {
		writeError(w, err)
		return
	}
	votos := make([]pokerVoteResponse, len(session.Votes))
	for i, voto := range session.Votes {
		votos[i] = pokerVoteResponse{Integrante: voto.Voter, Puntos: voto.Points}
	}
	// HU-24, RN-24.3: la comparación se calcula sobre los votos revelados.
	comparison := poker.CompareVotes(session.Votes)
	writeJSON(w, http.StatusOK, pokerSessionResponse{
		ID: session.ID, ItemID: session.ItemID, Estado: string(session.Status),
		Ronda: session.Round, Votos: votos, Cartas: poker.Cards(),
		Comparacion: &pokerComparisonResponse{
			HayDiferencias: comparison.HasDifferences,
			Menor:          pokerVoteGroupResponse{Puntos: comparison.Lowest.Points, Integrantes: comparison.Lowest.Voters},
			Mayor:          pokerVoteGroupResponse{Puntos: comparison.Highest.Points, Integrantes: comparison.Highest.Voters},
		},
	})
}

// repeatRoundPokerSession atiende POST /api/poker-sessions/{id}/rounds:
// repite la ronda de una sesión revelada (HU-24, CA-24.12). La ruta no recibe
// cuerpo; si llega uno, se ignora.
func (s *server) repeatRoundPokerSession(w http.ResponseWriter, r *http.Request) {
	id, err := pathID(r)
	if err != nil {
		writeError(w, err)
		return
	}
	session, err := s.poker.RepeatRound(id)
	if err != nil {
		writeError(w, err)
		return
	}
	writeJSON(w, http.StatusOK, pokerSessionResponse{
		ID: session.ID, ItemID: session.ItemID, Estado: string(session.Status),
		Ronda: session.Round, Votos: []pokerVoteResponse{}, Cartas: poker.Cards(),
	})
}

type pokerEstimateRequest struct {
	Puntos int `json:"puntos"`
}

// estimatePokerSession atiende POST /api/poker-sessions/{id}/estimate:
// registra la estimación acordada en el ítem y cierra la sesión (HU-25,
// CA-25.9). Los story points quedan visibles en el ítem.
func (s *server) estimatePokerSession(w http.ResponseWriter, r *http.Request) {
	id, err := pathID(r)
	if err != nil {
		writeError(w, err)
		return
	}
	var request pokerEstimateRequest
	if err := decodeJSON(w, r, &request); err != nil {
		writeError(w, err)
		return
	}
	session, err := s.poker.Estimate(id, request.Puntos)
	if err != nil {
		writeError(w, err)
		return
	}
	votos := make([]pokerVoteResponse, len(session.Votes))
	for i, voto := range session.Votes {
		votos[i] = pokerVoteResponse{Integrante: voto.Voter, Puntos: voto.Points}
	}
	writeJSON(w, http.StatusOK, pokerSessionResponse{
		ID: session.ID, ItemID: session.ItemID, Estado: string(session.Status),
		Ronda: session.Round, Votos: votos, Cartas: poker.Cards(),
	})
}
