// Copyright Envoy Gateway Authors
// SPDX-License-Identifier: Apache-2.0
// The full text of the Apache license is available in the LICENSE file at
// the root of the repo.

package main

import (
	"encoding/json"
	"fmt"
	"log"
	"net"

	"github.com/valyala/fasthttp"
)

func HandleFastHTTP(ctx *fasthttp.RequestCtx) {
	// /http10 responds with HTTP/1.0 and echoes all request headers as JSON.
	// Used to verify that Envoy proxies to an HTTP/1.0 backend without needing
	// accept_http_10 / default_host_for_http_10 set on the upstream cluster.
	if string(ctx.Path()) == "/http10" {
		ctx.Response.Header.SetProtocol([]byte("HTTP/1.0"))
		headers := map[string]string{}
		ctx.Request.Header.VisitAll(func(key, value []byte) {
			headers[string(key)] = string(value)
		})
		d, err := json.MarshalIndent(headers, "", "  ")
		if err != nil {
			ctx.Error(fmt.Sprintf("%s", err), fasthttp.StatusBadRequest)
			return
		}
		fmt.Fprintf(ctx, "%s\n", d)
		return
	}

	ctx.QueryArgs().VisitAll(func(key, value []byte) {
		if string(key) == "headers" {
			ctx.Response.Header.Add(string(value), "PrEsEnT")
		}
	})
	headers := map[string][]string{}
	ctx.Request.Header.VisitAll(func(key, value []byte) {
		headers[string(key)] = append(headers[string(key)], string(value))
	})
	if d, err := json.MarshalIndent(headers, "", "  "); err != nil {
		ctx.Error(fmt.Sprintf("%s", err), fasthttp.StatusBadRequest)
	} else {
		fmt.Fprintf(ctx, string(d)+"\n")
	}
}

func main() {
	s := fasthttp.Server{
		Handler:                       HandleFastHTTP,
		DisableHeaderNamesNormalizing: true,
	}
	log.Printf("Starting on port 8000")
	l, _ := net.Listen("tcp", ":8000")
	log.Fatal(s.Serve(l))
}
