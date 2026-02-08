# MedunkaOPBarcode2.0

## Overview
- End-to-end system for retail stations that scan barcodes and request product data over MQTT.
- Secure backend updates the product catalog over HTTPS (mTLS) and imports into PostgreSQL.
- Operator CLI converts legacy MDB files to CSV, uploads to the server, and waits for import completion.

## System Architecture

```
Station -> MQTT (TLS) -> mqtt_product_api -> PostgreSQL
   ^                                  |
   |                                  v
   +---- MQTT response <--------------+

CLI -> HTTPS (mTLS) -> http_db_update_server -> PostgreSQL
  |                               ^
  +---- job status polling -------+
```

**Core services (compose)**
- `postgres_db`: PostgreSQL with TLS and client cert auth (`pg_hba.conf`).
- `mosquitto_broker`: MQTT broker with mTLS client cert auth.
- `mqtt_product_api`: subscribes to product requests and publishes replies.
- `http_db_update_server`: accepts catalog uploads and runs async import jobs.
- `cli_control_app`: operator CLI to upload catalogs and send global station commands.

## Data Flows

### Product lookup (station -> MQTT -> database)
1. Station publishes an empty payload to `product/<station_mac>/<barcode>`.
2. `mqtt_product_api` consumes the request, queries PostgreSQL by `barcode`, and publishes a response to `product/<station_mac>`.
3. Stations display product data locally.

Request topic example:
```
product/246f2880a1b2/8595020340103
```

Response payload:
```json
{
  "name": "...",
  "barcode": "...",
  "price": "...",
  "stock": "...",
  "unitOfMeasure": "...",
  "unitOfMeasureCoef": "...",
  "valid": true
}
```

**Notes**
- If a product is missing, the response sets `valid=false` and echoes the requested `barcode`.
- `mqtt_product_api` normalizes diacritics on `name` and `price` before responding.
- Requests are processed by a worker pool; when the queue is full, new requests are dropped with a warning.

### Catalog update (CLI -> HTTPS -> async import)
1. CLI converts MDB to CSV using `cmd/cli_control_app/mdb_to_csv.sh`.
2. CLI uploads the CSV to `http_db_update_server` over HTTPS.
3. Server returns `202 Accepted` with a `job_id`.
4. CLI polls the status endpoint until the job is `completed` or `failed`.
5. Server imports into PostgreSQL (drop + create table + COPY stream).

## MQTT Topics and Global Commands

**Lookup topic**
- `MQTT_TOPIC_REQUEST` (default: `product/+/+`)
- Stations publish lookup requests as `product/<station_mac>/<barcode>`.

**Global commands (CLI)**
The CLI publishes directly to the topic configured in env:

| Command | Topic | Payload | Retained |
| --- | --- | --- | --- |
| `sfwe` | `MQTT_FIRMWARE_UPDATE_TOPIC` (default `station/firmware_update`) | `start` | `false` |
| `sfwd` | `MQTT_FIRMWARE_UPDATE_TOPIC` | `stop` | `false` |
| `sw` | `MQTT_STATUS_TOPIC` (default `station/status`) | `wake` | `true` |
| `ss` | `MQTT_STATUS_TOPIC` | `sleep` | `true` |

QoS is always 1.

## HTTP Endpoints (HTTPS + mTLS)

**Upload endpoint**
- `POST <HTTP_SERVER_UPDATE_ENDPOINT>` (default: `/update_db`)
- Accepts `multipart/form-data` with field name `file`
- Returns `202 Accepted` with JSON:
  ```json
  { "job_id": "<uuid>", "message": "Upload accepted" }
  ```

**Status endpoint**
- `GET <HTTP_SERVER_UPDATE_STATUS_ENDPOINT>?id=<job_id>` (default: `/update_status`)
- Returns JSON:
  ```json
  { "state": "pending|processing|completed|failed", "message": "..." }
  ```

**Important behavior**
- Job statuses live in memory (`sync.Map`). If `http_db_update_server` restarts, in-flight job data is lost.

## Parser and Database Layers

**Parser (`internal/parser`)**
- Interface-driven design: `CatalogParser` streams records from a source.
- Factory (`parser.New`) chooses a parser based on file extension.
- Current implementation: CSV parser only. Adding a new parser means extending the factory.

**Database (`internal/database`)**
- Interface-driven: `Handler` defines `Fetch` and `ImportCatalog`.
- Current implementation: PostgreSQL.
- Import behavior: drop existing table, recreate, then COPY stream into the table.

**How parser and DB align**
- `mdb_to_csv.sh` outputs six columns in this order:
  `Nazev, EAN, ProdejDPH, MJ2, MJ2Koef, StavZ`
- `internal/parser/csv` maps them to:
  `Name, Barcode, Price, UnitOfMeasure, UnitOfMeasureCoef, Stock`
- `internal/database/postgres` writes them into the table columns:
  `name, barcode, price, unit_of_measure, unit_of_measure_koef, stock`

## Scripts and Conversion

**`cmd/cli_control_app/mdb_to_csv.sh`**
- Uses `mdb-export` to dump the `SKz` table from an MDB file.
- Selects and normalizes required columns in a strict order.
- Outputs a semicolon-delimited CSV compatible with the CSV parser.
- The CLI container installs `mdbtools` and `mdbtools-utils` so `mdb-export` is available.

## TLS, Secrets, and `conf_gen.sh`

### What `conf_gen.sh` does
- Generates a CA, server certificate, and client certificate.
- Creates Podman secrets for:
  - `server_key` (TLS private key)
  - `client_key` (client TLS private key)
  - `db_user` and `db_password`
