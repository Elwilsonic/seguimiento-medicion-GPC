package httpapi

import (
	"net/http"

	"seguimiento-medicion-gpc/internal/poker"
)

type pokerSessionResponse struct {
	ID     int                 `json:"id"`
	ItemID int                 `json:"item_id"`
	Estado string              `json:"estado"`
	Ronda  int                 `json:"ronda"`
	Votos  []pokerVoteResponse `json:"votos"`
	Cartas []int               `json:"cartas"`
}

type pokerVoteResponse struct {
	Integrante string `json:"integrante"`
	Puntos     int    `json:"puntos"`
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
