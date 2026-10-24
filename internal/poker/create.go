package poker

// OpenSession abre una sesión de Planning Poker para un ítem existente
// (HU-21) y devuelve una copia. La sesión queda ABIERTA, en la ronda 1 y sin
// votos (RN-21.3 a RN-21.5).
func (m *Manager) OpenSession(itemID int) (Session, error) {
	return Session{}, errNotImplemented
}

// Cards devuelve una lista nueva con la escala de story points de HU-20, que
// son las cartas del Planning Poker. HU-22 la usa para validar los votos.
func Cards() []int {
	return nil
}
