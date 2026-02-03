# MedunkaOPBarcode2.0

**Overview**
- End-to-end system for retail stations that scan barcodes and request product data over MQTT.
- Secure backend updates the product catalog over HTTPS and imports into PostgreSQL.
- Operator CLI converts legacy MDB files to CSV, uploads to the server, and waits for import completion.

---

## System Architecture

```
Station -> MQTT (TLS) -> mqtt_product_api -> PostgreSQL
   ^                                  |
   |                                  v
   +---- MQTT response <--------------+

CLI -> HTTPS (TLS) -> http_db_update_server -> PostgreSQL
  |                             ^
  +---- job status polling -----+
```

**Core services (compose)**
- `postgres_db`: PostgreSQL with TLS enabled; stores product catalog.
- `mosquitto_broker`: MQTT broker (TLS + username/password auth).
- `mqtt_product_api`: subscribes to product requests and publishes replies.
- `http_db_update_server`: accepts catalog uploads and runs async import jobs.
- `cli_control_app`: operator CLI to upload catalogs and send global station commands.

---

## Data Flows

### Product lookup (station -> MQTT -> database)
1. Station publishes a `ProductDataRequest` JSON to `MQTT_TOPIC_REQUEST`.
2. `mqtt_product_api` consumes the request, queries PostgreSQL by `barcode`, and publishes a `Product` response to `clientTopic` from the request.
3. Stations display product data locally.

Request payload (from stations):
```json
{
  "clientTopic": "stations/response/<station_id>",
  "barcode": "8595020340103",
  "includeDiacritics": true
}
```

Response payload (from server):
```json
{
  "name": "...",
  "price": "...",
  "stock": "...",
  "unitOfMeasure": "...",
  "unitOfMeasureCoef": "..."
}
```

### Catalog update (CLI -> HTTPS -> async import)
1. CLI converts MDB to CSV using `cmd/cli_control_app/mdb_to_csv.sh`.
2. CLI uploads the CSV to `http_db_update_server` over HTTPS.
3. Server returns `202 Accepted` with a `job_id`.
4. CLI polls the status endpoint until the job is `completed` or `failed`.
5. Server imports into PostgreSQL (drop + create table + COPY stream).

---

## MQTT Topics and Global Commands

**Lookup topic**
- `MQTT_TOPIC_REQUEST` (example: `product/request`)
- Stations publish lookup requests here.

**Global commands (CLI)**
The CLI publishes to:
```
<MQTT_BASE_TOPIC>/<sub-topic>/<state>
```
- Firmware: `<MQTT_BASE_TOPIC>/<MQTT_FIRMWARE_UPDATE_TOPIC>/enable` or `disable`
- Status: `<MQTT_BASE_TOPIC>/<MQTT_STATUS_TOPIC>/wake` or `sleep`

Payload is a simple string (`"1"`) with QoS 1 and optional retain flag.

---

## HTTP Endpoints (HTTPS only)

**Upload endpoint**
- `POST <HTTP_SERVER_UPDATE_ENDPOINT>`
- Accepts `multipart/form-data` with field name `file`
- Returns `202 Accepted` with JSON:
  ```json
  { "job_id": "<uuid>", "message": "Upload accepted" }
  ```

**Status endpoint**
- `GET <HTTP_SERVER_UPDATE_STATUS_ENDPOINT>?id=<job_id>`
- Returns JSON:
  ```json
  { "state": "pending|processing|completed|failed", "message": "..." }
  ```

**Important behavior**
- Job statuses live in memory (`sync.Map`). If `http_db_update_server` restarts, in-flight job data is lost.

---

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

---

## Scripts and Conversion

**`cmd/cli_control_app/mdb_to_csv.sh`**
- Uses `mdb-export` to dump the `SKz` table from an MDB file.
- Selects and normalizes required columns in a strict order.
- Outputs a semicolon-delimited CSV compatible with the CSV parser.
- The CLI container installs `mdbtools` and `mdbtools-utils` so `mdb-export` is available.

---

## TLS, Secrets, and `conf_gen.sh`

### What `conf_gen.sh` does
- Generates a CA and server TLS certificate.
- Creates Podman secrets for:
  - `server_key` (TLS private key)
  - `db_user`, `db_password`
  - `mqtt_user`, `mqtt_password`
  - `mqtt_passwd_db` (mosquitto password file)
- Generates `mosquitto/mosquitto.conf` with TLS + password auth.