- Generates `mosquitto/mosquitto.conf` with mTLS client cert auth.

### Where secrets are used
- `server_key`: mounted into `postgres_db`, `mosquitto_broker`, and `http_db_update_server` as the TLS private key.
- `client_key`: mounted into `http_db_update_server`, `mqtt_product_api`, and `cli_control_app` for mTLS client auth.
- `db_user`/`db_password`: used by PostgreSQL and read by services via secret files.

### Critical caveat: IP address in certs
`conf_gen.sh` requires a server IP argument:
```
./conf_gen.sh <server_ip>
```
That IP is baked into the TLS certificate `subjectAltName`. If it does not match the IP used by clients (`MQTT_HOST_IP` and `HTTP_SERVER_IP`), TLS will fail with certificate errors. Update the IP and re-run the script any time the server IP changes.

## Configuration (`.env`)

Below is a practical guide to each field. Values shown are examples.

### Build and image config
- `GO_IMAGE`: `docker.io/library/golang:1.25.6-alpine3.23`
- `ALPINE_IMAGE`: `docker.io/library/alpine:3.23`
- `POSTGRES_IMAGE`: `docker.io/library/postgres:18.1-trixie`
- `MOSQUITTO_IMAGE`: `docker.io/eclipse-mosquitto:2.1.0-alpine`
- `TARGETARCH`: `amd64`
- `WORKDIR`: `/app` (must match container paths and volume mounts)

### PostgreSQL
- `POSTGRES_USER_FILE`: `/run/secrets/db_user`
- `POSTGRES_PASSWORD_FILE`: `/run/secrets/db_password`
- `POSTGRES_DB`: `products_db`
- `POSTGRES_HOST`: `postgres_db` (internal container DNS)
- `POSTGRES_PORT`: `5432`
- `DB_TABLE_NAME`: `products`
- `DB_MAX_OPEN_CONNS`: `10`
- `DB_MAX_IDLE_CONNS`: `10`
- `DB_CONN_MAX_LIFETIME`: `10m`

### TLS
- `TLS_CA_PATH`: `/app/certs/ca.crt`
- `TLS_CLIENT_CERT_PATH`: `/app/certs/client.crt`
- `TLS_CLIENT_KEY_PATH`: `/run/secrets/client_key`
- `TLS_SERVER_CERT_PATH`: `/app/certs/server.crt`
- `TLS_SERVER_KEY_PATH`: `/run/secrets/server_key`

### MQTT
- `MQTT_PROTOCOL`: `tcps`
- `MQTT_HOST`: `mosquitto_broker` (internal container DNS)
- `MQTT_PORT`: `8883`
- `MQTT_API_CLIENT_ID`: `mqtt_database_api` (used by `mqtt_product_api`)
- `MQTT_TOPIC_REQUEST`: `product/+/+`
- `MQTT_FIRMWARE_UPDATE_TOPIC`: `station/firmware_update`
- `MQTT_STATUS_TOPIC`: `station/status`
- `MQTT_CLIAPP_CLIENT_ID`: `cli_admin_console`

### HTTP update server
- `HTTP_SERVER_HOST`: `0.0.0.0` (server bind address inside container)
- `HTTP_SERVER_PORT`: `8443`
- `HTTP_SERVER_UPDATE_ENDPOINT`: `/update_db`
- `HTTP_SERVER_UPDATE_STATUS_ENDPOINT`: `/update_status`
- `HTTP_MAX_UPLOAD_SIZE`: `5368709120`

### mqtt_product_api worker tuning
- `APP_WORKER_COUNT`: `3` (goroutines processing MQTT requests)
- `APP_JOB_QUEUE_SIZE`: `100` (buffer size for incoming MQTT jobs)
- `APP_DB_TIMEOUT`: `2s` (per-request DB timeout)

### CLI-specific
- `MQTT_HOST_IP`: `192.168.1.19` (IP for CLI to reach MQTT API)
- `HTTP_SERVER_IP`: `192.168.1.19` (IP for CLI to reach HTTP server)
- `HOST_MDB_DIR`: `/home/adamhoof/Projects/MedunkaOPBarcode2.0`
- `CONTAINER_MDB_DIR`: `/data`
- `MDB_FILENAME`: `67668305_2025.mdb`
- `MDB_FILEPATH`: `/data/67668305_2025.mdb` (built from `CONTAINER_MDB_DIR` + `MDB_FILENAME`)

**Note:** any IP fields used by the CLI must match the IP passed to `conf_gen.sh`.

## Build and Run (Podman)

### 1) Generate TLS + secrets
Run:
```
./conf_gen.sh <server_ip>
```

### 2) Build all services (including CLI)
```
podman compose build
```

### 3) Start backend services
```
podman compose up -d postgres_db mosquitto_broker http_db_update_server mqtt_product_api
```

### 4) Run the CLI
```
podman compose run --rm cli_control_app
```

## CLI Commands (quick reference)
- `ls` list commands
- `upd` convert MDB to CSV and upload to server
- `sfwe` / `sfwd` enable or disable firmware update mode
- `sw` / `ss` mark stations wake/sleep (retained)
- `e` exit

## Expectations and Known Behavior
- HTTPS is enforced for uploads and status polling; certificates must match the server IP.
- Job status is in-memory; restarting `http_db_update_server` loses job history.
- The catalog import replaces the entire products table on each run.
- MQTT requests are processed in a worker pool; when the queue is full, requests are dropped with a warning.
