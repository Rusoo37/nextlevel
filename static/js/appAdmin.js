// ==========================================
// 1. AGENDA PRINCIPAL
// ==========================================
const fechaFiltro = document.getElementById('fechaFiltro');
const tablaTurnos = document.getElementById('tablaTurnos');

document.addEventListener('DOMContentLoaded', () => {
    const hoy = new Date().toISOString().split('T')[0];
    fechaFiltro.value = hoy;
    cargarTurnos(hoy);
});

fechaFiltro.addEventListener('change', (e) => cargarTurnos(e.target.value));

async function cargarTurnos(fecha) {
    tablaTurnos.innerHTML = '<tr><td colspan="5" style="text-align:center;">Buscando...</td></tr>';
    
    try {
        const respuesta = await fetch(`/api/admin/turnos?fecha=${fecha}`);
        const turnos = await respuesta.json();
        tablaTurnos.innerHTML = '';

        if (turnos.length === 0) {
            tablaTurnos.innerHTML = '<tr><td colspan="5" style="text-align:center;">No hay turnos.</td></tr>';
            return;
        }

        turnos.forEach(turno => {
            const tr = document.createElement('tr');
            
            const esPrivado = turno.hora.endsWith(":30");
            if (esPrivado) {
                tr.style.backgroundColor = "#eff6ff"; 
            }

            let badgeClass = 'bg-orange';
            let icon = '⏳';
            
            
            if (turno.estado === 'CONFIRMADO') {
                badgeClass = 'bg-green';
                icon = '✅';
            } else if (turno.estado === 'FIJO') {
                badgeClass = 'bg-gray'; 
                icon = '🔄';
                tr.style.backgroundColor = "#f3f4f6";
            } else if (turno.estado === 'MANUAL') {
                badgeClass = 'bg-gray';
                icon = '🔒';
            }

            let botonCancelar = '';
            if (turno.estado === 'CONFIRMADO' || turno.estado === 'MANUAL') {
                botonCancelar = `<button onclick="cancelarTurno(${turno.id})" style="background: #ef4444; color: white; border: none; padding: 5px 10px; border-radius: 5px; cursor: pointer; font-size: 0.8rem;">Cancelar</button>`;
            }

            tr.innerHTML = `
                <td><strong>${turno.hora}</strong></td>
                <td>${turno.nombre_cliente}</td>
                <td>${turno.telefono !== '-' ? `<a href="https://wa.me/549${turno.telefono}" target="_blank" style="color: black;">${turno.telefono}</a>` : '-'}</td>
                <td><span class="badge ${badgeClass}">${icon} ${turno.estado}</span></td>
                <td>${botonCancelar}</td> 
            `;
            tablaTurnos.appendChild(tr);
        });
    } catch (error) {
        tablaTurnos.innerHTML = '<tr><td colspan="4" style="text-align:center; color:red;">Error de conexión</td></tr>';
    }
}


// CANCELAR TURNO
async function cancelarTurno(idTurno) {
    if (!confirm("¿Seguro que querés cancelar este turno? El horario se va a liberar en la web automáticamente.")) {
        return;
    }

    try {
        const respuesta = await fetch(`/api/admin/turnos/cancelar?id=${idTurno}`, {
            method: 'POST'
        });

        if (respuesta.ok) {
            // Volvemos a cargar la tabla del mismo día que Ramón está mirando
            cargarTurnos(document.getElementById('fechaFiltro').value);
        } else {
            alert("Hubo un problema al cancelar el turno.");
        }
    } catch (error) {
        console.error("Error al cancelar:", error);
        alert("Error de conexión.");
    }
}

// ==========================================
// CERRAR MODALES HACIENDO CLIC AFUERA
// ==========================================
window.addEventListener('click', (e) => {
    if (e.target === document.getElementById('modalConfig')) {
        document.getElementById('modalConfig').classList.add('hidden');
    }
    if (e.target === document.getElementById('modalBloqueos')) {
        document.getElementById('modalBloqueos').classList.add('hidden');
    }
});


// ==========================================
// 2. MODAL CONFIGURACIÓN (Precios y Fijos)
// ==========================================
const modalConfig = document.getElementById('modalConfig');
const btnAbrirConfig = document.getElementById('btnAbrirConfig');
const btnCerrarConfig = document.getElementById('btnCerrarConfig');

// --- Lógica de Apertura/Cierre ---
btnAbrirConfig.addEventListener('click', () => {
    modalConfig.classList.remove('hidden');
    cargarPrecios(); 
    cargarTurnosFijos(); 
});

btnCerrarConfig.addEventListener('click', () => {
    modalConfig.classList.add('hidden');
});

// --- Lógica de Precios ---
const inputPrecio = document.getElementById('inputPrecio');
const inputSena = document.getElementById('inputSena');
const btnGuardarPrecios = document.getElementById('btnGuardarPrecios');
const msgPrecios = document.getElementById('msgPrecios');

