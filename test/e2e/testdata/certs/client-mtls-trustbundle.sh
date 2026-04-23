#!/bin/bash

# this's a script used to gen certs for client-mtls-trustbundle test

rm -rf www.* client.* example.com.*

# Generate a Private Key
openssl req -x509 -sha256 -nodes -days 365 -newkey rsa:2048 -subj '/O=example Inc./CN=example.com' -keyout example.com.key -out example.com.crt -addext "subjectAltName = DNS:*.example.com"

# Generate server cert for www.example.com
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

# Generate server cert for foo.example.com
cat > foo.ext  <<EOF
subjectKeyIdentifier   = hash
authorityKeyIdentifier = keyid:always,issuer:always
basicConstraints       = CA:TRUE
keyUsage               = digitalSignature, nonRepudiation, keyEncipherment, dataEncipherment, keyAgreement, keyCertSign
subjectAltName         = DNS:foo.example.com
issuerAltName          = issuer:copy
EOF

openssl req -out foo.example.com.csr -newkey rsa:2048 -nodes -keyout foo.example.com.key -subj "/CN=foo.example.com/O=example organization" -addext "subjectAltName = DNS:foo.example.com"
openssl x509 -req -in foo.example.com.csr -CA example.com.crt -CAkey example.com.key -CAcreateserial -out foo.example.com.crt -days 3650 -sha256 -extfile foo.ext

# Generate server cert for bar.example.com
cat > bar.ext  <<EOF
subjectKeyIdentifier   = hash
authorityKeyIdentifier = keyid:always,issuer:always
basicConstraints       = CA:TRUE
keyUsage               = digitalSignature, nonRepudiation, keyEncipherment, dataEncipherment, keyAgreement, keyCertSign
subjectAltName         = DNS:bar.example.com
issuerAltName          = issuer:copy
EOF

openssl req -out bar.example.com.csr -newkey rsa:2048 -nodes -keyout bar.example.com.key -subj "/CN=bar.example.com/O=example organization" -addext "subjectAltName = DNS:bar.example.com"
openssl x509 -req -in bar.example.com.csr -CA example.com.crt -CAkey example.com.key -CAcreateserial -out bar.example.com.crt -days 3650 -sha256 -extfile bar.ext

# Generate server cert for www.awesome.org
cat > www-awesome.ext  <<EOF
subjectKeyIdentifier   = hash
authorityKeyIdentifier = keyid:always,issuer:always
basicConstraints       = CA:TRUE
keyUsage               = digitalSignature, nonRepudiation, keyEncipherment, dataEncipherment, keyAgreement, keyCertSign
subjectAltName         = DNS:www.awesome.org
issuerAltName          = issuer:copy
EOF

openssl req -out www.awesome.org.csr -newkey rsa:2048 -nodes -keyout www.awesome.org.key -subj "/CN=www.awesome.org/O=example organization" -addext "subjectAltName = DNS:www.awesome.org"
openssl x509 -req -in www.awesome.org.csr -CA example.com.crt -CAkey example.com.key -CAcreateserial -out www.awesome.org.crt -days 3650 -sha256 -extfile www-awesome.ext

# Generate server cert for *.awesome.org
cat > awesome.ext  <<EOF
subjectKeyIdentifier   = hash
authorityKeyIdentifier = keyid:always,issuer:always
basicConstraints       = CA:TRUE
keyUsage               = digitalSignature, nonRepudiation, keyEncipherment, dataEncipherment, keyAgreement, keyCertSign
subjectAltName         = DNS:*.awesome.org
issuerAltName          = issuer:copy
EOF

openssl req -out awesome.org.csr -newkey rsa:2048 -nodes -keyout awesome.org.key -subj "/CN=awesome.org/O=example organization" -addext "subjectAltName = DNS:awesome.org"
openssl x509 -req -in awesome.org.csr -CA example.com.crt -CAkey example.com.key -CAcreateserial -out awesome.org.crt -days 3650 -sha256 -extfile awesome.ext

# Generate client cert
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
