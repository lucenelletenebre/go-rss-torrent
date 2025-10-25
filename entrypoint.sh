#!/bin/sh

# Imposta valori di default se non specificati
PUID=${PUID:-1000}
PGID=${PGID:-1000}

echo "Starting with PUID: $PUID and PGID: $PGID"

# 1. Crea il gruppo e l'utente
# -g $PGID: imposta l'ID del gruppo
# -u $PUID: imposta l'ID utente
# -s /bin/bash: shell non interattiva
# -D: non creare la directory home (l'utente è 'appuser')
addgroup -g "$PGID" appgroup 2>/dev/null
adduser -u "$PUID" -G appgroup -s /bin/bash -D appuser 2>/dev/null

# 2. Imposta i permessi sulla directory dei torrent
# Rendi appuser proprietario di /torrent_files e cambia i permessi
chown -R "$PUID":"$PGID" /torrent_files
chmod -R 775 /torrent_files

# 3. Gestione della Timezone (TZ)
# Se TZ è impostata, imposta la timezone del sistema. Alpine usa /usr/share/zoneinfo
if [ -n "$TZ" ]; then
    echo "Setting Timezone to $TZ"
    cp /usr/share/zoneinfo/"$TZ" /etc/localtime
    echo "$TZ" > /etc/timezone
fi

# 4. Esegui l'applicazione come l'utente specificato
# switchuser (su-exec) è essenziale per eseguire il binario Go come 'appuser'
# senza perdere le variabili d'ambiente.
exec su-exec appuser /torrent-downloader
