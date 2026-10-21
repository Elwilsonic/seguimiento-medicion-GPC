// Interfaz web de Software Metrics & Estimation (TEC-04).
// JavaScript sin frameworks: usa la API bajo /api/… con fetch y muestra los
// errores de la API ({"error": "..."}) junto al formulario que los produjo
// (CA-TEC04.2). Los datos del usuario se muestran siempre con textContent.
"use strict";

const $ = (id) => document.getElementById(id);

// api hace un pedido a la API y devuelve el JSON de la respuesta. Si la API
// responde un error, lanza un Error con su mensaje.
async function api(method, path, body) {
  const options = { method, headers: {} };
  if (body !== undefined) {
    options.headers["Content-Type"] = "application/json";
    options.body = JSON.stringify(body);
  }
  let response;
  try {
    response = await fetch("/api" + path, options);
  } catch {
    throw new Error("No se pudo conectar con el servidor. ¿Está en ejecución?");
  }
  let data = null;
  try {
    data = await response.json();
  } catch {
    // Una respuesta sin JSON (por ejemplo, una ruta inexistente) se informa por su código.
  }
  if (!response.ok) {
    throw new Error((data && data.error) || "Error " + response.status + " del servidor.");
  }
  return data;
}

// mostrar muestra un mensaje de éxito o de error en el elemento id.
function mostrar(id, texto, esError) {
  const el = $(id);
  el.textContent = texto;
  el.className = "mensaje " + (esError ? "error" : "ok");
  el.hidden = false;
}

// accion ejecuta fn y muestra su error, si lo hay, en el elemento msgId.
async function accion(msgId, fn) {
  $(msgId).hidden = true;
  try {
    await fn();
  } catch (err) {
    mostrar(msgId, err.message, true);
  }
}

// idActual lee el ID del proyecto o del sprint actual; si falta, lanza un
// error que explica qué hacer.
function idActual(inputId, falta) {
  const valor = $(inputId).value.trim();
  if (valor === "") {
    throw new Error(falta);
  }
  return valor;
}

const proyectoActual = () =>
  idActual("proyecto-actual", "Elegí un proyecto: creá uno o escribí su ID en «Proyecto actual».");
const sprintActual = () =>
  idActual("sprint-actual", "Elegí un sprint: creá uno o escribí su ID en «Sprint actual».");

// criterios convierte el texto de un textarea en una lista, una línea por
// criterio, sin líneas vacías.
function criterios(texto) {
  return texto.split("\n").map((c) => c.trim()).filter((c) => c !== "");
}

// --- Proyectos (HU-01, HU-02 y HU-03) ---

function mostrarProyecto(p) {
  $("proyecto-actual").value = p.id;
  const form = $("form-modificar-proyecto");
  form.nombre.value = p.nombre;
  form.descripcion.value = p.descripcion;
  form.estado.value = p.estado;
  $("sin-proyecto").textContent = "Proyecto " + p.id + ": " + p.nombre + " (" + p.estado + ")";
}

async function cargarProyecto() {
  const p = await api("GET", "/proyectos/" + proyectoActual());
  mostrarProyecto(p);
  await cargarBacklog();
}

$("form-contexto").addEventListener("submit", (e) => {
  e.preventDefault();
  accion("msg-contexto", cargarProyecto);
});

$("form-crear-proyecto").addEventListener("submit", (e) => {
  e.preventDefault();
  const form = e.target;
  accion("msg-proyectos", async () => {
    const p = await api("POST", "/proyectos", {
      nombre: form.nombre.value,
      descripcion: form.descripcion.value,
    });
    form.reset();
    mostrarProyecto(p);
    mostrar("msg-proyectos", "Proyecto " + p.id + " creado.", false);
    await cargarBacklog();
  });
});

$("form-modificar-proyecto").addEventListener("submit", (e) => {
  e.preventDefault();
  const form = e.target;
  accion("msg-proyectos", async () => {
    const p = await api("PUT", "/proyectos/" + proyectoActual(), {
      nombre: form.nombre.value,
      descripcion: form.descripcion.value,
      estado: form.estado.value,
    });
    mostrarProyecto(p);
    mostrar("msg-proyectos", "Proyecto " + p.id + " guardado.", false);
  });
});

// --- Product Backlog (HU-04 a HU-09) ---

async function cargarBacklog() {
  const orden = $("orden").value;
  const items = await api("GET", "/proyectos/" + proyectoActual() + "/backlog" + (orden ? "?orden=" + orden : ""));
  const tbody = $("items");
  tbody.replaceChildren(...items.map(filaItem));
  if (items.length === 0) {
    const tr = document.createElement("tr");
    const td = celda("El backlog del proyecto no tiene ítems.");
    td.colSpan = 7;
    td.className = "vacio";
    tr.append(td);
    tbody.append(tr);
  }
}

function celda(texto) {
  const td = document.createElement("td");
  td.textContent = texto;
  return td;
}

function boton(texto, fn, clase) {
  const b = document.createElement("button");
  b.type = "button";
  b.textContent = texto;
  if (clase) {
    b.className = clase;
  }
  b.addEventListener("click", () => accion("msg-backlog", fn));
  return b;
}