async function cargarPrecios() {
    try {
        const respuesta = await fetch('/api/admin/config/precios');
        const config = await respuesta.json();
        inputPrecio.value = config.precio_turno;
        inputSena.value = config.porcentaje_sena;
    } catch (error) {
        console.error("Error al cargar los precios:", error);
    }
}

btnGuardarPrecios.addEventListener('click', async () => {
    const precio = parseFloat(inputPrecio.value);
    const sena = parseInt(inputSena.value);

    if (isNaN(precio) || isNaN(sena) || precio <= 0 || sena < 0 || sena > 100) {
        alert("Por favor, ingresá valores válidos (La seña debe ser entre 0 y 100%).");
        return;
    }

    try {
        const respuesta = await fetch('/api/admin/config/precios', {
            method: 'POST',
            headers: { 'Content-Type': 'application/json' },
            body: JSON.stringify({ precio_turno: precio, porcentaje_sena: sena })
        });
        if (!respuesta.ok) throw new Error("Error al guardar los precios");

        msgPrecios.style.display = 'block';
        setTimeout(() => { msgPrecios.style.display = 'none'; }, 3000);
    } catch (error) {
        alert(error.message);
    }
});

// --- Lógica de Turnos Fijos ---
const tablaFijos = document.getElementById('tablaFijos');
const btnGuardarFijo = document.getElementById('btnGuardarFijo');
const diasSemana = ["Domingo", "Lunes", "Martes", "Miércoles", "Jueves", "Viernes", "Sábado"];

async function cargarTurnosFijos() {
    tablaFijos.innerHTML = '<tr><td colspan="4" style="text-align:center;">Cargando...</td></tr>';
    try {
        const respuesta = await fetch('/api/admin/config/fijos');
        const fijos = await respuesta.json();
        tablaFijos.innerHTML = '';
        
        if (fijos.length === 0) {
            tablaFijos.innerHTML = '<tr><td colspan="4" style="text-align:center; color: #6b7280;">No hay turnos fijos configurados.</td></tr>';
            return;
        }

        fijos.forEach(fijo => {
            const tr = document.createElement('tr');
            tr.innerHTML = `
                <td><strong>${diasSemana[fijo.dia_semana]}</strong></td>
                <td>${fijo.hora}</td>
                <td>${fijo.nombre_cliente}</td>
                <td>
                    <button onclick="eliminarTurnoFijo(${fijo.dia_semana}, '${fijo.hora}')" 
                            style="background: #ef4444; color: white; border: none; padding: 5px 10px; border-radius: 5px; cursor: pointer;">
                        Borrar
                    </button>
                </td>
            `;
            tablaFijos.appendChild(tr);
        });
    } catch (error) {
        tablaFijos.innerHTML = '<tr><td colspan="4" style="text-align:center; color: red;">Error al cargar.</td></tr>';
    }
}

btnGuardarFijo.addEventListener('click', async () => {
    const dia = document.getElementById('fijoDia').value;
    const hora = document.getElementById('fijoHora').value;
    const nombre = document.getElementById('fijoNombre').value.trim();

    if (!hora) {
        alert("Por favor, elegí una hora.");
        return;
    }

    if (!hora.endsWith(':00') && !hora.endsWith(':30')) {
        alert("🚨 Horario inválido para Turno Fijo. Ramón, acordate de agendar solo en horas en punto (:00) o medias horas (:30).");
        return;
    }

    try {
        const respuesta = await fetch('/api/admin/config/fijos', {
            method: 'POST',
            headers: { 'Content-Type': 'application/json' },
            body: JSON.stringify({ accion: "guardar", dia_semana: parseInt(dia), hora: hora, nombre_cliente: nombre })
        });
        if (!respuesta.ok) throw new Error("Error o turno ya existente");

        document.getElementById('fijoHora').value = '';
        document.getElementById('fijoNombre').value = '';
        cargarTurnosFijos();
        cargarTurnos(fechaFiltro.value); // Refresca la agenda principal por las dudas
    } catch (error) {
        alert(error.message);
    }
});

async function eliminarTurnoFijo(dia, hora) {
    if (!confirm(`¿Seguro que querés borrar el turno fijo del ${diasSemana[dia]} a las ${hora}?`)) return;
    try {
        const respuesta = await fetch('/api/admin/config/fijos', {
            method: 'POST',
            headers: { 'Content-Type': 'application/json' },
            body: JSON.stringify({ accion: "eliminar", dia_semana: parseInt(dia), hora: hora })
        });
        if (!respuesta.ok) throw new Error("Error al eliminar");
        
        cargarTurnosFijos();
        cargarTurnos(fechaFiltro.value);
    } catch (error) {
        alert(error.message);
    }
}


// ==========================================
// 3. MODAL BLOQUEOS (Manuales y Vacaciones)
// ==========================================
const modalBloqueos = document.getElementById('modalBloqueos');
const btnAbrirBloqueos = document.getElementById('btnAbrirBloqueos');
const btnCerrarBloqueos = document.getElementById('btnCerrarBloqueos');

