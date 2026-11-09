package member

import "testing"

// HU-15 · CA-15.6 · Verifica que el repositorio en memoria cumpla el contrato de Repository.
// El repositorio en memoria cumple el contrato de Repository (CA-15.6).
func TestMemoryRepository_Contrato(t *testing.T) {
	runRepositoryContract(t, func(t *testing.T) Repository {
		return NewMemoryRepository()
	})
}
