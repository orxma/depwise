# 💎 PRIVANOX BOT GO EDITION

<p align="center">
  <img src="https://img.shields.io/badge/Language-Go-00ADD8?style=for-the-badge&logo=go" alt="Go">
  <img src="https://img.shields.io/badge/OS-Ubuntu%2024.04-E95420?style=for-the-badge&logo=ubuntu" alt="Ubuntu 24.04">
  <img src="https://img.shields.io/badge/Platform-Linux-FCC624?style=for-the-badge&logo=linux" alt="Linux">
  <img src="https://img.shields.io/badge/Status-Stable-success?style=for-the-badge" alt="Status">
  <img src="https://img.shields.io/badge/Version-8.0.5-blue?style=for-the-badge" alt="Version">
</p>

---

## 🚀 ¿Qué es Privanox Bot?

**privanox Bot Go Edition** es una solución integral y de alto rendimiento para la gestión de servidores VPN y cuentas SSH a través de Telegram. Reescrito completamente en **Go** para garantizar la máxima velocidad, estabilidad y bajo consumo de recursos, este bot transforma tu VPS en un panel de control profesional y automatizado.

---

## 📥 Instalación Rápida (Universal)

> [!NOTE]
> **Compatibilidad OS:** Este bot fue desarrollado y probado rigurosamente en **Ubuntu 24.04**. Se recomienda encarecidamente utilizar esta versión (o distribuciones basadas en ella) para garantizar el correcto funcionamiento de todas las dependencias (Go, Systemd, SSH, Xray, SlowDNS, VayDNS, Slipstream, Dnsdist, etc).

Ejecuta el siguiente comando en tu terminal como usuario **root**:

```bash
bash <(curl -sL https://raw.githubusercontent.com/kevinaldaircama/privanox-code/main/install_go.sh)
```

> [!WARNING]
> **Usuarios de AWS, Oracle Cloud o VPS con usuario normal (`ubuntu`, `admin`, etc):**
> No puedes ejecutar el script directamente ni usando `sudo bash <(...)` debido a restricciones de seguridad. Debes iniciar sesión como Root primero:
> ```bash
> sudo su
> bash <(curl -sL https://raw.githubusercontent.com/kevinaldaircama/privanox-code/main/install_go.sh)
> ```

---

## 🔄 Cómo Actualizar

Si ya tienes el bot funcionando y quieres recibir parches y nuevas funciones **sin perder usuarios ni configuraciones**:

```bash
bash <(curl -sL https://raw.githubusercontent.com/kevinaldaircama/privanox-code/main/install_go.sh)
```
*(Selecciona la opción 1. El sistema actualizará el código preservando tu base de datos las 3 instalaciones son los mismos).*

---

## ✨ Características Principales

### 🛠️ Gestión de Protocolos (All-in-One)
- **Multiplexación DNS Avanzada (Nuevo):** Gracias a `dnsdist` y filtros `U32` a nivel de Kernel (Netfilter), el bot puede correr **SlowDNS, VayDNS y Slipstream** simultáneamente por el mismo puerto UDP 53 sin chocar. Soporte nativo para redes móviles (IPv6/NAT64).
- **SlowDNS / Noiz DNS:** Túneles DNS de altísima estabilidad y anti-bloqueos.
- **Slipstream:** Protocolo híbrido avanzado con alternativa nativa QUIC por el puerto 443.
- **SSH/Dropbear/WS TLS HTTP:** Gestión completa de cuentas con límites de conexión y banners HTML generados al vuelo.
- **Xray (VMess):** Protocolo de última generación sobre WebSocket compatible con Cloudflare y HAProxy.
- **ZiVPN & UDP Custom:** Soporte para protocolos de gaming UDP y bypass robusto (rango 6000:19999).
- **Falcon Proxy & ProxyDT:** Proxies HTTP optimizados.

### 🛡️ Administración Pro y Utilidades
- **Auto-Configuración Root:** Habilita el acceso Root por SSH de forma permanente en VPS de nube.
- **Reinicio Automático Resiliente:** El bot reconstruye toda la red, iptables, reglas u32, IPv6 y Dnsdist en cada reinicio para garantizar que ningún protocolo se caiga.
- **Mensaje Global (Broadcast):** Envía anuncios masivos a todos los usuarios.
- **Monitoreo en Tiempo Real:** Visualización en vivo de métricas de VPS (Cores, RAM, Disco, Uptime) y protocolos configurados.
- **Sistema de Baneo y Cuotas:** Evita abusos de revendedores y bloquea a usuarios indeseados.
- **Alertas de Expiración Automáticas:** Avisa a los administradores 1 día y 1 hora antes de que una cuenta SSH, ZiVPN o Xray expire.

### 🧹 Mantenimiento Inteligente
- **Persistencia de Datos Inquebrantable:** El tráfico y las configuraciones están seguras tras cualquier reinicio.
- **Resiliencia de Servicios:** Recuperación automática de HAProxy, Xray y DNS a través de rutinas automatizadas.
- **HAProxy Auto-Recovery:** El bot verifica que HAProxy esté corriendo, mata procesos invasores y reinicia si es necesario.

---

## 💸 Monetización Inteligente (Monetag + Vercel)

El bot cuenta con un sistema de monetización nativo. Cuando está activo, requiere que los usuarios públicos vean un anuncio (Rewarded Interstitial) antes de poder **crear** o **renovar** cualquier cuenta VPN. Esto te permite ganar dinero automáticamente por el uso de tus servidores.

El proceso de configuración ahora está **100% automatizado mediante un asistente interactivo** dentro del bot.

### 📝 Guía Paso a Paso para Activar la Monetización

