// Copyright Envoy Gateway Authors
// SPDX-License-Identifier: Apache-2.0
// The full text of the Apache license is available in the LICENSE file at
// the root of the repo.

package migrate

import (
	"fmt"
	"io"
	"text/tabwriter"
)

// reporter writes one row per finding.
type reporter struct {
	w                *tabwriter.Writer
	rows             int
	migrated, errors int
}

func newReporter(w io.Writer) *reporter {
	return &reporter{w: tabwriter.NewWriter(w, 0, 0, 2, ' ', 0)}
}

func (r *reporter) add(obj *Object, f Finding) {
	if r.rows == 0 {
		_, _ = fmt.Fprintln(r.w, "KIND\tNAMESPACE/NAME\tFIELD\tRESULT\tNOTE")
	}
	r.rows++
	switch f.Result {
	case ResultMigrated:
		r.migrated++
	case ResultError:
		r.errors++
	}
	_, _ = fmt.Fprintf(r.w, "%s\t%s\t%s\t%s\t%s\n", objectKind(obj), objectName(obj), f.Field, f.Result, f.Note)
}

func (r *reporter) flush() error {
	if r.rows == 0 {
		_, _ = fmt.Fprintln(r.w, "Nothing to migrate.")
	}
	return r.w.Flush()
}

func objectKind(obj *Object) string {
	if obj.Kind == "" {
		return "-"
	}
	return obj.Kind
}

func objectName(obj *Object) string {
	name := obj.Name
	if name == "" {
		name = "-"
	}
	if obj.Namespace != "" {
		name = obj.Namespace + "/" + name
	}
	if obj.Source != "" && obj.Source != "-" {
		name += " (" + obj.Source + ")"
	}
	return name
}
