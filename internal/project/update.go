package project

import (
	"errors"
	"strings"
)

// ErrInvalidStatus indica que el estado recibido, normalizado, no es uno de
// los estados válidos (RN-02.4, HU-02).
var ErrInvalidStatus = errors.New("el estado del proyecto no es válido")

// Update modifica el nombre, la descripción y el estado de un proyecto
// existente (HU-02) y devuelve una copia. El estado llega como texto y se
// normaliza adentro, igual que la prioridad en HU-04.
//
// El orden de validación es: nombre, estado, existencia, nombre duplicado y
// guardado; se informa el primer error encontrado y, si algo falla, el
// proyecto guardado queda como estaba.
func (m *Manager) Update(id int, name, description, status string) (Project, error) {
	// RN-02.1 y RN-02.2: el nombre se guarda sin espacios en blanco Unicode
	// en los extremos (strings.TrimSpace), y no puede quedar vacío.
	name = strings.TrimSpace(name)
	if name == "" {
		return Project{}, ErrEmptyName
	}

	// RN-02.4: se acepta el estado en cualquier combinación de mayúsculas y
	// minúsculas y con espacios en los extremos; se normaliza y se valida.
	normalized, err := normalizeStatus(status)
	if err != nil {
		return Project{}, err
	}

	m.mu.Lock()
	defer m.mu.Unlock()

	// RN-02.6: el proyecto tiene que existir. Se reutiliza la búsqueda por ID
	// de HU-03 y su error se propaga tal cual.
	current, err := m.repo.GetByID(id)
	if err != nil {
		return Project{}, err
	}

	// RN-02.3: el nombre nuevo no puede repetir el de otro proyecto (se
	// compara sin distinguir mayúsculas); el propio proyecto se excluye.
	projects, err := m.repo.List()
	if err != nil {
		return Project{}, err
	}
	for _, p := range projects {
		if p.ID != id && strings.EqualFold(p.Name, name) {
			return Project{}, ErrDuplicateName
		}
	}

	current.Name = name
	current.Description = description
	current.Status = normalized
	return m.repo.Update(current)
}

// normalizeStatus valida el estado contra PLANIFICADO, EN_CURSO y FINALIZADO
// (RN-02.4), sin distinguir mayúsculas ni contar los espacios de los extremos.
func normalizeStatus(status string) (Status, error) {
	normalized := Status(strings.ToUpper(strings.TrimSpace(status)))
	switch normalized {
	case StatusPlanned, StatusInProgress, StatusFinished:
		return normalized, nil
	default:
		return "", ErrInvalidStatus
	}
}
