#!/bin/bash
set -e


DB_USER_VAL="fill"
DB_PASS_VAL="fill" # needed for init

if [[ "$DB_USER_VAL" == "fill" || "$DB_PASS_VAL" == "fill" ]]; then
    echo "Error: Please update the DB_USER_VAL and DB_PASS_VAL in the script."
    exit 1
fi

if [ -z "$1" ]; then
    echo "usage: ./conf_gen.sh <server_ip>"
    exit 1
fi

SERVER_IP="$1"
ORG_NAME="MOB2"
ORG_NAME_LOWER=$(echo "$ORG_NAME" | tr '[:upper:]' '[:lower:]')

MOSQUITTO_CONF_DIR="mosquitto"
CERTS_DIR="certs"

mkdir -p $CERTS_DIR
mkdir -p $MOSQUITTO_CONF_DIR

CA_KEY="$CERTS_DIR/ca.key"
CA_CRT="$CERTS_DIR/ca.crt"
SERVER_KEY="$CERTS_DIR/server.key"
SERVER_CSR="$CERTS_DIR/server.csr"
SERVER_CRT="$CERTS_DIR/server.crt"
CLIENT_KEY="$CERTS_DIR/client.key"
CLIENT_CSR="$CERTS_DIR/client.csr"
CLIENT_CRT="$CERTS_DIR/client.crt"
EXT_FILE="$CERTS_DIR/v3.ext"
MOSQUITTO_CONF="$MOSQUITTO_CONF_DIR/mosquitto.conf"

SECURE_VAULT_VOLUME="secure_vault"

echo "gen prime256v1 certificates..."

# CA
openssl req -x509 -new -newkey ec -pkeyopt ec_paramgen_curve:prime256v1 -nodes \
  -days 3650 -sha256 \
  -keyout "$CA_KEY" -out "$CA_CRT" \
  -subj "/O=$ORG_NAME/CN=$ORG_NAME-Root-CA"

# server key and CSR
openssl req -new -newkey ec -pkeyopt ec_paramgen_curve:prime256v1 -nodes \
  -keyout "$SERVER_KEY" -out "$SERVER_CSR" \
  -subj "/O=$ORG_NAME/CN=$ORG_NAME_LOWER-server"

# client key and CSR
openssl req -new -newkey ec -pkeyopt ec_paramgen_curve:prime256v1 -nodes \
  -keyout "$CLIENT_KEY" -out "$CLIENT_CSR" \
  -subj "/O=$ORG_NAME/CN=client"

cat > "$EXT_FILE" << EOF
authorityKeyIdentifier=keyid,issuer
basicConstraints=CA:FALSE
keyUsage = digitalSignature, nonRepudiation, keyEncipherment, dataEncipherment
subjectAltName = @alt_names

[alt_names]
DNS.1 = localhost
DNS.2 = postgres_db
DNS.3 = mosquitto_broker
DNS.4 = http_db_update_server
IP.1 = 127.0.0.1
IP.2 = $SERVER_IP
EOF

# sign server cert
openssl x509 -req -in "$SERVER_CSR" -CA "$CA_CRT" -CAkey "$CA_KEY" -CAcreateserial \
  -out "$SERVER_CRT" -days 365 -sha256 -extfile "$EXT_FILE"

# sign client cert
openssl x509 -req -in "$CLIENT_CSR" -CA "$CA_CRT" -CAkey "$CA_KEY" -CAcreateserial \
  -out "$CLIENT_CRT" -days 365 -sha256

echo "populating secure_vault volume..."

podman volume create "$SECURE_VAULT_VOLUME" >/dev/null

podman run --rm \
  -v "$SECURE_VAULT_VOLUME":/vault \
  -v "$(pwd)/$CERTS_DIR":/source:ro \
  alpine:latest \
  /bin/sh -c "set -e; \
    cp /source/server.key /vault/server_key; \
    cp /source/client.key /vault/client_key; \
    printf '%s' \"$DB_USER_VAL\" > /vault/db_user; \
    printf '%s' \"$DB_PASS_VAL\" > /vault/db_password; \
    chown -R 1000:1000 /vault; \
    chmod -R 0400 /vault/*"

echo "gen mosquitto.conf..."
cat > "$MOSQUITTO_CONF" << EOF
listener 8883

# TLS conf
certfile /mosquitto/$CERTS_DIR/server.crt
keyfile /run/secrets/server_key
cafile /mosquitto/$CERTS_DIR/ca.crt

# Client auth
require_certificate true
use_identity_as_username true
EOF

# cleanup
rm "$SERVER_CSR" "$CLIENT_CSR" "$EXT_FILE"

echo "done."