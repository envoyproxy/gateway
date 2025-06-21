#!/bin/bash

# this's a script used to gen certs for client-mtls-trustbundle test

rm -rf www.* client.* example.com.*

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
