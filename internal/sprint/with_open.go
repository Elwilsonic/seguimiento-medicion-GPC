package sprint

import "errors"

func (m *Manager) WithOpenSprint(sprintID int, fn func(Sprint) error) error {
	return errors.New("no implementado")
}
