#!/bin/bash

# this's a script used to gen certs for client-mtls-trustbundle test

rm -rf www.* client.* example.com.* revoked-client.* index.txt* crlnumber* ca.conf *.old *.attr

# Generate a Private Key
openssl req -x509 -sha256 -nodes -days 365 -newkey rsa:2048 -subj '/O=example Inc./CN=example.com' -keyout example.com.key -out example.com.crt -addext "subjectAltName = DNS:*.example.com"

cat > www.ext  <<EOF
subjectKeyIdentifier   = hash
authorityKeyIdentifier = keyid:always,issuer:always
basicConstraints       = CA:TRUE
keyUsage               = digitalSignature, nonRepudiation, keyEncipherment, dataEncipherment, keyAgreement, keyCertSign
subjectAltName         = DNS:www.example.com
issuerAltName          = issuer:copy
EOF

openssl req -out www.example.com.csr -newkey rsa:2048 -nodes -keyout www.example.com.key -subj "/CN=www.example.com/O=example organization" -addext "subjectAltName = DNS:www.example.com"
openssl x509 -req -in www.example.com.csr -CA example.com.crt -CAkey example.com.key -CAcreateserial -out www.example.com.crt -days 3650 -sha256 -extfile www.ext

cat > client.ext  <<EOF
subjectKeyIdentifier   = hash
authorityKeyIdentifier = keyid:always,issuer:always
basicConstraints       = CA:TRUE
keyUsage               = digitalSignature, nonRepudiation, keyEncipherment, dataEncipherment, keyAgreement, keyCertSign
subjectAltName         = DNS:client.example.com
issuerAltName          = issuer:copy
EOF

openssl req -out client.example.com.csr -newkey rsa:2048 -nodes -keyout client.example.com.key -subj "/CN=client.example.com/O=example organization" -addext "subjectAltName = DNS:client.example.com"
openssl x509 -req -in client.example.com.csr -CA example.com.crt -CAkey example.com.key -CAcreateserial -out client.example.com.crt -days 3650 -sha256 -extfile client.ext

# Generate a revoked client certificate for CRL testing
cat > revoked-client.ext  <<EOF
subjectKeyIdentifier   = hash
authorityKeyIdentifier = keyid:always,issuer:always
basicConstraints       = CA:TRUE
keyUsage               = digitalSignature, nonRepudiation, keyEncipherment, dataEncipherment, keyAgreement, keyCertSign
subjectAltName         = DNS:revoked-client.example.com
issuerAltName          = issuer:copy
EOF

openssl req -out revoked-client.example.com.csr -newkey rsa:2048 -nodes -keyout revoked-client.example.com.key -subj "/CN=revoked-client.example.com/O=example organization" -addext "subjectAltName = DNS:revoked-client.example.com"
openssl x509 -req -in revoked-client.example.com.csr -CA example.com.crt -CAkey example.com.key -CAcreateserial -out revoked-client.example.com.crt -days 3650 -sha256 -extfile revoked-client.ext

# Create CRL
touch index.txt
echo "01" > crlnumber

cat > ca.conf <<EOF
[ca]
default_ca = CA_default

[CA_default]
database = index.txt
crlnumber = crlnumber
default_crl_days = 3650
default_md = sha256

[crl_ext]
authorityKeyIdentifier=keyid:always
EOF

# Revoke the client certificate and generate CRL
openssl ca -config ca.conf -revoke revoked-client.example.com.crt -keyfile example.com.key -cert example.com.crt
openssl ca -config ca.conf -gencrl -keyfile example.com.key -cert example.com.crt -out example.com.crl

echo ""
echo "Certificates and CRL generated. Base64 encoded values:"
echo ""
echo "===== FOR example-com-tls Secret ====="
echo "CA Certificate (ca.crt):"
base64 < example.com.crt | tr -d '\n' && echo
echo ""
echo "Server Certificate (tls.crt):"
base64 < www.example.com.crt | tr -d '\n' && echo
echo ""
echo "Server Key (tls.key):"
base64 < www.example.com.key | tr -d '\n' && echo
echo ""
echo "===== FOR ClusterTrustBundle ====="
echo "Trust Bundle (plain text - paste into trustBundle field):"
cat example.com.crt
echo ""
echo "===== FOR client-example-com Secret (valid client) ====="
echo "CA Certificate (ca.crt):"
base64 < example.com.crt | tr -d '\n' && echo
echo ""
echo "Client Certificate (tls.crt):"
base64 < client.example.com.crt | tr -d '\n' && echo
echo ""
echo "Client Key (tls.key):"
base64 < client.example.com.key | tr -d '\n' && echo
echo ""
echo "===== FOR client-crl Secret ====="
echo "CA CRL (ca.crl):"
base64 < example.com.crl | tr -d '\n' && echo
echo ""
echo "===== FOR revoked-client-example-com Secret ====="
echo "CA Certificate (ca.crt):"
base64 < example.com.crt | tr -d '\n' && echo
echo ""
echo "Revoked Client Certificate (tls.crt):"
base64 < revoked-client.example.com.crt | tr -d '\n' && echo
echo ""
echo "Revoked Client Key (tls.key):"
base64 < revoked-client.example.com.key | tr -d '\n' && echo

echo ""
echo "==> Cleaning up temporary OpenSSL CA database files..."
rm -f index.txt* crlnumber* ca.conf *.old *.attr *.srl *.csr *.ext *.crl

echo "==> Done! Certificates and CRL generated successfully."
