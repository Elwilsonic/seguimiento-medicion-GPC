package backlog

import (
	"errors"
	"strings"
)

// ErrInvalidStatus indica que el estado destino no es PENDIENTE, EN_CURSO ni
// TERMINADO después de normalizarlo (RN-08.1).
var ErrInvalidStatus = errors.New("el estado del ítem no es válido")

// ErrInvalidTransition indica que el paso entre el estado actual y el destino
// no está permitido (RN-08.2).
var ErrInvalidTransition = errors.New("la transición de estado no está permitida")

// allowedTransitions define los cambios permitidos: un paso por vez, hacia
// adelante o hacia atrás (RN-08.2).
var allowedTransitions = map[Status][]Status{
	StatusPending:    {StatusInProgress},
	StatusInProgress: {StatusFinished, StatusPending},
	StatusFinished:   {StatusInProgress},
}

// ChangeStatus cambia el estado de un ítem existente (HU-08) y devuelve una
// copia con su propia lista de criterios. El estado destino se normaliza y se
// valida antes de consultar el repositorio (RN-08.1); después, con el mutex, se
// verifica la existencia (RN-08.4), el mismo estado (RN-08.3), la transición
// (RN-08.2) y se escribe con Update (RN-08.6). Solo cambia el estado (RN-08.5).
func (m *Manager) ChangeStatus(id int, status string) (Item, error) {
	normalized, err := normalizeStatus(status)
	if err != nil {
		return Item{}, err
	}

	m.mu.Lock()
	defer m.mu.Unlock()

	current, err := m.repo.GetByID(id)
	if err != nil {
		return Item{}, err
	}

	// RN-08.3: el mismo estado es un no-op válido, sin llamar a Update.
	if current.Status == normalized {
		return current, nil
	}

	if !isTransitionAllowed(current.Status, normalized) {
		return Item{}, ErrInvalidTransition
	}

	current.Status = normalized
	return m.repo.Update(current)
}

// normalizeStatus valida el estado destino contra PENDIENTE, EN_CURSO y
// TERMINADO, sin distinguir mayúsculas ni contar los espacios de los extremos
// (RN-08.1).
func normalizeStatus(status string) (Status, error) {
	normalized := Status(strings.ToUpper(strings.TrimSpace(status)))
	switch normalized {
	case StatusPending, StatusInProgress, StatusFinished:
		return normalized, nil
	default:
		return "", ErrInvalidStatus
	}
}

// isTransitionAllowed indica si se puede pasar de from a to (RN-08.2).
func isTransitionAllowed(from, to Status) bool {
	for _, allowed := range allowedTransitions[from] {
		if allowed == to {
			return true
		}
	}
	return false
}
