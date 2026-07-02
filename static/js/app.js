// Variables globales para guardar la selección
let fechaSeleccionada = '';
let horaSeleccionada = '';

document.addEventListener('DOMContentLoaded', () => {
    cargarMontoSena();
});

const fechaInput = document.getElementById('fechaTurno');
const grillaHorarios = document.getElementById('grillaHorarios');
const contenedorHorarios = document.getElementById('contenedorHorarios');
const formularioReserva = document.getElementById('formularioReserva');
const btnReservar = document.getElementById('btnReservar');
const mensajeError = document.getElementById('mensajeError');

// Restringir el calendario para que no elijan fechas pasadas
const hoy = new Date().toISOString().split('T')[0];
fechaInput.setAttribute('min', hoy);

// Cuando el usuario elige un día
fechaInput.addEventListener('change', async (e) => {
    fechaSeleccionada = e.target.value;
    horaSeleccionada = ''; 
    formularioReserva.classList.add('hidden');
    grillaHorarios.innerHTML = '<p style="grid-column: span 3; text-align: center;">Cargando...</p>';
    contenedorHorarios.classList.remove('hidden');

    try {
        const respuesta = await fetch(`/api/disponibilidad?fecha=${fechaSeleccionada}`);
        const data = await respuesta.json();

        grillaHorarios.innerHTML = '';
        
        if (!data.disponibles || data.disponibles.length === 0) {
            grillaHorarios.innerHTML = '<p style="grid-column: span 3; text-align: center; color: red;">No hay turnos para este día.</p>';
            return;
        }

        // Dibujamos los botones de horarios
        data.disponibles.forEach(hora => {
            const btn = document.createElement('button');
            btn.className = 'btn-horario';
            btn.textContent = hora;
            btn.onclick = () => seleccionarHora(btn, hora);
            grillaHorarios.appendChild(btn);
        });
    } catch (error) {
        grillaHorarios.innerHTML = '<p style="grid-column: span 3; color: red;">Error al cargar horarios.</p>';
    }
});

// Lógica visual al clickear un horario
function seleccionarHora(botonClickeado, hora) {
    document.querySelectorAll('.btn-horario').forEach(b => b.classList.remove('selected'));
    botonClickeado.classList.add('selected');
    
    horaSeleccionada = hora;
    formularioReserva.classList.remove('hidden');
    mensajeError.textContent = '';
}

// Cuando hace clic en "Abonar Seña"
btnReservar.addEventListener('click', async () => {
    const nombre = document.getElementById('nombreCliente').value.trim();
    const telefono = document.getElementById('telefonoCliente').value.trim();
    const errorSpan = document.getElementById('errorTelefono');
    const regex = /^\+?[0-9]{7,20}$/;

    if (!nombre || !telefono) {
        mensajeError.textContent = "Por favor, completá nombre y teléfono.";
        return;
    }
    if (!regex.test(telefono)) {
        errorSpan.innerText = "⚠️ Número inválido";
        errorSpan.style.display = 'inline'; 
        return; 
    } else {
        errorSpan.style.display = 'none'; 
    }

    btnReservar.disabled = true;
    btnReservar.textContent = "Redirigiendo a Mercado Pago...";
    mensajeError.textContent = "";

    const fechaISO = `${fechaSeleccionada}T${horaSeleccionada}:00Z`;

    try {
        const respuesta = await fetch('/api/reservar', {
            method: 'POST',
            headers: { 'Content-Type': 'application/json' },
            body: JSON.stringify({
                fecha_hora_inicio: fechaISO,
                nombre_cliente: nombre,
                telefono: telefono,
            })
        });

        const data = await respuesta.json();

        if (!respuesta.ok) {
            throw new Error(data.mensaje || "Error al procesar el turno");
        }

        // Redirigimos al cliente al link de Mercado Pago
        window.location.href = data.link_pago;

    } catch (error) {
        const respuesta = await fetch('/api/admin/config/precios');
        const config = await respuesta.json();
        
        // Calculamos la seña tal cual lo hace Go
        const montoSena = (config.precio_turno * config.porcentaje_sena) / 100;
        mensajeError.textContent = error.message;
        btnReservar.disabled = false;
        btnReservar.innerText = `Abonar Seña ($${montoSena})`;
    }
});

async function cargarMontoSena() {
    try {
        const respuesta = await fetch('/api/admin/config/precios');
        const config = await respuesta.json();
        
        // Calculamos la seña tal cual lo hace Go
        const montoSena = (config.precio_turno * config.porcentaje_sena) / 100;
        
        
        const btnReservar = document.getElementById('btnReservar');
        if (btnReservar) {
            btnReservar.innerText = `Abonar Seña ($${montoSena})`;
        }
    } catch (error) {
        console.error("Error obteniendo el precio de la seña:", error);
    }
}