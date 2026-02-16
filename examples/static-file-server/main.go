// Copyright Envoy Gateway Authors
// SPDX-License-Identifier: Apache-2.0
// The full text of the Apache license is available in the LICENSE file at
// the root of the repo.

package main

import (
	"flag"
	"log"
	"net/http"
	"os"
	"path"
)

var (
	port      string
	directory string
	certPath  string
)

func main() {
	flag.StringVar(&port, "port", "8080", "port to serve on")
	flag.StringVar(&directory, "dir", "./files", "the directory of static file to host")
	flag.StringVar(&certPath, "certPath", "/etc/certs", "path to extProcServer certificate and private key")
	flag.Parse()

	http.Handle("/", http.FileServer(http.Dir(directory)))

	if _, err := os.Stat(path.Join(certPath, "tls.crt")); err != nil {
		log.Printf("Serving %s on HTTP port: %s\n", directory, port)
		log.Fatal(http.ListenAndServe(":"+port, nil))
		return
	}

	log.Printf("Serving %s on HTTPS port: %s\n", directory, port)
	log.Fatal(http.ListenAndServeTLS(":"+port,
		path.Join(certPath, "tls.crt"), path.Join(certPath, "tls.key"), nil))
}
