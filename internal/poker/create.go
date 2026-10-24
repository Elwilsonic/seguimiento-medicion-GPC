package poker

import (
	"seguimiento-medicion-gpc/internal/backlog"
)

// OpenSession abre una sesión de Planning Poker para un ítem existente
// (HU-21) y devuelve una copia. La sesión queda ABIERTA, en la ronda 1 y sin
// votos (RN-21.3 a RN-21.5).
//
// El orden de validación es: existencia del ítem, ítem TERMINADO, sesión
// abierta y guardado; se informa el primer error encontrado.
func (m *Manager) OpenSession(itemID int) (Session, error) {
	// RN-21.1 y RN-21.6: el ítem se valida antes de tomar el mutex de poker,
	// para no tomar dos mutex a la vez. Si no existe, se propaga el error de
	// HU-04 tal cual.
	item, err := m.items.GetItem(itemID)
	if err != nil {
		return Session{}, err
	}
	if item.Status == backlog.StatusFinished {
		return Session{}, ErrItemFinished
	}

	// RN-21.2: la consulta y el guardado van bajo el mutex para que dos
	// aperturas simultáneas no dejen dos sesiones abiertas (CA-21.6).
	m.mu.Lock()
	defer m.mu.Unlock()

	if _, exists, err := m.repo.OpenByItem(itemID); err != nil {
		return Session{}, err
	} else if exists {
		return Session{}, ErrSessionAlreadyOpen
	}

	return m.repo.Save(Session{
		ItemID: itemID,
		Status: StatusOpen,
		Round:  1,
		Votes:  []Vote{},
	})
}

// Cards devuelve una lista nueva con la escala de story points de HU-20, que
// son las cartas del Planning Poker. HU-22 la usa para validar los votos.
func Cards() []int {
	return backlog.StoryPointScale()
}
