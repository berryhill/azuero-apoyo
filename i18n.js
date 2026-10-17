// Inicio en inglés (público): el español sigue siendo el texto de la página; esto lo cambia en pantalla al inglés con un
// diccionario de frases exactas de la interfaz (nunca toca lo que escribió la gente: nombres, lugares, notas) y unas
// reglas para textos con números («3 personas», «hace 2 h»). Recorre los textos y los atributos visibles (placeholder,
// title, aria-label, alt), guarda el original de cada uno para volver al español tal cual, y sigue los cambios de la
// página con un MutationObserver. nav.js pone el selector ES | EN junto a «Aa» y llama I18n.set(lang).
(() => {
  const EN = {
    // Encabezado, pestañas y navegación
    'Azuero Apoyo': 'Azuero Apoyo',
    'Mapa de apoyo · Azuero': 'Support map · Azuero',
    'Soy voluntario': "I'm a volunteer",
    'Iniciar sesión': 'Sign in',
    'Soy voluntario: iniciar sesión': "I'm a volunteer: sign in",
    '¿Cómo funciona?': 'How does it work?',
    'Buscar…': 'Search…',
    'Buscar': 'Search',
    'Buscar puntos o una ubicación de Google Maps': 'Search places or a Google Maps location',
    'Borrar búsqueda': 'Clear search',
    'Borrar búsqueda (en todas las páginas)': 'Clear search (on every page)',
    'Ver o cambiar la búsqueda': 'See or change the search',
    'Filtrado por «': 'Filtered by «',
    'Mapa': 'Map',
    'Resumen': 'Summary',
    'Esta donación': 'This donation',
    'Vista': 'View',
    'Vista móvil': 'Mobile view',
    'Menú': 'Menu',
    'Más': 'More',
    'Páginas': 'Pages',
    'Cambiar de página': 'Change page',
    'Tamaño del texto': 'Text size',
    'Texto más pequeño': 'Smaller text',
    'Texto más grande': 'Larger text',
    'Cerrar': 'Close',
    'Cancelar': 'Cancel',
    'Listo': 'Done',
    'Atrás': 'Back',
    '‹ Volver': '‹ Back',
    'Volver': 'Back',
    'Siguiente': 'Next',
    'Salir': 'Sign out',
    'Cerrar sesión': 'Sign out',
    '¿Cerrar sesión?': 'Sign out?',
    'Tendrás que volver a entrar con tu código de acceso.': "You'll need to sign in again with your access code.",
    'Login': 'Login',
    'Registro ciudadano · corte descargado 12 oct 2026': 'Citizen registry · snapshot downloaded Oct 12, 2026',
    'Búsqueda global': 'Global search',
    'Buscar en todos los puntos': 'Search all places',
    'Código, zona, lugar, necesidad, quién reporta…': 'Code, area, place, need, who reported…',
    'Escribe para buscar en todos los puntos: código, zona, lugar, necesidad, quién reporta o un contacto.': 'Type to search all places: code, area, place, need, who reported, or a contact.',
    'Ficha ›': 'Details ›',
    'Entorno de prueba (sandbox)': 'Test environment (sandbox)',
    'PRUEBA · sandbox': 'TEST · sandbox',
    'Entorno de prueba: los datos no se guardan en el sitio real. Toca para saber más.': "Test environment: data isn't saved on the real site. Tap to learn more.",
    'Tu acceso cambió: ya no puedes usar esa página. Tus páginas están en el menú.': "Your access changed: you can't use that page anymore. Your pages are in the menu.",

    // Reportar y botones del mapa
    'Reportar un lugar que necesita ayuda': 'Report a place that needs help',
    'Reportar a alguien que necesita ayuda': 'Report someone who needs help',
    'Reportar ayuda': 'Report help',
    'Reportar punto': 'Report a place',
    'Reportar necesidad': 'Report a need',
    'Reportar otra necesidad': 'Report another need',
    'Reportar': 'Report',
    'Registrar punto aquí': 'Register a place here',
    'Mostrar mi ubicación': 'Show my location',
    'Tu ubicación': 'Your location',
    'Compartir mi ubicación': 'Share my location',
    'Ver el sismo en el mapa': 'See the earthquake on the map',
    'Mapa de puntos de entrega': 'Map of delivery places',
    'Mapa base requiere conexión a internet.': 'The base map needs an internet connection.',
    'Mapa no disponible sin conexión. La lista de puntos sigue accesible.': 'Map unavailable offline. The list of places is still available.',
    'Puntos sin GPS no salen en el mapa.': "Places without GPS don't show on the map.",
    'Ubicación buscada': 'Searched location',
    'No es un punto registrado. ¿Quieres registrarlo?': "It's not a registered place. Do you want to register it?",
    'Enlace corto sin coordenadas: ábrelo, espera a que cargue el mapa y pega aquí la URL completa del navegador.': "Short link without coordinates: open it, wait for the map to load, and paste the browser's full URL here.",
    'Este navegador no permite obtener tu ubicación.': "This browser can't get your location.",
    'No se pudo obtener tu ubicación': "Couldn't get your location",
    'Permiso de ubicación denegado. Actívalo en la configuración del navegador para este sitio.': "Location permission denied. Turn it on in your browser's settings for this site.",
    'Acepta el aviso de ubicación para volver a compartirla.': 'Accept the location prompt to share it again.',
    'Arrastrar para mover': 'Drag to move',
    'Epicentro': 'Epicenter',
    'Réplica fuerte': 'Strong aftershock',
    '© OpenStreetMap contributors': '© OpenStreetMap contributors',

    // Filtros y lista
    'Filtros': 'Filters',
    'Puntos': 'Places',
    'En mapa': 'On map',
    'Necesidades': 'Needs',
    'Necesidad': 'Need',
    'Con necesidades': 'With needs',
    'Abiertas': 'Open',
    'En camino': 'On the way',
    'Completadas': 'Completed',
    'Atendidos': 'Served',
    'Urgentes': 'Urgent',
    'Normales': 'Normal',
    'Urgente': 'Urgent',
    'Normal': 'Normal',
    'Todas': 'All',
    'Todos los puntos': 'All places',
    'Solo nuestro equipo': 'Only our team',
    'Coordinadores': 'Coordinators',
    'Buscar quién reporta': 'Search who reported',
    'Buscar quién reporta…': 'Search who reported…',
    'Sin quién reporta': 'No reporter',
    'Orden': 'Sort',
    'Más recientes': 'Newest',
    'Más antiguos': 'Oldest',
    'Sin coincidencias': 'No matches',
    'Sin resultados.': 'No results.',
    'Ningún punto con necesidades abiertas o en camino.': 'No places with open or on-the-way needs.',
    'Ningún punto con necesidades urgentes ahora.': 'No places with urgent needs right now.',
    'Ningún punto con necesidades normales ahora.': 'No places with normal needs right now.',
    'Ningún punto atendido todavía.': 'No places served yet.',
    'Sin necesidades en este estado.': 'No needs in this state.',
    'Sin necesidades abiertas ni en camino.': 'No open or on-the-way needs.',
    'Sin necesidades abiertas': 'No open needs',
    'Sin necesidades abiertas ahora.': 'No open needs right now.',
    'Sin necesidades abiertas ahora': 'No open needs right now',
    'Sin necesidades pendientes': 'No pending needs',
    'Sin necesidades de servicio': 'No service needs',
    'Sin pendientes': 'Nothing pending',
    'Sin datos': 'No data',
    'Sin GPS': 'No GPS',
    'sin GPS': 'no GPS',
    'Sin GPS: no aparece en el mapa': "No GPS: it doesn't appear on the map",
    'Sin GPS: confirmar ubicación antes de salir.': 'No GPS: confirm the location before heading out.',
    'Sin ubicación para Maps': 'No location for Maps',
    'Cargando…': 'Loading…',
    'Cargando necesidades…': 'Loading needs…',
    'No se pudieron cargar las necesidades.': "Couldn't load the needs.",

    // Ficha del punto
    'Ubicación no especificada': 'Location not specified',
    'Personas': 'People',
    'Adultos': 'Adults',
    'Niños': 'Children',
    'Bebés': 'Babies',
    'Mayores': 'Seniors',
    'Adultos mayores': 'Seniors',
    'Discapacidad': 'Disability',
    'Discapacitados': 'People with disabilities',
    'Con discapacidad': 'With disabilities',
    'Familias': 'Families',
    'Generador': 'Generator',
    'Fila': 'Row',
    'Notas de grupo': 'Group notes',
    'Indicaciones': 'Directions',
    'Otras notas': 'Other notes',
    'Notas de servicio': 'Service notes',
    'Registrado': 'Registered',
    'Reportado por': 'Reported by',
    'Servicio reportado': 'Service reported',
    'Servicio incierto': 'Service uncertain',
    'Sin confirmación': 'Unconfirmed',
    'Compartir': 'Share',
    'Compartir enlace': 'Share link',
    'Compartir enlace del punto': 'Share link to this place',
    '✓ Enlace copiado': '✓ Link copied',
    'Enlace copiado': 'Link copied',
    'Copia este enlace:': 'Copy this link:',
    'Ver en el mapa': 'See on the map',
    'Abrir en Google Maps ↗': 'Open in Google Maps ↗',
    'Abrir en Apple Maps ↗': 'Open in Apple Maps ↗',
    'Abrir en Google Maps': 'Open in Google Maps',
    'Abrir en Apple Maps': 'Open in Apple Maps',
    'Abrir en Maps ↗': 'Open in Maps ↗',
    'Google Maps': 'Google Maps',
    'Apple Maps': 'Apple Maps',
    'Pedir comida preparada': 'Request prepared food',
    'Llamar': 'Call',
    'Copiar': 'Copy',
    'WhatsApp': 'WhatsApp',
    'Llamar o escribir por WhatsApp': 'Call or message on WhatsApp',
    'Selecciona el número y cópialo': 'Select the number and copy it',

    // Resumen
    'Cifras reunidas por voluntarios: úsalas como guía, no como datos oficiales.': 'Figures gathered by volunteers: use them as a guide, not as official data.',
    'Alcance': 'Reach',
    'Zonas': 'Areas',
    'Quiénes son': 'Who they are',
    'Necesidades ahora': 'Needs right now',
    'por llevar o en camino': 'to deliver or on the way',
    'Por llevar': 'To deliver',
    'Qué se pide': "What's being asked for",
    'pendiente': 'pending',
    'pendientes': 'pending',
    'Nada pendiente ahora.': 'Nothing pending right now.',
    'Zonas con necesidades': 'Areas with needs',
    'Ninguna zona con necesidades abiertas.': 'No areas with open needs.',
    'Sin zona': 'No area',
    'punto': 'place',
    'puntos': 'places',
    'Materiales para reparaciones de casa (': 'Materials for house repairs (',
    'Copiar lista': 'Copy list',
    '✓ Copiada': '✓ Copied',
    'No se pudo copiar': "Couldn't copy",

    // Sismo
    'Terremoto en la península de Azuero': 'Earthquake on the Azuero Peninsula',
    '9 de octubre de 2026 · 12:56 p. m. · magnitud 7,7': 'October 9, 2026 · 12:56 p.m. · magnitude 7.7',
    'Un sismo con epicentro a 12 km de': 'An earthquake centered 12 km from',
    ', entre Macaracas y Tonosí, sacudió Los Santos, Herrera y Veraguas. Hay más de': ', between Macaracas and Tonosí, shook Los Santos, Herrera and Veraguas. There are more than',
    '13.000 damnificados': '13,000 people affected',
    ', unas': ', about',
    '3.700 personas en albergues': '3,700 people in shelters',
    'y casi': 'and almost',
    '900 familias perdieron su casa': '900 families lost their homes',
    '; colapsaron 35 escuelas y hay caminos y puentes dañados. Las réplicas siguen (más de 600).': '; 35 schools collapsed and roads and bridges are damaged. Aftershocks continue (more than 600).',
    'Comunidades como Tonosí, Cañas y La Pitaloza aún necesitan': 'Communities like Tonosí, Cañas and La Pitaloza still need',
    'agua, comida, colchones y medicinas': 'water, food, mattresses and medicine',
    ', sobre todo para embarazadas, bebés y adultos mayores. Si conoces un lugar que necesita ayuda, repórtalo: cada reporte llega a los equipos que llevan la ayuda.': ', especially for pregnant women, babies and seniors. If you know a place that needs help, report it: every report reaches the teams that deliver aid.',
    'Sismo:': 'Earthquake:',
    'Sismo': 'Earthquake',
    'cifras oficiales al 16 oct 2026:': 'official figures as of Oct 16, 2026:',

    // ¿Quieres ayudar? / donar
    '¿Quieres ayudar?': 'Want to help?',
    'Hay mucho por hacer: reportar a quien necesita ayuda, llevar donaciones a las comunidades, organizar centros de acopio, cocinar, verificar reportes por teléfono o reparar casas. Déjanos tus datos y te contactamos.': 'There is a lot to do: report people who need help, bring donations to communities, organize collection centers, cook, verify reports by phone or repair houses. Leave us your details and we will contact you.',
    '¿Prefieres donar?': 'Prefer to donate?',
    'Dona con Unidos por Azuero.': 'Donate through Unidos por Azuero.',
    'unidosporazuero.com ↗': 'unidosporazuero.com ↗',
    'Unidos por Azuero': 'Unidos por Azuero',
    'Sí, quiero ayudar': 'Yes, I want to help',
    'Déjanos tus datos y te contactamos para coordinar.': "Leave us your details and we'll contact you to coordinate.",
    'obligatorio · teléfono o correo: al menos uno': 'required · phone or email: at least one',
    'Nombre': 'Name',
    'Teléfono': 'Phone',
    'Correo': 'Email',
    '(o correo)': '(or email)',
    '(o teléfono)': '(or phone)',
    '(opcional)': '(optional)',
    '(opcional, para recuperar tu código)': '(optional, to recover your code)',
    '¿Cómo puedes ayudar?': 'How can you help?',
    'Reportar necesidades': 'Report needs',
    'Llevar donaciones a la gente': 'Bring donations to people',
    'Organizar centros de acopio y despachos': 'Organize collection centers and dispatches',
    'Cocinar': 'Cook',
    'Centro de llamadas: verificar necesidades reportadas': 'Call center: verify reported needs',
    'Reparar casas': 'Repair houses',
    'Donar': 'Donate',
    'Mensaje': 'Message',
    'Cuéntanos de ti o cómo quieres ayudar': 'Tell us about yourself or how you want to help',
    'Ej. María Pérez': 'E.g. María Pérez',
    'Ej. 6000-0000': 'E.g. 6000-0000',
    'Enviar': 'Send',
    'Acepto que me contacten por teléfono o correo.': 'I agree to be contacted by phone or email.',
    'Acepto que mi nombre y teléfono se compartan con quienes ayudan en la red, para que me contacten sobre los puntos o necesidades que yo reporte.': 'I agree to share my name and phone with the people helping in the network, so they can contact me about the places or needs I report.',
    'Marca al menos una forma de ayudar.': 'Choose at least one way to help.',
    'Para esta forma de ayudar necesitamos tu teléfono.': 'For this way of helping we need your phone.',
    'Marca la casilla para que te podamos contactar.': 'Check the box so we can contact you.',
    'Marca la casilla para compartir tu nombre y teléfono con quienes ayudan.': 'Check the box to share your name and phone with the people helping.',
    'Marca la casilla para aceptar que tu nombre y teléfono se compartan.': 'Check the box to agree to share your name and phone.',
    'Escribe tu nombre.': 'Write your name.',
    'Escribe tu teléfono o tu correo.': 'Write your phone or email.',
    'Escribe tu número de teléfono (7 dígitos o más).': 'Write your phone number (7 digits or more).',
    'Teléfono incompleto: 7 dígitos o más (ej. 6000-0000).': 'Incomplete phone: 7 digits or more (e.g. 6000-0000).',
    'Revisa el correo (ej. nombre@gmail.com).': 'Check the email (e.g. name@gmail.com).',
    'Revisa el correo (por ejemplo nombre@gmail.com) o déjalo vacío.': 'Check the email (for example name@gmail.com) or leave it empty.',
    'Revisa el campo marcado.': 'Check the highlighted field.',
    'Recibimos tu mensaje. Te vamos a llamar o escribir para coordinar cómo ayudar.': "We got your message. We'll call or write to you to coordinate how to help.",
    '✓ Ya recibimos tus datos. ¡Gracias!': '✓ We got your details. Thank you!',
    'No se pudo enviar.': "Couldn't send.",
    'No se pudo enviar:': "Couldn't send:",
    'Error de envío': 'Sending error',

    // «Esta donación» (/runner/<código>)
    'Hazte runner para llevar esta donación': 'Become a runner to deliver this donation',
    'Los runners llevan a cada punto lo que necesita.': 'Runners bring each place what it needs.',
    'Inicia sesión con tu código o crea tu cuenta para llevarla.': 'Sign in with your code or create your account to deliver it.',
    'Pasa a Runners para hacerte runner y llevarla.': 'Go to Runners to become a runner and deliver it.',
    'Quiero ser runner': 'I want to be a runner',
    'NECESITA': 'NEEDS',
    'Esta donación ya no está publicada.': 'This donation is no longer published.',

    // ¿Cómo funciona? (nav.js)
    'Cada parte es igual de importante. Juntos, la ayuda llega.': 'Every part matters just as much. Together, help gets there.',
    'Reportan quién necesita ayuda': 'They report who needs help',
    'Solo con reportar ya estás ayudando': 'Just by reporting you are already helping',
    'El lugar y su contacto': 'The place and its contact',
    'Qué hace falta: comida, agua, una silla de ruedas, gas, una medicina…': "What's missing: food, water, a wheelchair, gas, a medicine…",
    'Ven lo que reportaron los scouts': 'They see what the scouts reported',
    'Conectan la donación con quien la necesita': 'They connect the donation with whoever needs it',
    'Centros de acopio': 'Collection centers',
    'Preparan la carga y piden un runner': 'They prepare the load and ask for a runner',
    'Llevan la cuenta de todo lo que sale': 'They keep track of everything that goes out',
    'Llevan la ayuda hasta allá': 'They take the help all the way there',
    'Tienes carro, pickup o moto': 'You have a car, pickup or motorbike',
    'Recoges donaciones: no tienen que ser tuyas': "You pick up donations: they don't have to be yours",
    'Las llevas a donde hacen falta': 'You take them where they are needed',

    // Entrar / crear cuenta
    'Ya tengo código': 'I have a code',
    'Código de acceso': 'Access code',
    'Tu código de acceso': 'Your access code',
    'Tu código de acceso:': 'Your access code:',
    'Recordar mi código de acceso': 'Remember my access code',
    'Entrar': 'Sign in',
    'Crear cuenta nueva': 'Create a new account',
    'Crear mi código': 'Create my code',
    'Nombre completo': 'Full name',
    'Nombre y apellido': 'First and last name',
    'Tu nombre': 'Your name',
    'Tu teléfono': 'Your phone',
    'nombre@gmail.com': 'name@gmail.com',
    'Escribe tu código de acceso, o crea una cuenta nueva abajo.': 'Write your access code, or create a new account below.',
    'En «¿Quién eres?», escribe tu nombre y apellido.': 'In «Who are you?», write your first and last name.',
    '¿Olvidaste tu código?': 'Forgot your code?',
    'Pide uno nuevo a administración o a tu coordinador, con tu nombre y el teléfono o correo con que te registraste.': 'Ask the admins or your coordinator for a new one, with your name and the phone or email you registered with.',
    'Cópialo y guárdalo. En este teléfono ya queda guardado: la próxima vez en «Login» aparece escrito y solo tocas «Entrar». En otro teléfono, pégalo en «Login».': 'Copy it and keep it. It stays saved on this phone: next time it appears in «Login» and you just tap «Sign in». On another phone, paste it in «Login».',
    'Copiar código': 'Copy code',
    'Selecciónalo y cópialo': 'Select it and copy it',
    '¡Listo, ya entraste!': "Done, you're in!",
    'Tu cuenta:': 'Your account:',
    'No se pudo entrar.': "Couldn't sign in.",
    'No se pudo crear tu cuenta.': "Couldn't create your account.",
    'No se pudo crear la cuenta.': "Couldn't create the account.",
    'No pudimos crear tu cuenta:': "We couldn't create your account:",
    'No se pudo obtener tu ID. Vuelve a entrar.': "Couldn't get your ID. Sign in again.",
    'Escribe tu código de acceso para retomar tus últimas carreras.': 'Write your access code to resume your latest runs.',

    // Formulario de punto (pointform.js)
    '¿Dónde estás?': 'Where are you?',
    '¿Con quién estás?': 'Who are you with?',
    '¿Quién eres?': 'Who are you?',
    '¿Quién necesita ayuda?': 'Who needs help?',
    '¿Qué necesitas?': 'What do you need?',
    '¿Qué necesitan?': 'What do they need?',
    '¿Qué necesitan ahora?': 'What do they need now?',
    '¿Cuántas personas?': 'How many people?',
    '¿Es urgente?': 'Is it urgent?',
    'Márcalo si no puede esperar.': "Check it if it can't wait.",
    'Zona (pueblo o comunidad)': 'Area (town or community)',
    'Lugar': 'Place',
    'Ej. Río Viejo': 'E.g. Río Viejo',
    'Ej. Escuela de la comunidad': 'E.g. Community school',
    'Ej. Casa de la familia Nieto': 'E.g. The Nieto family home',
    'Ej. después de la iglesia': 'E.g. after the church',
    'Indicaciones de llegada': 'Directions to get there',
    'Cómo llegar, puntos de referencia…': 'How to get there, landmarks…',
    'Usar mi ubicación': 'Use my location',
    'o pega un enlace de Google Maps': 'or paste a Google Maps link',
    'o un enlace de Google Maps': 'or a Google Maps link',
    'Buscando tu ubicación…': 'Finding your location…',
    'Buscando señal GPS': 'Looking for GPS signal',
    'Buscando señal GPS…': 'Looking for GPS signal…',
    'Todavía sin señal GPS…': 'Still no GPS signal…',
    'Buena precisión': 'Good accuracy',
    'Precisión media: espera un poco más': 'Medium accuracy: wait a little longer',
    'Poca precisión: sal a cielo abierto si puedes': 'Low accuracy: go outside if you can',
    'Usar esta ubicación': 'Use this location',
    'Guardar ubicación': 'Save location',
    'Toca para mover la ubicación': 'Tap to move the location',
    'Mueve el mapa para ajustar la cruz.': 'Move the map to adjust the crosshair.',
    'Mover la cruz a tu ubicación GPS': 'Move the crosshair to your GPS location',
    'Sin pin todavía: toca el mapa o usa tu ubicación.': 'No pin yet: tap the map or use your location.',
    'Mapa de puntos y tu ubicación': 'Map of places and your location',
    'Vista previa de la ubicación elegida': 'Preview of the chosen location',
    'Cerrar mapa': 'Close map',
    'Tu GPS queda fuera de Panamá.': 'Your GPS is outside Panama.',
    'No se pudo obtener tu GPS.': "Couldn't get your GPS.",
    'No se pudo obtener tu GPS. Intenta de nuevo.': "Couldn't get your GPS. Try again.",
    'GPS no disponible: permiso de ubicación denegado.': 'GPS unavailable: location permission denied.',
    'Permiso de ubicación denegado.': 'Location permission denied.',
    'Permiso de ubicación denegado. Actívalo en el navegador.': 'Location permission denied. Turn it on in your browser.',
    'Permiso de ubicación denegado. Actívalo en el navegador o pega un enlace de mapa.': 'Location permission denied. Turn it on in your browser or paste a map link.',
    'Este dispositivo no permite obtener la ubicación.': "This device can't get the location.",
    'Este dispositivo no permite obtener la ubicación. Pega un enlace de mapa.': "This device can't get the location. Paste a map link.",
    'Tu nombre y apellido': 'Your first and last name',
    'Teléfono de quien reporta': "Reporter's phone",
    'Ponerme como quien reporta': 'Set me as the reporter',
    'Quien más sabe de este punto': 'Whoever knows this place best',
    'Opcional: otras personas del grupo y su teléfono.': 'Optional: other people in the group and their phone.',
    '＋ Agregar otro contacto': '＋ Add another contact',
    'Agrega al menos un contacto con teléfono (7 dígitos o más).': 'Add at least one contact with a phone (7 digits or more).',
    'Al menos un contacto con teléfono es obligatorio.': 'At least one contact with a phone is required.',
    'Edad': 'Age',
    'Cédula': 'ID number',
    'Quitar': 'Remove',
    'Agua y comida': 'Water and food',
    'Refugio y cuidado': 'Shelter and care',
    'Energía y comunicación': 'Power and communication',
    'Casa': 'House',
    '¿Algo más?': 'Anything else?',
    'Otros': 'Other',
    'Otros: ¿qué más necesitan?': 'Other: what else do they need?',
    'Medicinas, ropa, colchones…': 'Medicine, clothes, mattresses…',
    'Marca al menos una necesidad o escribe en «Otros».': 'Choose at least one need or write in «Other».',
    'Notas adicionales': 'Additional notes',
    'Con tu nombre y teléfono también te creamos una cuenta al enviar, para que sigas tu reporte.': 'With your name and phone we also create an account for you when you send it, so you can follow your report.',
    'Revisa antes de guardar': 'Check before saving',
    'Guardar': 'Save',
    'Guardar punto': 'Save place',
    'Enviar necesidad': 'Send need',
    'Revisar': 'Review',
    '‹ Atrás': '‹ Back',
    'Punto publicado. Ya aparece en el mapa y en la lista.': 'Place published. It now appears on the map and in the list.',
    'Necesidad reportada. Ya está abierta para los corredores.': 'Need reported. It is now open for runners.',
    'Se ha enviado tu solicitud.': 'Your request has been sent.',
    'Sin conexión: el punto quedó guardado en este teléfono y se enviará solo cuando haya internet. Sus datos también se copiaron al portapapeles.': 'Offline: the place was saved on this phone and will be sent automatically when there is internet. Its details were also copied to the clipboard.',
    'Sin conexión': 'Offline',
    'Sin conexión. Intenta de nuevo cuando tengas señal.': 'Offline. Try again when you have signal.',
    'El servidor no respondió. Intenta de nuevo en un momento.': "The server didn't respond. Try again in a moment.",
    'Intentando enviar…': 'Trying to send…',
    'Enviando…': 'Sending…',
    'Puntos por enviar': 'Places to send',
    'Reintentar envío': 'Retry sending',
    'Reintentar': 'Retry',
    'Guardado en el teléfono.': 'Saved on the phone.',
    'No se pudo guardar.': "Couldn't save.",
    'No se pudo guardar:': "Couldn't save:",
    'Acepto que mi nombre y teléfono se compartan con los voluntarios y equipos que ayudan, para que me contacten y puedan llevar la ayuda.': 'I agree to share my name and phone with the volunteers and teams helping, so they can contact me and deliver the help.',
    'Cancelar y volver': 'Cancel and go back',

    // Necesidades (pointform.js NEEDS / BRANDED)
    'Agua': 'Water',
    'Alimentos': 'Food',
    'Refugio (carpa / lona / toldo)': 'Shelter (tent / tarp / canopy)',
    'Refugio': 'Shelter',
    'Kit de cocina': 'Kitchen kit',
    'Kit de bebé': 'Baby kit',
    'Combustible para generador': 'Generator fuel',
    'Tarjeta de datos prepago': 'Prepaid data card',
    'Gas de cocina': 'Cooking gas',
    'Reparación de casa': 'House repair',
    '¿Para qué?': 'What for?',
    'Para beber': 'For drinking',
    'Para cocinar': 'For cooking',
    'para beber': 'for drinking',
    'para cocinar': 'for cooking',
    '¿Qué tipo?': 'What kind?',
    'Comida preparada': 'Prepared food',
    'comida preparada': 'prepared food',
    'Ingredientes para cocinar': 'Ingredients to cook',
    'ingredientes': 'ingredients',
    'Combustible': 'Fuel',
    'Gasolina': 'Gasoline',
    'Diésel': 'Diesel',
    'Red': 'Network',
    'Empresa': 'Company',
    'Otro': 'Other',
    'Otra': 'Other',
    'Escribe cuál': 'Write which one',
    'Escribe la red': 'Write the network',
    'Escribe la empresa': 'Write the company',
    'toca una o varias': 'tap one or more',
    'Comida preparada para': 'Prepared food for',
    '¿Para cuántas personas?': 'For how many people?',
    '¿Qué comida?': 'What food?',
    'Ej. 5 niños pequeños, sin picante': 'E.g. 5 small children, not spicy',
    'Pedir': 'Request',
    'Esperando una cocina': 'Waiting for a kitchen',
    'Una cocina la está preparando': 'A kitchen is preparing it',

    // Avisos y otros
    'Atención': 'Attention',
    'Compartida solo con tu equipo.': 'Shared only with your team.',
    'Solo tu equipo': 'Only your team',
    'Solo visible en este dispositivo.': 'Only visible on this device.',
    'Verificado': 'Verified',
    'Verificado: ver qué equipo': 'Verified: see which team',
    'Verificado por otro equipo': 'Verified by another team',
    'Verificado por un equipo': 'Verified by a team',
    'Última revisión': 'Last check',
    'Daño reportado': 'Reported damage',
    'Foto de partida': 'Departure photo',
    'Galería': 'Gallery',
    'Tomar foto': 'Take photo',
    'Grabar video': 'Record video',
    'Para el punto': 'For the place',
    '. Queda abierta de inmediato para los corredores. Tu nombre y teléfono se muestran con la necesidad.': '. It is open for runners right away. Your name and phone are shown with the need.',
    'Zona / área': 'Area',
    'Ubicación / referencia': 'Location / landmark',
    'Hojas de zinc': 'Zinc sheets', 'Tubo de hierro': 'Iron pipe', 'Bolsas de cemento': 'Bags of cement', 'Carriola': 'Purlin', 'Varilla': 'Rebar', 'Bloques': 'Blocks', 'Tornillos': 'Screws', 'Clavos': 'Nails', 'Soldadura': 'Welding',
    'Materiales (opcional)': 'Materials (optional)', 'Escribe el material': 'Write the material',
    'Saneamiento / baño': 'Sanitation / toilet',
    'Tu ubicación en vivo la ve solo tu equipo': 'Only your team sees your live location',
    'Elegir…': 'Choose…',
    'Área': 'Area',
    'Necesita': 'Needs',
    'Inicio': 'Home',
    'Español': 'Español',
    'Todos': 'All',
    'Equipo': 'Team',
    'Nota': 'Note',
    'Urgencia': 'Urgency',
    'Para': 'For',
    'Conductor:': 'Driver:',
    'Lo que necesitan': 'What they need',
    'Por ubicar': 'To locate',
    // Palabras sueltas tras un número en negrita («<b>3</b> personas»)
    'personas': 'people', 'persona': 'person', 'adultos': 'adults', 'niños': 'children', 'bebés': 'babies', 'mayores': 'seniors', 'discapacidad': 'disability', 'familias': 'families',
  };

  // Textos con números o nombres: la parte que cambia queda igual.
  const WORDS = {personas: 'people', persona: 'person', adultos: 'adults', niños: 'children', bebés: 'babies', mayores: 'seniors', familias: 'families', puntos: 'places', punto: 'place'};
  const RULES = [
    [/^hace (\d+) min$/, '$1 min ago'], [/^hace (\d+) h$/, '$1 h ago'], [/^hace (\d+) d$/, '$1 d ago'], [/^hace (\d+) s$/, '$1 s ago'],
    [/^ahora$/i, 'now'],
    [/^No es el sitio real \((.+)\)\. Lo que hagas aquí no llega al sitio público y los datos se pueden borrar en cualquier momento\.$/, "This isn't the real site ($1). What you do here doesn't reach the public site and the data can be deleted at any time."],
    [/^(.+) sabe más$/, '$1 knows more'],
    [/^(\d+) (puntos|punto) por enviar$/, (m, n, w) => `${n} ${w === 'punto' ? 'place' : 'places'} to send`],
    [/^(Cédula|Teléfono|Nombre|Edad) de la persona (\d+)$/, (m, f, n) => `${({Cédula: 'ID number', Teléfono: 'Phone', Nombre: 'Name', Edad: 'Age'})[f]} of person ${n}`],
    [/^(Sumar|Restar) 1 a (.+)$/, (m, op, what) => `${op === 'Sumar' ? 'Add 1 to' : 'Subtract 1 from'} ${tr(what)}`],
    [/^Ej\. (.+)$/, 'E.g. $1'],
    [/^([\d.,]+) (personas|persona|adultos|niños|bebés|mayores|familias|puntos|punto)$/, (m, n, w) => `${n} ${WORDS[w]}`],
    [/^([\d.,]+) en mapa$/, '$1 on map'],
    [/^Otras (\d+) cosas$/, '$1 other things'],
    [/^([\d.,]+) nuevas en 24 h$/, '$1 new in 24 h'],
    [/^([\d.,]+) en 24 h · ([\d.,]+) en 7 días$/, '$1 in 24 h · $2 in 7 days'],
    [/^de ([\d.,]+) (puntos|punto)$/, (m, n, w) => `of ${n} ${WORDS[w]}`],
    [/^Revisa los (\d+) campos marcados\.$/, 'Check the $1 highlighted fields.'],
    [/^Precisión aprox\. ±(\d+) m$/, 'Approx. accuracy ±$1 m'],
    [/^¡Gracias, (.+)!$/, 'Thank you, $1!'],
    [/^Ningún punto coincide con «(.+)»\.$/, 'No place matches «$1».'],
    [/^Réplica · magnitud (.+)$/, 'Aftershock · magnitude $1'],
    [/^Las coordenadas (.+) quedan fuera de Panamá\. Revisa el enlace\.$/, 'The coordinates $1 are outside Panama. Check the link.'],
    [/^Ubicación (.+)\. Mapa no disponible sin conexión\.$/, 'Location $1. Map unavailable offline.'],
    [/^Última revisión: (.+)$/, 'Last check: $1'],
    [/^Buscando «(.+)»$/, 'Searching «$1»'],
    [/^Buscar \(buscando «(.+)»\)$/, 'Search (searching «$1»)'],
    [/^y (\d+) más: escribe algo más preciso\.$/, 'and $1 more: type something more specific.'],
    [/^Los runners llevan a cada punto lo que necesita\. (.+)$/, (m, rest) => 'Runners bring each place what it needs. ' + tr(rest)],
    [/^(.+) · Punto de apoyo · Azuero$/, '$1 · Support point · Azuero'],
    [/^PRUEBA · (.+)$/, (m, rest) => 'TEST · ' + tr(rest)],
    // «3 · 2 niños · 1 bebés» (personas en la ficha)
    [/^[\d.,]+( · [\d.,]+ (niños|bebés))+$/, m => m.replace(/(niños|bebés)/g, w => WORDS[w])],
    // Fechas «Oct 13, 2026 7:33pm»: los meses que no son iguales en inglés.
    [/^(Ene|Abr|Ago|Dic) (\d{1,2}, \d{4}.*)$/, (m, mo, rest) => ({Ene: 'Jan', Abr: 'Apr', Ago: 'Aug', Dic: 'Dec'})[mo] + ' ' + rest],
  ];

  // Frase exacta, regla, o la misma sin el ícono/emoji del inicio o la puntuación final («📍 Reportar…», «Tu código:»).
  function look(core, depth = 0) {
    if (!core) return undefined;
    if (Object.prototype.hasOwnProperty.call(EN, core)) return EN[core];
    for (const [re, rep] of RULES) if (re.test(core)) return core.replace(re, rep);
    if (depth > 2) return undefined;
    // «Agua: ¿Para qué?» o «Con necesidades (24)»: solo si cada parte es una frase conocida.
    const pair = core.match(/^([^:]+): (.+)$/);
    if (pair) { const x = look(pair[1], depth + 1), y = look(pair[2], depth + 1); if (x !== undefined && y !== undefined) return x + ': ' + y; }
    const count = core.match(/^(.+) \((\d+)\)$/);
    if (count) { const x = look(count[1], depth + 1); if (x !== undefined) return x + ' (' + count[2] + ')'; }
    const pre = core.match(/^([^\p{L}\p{N}¿¡«("]+)([\s\S]+)$/u);
    if (pre) { const t = look(pre[2], depth + 1); if (t !== undefined) return pre[1] + t; }
    const post = core.match(/^([\s\S]+?)([\s:…·,;]+)$/u);
    if (post) { const t = look(post[1], depth + 1); if (t !== undefined) return t + post[2]; }
    return undefined;
  }
  function tr(s) {
    const m = String(s).match(/^(\s*)([\s\S]*?)(\s*)$/);
    const t = look(m[2]);
    return t === undefined ? s : m[1] + t + m[3];
  }

  let lang = 'es', obs = null, title = null;
  const origText = new WeakMap(), wrote = new WeakMap(), origAttr = new WeakMap();
  const ATTRS = ['placeholder', 'title', 'aria-label', 'alt'];
  const SKIP = new Set(['SCRIPT', 'STYLE', 'TEXTAREA', 'NOSCRIPT', 'CODE', 'PRE', 'svg']);
  const skipped = el => SKIP.has(el.tagName) || el.isContentEditable || el.getAttribute('translate') === 'no';
  function doText(n) {
    const cur = n.data;
    let o = origText.get(n);
    if (o === undefined || wrote.get(n) !== cur) {
      if (lang === 'es') return;
      o = cur; origText.set(n, o);
    }
    const want = lang === 'en' ? tr(o) : o;
    if (want !== cur) n.data = want;
    wrote.set(n, want);
  }
  function doAttrs(el) {
    let saved = origAttr.get(el);
    const list = el.tagName === 'INPUT' && /^(button|submit)$/i.test(el.type) ? ATTRS.concat('value') : ATTRS;
    for (const a of list) {
      if (!el.hasAttribute(a)) continue;
      const cur = el.getAttribute(a);
      let rec = saved && saved[a];
      if (!rec || rec.w !== cur) {
        if (lang === 'es') continue;
        if (!saved) { saved = {}; origAttr.set(el, saved); }
        rec = saved[a] = {o: cur, w: cur};
      }
      const want = lang === 'en' ? tr(rec.o) : rec.o;
      if (want !== cur) el.setAttribute(a, want);
      rec.w = want;
    }
  }
  function walk(root) {
    if (!root) return;
    if (root.nodeType === 3) { if (root.parentElement && !root.parentElement.closest('script,style,textarea,[translate="no"],[contenteditable="true"]')) doText(root); return; }
    if (root.nodeType !== 1 || root.closest?.('[translate="no"],[contenteditable="true"]') || skipped(root)) return;
    doAttrs(root);
    const tw = document.createTreeWalker(root, NodeFilter.SHOW_ELEMENT | NodeFilter.SHOW_TEXT, {acceptNode: n => n.nodeType === 1 && skipped(n) ? NodeFilter.FILTER_REJECT : NodeFilter.FILTER_ACCEPT});
    for (let n = tw.nextNode(); n; n = tw.nextNode()) n.nodeType === 3 ? doText(n) : doAttrs(n);
  }
  function doTitle() {
    if (title === null || (lang === 'en' && document.title !== tr(title) && document.title !== title)) title = document.title;
    const want = lang === 'en' ? tr(title) : title;
    if (document.title !== want) document.title = want;
  }
  function onMutations(list) {
    for (const m of list) {
      if (m.type === 'childList') m.addedNodes.forEach(walk);
      else if (m.type === 'characterData') walk(m.target);
      else if (m.type === 'attributes') walk(m.target.nodeType === 1 ? m.target : null);
    }
    doTitle();
    obs.takeRecords(); // lo que se acaba de cambiar aquí no vuelve a entrar
  }
  // Avisos del navegador (alert / confirm / prompt) en el idioma elegido.
  for (const k of ['alert', 'confirm', 'prompt']) {
    const orig = window[k];
    if (typeof orig === 'function') window[k] = function (msg, ...rest) { return orig.call(window, lang === 'en' && msg != null ? tr(String(msg)) : msg, ...rest); };
  }
  function set(l) {
    lang = l === 'en' ? 'en' : 'es';
    document.documentElement.lang = lang;
    walk(document.body);
    doTitle();
    if (lang === 'en' && !obs) {
      obs = new MutationObserver(onMutations);
      obs.observe(document.documentElement, {subtree: true, childList: true, characterData: true, attributes: true, attributeFilter: ATTRS.concat('value')});
    } else if (lang === 'es' && obs) { obs.disconnect(); obs = null; }
  }
  window.I18n = {set, tr, lang: () => lang, EN};
})();