### Where secrets are used
- `server_key`: mounted into `postgres_db`, `mosquitto_broker`, and `http_db_update_server` as the TLS private key. `mqtt_product_api` and `cli_control_app` load `TLS_CERT_PATH` + `TLS_KEY_PATH` as the MQTT client certificate/key.
- `db_user` / `db_password`: used by PostgreSQL and by Go services through `POSTGRES_USER_FILE` and `POSTGRES_PASSWORD_FILE`.
- `mqtt_user` / `mqtt_password`: used by MQTT clients via `MQTT_USER_FILE` / `MQTT_PASSWORD_FILE`.
- `mqtt_passwd_db`: mosquitto password file used by the broker.

### Critical caveat: IP address in certs
`conf_gen.sh` requires a server IP argument:
```
./conf_gen.sh <server_ip>
```
That IP is baked into the TLS certificate `subjectAltName`. If it does not match the IP used by clients (`MQTT_HOST_IP` and the overridden `HTTP_SERVER_HOST` for CLI), TLS will fail with certificate errors. Update the IP and re-run the script any time the server IP changes.

---

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
- `POSTGRES_HOST`: `postgres_db` (internal container DNS)
- `POSTGRES_PORT`: `5432`
- `POSTGRES_DB`: `products_db`
- `POSTGRES_SSLMODE`: `verify-full`
- `DB_TABLE_NAME`: `products`
- `DB_MAX_OPEN_CONNS`: `10`
- `DB_MAX_IDLE_CONNS`: `10`
- `DB_CONN_MAX_LIFETIME`: `10m`
- `POSTGRES_USER_FILE`: `/run/secrets/db_user`
- `POSTGRES_PASSWORD_FILE`: `/run/secrets/db_password`

### TLS
- `TLS_CA_PATH`: `/app/certs/ca.crt`
- `TLS_CERT_PATH`: `/app/certs/server.crt`
- `TLS_KEY_PATH`: `/run/secrets/server_key`

### MQTT
- `MQTT_PROTOCOL`: `tcps`
- `MQTT_HOST`: `mosquitto_broker` (internal container DNS)
- `MQTT_PORT`: `8883`
- `MQTT_API_CLIENT_ID`: `mqttDatabaseAPI` (used by `mqtt_product_api`)
- `MQTT_TOPIC_REQUEST`: `product/request`
- `MQTT_BASE_TOPIC`: `stations`
- `MQTT_FIRMWARE_UPDATE_TOPIC`: `firmware_update`
- `MQTT_STATUS_TOPIC`: `status`
- `MQTT_USER_FILE`: `/run/secrets/mqtt_user`
- `MQTT_PASSWORD_FILE`: `/run/secrets/mqtt_password`

### HTTP update server
- `HTTP_SERVER_HOST`: `0.0.0.0` (server bind address inside container)
- `HTTP_SERVER_PORT`: `8443`
- `HTTP_SERVER_UPDATE_ENDPOINT`: `/update-db`
- `HTTP_SERVER_UPDATE_STATUS_ENDPOINT`: `/update-status`
- `HTTP_MAX_UPLOAD_SIZE`: `5368709120`

### mqtt_product_api worker tuning
- `APP_WORKER_COUNT`: `3` (goroutines processing MQTT requests)
- `APP_JOB_QUEUE_SIZE`: `100` (buffer size for incoming MQTT jobs)
- `APP_DB_TIMEOUT`: `2s` (per-request DB timeout)

### CLI-specific
- `MDB_FILEPATH`: `/data/67668305_2025.mdb` (path inside CLI container)
- `MQTT_CLIAPP_CLIENT_ID`: `cli_admin_console`
- `MQTT_HOST_IP`: `192.168.1.19` (IP for CLI to reach MQTT API)
- `HTTP_SERVER_IP`: `192.168.1.19` (IP for CLI to reach HTTP server)
- `HOST_MDB_DIR`: `/home/user/Projects/MedunkaOPBarcode2.0`
- `CONTAINER_MDB_DIR`: `/data`
- `MDB_FILENAME`: `67668305_2025.mdb`

**Note:** any IP fields used by the CLI must match the IP passed to `conf_gen.sh`.

---

## Build and Run (Podman)

### 1) Generate TLS + secrets
Edit `conf_gen.sh` and set:
- `DB_USER_VAL`, `DB_PASS_VAL`
- `MQTT_USER_VAL`, `MQTT_PASS_VAL`

Then run:
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

---

## CLI Commands (quick reference)
- `ls` list commands
- `dbu` upload catalog and wait for import completion
- `sfwe` / `sfwd` enable or disable firmware update mode
- `sw` / `ss` mark stations wake/sleep (retained)
- `e` exit

---

## Expectations and Known Behavior
- HTTPS is enforced for uploads and status polling; certificates must match the server IP.
- Job status is in-memory; restarting `http_db_update_server` loses job history.
- The catalog import replaces the entire products table on each run.
- MQTT requests are processed in a worker pool; when the queue is full, requests are dropped with a warning.