// --- Lógica de Apertura/Cierre ---
btnAbrirBloqueos.addEventListener('click', () => {
    modalBloqueos.classList.remove('hidden');
    // Pre-llenamos el input manual con el día que Ramón está mirando
    document.getElementById('fechaBloqueoManual').value = fechaFiltro.value; 
    cargarExcepciones();
});

btnCerrarBloqueos.addEventListener('click', () => {
    modalBloqueos.classList.add('hidden');
});

// --- Lógica de Bloqueo Específico (Hora puntual) ---
const btnEjecutarBloqueo = document.getElementById('btnEjecutarBloqueo');
btnEjecutarBloqueo.addEventListener('click', async () => {
    const fecha = document.getElementById('fechaBloqueoManual').value;
    const hora = document.getElementById('horaBloqueoManual').value;

    if (!fecha || !hora) {
        alert("Elegí día y hora para bloquear.");
        return;
    }

    if (!hora.endsWith(':00') && !hora.endsWith(':30')) {
        alert("🚨 Bloqueo inválido. Los horarios de la peluquería deben gestionarse en punto (:00) o en media hora (:30).");
        return;
    }

    try {
        const respuesta = await fetch('/api/admin/bloquear', {
            method: 'POST',
            headers: { 'Content-Type': 'application/json' },
            body: JSON.stringify({ fecha_hora: `${fecha}T${hora}:00-03:00` })
        });

        if (!respuesta.ok) throw new Error(await respuesta.text());
        
        document.getElementById('horaBloqueoManual').value = '';
        alert("Horario bloqueado con éxito.");
        
        // Refrescamos la tabla principal si Ramón bloqueó algo en el día que está mirando
        if (fecha === fechaFiltro.value) {
            cargarTurnos(fecha);
        }
    } catch (error) {
        alert("Error: " + error.message);
    }
});

// --- Lógica de Excepciones (Vacaciones / Día completo) ---
const excFecha = document.getElementById('excFecha');
const excMotivo = document.getElementById('excMotivo');
const btnGuardarExc = document.getElementById('btnGuardarExc');
const tablaExcepciones = document.getElementById('tablaExcepciones');

async function cargarExcepciones() {
    tablaExcepciones.innerHTML = '<tr><td colspan="3" style="text-align:center;">Cargando...</td></tr>';
    try {
        const respuesta = await fetch('/api/admin/config/excepciones');
        const excepciones = await respuesta.json();

        tablaExcepciones.innerHTML = '';
        
        if (excepciones.length === 0) {
            tablaExcepciones.innerHTML = '<tr><td colspan="3" style="text-align:center; color: #6b7280;">No hay días bloqueados.</td></tr>';
            return;
        }

        excepciones.forEach(exc => {
            const [anio, mes, dia] = exc.fecha.split('T')[0].split('-');
            const fechaLinda = `${dia}/${mes}/${anio}`;

            const tr = document.createElement('tr');
            tr.innerHTML = `
                <td><strong>${fechaLinda}</strong></td>
                <td>${exc.motivo}</td>
                <td>
                    <button onclick="eliminarExcepcion(${exc.id})" 
                            style="background: #ef4444; color: white; border: none; padding: 5px 10px; border-radius: 5px; cursor: pointer;">
                        Desbloquear
                    </button>
                </td>
            `;
            tablaExcepciones.appendChild(tr);
        });
    } catch (error) {
        tablaExcepciones.innerHTML = '<tr><td colspan="3" style="text-align:center; color: red;">Error al cargar.</td></tr>';
    }
}

btnGuardarExc.addEventListener('click', async () => {
    const fecha = excFecha.value;
    const motivo = excMotivo.value.trim();

    if (!fecha) {
        alert("Por favor, elegí una fecha del calendario.");
        return;
    }

    try {
        const respuesta = await fetch('/api/admin/config/excepciones', {
            method: 'POST',
            headers: { 'Content-Type': 'application/json' },
            body: JSON.stringify({ accion: "guardar", fecha: fecha, motivo: motivo })
        });

        if (!respuesta.ok) throw new Error("Error al bloquear el día.");

        excFecha.value = '';
        excMotivo.value = '';
        cargarExcepciones();
    } catch (error) {
        alert(error.message);
    }
});

async function eliminarExcepcion(id) {
    if (!confirm(`¿Seguro que querés volver a abrir este día para reservas?`)) return;

    try {
        const respuesta = await fetch('/api/admin/config/excepciones', {
            method: 'POST',
            headers: { 'Content-Type': 'application/json' },
            body: JSON.stringify({ accion: "eliminar", id: id })
        });

        if (!respuesta.ok) throw new Error("Error al eliminar.");
        
        cargarExcepciones();
    } catch (error) {
        alert(error.message);
    }
}