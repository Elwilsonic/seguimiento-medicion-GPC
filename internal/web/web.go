// Package web sirve la interfaz web de la aplicación (TEC-04): HTML,
// JavaScript y CSS sin frameworks, embebidos en el binario con embed
// (CA-TEC04.4). La interfaz usa la API de internal/httpapi bajo /api/….
package web

import (
	"embed"
	"net/http"
)

//go:embed index.html app.js styles.css
var files embed.FS

// Handler sirve los archivos de la interfaz desde la raíz: / devuelve
// index.html (CA-TEC04.8).
func Handler() http.Handler {
	return http.FileServerFS(files)
}
