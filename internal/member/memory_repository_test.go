package member

import "testing"

// El repositorio en memoria cumple el contrato de Repository (CA-15.6).
func TestMemoryRepository_Contrato(t *testing.T) {
	runRepositoryContract(t, func(t *testing.T) Repository {
		return NewMemoryRepository()
	})
}