**1. Crear cuenta en Monetag**
- Regístrate como publisher en [Monetag.com](https://monetag.com/?ref_id=tlyy) (enlace de referido).
- En el panel, añade una nueva aplicación y selecciona el formato **Telegram Mini App**.
- El sistema te pedirá el **nombre de usuario (Username)** de tu bot (ej: `@MiBot_bot` o `@Depwise_bot` dependiendo de cómo se llame el bot que estás configurando).
- Genera un bloque de anuncios de tipo **Rewarded Interstitial**.

**2. Obtener los Códigos**
Monetag te mostrará las instrucciones de integración con dos fragmentos de código:
- **El Script SDK:** Una etiqueta `<script>` (ej: `<script src='//libtl.com/sdk.js' data-zone='1234567' data-sdk='show_1234567'></script>`).
- **El Bloque Rewarded:** Un bloque de JavaScript (ej: `show_1234567().then(() => { ... })`).

**3. Usar el Asistente del Bot**
- En tu bot de Telegram, entra a **⚙️ Ajustes Pro** y haz clic en **⚙️ Configurar MiniApp Ads**.
- El bot te pedirá primero la etiqueta `<script>`. Cópiala y pégala en el chat.
- Luego te pedirá el bloque **Rewarded**. Cópialo y pégalo en el chat.
- **¡Magia!** El bot leerá tus códigos, configurará el HTML internamente de forma automática y te enviará por el chat un archivo `.zip` (`monetag_miniapp.zip`) totalmente listo para usar, vinculado a tu cuenta. **Debes descargar este `.zip`, descomprimirlo, y utilizar el archivo `.html` que viene adentro.**

**4. Subir a Vercel**
- Crea una cuenta gratuita en [Vercel.com](https://vercel.com/) o [GitHub Pages](https://pages.github.com/).
- Arrastra la carpeta descomprimida (o sube directamente el archivo `monetag_miniapp.html`) para crear un nuevo proyecto en Vercel (o tu hosting preferido).
- Una vez publicado, asegúrate de obtener el enlace público directo al archivo HTML (ej: `https://mi-miniapp.vercel.app/monetag_miniapp.html`).

**5. Activación Final**
- Envía ese enlace a tu bot en Telegram (es el último paso que te estará pidiendo el asistente).
- El bot guardará el enlace y **activará automáticamente** el AdWall para todos los usuarios. ¡A partir de ese momento empezarás a generar ingresos en piloto automático!

---

## 📈 Historial de Novedades

### 🚀 v8.0.5 — Soporte Multi-idioma y Traducciones Completas
- **Multi-idioma (Inglés y Español):** El bot ahora detecta o permite elegir el idioma preferido para todos los menús.
- **Traducciones 100% Mapeadas:** Los menús de Edición (contraseña, renovación, límite), Monitor de Estado, y Creador Xray/ZiVPN ahora soportan traducciones impecables.

### 🚀 v8.0 — El Mayor Salto Técnico (Multiplexación Dnsdist & IPv6)
- **Multiplexación Total:** Ahora puedes instalar y correr SlowDNS (Noiz DNS), VayDNS y Slipstream de forma simultánea. El bot levanta un loadbalancer `dnsdist` de grado de servidor.
- **Filtro Kernel U32:** Se han diseñado reglas Hexadecimales U32 personalizadas para separar el tráfico QUIC del DNS estándar directamente en Netfilter, bajando la latencia a 0 y reduciendo el consumo de CPU.
- **Soporte Datos Móviles (IPv6):** Implementado soporte integral `ip6tables` y ACLs universales `0.0.0.0/0` y `::/0` en Dnsdist para redes NAT64 (Telcel, Claro, Movistar, etc.).
- **Resiliencia Anti-Reboot:** El bot reconstruye dinámicamente toda la red compleja (iptables + dnsdist) de forma concurrente tras cualquier reinicio.
- **UI Profesional:** Los datos de conexión SSH y VPN se envían al usuario de forma bellamente estructurada. Prevención de colisiones de dominios (NS) entre protocolos.

### 🚀 v7.9 y Anteriores
- Fix de deadlocks y cuelgues.
- Payloads automáticos y soporte Cloudflare/Cloudfront.
- Módulo avanzado de Copias de Seguridad en Google Drive (cada 24h).

---

## ☁️ Copias de Seguridad Nativas (Telegram)

El bot incluye un sistema de respaldos nativo en Telegram. Permite copias **inmediatas** desde el panel y **automáticas (cada 1, 3, 7 o 30 días)** enviando el archivo directamente al chat del administrador. No requiere configuración externa.

---

## 🛠️ Solución de Problemas (Troubleshooting)

| Síntoma | Causa Probable | Solución |
| :--- | :--- | :--- |
| **No conecta con Datos Móviles** | El dominio NS no tiene IP o IPv6 falla | Actualiza a la V8.0 y reinstala los protocolos. |
| **El bot no responde** | Proceso congelado | `systemctl restart depwise` |
| **Xray/VMess no conecta** | HAProxy o Xray no iniciaron | `systemctl status haproxy xray` |
| **Error Backup Automático** | ID de Chat inválido | Volver a configurar intervalo de backup en menú |

---

## 💎 Créditos y Soporte

Este proyecto es desarrollado y mantenido con pasión por:

- **👨‍💻 Desarrollador:** [@Dan3651](https://t.me/Dan3651)
- **📢 Canal Oficial:** [Depwise Channel](https://t.me/orxtunnel)

---

<p align="center">
  <i>"Potenciando tu VPS con la velocidad de Go y la potencia del Kernel."</i><br>
  <b>© 2026 Depwise Project regalo por Kevin tech tutorials</b>
</p>
