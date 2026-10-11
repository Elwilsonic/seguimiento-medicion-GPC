package project

import "slices"

// MemoryRepository guarda los proyectos en memoria (Sprint 1). No es seguro
// para uso concurrente: se accede a través del Manager, que lo protege.
type MemoryRepository struct {
	projects []Project
	lastID   int
}

// NewMemoryRepository crea un repositorio en memoria vacío.
func NewMemoryRepository() *MemoryRepository {
	return &MemoryRepository{}
}

// Save asigna el siguiente ID correlativo (RN-01.5) y guarda una copia del proyecto.
func (r *MemoryRepository) Save(p Project) (Project, error) {
	r.lastID++
	p.ID = r.lastID
	r.projects = append(r.projects, p)
	return p, nil
}

// List devuelve una copia de los proyectos guardados.
func (r *MemoryRepository) List() ([]Project, error) {
	return slices.Clone(r.projects), nil
}

// GetByID devuelve una copia del proyecto con ese ID, o ErrProjectNotFound
// si no existe (HU-03).
func (r *MemoryRepository) GetByID(id int) (Project, error) {
	for _, p := range r.projects {
		if p.ID == id {
			return p, nil
		}
	}
	return Project{}, ErrProjectNotFound
}
