package project

import (
	"slices"
	"strings"
)

// List devuelve todos los proyectos (HU-50), como copias, ordenados por nombre
// sin distinguir mayúsculas. Sin proyectos devuelve una lista vacía, no nil.
// Si el repositorio falla, devuelve nil y ese error, sin una lista parcial.
func (m *Manager) List() ([]Project, error) {
	// CA-50.7: la lectura va bajo el mismo mutex con el que Create y Update
	// escriben, porque MemoryRepository no es seguro para uso concurrente.
	m.mu.Lock()
	defer m.mu.Unlock()

	// El repositorio ya devuelve una lista nueva de copias (CA-50.4), así que
	// ordenarla no toca el estado guardado.
	projects, err := m.repo.List()
	if err != nil {
		return nil, err
	}
	// CA-50.2: MemoryRepository devuelve nil cuando no hay proyectos.
	if projects == nil {
		projects = []Project{}
	}
	// CA-50.6 (D2): el orden lo aplica el Manager, no la base, para que sea el
	// mismo en memoria y en PostgreSQL y no dependa del ID (HU-51). Compara
	// bytes en minúscula: los nombres con acento o ñ inicial quedan después
	// de la "z" (limitación aceptada en el design).
	slices.SortStableFunc(projects, func(a, b Project) int {
		return strings.Compare(strings.ToLower(a.Name), strings.ToLower(b.Name))
	})
	return projects, nil
}
