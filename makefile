# Nome e tag dell'immagine Docker
DOCKER_IMAGE_NAME = go-rss-torrent
DOCKER_TAG = latest

# Modulo Go predefinito (usato se go.mod viene creato da zero)
GO_MODULE_PATH = github.com/my-project/torrent-rss

# Immagine builder Go
GO_BUILDER_IMAGE = golang:1.22-alpine

.PHONY: all build-docker clean

# Target principale: compila i moduli Go se necessario, poi compila Docker.
all: build-docker

# ------------------------------------------------------------------------------
# FASE 1: GESTIONE MODULI GO (go.mod & go.sum)
# ------------------------------------------------------------------------------

# Il target 'go.mod' viene eseguito solo se il file non esiste.
go.mod:
	@echo "go.mod non trovato. Inizializzo il modulo Go..."
	docker run --rm \
		-v "$(shell pwd):/app" \
		-w /app \
		$(GO_BUILDER_IMAGE) \
		go mod init $(GO_MODULE_PATH)

# Il target 'go.sum' dipende da 'go.mod' e 'main.go'.
# Sarà eseguito se go.sum non esiste, go.mod è più recente o main.go è più recente.
go.sum: go.mod main.go
	@echo "Aggiorno go.mod e genero go.sum basandomi su main.go..."
	# Questo comando legge main.go, aggiorna go.mod con le dipendenze mancanti e genera go.sum
	docker run --rm \
		-v "$(shell pwd):/app" \
		-w /app \
		$(GO_BUILDER_IMAGE) \
		go mod tidy

# ------------------------------------------------------------------------------
# FASE 2: BUILD DOCKER
# ------------------------------------------------------------------------------

# Il target 'build-docker' dipende da 'go.sum', garantendo che i moduli siano aggiornati.
build-docker: go.sum Dockerfile
	@echo "--- Costruisco l'immagine Docker: $(DOCKER_IMAGE_NAME):$(DOCKER_TAG) ---"
	docker build \
		-t $(DOCKER_IMAGE_NAME):$(DOCKER_TAG) \
		.
	@echo "Build completata. Immagine: $(DOCKER_IMAGE_NAME):$(DOCKER_TAG)"

# ------------------------------------------------------------------------------
# FASE 3: test
# ------------------------------------------------------------------------------
.PHONY: test
test: build-docker
	# Avvia il container
	docker run -it --rm\
		--name "rss-test-temp" \
		-p 8080:8080 \
		-e RSSURL="https://nyaa.si/?page=rss&q=%5BErai-raws%5D+Ranma+1%2F2+%282024%29+2nd+Season+%2B1080p+%2Bhevc&c=0_0&f=0" \
		-e CRONTAB="* * * * *" \
		-e TZ="Europe/Rome" \
		-e PUID=$$(id -u) \
		-e PGID=$$(id -g) \
		-v $$(pwd)/test_files:/torrent_files \
		$(DOCKER_IMAGE_NAME)

# ------------------------------------------------------------------------------
# PULIZIA
# ------------------------------------------------------------------------------

clean:
	@echo "Pulizia: Rimuovo i file Go generati (go.mod, go.sum) e l'immagine Docker"
	@rm -f go.mod go.sum
	-docker rmi $(DOCKER_IMAGE_NAME):$(DOCKER_TAG) || true