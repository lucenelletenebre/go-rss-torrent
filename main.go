package main

import (
	"fmt"
	"io"
	"log"
	"net/http"
	"net/url"
	"os"
	"path/filepath"
	"strings"
	"time"

	"github.com/mmcdole/gofeed"
	"github.com/robfig/cron/v3"
)

const (
	// Definisci la directory per i file torrent
	TorrentDir = "/torrent_files"
	// Porta su cui è in ascolto il web server
	ServerPort = "8080"
)

// Funzione principale del cron job: scarica i torrent dal feed RSS
func downloadTorrents(rssURL string) {
	log.Printf("Avvio del download dei torrent da %s...", rssURL)

	fp := gofeed.NewParser()
	feed, err := fp.ParseURL(rssURL)
	if err != nil {
		log.Printf("ERRORE: Impossibile parsare il feed RSS: %v", err)
		return
	}

	for _, item := range feed.Items {
		// La logica per trovare il link al .torrent dipende dal feed.
		// Assumiamo che l'URL del torrent sia nel campo Link.
		torrentURL := item.Link
		if !strings.HasSuffix(strings.ToLower(torrentURL), ".torrent") {
			// Potrebbe essere necessario un parsing più complesso o un link alternativo
			// Se il link principale non è il .torrent
			log.Printf("SKIP: Il link '%s' non sembra essere un file .torrent per '%s'", torrentURL, item.Title)
			continue
		}

		// Estrai il nome del file dal link
		u, err := url.Parse(torrentURL)
		if err != nil {
			log.Printf("ERRORE: Impossibile parsare l'URL del torrent '%s': %v", torrentURL, err)
			continue
		}

		// Usa il percorso URL o il titolo per il nome del file
		fileName := filepath.Base(u.Path)
		if fileName == "." || fileName == "/" || fileName == "" {
			fileName = strings.ReplaceAll(item.Title, " ", "_") + ".torrent"
		}

		filePath := filepath.Join(TorrentDir, fileName)

		// Verifica se il file esiste già
		if _, err := os.Stat(filePath); err == nil {
			log.Printf("SKIP: Il file '%s' esiste già.", fileName)
			continue
		} else if !os.IsNotExist(err) {
			log.Printf("ERRORE: Errore di accesso al file '%s': %v", filePath, err)
			continue
		}

		// Scarica il file torrent
		log.Printf("DOWNLOAD: Scarico '%s' da '%s'...", fileName, torrentURL)

		resp, err := http.Get(torrentURL)
		if err != nil {
			log.Printf("ERRORE: Impossibile scaricare '%s': %v", torrentURL, err)
			continue
		}
		defer resp.Body.Close()

		if resp.StatusCode != http.StatusOK {
			log.Printf("ERRORE: Stato HTTP non OK (%d) per '%s'.", resp.StatusCode, torrentURL)
			continue
		}

		outFile, err := os.Create(filePath)
		if err != nil {
			log.Printf("ERRORE: Impossibile creare il file '%s': %v", filePath, err)
			continue
		}
		defer outFile.Close()

		_, err = io.Copy(outFile, resp.Body)
		if err != nil {
			log.Printf("ERRORE: Impossibile scrivere il file '%s': %v", filePath, err)
			// Rimuovi il file parziale in caso di errore
			os.Remove(filePath)
			continue
		}

		log.Printf("SUCCESSO: '%s' scaricato e salvato.", fileName)
	}

	log.Println("Download dei torrent completato.")
}

// Handler per l'endpoint /rss: genera un nuovo feed RSS
func generateRssFeed(w http.ResponseWriter, r *http.Request) {
	log.Println("Richiesta per la generazione del feed RSS.")

	// Imposta gli header per un feed RSS
	w.Header().Set("Content-Type", "application/rss+xml")
	w.WriteHeader(http.StatusOK)

	// Inizio del documento RSS (sostituisci con informazioni pertinenti)
	rssHeader := `<?xml version="1.0" encoding="UTF-8"?>
<rss version="2.0">
<channel>
<title>Torrent Files Feed</title>
<link>http://` + r.Host + `/rss</link>
<description>Torrent files saved in the container.</description>
<lastBuildDate>` + time.Now().Format(time.RFC1123Z) + `</lastBuildDate>
`
	fmt.Fprint(w, rssHeader)

	// Scansiona la directory dei torrent e crea un elemento per ogni file
	files, err := os.ReadDir(TorrentDir)
	if err != nil {
		log.Printf("ERRORE: Impossibile leggere la directory dei torrent: %v", err)
		// Gestione errore (opzionale, per semplicità qui si salta l'output dei file)
	}

	for _, file := range files {
		if file.IsDir() || !strings.HasSuffix(strings.ToLower(file.Name()), ".torrent") {
			continue // Salta le directory o i file non .torrent
		}

		// L'URL per scaricare il torrent. Supponiamo che il web server lo serva tramite /files/nome_file
		// NOTA: Devi aggiungere un handler per /files/nome_file, vedi nota sotto.
		torrentDownloadURL := "http://" + r.Host + "/files/" + file.Name()

		item := fmt.Sprintf(`
<item>
<title>%s</title>
<link>%s</link>
<guid>%s</guid>
<pubDate>%s</pubDate>
<description>Downloaded torrent file: %s</description>
</item>
`, file.Name(), torrentDownloadURL, file.Name(), time.Now().Format(time.RFC1123Z), file.Name())

		fmt.Fprint(w, item)
	}

	// Chiusura del documento RSS
	fmt.Fprint(w, `
</channel>
</rss>`)
}

func main() {
	// 1. Inizializzazione
	rssURL := os.Getenv("RSSURL")
	cronTab := os.Getenv("CRONTAB")

	if rssURL == "" || cronTab == "" {
		log.Fatal("Le variabili d'ambiente RSSURL e CRONTAB devono essere impostate.")
	}

	// Crea la directory dei torrent se non esiste
	if err := os.MkdirAll(TorrentDir, 0755); err != nil {
		log.Fatalf("Impossibile creare la directory dei torrent: %v", err)
	}

	// 2. Avvio dello Scheduler Cron Go
	c := cron.New()

	// L'uso di una closure permette di passare la variabile rssURL al job
	_, err := c.AddFunc(cronTab, func() {
		downloadTorrents(rssURL)
	})
	if err != nil {
		log.Fatalf("ERRORE: Impossibile schedulare il job con CRONTAB '%s': %v", cronTab, err)
	}

	// Esegui il primo job immediatamente per test e popolamento iniziale
	go downloadTorrents(rssURL)

	c.Start()
	log.Printf("Scheduler avviato con CRONTAB: '%s'", cronTab)

	// 3. Configurazione del Web Server
	router := http.NewServeMux()

	// Endpoint per il feed RSS
	router.HandleFunc("/rss", generateRssFeed)

	// Endpoint per servire i file torrent
	// NOTA: http.StripPrefix è fondamentale per servire i file dalla directory interna
	router.Handle("/files/", http.StripPrefix("/files/", http.FileServer(http.Dir(TorrentDir))))

	log.Printf("Web server avviato su http://0.0.0.0:%s", ServerPort)

	// Il server rimarrà in esecuzione e bloccherà main(), mantenendo attivo anche il cron
	err = http.ListenAndServe(":"+ServerPort, router)
	if err != nil {
		log.Fatalf("Web server terminato con errore: %v", err)
	}
}
