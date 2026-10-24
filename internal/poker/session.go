package poker

import (
	"errors"
	"slices"
)

// Status es el estado de una sesión de Planning Poker (RN-21.3). Esta
// historia solo define ABIERTA; HU-23 y HU-25 agregan los demás estados.
type Status string

const (
	// StatusOpen indica que la sesión está en curso y acepta votos.
	StatusOpen Status = "ABIERTA"
)

// Vote es un voto de un integrante en una ronda de Planning Poker. El votante
// se identifica por su nombre libre porque la épica E4 (integrantes) no está
// implementada en esta entrega (decisión 4 de la SDD de HU-21).
type Vote struct {
	Voter  string
	Points int
}

// Session es una sesión de Planning Poker de un ítem del backlog. Tiene un
// slice de votos, así que copiarla no copia la lista: el repositorio y el
// Manager siempre devuelven listas nuevas (RN-21.8).
type Session struct {
	ID     int
	ItemID int
	Status Status
	Round  int
	Votes  []Vote
}

// cloneSession devuelve una copia por valor con su propia lista de votos. Una
// lista nil se normaliza a una lista vacía no nil (RN-21.5 y RN-21.8).
func cloneSession(session Session) Session {
	session.Votes = slices.Clone(session.Votes)
	if session.Votes == nil {
		session.Votes = []Vote{}
	}
	return session
}

var (
	// ErrSessionAlreadyOpen indica que ya hay una sesión ABIERTA para el ítem
	// (RN-21.2).
	ErrSessionAlreadyOpen = errors.New("ya hay una sesión de Planning Poker abierta para ese ítem")
	// ErrItemFinished indica que el ítem está TERMINADO y no se puede abrir una
	// sesión para él (RN-21.6).
	ErrItemFinished = errors.New("no se puede abrir una sesión de Planning Poker para un ítem terminado")
)
