package member

import "errors"

// Role es el rol de un integrante en la metodología (AGENTS.md).
type Role string

// Roles válidos de un integrante (RN-15.2).
const (
	RoleProductArchitect Role = "PRODUCT_ARCHITECT"
	RoleAgileEnabler     Role = "AGILE_ENABLER"
	RoleProductBuilder   Role = "PRODUCT_BUILDER"
)

// Member es un integrante de un proyecto (HU-15). No tiene slices ni mapas,
// así que copiar el struct alcanza para devolver una copia (RN-15.10).
type Member struct {
	ID        int
	ProjectID int
	Name      string
	Role      Role
}

var (
	// ErrEmptyName indica que el nombre está vacío o tiene solo espacios
	// (RN-15.1).
	ErrEmptyName = errors.New("el nombre del integrante es obligatorio")
	// ErrMissingRole indica que el rol está vacío (RN-15.2).
	ErrMissingRole = errors.New("el rol del integrante es obligatorio")
	// ErrInvalidRole indica que el rol no es uno de los tres válidos
	// (RN-15.2).
	ErrInvalidRole = errors.New("el rol debe ser PRODUCT_ARCHITECT, AGILE_ENABLER o PRODUCT_BUILDER")
	// ErrDuplicateMember indica que ya hay un integrante con ese nombre en el
	// proyecto (RN-15.4).
	ErrDuplicateMember = errors.New("ya hay un integrante con ese nombre en el proyecto")
)
