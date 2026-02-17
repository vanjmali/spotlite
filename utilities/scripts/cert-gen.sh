#!/bin/sh
set -e # Exit on error

CERT_DIR="/certs"
mkdir -p "$CERT_DIR"

if [ ! -s "$CERT_DIR/rootCA.crt" ]; then
    echo "Generating new Root CA..."
    openssl genpkey -algorithm RSA -pkeyopt rsa_keygen_bits:3072 -out "$CERT_DIR/rootCA.key"

    openssl req -x509 -new -nodes -key "$CERT_DIR/rootCA.key" -sha256 -days 3650 \
        -out "$CERT_DIR/rootCA.crt" \
        -subj "/CN=Spotlite Root CA" \
        -addext "basicConstraints=critical,CA:TRUE" \
        -addext "keyUsage=critical,keyCertSign,cRLSign"
else
    echo "Root CA already exists, skipping..."
fi

# function used for certificate generation
generate_service_cert() {
    local NAME=$1
    echo "Generating certificate for: $NAME"

    # generate private key
    openssl genpkey -algorithm RSA -pkeyopt rsa_keygen_bits:3072 -out "$CERT_DIR/$NAME.key"

    # create CSR with SANs (Subject Alternative Names) using -addext
    openssl req -new -key "$CERT_DIR/$NAME.key" -out "$CERT_DIR/$NAME.csr" \
        -subj "/CN=$NAME" \
        -addext "subjectAltName=DNS:$NAME,DNS:localhost" \
        -addext "keyUsage=critical,digitalSignature,keyEncipherment" \
        -addext "extendedKeyUsage=serverAuth,clientAuth"

    # certificate signing
    openssl x509 -req -in "$CERT_DIR/$NAME.csr" \
        -CA "$CERT_DIR/rootCA.crt" -CAkey "$CERT_DIR/rootCA.key" \
        -CAcreateserial -out "$CERT_DIR/$NAME.crt" \
        -days 365 -sha256 -copy_extensions copy

    # clean up CSR 
    rm "$CERT_DIR/$NAME.csr"
}

# generate certs for all services
SERVICES="api-gateway user-service notification-service subscription-service content-service frontend"

# repeat for every service
for SERVICE in $SERVICES; do
    if [ ! -s "$CERT_DIR/$SERVICE.crt" ]; then
        generate_service_cert "$SERVICE"
        echo "$SERVICE certificate is generated!"
    else
        echo "$SERVICE certificate already exists and is valid, skipping..."
    fi
done

echo "Certificates generated successfully!"


chmod 644 $CERT_DIR/*.key
chmod 644 $CERT_DIR/*.crt