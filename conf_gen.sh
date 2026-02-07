#!/bin/bash
set -e

# credentials
DB_USER_VAL="fill"
DB_PASS_VAL="fill"
MQTT_USER_VAL="fill"
MQTT_PASS_VAL="fill"

if [[ "$DB_USER_VAL" == "fill" || "$DB_PASS_VAL" == "fill" || "$MQTT_USER_VAL" == "fill" || "$MQTT_PASS_VAL" == "fill" ]]; then
    echo "Error: One or more credentials are set to 'fill'. Please update the script variables."
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
EXT_FILE="$CERTS_DIR/v3.ext"
MOSQUITTO_PASS_FILE="mosquitto.passwd"
MOSQUITTO_CONF="$MOSQUITTO_CONF_DIR/mosquitto.conf"

SECRET_SERVER_KEY="server_key"
SECRET_DB_USER="db_user"
SECRET_DB_PASS="db_password"
SECRET_MQTT_USER="mqtt_user"
SECRET_MQTT_PASS="mqtt_password"
SECRET_MQTT_DB="mqtt_passwd_db"

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

# sign cert
openssl x509 -req -in "$SERVER_CSR" -CA "$CA_CRT" -CAkey "$CA_KEY" -CAcreateserial \
  -out "$SERVER_CRT" -days 365 -sha256 -extfile "$EXT_FILE"

echo "re-creating podman secrets..."

podman secret rm "$SECRET_SERVER_KEY" "$SECRET_DB_USER" "$SECRET_DB_PASS" \
  "$SECRET_MQTT_USER" "$SECRET_MQTT_PASS" "$SECRET_MQTT_DB" 2>/dev/null || true

podman secret create "$SECRET_SERVER_KEY" "$SERVER_KEY"

echo -n "$DB_USER_VAL" | podman secret create "$SECRET_DB_USER" -
echo -n "$DB_PASS_VAL" | podman secret create "$SECRET_DB_PASS" -

echo -n "$MQTT_USER_VAL" | podman secret create "$SECRET_MQTT_USER" -
echo -n "$MQTT_PASS_VAL" | podman secret create "$SECRET_MQTT_PASS" -

echo "gen mosquitto password hash..."
podman run --rm docker.io/eclipse-mosquitto:2.1.0-alpine \
  sh -c "mosquitto_passwd -c -b /tmp/passwd '$MQTT_USER_VAL' '$MQTT_PASS_VAL' > /dev/null 2>&1 && cat /tmp/passwd" > "$MOSQUITTO_PASS_FILE"

podman secret create "$SECRET_MQTT_DB" "$MOSQUITTO_PASS_FILE"

echo "gen mosquitto.conf..."
cat > "$MOSQUITTO_CONF" << EOF
listener 8883

# TLS conf
certfile /mosquitto/$CERTS_DIR/server.crt
keyfile /run/secrets/$SECRET_SERVER_KEY

# Client auth
allow_anonymous false
password_file /run/secrets/$SECRET_MQTT_DB
EOF

# cleanup
rm "$MOSQUITTO_PASS_FILE" "$SERVER_CSR" "$EXT_FILE"

echo "done."