const fechaInput = document.getElementById('fechaFiltro');
const tablaTurnos = document.getElementById('tablaTurnos');
const horaBloqueo = document.getElementById('horaBloqueo');
const btnBloquear = document.getElementById('btnBloquear');

document.addEventListener('DOMContentLoaded', () => {
    const hoy = new Date().toISOString().split('T')[0];
    fechaInput.value = hoy;
    cargarTurnos(hoy);
});

fechaInput.addEventListener('change', (e) => cargarTurnos(e.target.value));

btnBloquear.addEventListener('click', async () => {
    const fecha = fechaInput.value;
    const hora = horaBloqueo.value;

    if (!fecha || !hora) {
        alert("Elegí día y hora para bloquear.");
        return;
    }

    try {
        const respuesta = await fetch('/api/admin/bloquear', {
            method: 'POST',
            headers: { 'Content-Type': 'application/json' },
            body: JSON.stringify({ fecha_hora: `${fecha}T${hora}:00Z` })
        });

        if (!respuesta.ok) throw new Error(await respuesta.text());
        horaBloqueo.value = '';
        cargarTurnos(fecha);
    } catch (error) {
        alert("Error: " + error.message);
    }
});

async function cargarTurnos(fecha) {
    tablaTurnos.innerHTML = '<tr><td colspan="4" style="text-align:center;">Buscando...</td></tr>';
    
    try {
        const respuesta = await fetch(`/api/admin/turnos?fecha=${fecha}`);
        const turnos = await respuesta.json();
        tablaTurnos.innerHTML = '';

        if (turnos.length === 0) {
            tablaTurnos.innerHTML = '<tr><td colspan="4" style="text-align:center;">No hay turnos.</td></tr>';
            return;
        }

        turnos.forEach(turno => {
            const tr = document.createElement('tr');
            
            // Lógica visual: si es "y media", resaltamos para Ramón
            const esPrivado = turno.hora.endsWith(":30");
            if (esPrivado) {
                tr.style.backgroundColor = "#eff6ff"; // Azul muy clarito para bloques de Ramón
            }

            let badgeClass = turno.estado === 'CONFIRMADO' ? 'bg-green' : 'bg-orange';
            let icon = turno.estado === 'CONFIRMADO' ? '✅' : '⏳';

            tr.innerHTML = `
                <td><strong>${turno.hora}</strong> ${esPrivado ? '🔒' : ''}</td>
                <td>${turno.nombre_cliente}</td>
                <td><a href="https://wa.me/549${turno.telefono}" target="_blank">📱 ${turno.telefono}</a></td>
                <td><span class="badge ${badgeClass}">${icon} ${turno.estado}</span></td>
            `;
            tablaTurnos.appendChild(tr);
        });
    } catch (error) {
        tablaTurnos.innerHTML = '<tr><td colspan="4" style="text-align:center; color:red;">Error de conexión</td></tr>';
    }
}