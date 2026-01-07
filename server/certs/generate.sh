#!/bin/bash
# Generate mTLS certificates for Travio services
# Run from server/ directory

set -e

CERT_DIR="certs"
DAYS=365

# Services that need certificates
SERVICES=("gateway" "catalog" "inventory" "order" "payment" "fulfillment" "pricing" "queue" "identity" "search" "notification")

echo "=== Travio mTLS Certificate Generator ==="

# Generate CA if not exists
if [ ! -f "$CERT_DIR/ca.key" ]; then
    echo "Generating CA..."
    openssl genrsa -out "$CERT_DIR/ca.key" 2048 2>/dev/null
    openssl req -new -x509 -key "$CERT_DIR/ca.key" -out "$CERT_DIR/ca.crt" -days $DAYS -subj "/CN=Travio CA" 2>/dev/null
    echo "✓ CA generated"
fi

# Generate cert for each service
for service in "${SERVICES[@]}"; do
    if [ ! -f "$CERT_DIR/$service.key" ]; then
        echo "Generating cert for $service..."
        openssl genrsa -out "$CERT_DIR/$service.key" 2048 2>/dev/null
        openssl req -new -key "$CERT_DIR/$service.key" -out "$CERT_DIR/$service.csr" -subj "/CN=$service" 2>/dev/null
        openssl x509 -req -in "$CERT_DIR/$service.csr" -CA "$CERT_DIR/ca.crt" -CAkey "$CERT_DIR/ca.key" -CAcreateserial -out "$CERT_DIR/$service.crt" -days $DAYS 2>/dev/null
        rm "$CERT_DIR/$service.csr"
        echo "✓ $service certificate generated"
    else
        echo "· $service certificate exists"
    fi
done

echo ""
echo "=== Certificate Generation Complete ==="
echo "Certificates are in: $CERT_DIR/"
ls -la "$CERT_DIR/"
