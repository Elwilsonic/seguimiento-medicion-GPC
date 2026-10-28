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

// --- Proyectos (HU-01, HU-02, HU-03 y HU-18) ---

function mostrarProyecto(p) {
  $("proyecto-actual").value = p.id;
  const form = $("form-modificar-proyecto");
  form.nombre.value = p.nombre;
  form.descripcion.value = p.descripcion;
  form.estado.value = p.estado;
  // HU-18: fecha_inicio es "" si no está registrada.
  $("form-fecha-inicio-proyecto").fecha_inicio.value = p.fecha_inicio;
  const inicio = p.fecha_inicio ? "inicio: " + p.fecha_inicio : "sin fecha de inicio";
  $("sin-proyecto").textContent = "Proyecto " + p.id + ": " + p.nombre + " (" + p.estado + "), " + inicio;
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

// HU-18: registra la fecha de inicio del proyecto actual. El campo de fecha
// del navegador ya envía AAAA-MM-DD.
$("form-fecha-inicio-proyecto").addEventListener("submit", (e) => {
  e.preventDefault();
  const form = e.target;
  accion("msg-proyectos", async () => {
    const p = await api("PATCH", "/proyectos/" + proyectoActual() + "/fecha-inicio", {
      fecha_inicio: form.fecha_inicio.value,
    });
    mostrarProyecto(p);
    mostrar("msg-proyectos", "Fecha de inicio del proyecto " + p.id + " registrada: " + p.fecha_inicio + ".", false);
  });
});

// --- Product Backlog (HU-04 a HU-09 y HU-20) ---

// escalaStoryPoints es la escala de backlog.StoryPointScale() (HU-20). Un test
// de internal/httpapi verifica que coincida con la del paquete Go.
const escalaStoryPoints = [1, 2, 3, 5, 8, 13, 21];

async function cargarBacklog() {
  const orden = $("orden").value;
  const items = await api("GET", "/proyectos/" + proyectoActual() + "/backlog" + (orden ? "?orden=" + orden : ""));
  const tbody = $("items");
  tbody.replaceChildren(...items.map(filaItem));
  if (items.length === 0) {
    const tr = document.createElement("tr");
    const td = celda("El backlog del proyecto no tiene ítems.");
    td.colSpan = 8;
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

// mostrarPoker muestra la última sesión de Planning Poker abierta (HU-21) con
// sus cartas.
function mostrarPoker(s) {
  mostrarDatosPoker(s);
  $("cartas-poker").textContent = "Cartas: " + s.cartas.join(", ");
  $("cartas-poker").className = "";
  $("votos-poker").hidden = true;
  $("votos-poker").replaceChildren();
  $("form-estimar-poker").hidden = true;
  const form = $("form-votar-poker");
  form.hidden = false;
  form.dataset.sessionId = s.id;
  const botones = $("botones-cartas-poker");
  botones.replaceChildren();
  for (const carta of s.cartas) {
    botones.append(boton("Votar " + carta, async () => {
      const integrante = form.integrante.value;
      await api("POST", "/poker-sessions/" + form.dataset.sessionId + "/votes", {
        integrante,
        puntos: carta,
      });
      mostrar("msg-poker", "Voto registrado en la sesión " + s.id + ".", false);
    }, "secundario"));
  }
}

function mostrarDatosPoker(s) {
  $("datos-poker").textContent =
    "Sesión " + s.id + " · ítem " + s.item_id + " · " + s.estado + " · ronda " + s.ronda;
  $("datos-poker").className = "";
}

// "Revelar votos" revela la sesión mostrada (HU-23): oculta el formulario de
// voto, porque la sesión ya no acepta votos, y lista todos los votos.
$("btn-revelar-poker").addEventListener("click", () =>
  accion("msg-poker", async () => {
    const form = $("form-votar-poker");
    const s = await api("POST", "/poker-sessions/" + form.dataset.sessionId + "/reveal");
    mostrarDatosPoker(s);
    form.hidden = true;
    const lista = $("votos-poker");
    lista.replaceChildren(...s.votos.map((v) => {
      const li = document.createElement("li");
      li.textContent = v.integrante + ": " + v.puntos;
      return li;
    }));
    lista.hidden = false;
    mostrarEstimarPoker(s);
    mostrar("msg-poker", "Votos revelados en la sesión " + s.id + ".", false);
  }));

// mostrarEstimarPoker muestra un botón por carta para registrar la estimación
// acordada de la sesión revelada (HU-25). Al registrarla, la sesión queda
// CERRADA y se recarga el backlog para ver los story points en la fila del
// ítem.
function mostrarEstimarPoker(s) {
  const form = $("form-estimar-poker");
  const botones = $("botones-estimar-poker");
  botones.replaceChildren(...s.cartas.map((carta) => {
    const b = document.createElement("button");
    b.type = "button";
    b.className = "secundario";
    b.textContent = String(carta);
    b.addEventListener("click", () => accion("msg-poker", async () => {
      const cerrada = await api("POST", "/poker-sessions/" + s.id + "/estimate", { puntos: carta });
      mostrarDatosPoker(cerrada);
      form.hidden = true;
      await cargarBacklog();
      mostrar("msg-poker", "Estimación registrada: " + carta + " story points en la sesión " + cerrada.id + ".", false);
    }));
    return b;
  }));
  form.hidden = false;
}

// filaItem arma la fila de un ítem con sus acciones: editar (HU-07), cambiar
// estado (HU-08), asignar story points (HU-20), abrir Planning Poker (HU-21),
// asignar al sprint actual (HU-11) y completar (HU-12).
function filaItem(it) {
  const tr = document.createElement("tr");
  tr.append(
    celda(it.id),
    celda(it.titulo),
    celda(it.prioridad),
    celda(it.estado),
    // 0 significa sin estimar (TEC-03).
    celda(it.story_points === 0 ? "—" : it.story_points),
    celda(it.horas_estimadas),
    celda(it.sprint_id === 0 ? "—" : it.sprint_id),
  );

  const estado = document.createElement("select");
  estado.setAttribute("aria-label", "Nuevo estado del ítem " + it.id);
  for (const s of ["PENDIENTE", "EN_CURSO", "TERMINADO"]) {
    estado.append(new Option(s, s, false, s === it.estado));
  }

  const puntos = document.createElement("select");
  puntos.setAttribute("aria-label", "Story points del ítem " + it.id);
  for (const p of escalaStoryPoints) {
    puntos.append(new Option(p, p, false, p === it.story_points));
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
    puntos,
    boton("Asignar SP", async () => {
      const valor = Number(puntos.value);
      await api("PATCH", "/items/" + it.id + "/story-points", { story_points: valor });
      await cargarBacklog();
      mostrar("msg-backlog", "Ítem " + it.id + " con " + valor + " story points.", false);
    }, "secundario"),
    boton("Abrir Poker", async () => {
      const sesion = await api("POST", "/items/" + it.id + "/poker-sessions");
      mostrarPoker(sesion);
      mostrar("msg-backlog", "Sesión " + sesion.id + " abierta para el ítem " + sesion.item_id + ".", false);
    }),
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

// HU-26 · Registrar esfuerzo (CA-26.11). Usa el sprint actual y el ítem
// indicado; al terminar recarga el backlog y muestra el ID del esfuerzo.
const formEsfuerzo = $("form-registrar-esfuerzo");
formEsfuerzo.setAttribute("aria-label", "Registrar esfuerzo");
formEsfuerzo.addEventListener("submit", (e) => {
  e.preventDefault();
  const form = e.target;
  accion("msg-sprints", async () => {
    const sprint = sprintActual();
    const item = form.item_id.value.trim();
    const esfuerzo = await api("POST", "/sprints/" + sprint + "/items/" + item + "/esfuerzos", {
      integrante: form.integrante.value,
      fecha: form.fecha.value,
      actividad: form.actividad.value,
      horas: Number(form.horas.value),
    });
    form.reset();
    await cargarBacklog();
    mostrar("msg-sprints", "Esfuerzo " + esfuerzo.id + " registrado para el ítem " + esfuerzo.item_id + ".", false);
  });
});

// HU-27 · Consultar esfuerzos (CA-27.10). Consulta por sprint o por historia,
// llena la tabla y muestra el total con un decimal.
const formConsultaEsfuerzos = $("form-consultar-esfuerzos");
formConsultaEsfuerzos.setAttribute("aria-label", "Consultar esfuerzos");
formConsultaEsfuerzos.addEventListener("submit", (e) => {
  e.preventDefault();
  const form = e.target;
  accion("msg-esfuerzos", async () => {
    const referencia = form.referencia.value.trim();
    const ruta =
      form.tipo.value === "sprint"
        ? "/sprints/" + referencia + "/esfuerzos"
        : "/items/" + referencia + "/esfuerzos";
    const resumen = await api("GET", ruta);
    const tbody = $("esfuerzos");
    tbody.replaceChildren(...resumen.esfuerzos.map(filaEsfuerzo));
    if (resumen.esfuerzos.length === 0) {
      const tr = document.createElement("tr");
      const td = celda("No hay esfuerzos registrados.");
      td.colSpan = 7;
      td.className = "vacio";
      tr.append(td);
      tbody.append(tr);
    }
    $("total-esfuerzos").textContent = "Total: " + resumen.total_horas.toFixed(1) + " horas";
    mostrar("msg-esfuerzos", "Se listaron " + resumen.esfuerzos.length + " esfuerzo(s).", false);
  });
});

// filaEsfuerzo arma una fila de la tabla de consulta de esfuerzos.
function filaEsfuerzo(e) {
  const tr = document.createElement("tr");
  tr.append(
    celda(e.id),
    celda(e.fecha),
    celda(e.integrante),
    celda(e.actividad),
    celda(e.item_id),
    celda(e.sprint_id),
    celda(e.horas),
  );
  return tr;
}

// HU-32 · Métricas (CA-32.9). Muestra los story points planificados y
// completados, los ítems sin estimar y la velocidad del proyecto actual, y si
// hay sprint actual, los de ese sprint.
// HU-34 · Porcentaje de historias completadas (CA-34.6), con un decimal y las
// cantidades, o "sin datos" si no hay ítems.
function historiasCompletadas(m) {
  if (m.porcentaje_completadas === null) {
    return "Historias completadas: sin datos.";
  }
  return "Historias completadas: " + m.porcentaje_completadas.toFixed(1) + " % (" +
    m.items_terminados + " de " + m.items_totales + ").";
}

function numero1(valor) {
  const texto = valor.toFixed(1);
  return texto === "-0.0" ? "0.0" : texto;
}

function horasYDesviacion(m) {
  let texto = "Horas: " + numero1(m.horas_estimadas) + " estimadas, " +
    numero1(m.horas_reales) + " reales, " + m.items_sin_horas_estimadas +
    " ítem(s) sin horas estimadas. Desviación: ";
  if (m.desviacion_horas === null || m.desviacion_porcentual === null) {
    return texto + "sin datos.";
  }
  return texto + numero1(m.desviacion_horas) + " h (" +
    numero1(m.desviacion_porcentual) + " %).";
}

$("ver-metricas").addEventListener("click", () =>
  accion("msg-metricas", async () => {
    const p = await api("GET", "/proyectos/" + proyectoActual() + "/metricas");
    const velocidad = p.velocidad === null ? "sin datos" : p.velocidad.toFixed(1) + " SP por sprint";
    let texto =
      "Proyecto " + p.proyecto_id + ": " + p.sp_planificados + " SP planificados, " +
      p.sp_completados + " SP completados, " + p.items_sin_estimar + " ítem(s) sin estimar. " +
      "Velocidad: " + velocidad + " (" + p.sprints_cerrados + " sprint(s) cerrado(s)). " +
      historiasCompletadas(p) + " " + horasYDesviacion(p);
    if ($("sprint-actual").value.trim() !== "") {
      const s = await api("GET", "/sprints/" + sprintActual() + "/metricas");
      texto +=
        " Sprint " + s.sprint_id + ": " + s.sp_planificados + " SP planificados, " +
        s.sp_completados + " SP completados, " + s.items_sin_estimar + " ítem(s) sin estimar. " +
        historiasCompletadas(s) + " " + horasYDesviacion(s);
    }
    $("datos-metricas").textContent = texto;
    $("datos-metricas").className = "";
  }),
);

// HU-29 · Registrar defecto (CA-29.11). Usa el ítem y el sprint de detección
// indicados en el formulario; al terminar muestra el ID del defecto.
const severidadesDefecto = ["CRITICA", "ALTA", "MEDIA", "BAJA"];
const formDefecto = $("form-registrar-defecto");
formDefecto.setAttribute("aria-label", "Registrar defecto");
for (const severidad of severidadesDefecto) {
  formDefecto.severidad.append(new Option(severidad, severidad));
}
formDefecto.addEventListener("submit", (e) => {
  e.preventDefault();
  const form = e.target;
  accion("msg-defectos", async () => {
    const item = form.item_id.value.trim();
    const defecto = await api("POST", "/items/" + item + "/defectos", {
      descripcion: form.descripcion.value,
      severidad: form.severidad.value,
      sprint_deteccion_id: Number(form.sprint_deteccion_id.value),
    });
    form.reset();
    mostrar("msg-defectos", "Defecto " + defecto.id + " registrado para el ítem " + defecto.item_id + ".", false);
  });
});

// HU-30 · Cambiar el estado de un defecto (CA-30.9). Envía el sprint de
// resolución solo si el campo tiene valor; la API decide si es obligatorio.
const estadosDefecto = ["ABIERTO", "EN_CORRECCION", "RESUELTO"];
const formEstadoDefecto = $("form-cambiar-estado-defecto");
formEstadoDefecto.setAttribute("aria-label", "Cambiar estado de defecto");
for (const estado of estadosDefecto) {
  formEstadoDefecto.estado.append(new Option(estado, estado));
}
formEstadoDefecto.addEventListener("submit", (e) => {
  e.preventDefault();
  const form = e.target;
  accion("msg-defectos", async () => {
    const defecto = form.defecto_id.value.trim();
    const cuerpo = { estado: form.estado.value };
    const sprint = form.sprint_resolucion_id.value.trim();
    if (sprint !== "") {
      cuerpo.sprint_resolucion_id = Number(sprint);
    }
    const cambiado = await api("PATCH", "/defectos/" + defecto + "/estado", cuerpo);
    form.reset();
    let texto = "Defecto " + cambiado.id + " ahora está en " + cambiado.estado;
    if (cambiado.estado === "RESUELTO") {
      texto += "," + " resuelto en el sprint " + cambiado.sprint_resolucion_id;
    }
    mostrar("msg-defectos", texto + ".", false);
  });
});

// HU-31 · Consultar defectos (CA-31.9). Consulta el proyecto o el sprint
// actual; el filtro de estado se envía solo si se eligió uno.
const formConsultaDefectos = $("form-consultar-defectos");
formConsultaDefectos.setAttribute("aria-label", "Consultar defectos");
for (const estado of estadosDefecto) {
  formConsultaDefectos.estado.append(new Option(estado, estado));
}
formConsultaDefectos.addEventListener("submit", (e) => {
  e.preventDefault();
  const form = e.target;
  accion("msg-consultar-defectos", async () => {
    let ruta =
      form.tipo.value === "sprint"
        ? "/sprints/" + sprintActual() + "/defectos"
        : "/proyectos/" + proyectoActual() + "/defectos";
    if (form.estado.value !== "") {
      ruta += "?estado=" + encodeURIComponent(form.estado.value);
    }
    const resultado = await api("GET", ruta);
    const tbody = $("lista-defectos");
    tbody.replaceChildren(...resultado.defectos.map(filaDefecto));
    if (resultado.defectos.length === 0) {
      const tr = document.createElement("tr");
      const td = celda("Sin defectos.");
      td.colSpan = 7;
      td.className = "vacio";
      tr.append(td);
      tbody.append(tr);
    }
    $("cantidad-defectos").textContent = "Cantidad: " + resultado.cantidad + " defecto(s)";
    mostrar("msg-consultar-defectos", "Se listaron " + resultado.cantidad + " defecto(s).", false);
  });
});

// filaDefecto arma una fila de la tabla de consulta de defectos. Un sprint de
// resolución 0 (sin resolver) se muestra como "—".
function filaDefecto(d) {
  const tr = document.createElement("tr");
  tr.append(
    celda(d.id),
    celda(d.descripcion),
    celda(d.severidad),
    celda(d.estado),
    celda(d.item_id),
    celda(d.sprint_deteccion_id),
    celda(d.sprint_resolucion_id === 0 ? "—" : d.sprint_resolucion_id),
  );
  return tr;
}