// filaItem arma la fila de un ítem con sus acciones: editar (HU-07), cambiar
// estado (HU-08), asignar al sprint actual (HU-11) y completar (HU-12).
function filaItem(it) {
  const tr = document.createElement("tr");
  tr.append(
    celda(it.id),
    celda(it.titulo),
    celda(it.prioridad),
    celda(it.estado),
    celda(it.horas_estimadas),
    celda(it.sprint_id === 0 ? "—" : it.sprint_id),
  );

  const estado = document.createElement("select");
  estado.setAttribute("aria-label", "Nuevo estado del ítem " + it.id);
  for (const s of ["PENDIENTE", "EN_CURSO", "TERMINADO"]) {
    estado.append(new Option(s, s, false, s === it.estado));
  }

  const acciones = document.createElement("td");
  acciones.className = "acciones";
  acciones.append(
    boton("Editar", () => editarItem(it.id), "secundario"),
    estado,
    boton("Cambiar estado", async () => {
      await api("PATCH", "/items/" + it.id + "/estado", { estado: estado.value });
      await cargarBacklog();
      mostrar("msg-backlog", "Ítem " + it.id + " en estado " + estado.value + ".", false);
    }, "secundario"),
    boton("Asignar al sprint", async () => {
      const sprint = sprintActual();
      await api("POST", "/sprints/" + sprint + "/items/" + it.id);
      await cargarBacklog();
      mostrar("msg-backlog", "Ítem " + it.id + " asignado al sprint " + sprint + ".", false);
    }, "secundario"),
    boton("Completar", async () => {
      const sprint = sprintActual();
      await api("POST", "/sprints/" + sprint + "/items/" + it.id + "/completar");
      await cargarBacklog();
      mostrar("msg-backlog", "Ítem " + it.id + " completado en el sprint " + sprint + ".", false);
    }),
  );
  tr.append(acciones);
  return tr;
}

$("form-crear-item").addEventListener("submit", (e) => {
  e.preventDefault();
  const form = e.target;
  accion("msg-backlog", async () => {
    const it = await api("POST", "/proyectos/" + proyectoActual() + "/backlog", {
      titulo: form.titulo.value,
      descripcion: form.descripcion.value,
      prioridad: form.prioridad.value,
      criterios: criterios(form.criterios.value),
    });
    form.reset();
    await cargarBacklog();
    mostrar("msg-backlog", "Ítem " + it.id + " creado.", false);
  });
});

// editarItem consulta el ítem (HU-05) y lo carga en el formulario de edición.
async function editarItem(id) {
  const it = await api("GET", "/items/" + id);
  const form = $("form-editar-item");
  form.dataset.id = it.id;
  $("editar-item-id").textContent = it.id;
  form.titulo.value = it.titulo;
  form.descripcion.value = it.descripcion;
  form.prioridad.value = it.prioridad;
  form.criterios.value = it.criterios.join("\n");
  form.horas_estimadas.value = it.horas_estimadas;
  form.hidden = false;
  form.titulo.focus();
}

$("form-editar-item").addEventListener("submit", (e) => {
  e.preventDefault();
  const form = e.target;
  accion("msg-backlog", async () => {
    const horas = form.horas_estimadas.value.trim();
    const it = await api("PUT", "/items/" + form.dataset.id, {
      titulo: form.titulo.value,
      descripcion: form.descripcion.value,
      prioridad: form.prioridad.value,
      criterios: criterios(form.criterios.value),
      horas_estimadas: horas === "" ? 0 : Number(horas),
    });
    form.hidden = true;
    await cargarBacklog();
    mostrar("msg-backlog", "Ítem " + it.id + " guardado.", false);
  });
});

$("cancelar-edicion").addEventListener("click", () => {
  $("form-editar-item").hidden = true;
});

$("orden").addEventListener("change", () => accion("msg-backlog", cargarBacklog));
$("actualizar-backlog").addEventListener("click", () => accion("msg-backlog", cargarBacklog));

// --- Sprints (HU-10 a HU-13) ---

function mostrarSprint(s) {
  $("sprint-actual").value = s.id;
  $("datos-sprint").textContent =
    "Sprint " + s.id + ": " + s.nombre + " (" + s.estado + "), del " + s.fecha_inicio +
    " al " + s.fecha_fin + ". Goal: " + s.objetivo;
  $("datos-sprint").className = "";
}

$("form-crear-sprint").addEventListener("submit", (e) => {
  e.preventDefault();
  const form = e.target;
  accion("msg-sprints", async () => {
    const s = await api("POST", "/proyectos/" + proyectoActual() + "/sprints", {
      nombre: form.nombre.value,
      fecha_inicio: form.fecha_inicio.value,
      fecha_fin: form.fecha_fin.value,
      objetivo: form.objetivo.value,
    });
    form.reset();
    mostrarSprint(s);
    mostrar("msg-sprints", "Sprint " + s.id + " creado.", false);
  });
});

$("ver-completados").addEventListener("click", () =>
  accion("msg-sprints", async () => {
    const sprint = sprintActual();
    const items = await api("GET", "/sprints/" + sprint + "/completados");
    const lista = $("completados");
    lista.replaceChildren(
      ...items.map((it) => {
        const li = document.createElement("li");
        li.textContent = it.id + " · " + it.titulo;
        return li;
      }),
    );
    mostrar("msg-sprints", items.length + " ítem(s) completado(s) en el sprint " + sprint + ".", false);
  }),
);

$("cerrar-sprint").addEventListener("click", () =>
  accion("msg-sprints", async () => {
    const s = await api("POST", "/sprints/" + sprintActual() + "/cerrar");
    mostrarSprint(s);
    mostrar("msg-sprints", "Sprint " + s.id + " cerrado.", false);
  }),
);
