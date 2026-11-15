package httpapi

import (
	"errors"
	"strings"
	"testing"
	"uuid"
)

// HU-51 · Identificadores UUID, parte de proyectos (Issue #193). Test interno
// del paquete, porque parseProjectID no se exporta.

const canonicalProjectID = "3f2b8c1d-5e6a-4b7c-8d9e-0f1a2b3c4d5e"

// HU-51 · CA-51.4 · Verifica que parseProjectID acepte un UUID en forma canónica y devuelva ese mismo UUID.
func TestParseProjectID_AceptaLaFormaCanonica(t *testing.T) {
	got, err := parseProjectID(canonicalProjectID)

	if err != nil {
		t.Fatalf("parseProjectID(%q) devolvió un error inesperado: %v", canonicalProjectID, err)
	}
	if want := uuid.MustParse(canonicalProjectID); got != want {
		t.Errorf("parseProjectID(%q) = %s, se esperaba %s", canonicalProjectID, got, want)
	}
}

// HU-51 · CA-51.3 · Verifica que parseProjectID acepte el UUID en mayúsculas y que su forma de salida quede en minúsculas.
func TestParseProjectID_AceptaMayusculas(t *testing.T) {
	upper := strings.ToUpper(canonicalProjectID)

	got, err := parseProjectID(upper)

	if err != nil {
		t.Fatalf("parseProjectID(%q) devolvió un error inesperado: %v", upper, err)
	}
	if got.String() != canonicalProjectID {
		t.Errorf("parseProjectID(%q).String() = %q, se esperaba %q", upper, got.String(), canonicalProjectID)
	}
}

// HU-51 · CA-51.4 · Verifica que parseProjectID rechace con errInvalidProjectID todo lo que no es un UUID en forma canónica: números, llaves, urn:uuid:, sin guiones, vacío y dígitos no hexadecimales.
func TestParseProjectID_RechazaLoQueNoEsLaFormaCanonica(t *testing.T) {
	cases := map[string]string{
		"número":                      "1",
		"cero":                        "0",
		"negativo":                    "-1",
		"texto":                       "abc",
		"vacío":                       "",
		"entre llaves":                "{" + canonicalProjectID + "}",
		"con urn:uuid:":               "urn:uuid:" + canonicalProjectID,
		"sin guiones":                 strings.ReplaceAll(canonicalProjectID, "-", ""),
		"dígito no hexadecimal":       "3f2b8c1d-5e6a-4b7c-8d9e-0f1a2b3c4d5g",
		"un carácter de menos":        canonicalProjectID[:35],
		"un carácter de más":          canonicalProjectID + "0",
		"guiones en otras posiciones": "3f2b8c1d5-e6a-4b7c-8d9e-0f1a2b3c4d5e",
		"espacios alrededor":          " " + canonicalProjectID + " ",
	}
	for name, s := range cases {
		t.Run(name, func(t *testing.T) {
			_, err := parseProjectID(s)

			if !errors.Is(err, errInvalidProjectID) {
				t.Errorf("parseProjectID(%q): err = %v, se esperaba %v", s, err, errInvalidProjectID)
			}
		})
	}
}

// HU-51 · CA-51.4 · Verifica que errInvalidProjectID responda 400 con el mensaje "el id del proyecto debe ser un UUID".
func TestStatusFor_IDDeProyectoInvalidoEs400(t *testing.T) {
	if got := statusFor(errInvalidProjectID); got != 400 {
		t.Errorf("statusFor(errInvalidProjectID) = %d, se esperaba 400", got)
	}
	if got, want := errInvalidProjectID.Error(), "el id del proyecto debe ser un UUID"; got != want {
		t.Errorf("errInvalidProjectID = %q, se esperaba %q", got, want)
	}
}
