package httpapi_test

import (
	"context"
	"fmt"
	"net/http"
	"net/http/httptest"
	"os"
	"os/exec"
	"strings"
	"testing"
	"time"

	"github.com/chromedp/chromedp"
)

// HU-53 · Navegar la interfaz por secciones y elegir de listas (Issue #211).
//
// Los tests de la interfaz de HU-53 hacen clics en un navegador real con
// chromedp (decisión técnica 23, D7 de design.md). chromedp se usa solo en
// archivos _test.go, así que no entra al binario.

// uiTimeout es el tiempo máximo de cada test en el navegador (R1).
const uiTimeout = 30 * time.Second

// browserPath devuelve el ejecutable de Chrome o Chromium: el de la variable
// CHROME_PATH o el primero que encuentre en el PATH. Devuelve "" si no hay.
func browserPath() string {
	if p := os.Getenv("CHROME_PATH"); p != "" {
		return p
	}
	for _, name := range []string{"google-chrome", "google-chrome-stable", "chromium", "chromium-browser", "chrome"} {
		if p, err := exec.LookPath(name); err == nil {
			return p
		}
	}
	return ""
}

// failer es lo que los helpers de la interfaz necesitan para fallar: un
// *testing.T en los tests de Go, o stepFailer en los pasos de godog.
type failer interface {
	Helper()
	Fatalf(format string, args ...any)
}

// stepFailer hace fallar un paso de godog: godog recupera el panic y lo
// informa como el error del paso.
type stepFailer struct{}

func (stepFailer) Helper() {}

func (stepFailer) Fatalf(format string, args ...any) { panic(fmt.Sprintf(format, args...)) }

// requireBrowser devuelve el ejecutable del navegador. Sin navegador, el test
// se saltea con un aviso, salvo en la CI (CI=true), donde es un error (D7).
func requireBrowser(t *testing.T) string {
	t.Helper()
	path := browserPath()
	if path == "" {
		msg := "no se encontró Chrome ni Chromium: instalalo o indicá su ruta en CHROME_PATH para correr los tests de la interfaz de HU-53"
		if os.Getenv("CI") != "" {
			t.Fatal(msg)
		}
		t.Skip(msg)
	}
	return path
}

// startBrowser sirve h en un servidor de prueba y abre el navegador path sin
// ventana, de width x 800 px. Devuelve el contexto de chromedp, la dirección
// del servidor y la función que cierra los dos.
func startBrowser(f failer, path string, h http.Handler, width int) (context.Context, string, func()) {
	f.Helper()
	srv := httptest.NewServer(h)
	opts := append(chromedp.DefaultExecAllocatorOptions[:],
		chromedp.ExecPath(path),
		chromedp.NoSandbox,
		chromedp.WindowSize(width, 800),
	)
	allocCtx, cancelAlloc := chromedp.NewExecAllocator(context.Background(), opts...)
	browserCtx, cancelBrowser := chromedp.NewContext(allocCtx)
	stop := func() {
		cancelBrowser()
		cancelAlloc()
		srv.Close()
	}
	// La primera ejecución abre el navegador y no puede tener tiempo máximo,
	// porque al vencer cerraría el navegador; el tiempo máximo va en un
	// contexto derivado.
	if err := chromedp.Do(browserCtx, chromedp.EmulateViewport(int64(width), 800)); err != nil {
		stop()
		f.Fatalf("no se pudo abrir el navegador %s: %v", path, err)
	}
	ctx, cancelTimeout := context.WithTimeout(browserCtx, uiTimeout)
	return ctx, srv.URL, func() {
		cancelTimeout()
		stop()
	}
}

// newBrowser abre la interfaz servida por h en un Chrome de 1280 x 800 px,
// que se cierra al terminar el test.
func newBrowser(t *testing.T, h http.Handler) (context.Context, string) {
	t.Helper()
	ctx, url, stop := startBrowser(t, requireBrowser(t), h, 1280)
	t.Cleanup(stop)
	return ctx, url
}

// browse ejecuta los pasos en el navegador, en orden, y falla si alguno da
// error.
func browse(t failer, ctx context.Context, steps ...chromedp.Action[chromedp.Void]) {
	t.Helper()
	if err := chromedp.Do(ctx, steps...); err != nil {
		t.Fatalf("navegador: %v", err)
	}
}

// value ejecuta una acción que devuelve un valor y falla si da error.
func value[T any](t failer, ctx context.Context, a chromedp.Action[T]) T {
	t.Helper()
	v, err := chromedp.Run(ctx, a)
	if err != nil {
		t.Fatalf("navegador: %v", err)
	}
	return v
}

// HU-53 · CA-53.7 · Verifica que la interfaz se abra en un navegador real desde el mismo servidor de la API, con su título.
func TestInterfazNavegador_AbreLaPaginaDesdeElServidor(t *testing.T) {
	ctx, url := newBrowser(t, newHandler())

	browse(t, ctx, chromedp.Navigate(url+"/"))
	title := value(t, ctx, chromedp.Title())

	if !strings.Contains(title, "Software Metrics") {
		t.Errorf("título = %q, se esperaba que contenga \"Software Metrics\"", title)
	}
}
