#!/bin/bash
set -e


DB_USER_VAL="client"
DB_PASS_VAL="securepass" # needed for init

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

# secret names
SECRET_SERVER_KEY="server_key"
SECRET_CLIENT_KEY="client_key"
SECRET_DB_USER="db_user"
SECRET_DB_PASS="db_password"

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

echo "re-creating podman secrets..."

# remove old secrets if they exist
podman secret rm "$SECRET_SERVER_KEY" "$SECRET_CLIENT_KEY" "$SECRET_DB_USER" "$SECRET_DB_PASS" 2>/dev/null || true

# create cert secrets
podman secret create "$SECRET_SERVER_KEY" "$SERVER_KEY"
podman secret create "$SECRET_CLIENT_KEY" "$CLIENT_KEY"

# create db auth secrets
echo -n "$DB_USER_VAL" | podman secret create "$SECRET_DB_USER" -
echo -n "$DB_PASS_VAL" | podman secret create "$SECRET_DB_PASS" -

echo "gen mosquitto.conf..."
cat > "$MOSQUITTO_CONF" << EOF
listener 8883

# TLS conf
certfile /mosquitto/$CERTS_DIR/server.crt
keyfile /run/secrets/$SECRET_SERVER_KEY
cafile /mosquitto/$CERTS_DIR/ca.crt

# Client auth
require_certificate true
use_identity_as_username true
EOF

# cleanup
rm "$SERVER_CSR" "$CLIENT_CSR" "$EXT_FILE"

echo "done."