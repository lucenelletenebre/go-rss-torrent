# --- FASE 1: BUILD ---
FROM golang:1.22-alpine AS builder

# Imposta la directory di lavoro all'interno del container
WORKDIR /app

# Copia i file di dipendenza e scarica i moduli
COPY go.mod go.sum ./
RUN go mod download

# Copia il codice sorgente
COPY main.go .

# Compila l'applicazione Go (binario statico)
RUN CGO_ENABLED=0 go build -o /torrent-downloader main.go

# --- FASE 2: RUNTIME (Immagine finale leggera) ---
FROM alpine:latest

# Installa pacchetti necessari:
# ca-certificates: per HTTPS
# tzdata: per la configurazione del fuso orario (TZ)
# su-exec: per eseguire il processo come utente non root (PUID/PGID)
# shadow: per i comandi adduser/addgroup
RUN apk --no-cache add ca-certificates tzdata su-exec shadow

# Imposta la directory di lavoro
# WORKDIR /root/

# Crea la directory dove verranno salvati i file torrent
ENV TORRENT_DIR=/torrent_files
RUN mkdir -p ${TORRENT_DIR}

# Copia il binario compilato dalla fase di build
COPY --from=builder /torrent-downloader .

# Copia lo script di ingresso e rendilo eseguibile
COPY entrypoint.sh /entrypoint.sh
RUN chmod +x /entrypoint.sh

# Espone la porta del web server
EXPOSE 8080

# Definisce il punto di ingresso
ENTRYPOINT ["/entrypoint.sh"]

# Il CMD è l'argomento passato all'ENTRYPOINT
CMD [""]
