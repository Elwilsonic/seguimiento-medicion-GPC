package member

import (
	"errors"
	"testing"
)

// HU-54 · Votar y registrar esfuerzo solo con integrantes del proyecto (Issue #219).
// Los proyectos de memberEnv son el 1 ("Sistema de ventas") y el 2 ("Inventario").

// HU-54 · CA-54.4 · Verifica que CanonicalName devuelva el nombre registrado de un integrante del proyecto.
func TestCanonicalName_DevuelveElNombreRegistrado(t *testing.T) {
	env := newMemberEnv(t)
	env.register(t, 1, "Sosa Ricardo", "PRODUCT_BUILDER")

	got, err := env.members.CanonicalName(1, "Sosa Ricardo")

	if err != nil || got != "Sosa Ricardo" {
		t.Fatalf("CanonicalName = %q, %v; se esperaba \"Sosa Ricardo\", nil", got, err)
	}
}

// HU-54 · CA-54.4 · Verifica que CanonicalName ignore las mayúsculas y los espacios de los extremos, como RN-15.4.
func TestCanonicalName_IgnoraMayusculasYEspacios(t *testing.T) {
	env := newMemberEnv(t)
	env.register(t, 1, "García Santiago", "PRODUCT_BUILDER")

	for _, written := range []string{"  garcía santiago ", "GARCÍA SANTIAGO"} {
		got, err := env.members.CanonicalName(1, written)
		if err != nil || got != "García Santiago" {
			t.Errorf("CanonicalName(%q) = %q, %v; se esperaba \"García Santiago\", nil", written, got, err)
		}
	}
}

// HU-54 · CA-54.2 · Verifica que CanonicalName devuelva ErrNotProjectMember con un nombre que no es integrante del proyecto.
func TestCanonicalName_NoIntegrante(t *testing.T) {
	env := newMemberEnv(t)
	env.register(t, 1, "Sosa Ricardo", "PRODUCT_BUILDER")

	got, err := env.members.CanonicalName(1, "hernan")

	if !errors.Is(err, ErrNotProjectMember) || got != "" {
		t.Fatalf("CanonicalName = %q, %v; se esperaba \"\", ErrNotProjectMember", got, err)
	}
	if ErrNotProjectMember.Error() != "el integrante no pertenece al proyecto" {
		t.Errorf("mensaje = %q", ErrNotProjectMember.Error())
	}
}

// HU-54 · CA-54.2 · Verifica que un integrante de otro proyecto no cuente como integrante del proyecto pedido.
func TestCanonicalName_IntegranteDeOtroProyecto(t *testing.T) {
	env := newMemberEnv(t)
	env.register(t, 2, "Cardozo Leandro", "PRODUCT_BUILDER")

	if _, err := env.members.CanonicalName(1, "Cardozo Leandro"); !errors.Is(err, ErrNotProjectMember) {
		t.Fatalf("CanonicalName devolvió %v, se esperaba ErrNotProjectMember", err)
	}
}

// HU-54 · CA-54.2 · Verifica que en un proyecto sin integrantes ningún nombre sea integrante.
func TestCanonicalName_ProyectoSinIntegrantes(t *testing.T) {
	env := newMemberEnv(t)

	if _, err := env.members.CanonicalName(1, "Sosa Ricardo"); !errors.Is(err, ErrNotProjectMember) {
		t.Fatalf("CanonicalName devolvió %v, se esperaba ErrNotProjectMember", err)
	}
}

// HU-54 · Error · Verifica que CanonicalName devuelva el error del repositorio si no puede listar los integrantes.
func TestCanonicalName_ErrorDelRepositorio(t *testing.T) {
	env := newMemberEnv(t)
	env.repo.failList = true

	if _, err := env.members.CanonicalName(1, "Sosa Ricardo"); !errors.Is(err, errStorage) {
		t.Fatalf("CanonicalName devolvió %v, se esperaba el error del repositorio", err)
	}
}
