package effort

import "time"

func (m *Manager) Register(sprintID, itemID int, member string, date time.Time, activity string, hours float64) (Effort, error) {
	return Effort{}, errNotImplemented
}
