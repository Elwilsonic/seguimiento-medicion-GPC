package metrics

import "strings"

// anyMember acepta cualquier nombre tal como llega. Lo usan los tests de las
// historias anteriores a HU-54 que arman su propio handler sin registrar
// integrantes (D5 de HU-54).
type anyMember struct{}

func (anyMember) CanonicalName(_ int, name string) (string, error) {
	return strings.TrimSpace(name), nil
}
