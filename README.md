# Go RSS Torrent Downloader

This project is a lightweight, Go-based service containerized with Docker. It is designed to monitor a specified RSS feed, automatically download the `.torrent` files, and serve them via a new, self-generated RSS feed.

The service prioritizes secure and efficient operation by using dynamic user management (`PUID`/`PGID`) to ensure correct file ownership on host volumes.

## Key Features

  * **Automatic Download:** Monitors the `RSSURL` feed based on the defined `CRONTAB` schedule.

  * **Duplicate Management:** Only downloads `.torrent` files if they do not already exist in the target directory.

  * **Self-Generated RSS Feed:** Starts a web server on port `8080` that exposes a new RSS feed at the `/rss` endpoint, listing the currently downloaded `.torrent` files.

  * **Permission Handling (PUID/PGID):** Uses the `PUID` and `PGID` environment variables to run the application as a non-root user, ensuring proper file ownership on mounted host volumes.

  * **Time Zone Support:** Configures the container's internal time zone using the `TZ` environment variable.

  * **Automated Build:** Includes a `Makefile` to simplify dependency management and Docker image compilation.

## Prerequisites

To build and run this project, you will need:

  * [Docker](https://www.docker.com/get-started) (or Docker Desktop)

  * [GNU Make](https://www.gnu.org/software/make/) (Standard on most Linux/macOS systems)

## Project Build

The `Makefile` automates the entire setup process.

1.  Clone this repository.

2.  Run the following command from the project root directory:

<!-- end list -->

```bash
make
```

This command will:

1.  Check and update Go dependencies (`go mod tidy`).

2.  Build the Docker image, tagged as `go-rss-torrent-puid:latest`.

## Running the Container

To run the container, you must provide the mandatory environment variables (`RSSURL`, `CRONTAB`) and mount a host volume to the internal `TORRENT_DIR` path (`/torrent_files`) for file persistence.

### Example `docker run` Command

This example sets the container to check the feed every 10 minutes, uses the `Europe/Rome` timezone, and maps the local `./my_torrents` folder for downloads.

```bash
# Replace PUID/PGID with your host user and group IDs (use 'id -u' and 'id -g' on your host)
MY_PUID=$(id -u)
MY_PGID=$(id -g)

docker run -d \
  --name go-rss-downloader \
  -p 8080:8080 \
  -v "$(pwd)/my_torrents":/torrent_files \
  -e RSSURL="https://YOUR_RSS_FEED_URL.xml" \
  -e CRONTAB="*/10 * * * *" \
  -e PUID=$MY_PUID \
  -e PGID=$MY_PGID \
  -e TZ="Europe/Rome" \
  go-rss-torrent-puid:latest
```

### Accessing Services

  * **Downloaded Files:** Will appear in your local `./my_torrents` folder.

  * **Generated RSS Feed:** The new feed will be available at `http://localhost:8080/rss`.

## Configuration (Environment Variables)

| Variable | Description | Required | Default |
| :--- | :--- | :--- | :--- |
| `RSSURL` | The complete URL of the RSS feed to monitor. | **Yes** | (none) |
| `CRONTAB` | The cron string defining the download scheduling frequency. | **Yes** | (none) |
| `PUID` | The User ID used to execute the process inside the container. | No | `1000` |
| `PGID` | The Group ID used to execute the process inside the container. | No | `1000` |
| `TZ` | Sets the container's time zone (e.g., `America/New_York`). | No | `UTC` |
| `TORRENT_DIR` | Internal container path for saving the files. | No | `/torrent_files` |

## Maintenance

To clean up locally generated Go module files and the built Docker image, run:

```bash
make clean
```
