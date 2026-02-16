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
