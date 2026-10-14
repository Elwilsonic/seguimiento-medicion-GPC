package main

import "testing"

// CA-TEC01.1 · El puerto se configura con la variable de entorno PORT.
func TestAddress_UsaLaVariablePORT(t *testing.T) {
	getenv := func(key string) string {
		if key == "PORT" {
			return "9090"
		}
		return ""
	}

	if got := address(getenv); got != ":9090" {
		t.Errorf("address = %q, se esperaba %q", got, ":9090")
	}
}

func TestAddress_SinPORTUsa8080(t *testing.T) {
	getenv := func(string) string { return "" }

	if got := address(getenv); got != ":8080" {
		t.Errorf("address = %q, se esperaba %q", got, ":8080")
	}
}
