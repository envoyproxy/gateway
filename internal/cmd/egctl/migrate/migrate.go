// Copyright Envoy Gateway Authors
// SPDX-License-Identifier: Apache-2.0
// The full text of the Apache license is available in the LICENSE file at
// the root of the repo.

// Package migrate implements the shared driver of the `egctl x migrate` subcommands: it loads
// objects from files or a cluster, runs the rewrite rules of a migration Path on them, writes the
// result according to the Mode, and reports what it did.
package migrate

import (
	"context"
	"errors"
	"fmt"
	"io"
	"strings"

	"go.yaml.in/yaml/v3"
	"sigs.k8s.io/controller-runtime/pkg/client"
)

// Mode selects what migrate does with the rewritten objects.
type Mode string

const (
	// ModePrint prints the migrated YAML to stdout.
	ModePrint Mode = "print"
	// ModeInPlace writes the migrated YAML back to its source.
	ModeInPlace Mode = "in-place"
	// ModeCheckOnly only reports what would change.
	ModeCheckOnly Mode = "check-only"
)

// Modes lists the supported modes.
var Modes = []Mode{ModePrint, ModeInPlace, ModeCheckOnly}

// Result is the outcome of the migration of one field.
type Result string

const (
	// ResultMigrated means the field was rewritten automatically.
	ResultMigrated Result = "migrated"
	// ResultError means the field can't be migrated automatically and was left unchanged.
	ResultError Result = "error"
)

// Finding reports the migration of one field.
type Finding struct {
	// Field is the path of the field, e.g. "spec.targetRef".
	Field string
	// Result is the outcome.
	Result Result
	// Note explains the rewrite, or what the user has to do for an error.
	Note string
}

// Object is a Kubernetes object, or another YAML document, being migrated.
type Object struct {
	// Root is the top-level mapping node of the document. Rules rewrite it in place.
	Root *yaml.Node
	// APIVersion, Kind, Namespace and Name identify the object. They're empty for a document
	// that isn't a Kubernetes object, such as Helm values.
	APIVersion, Kind, Namespace, Name string
	// Source is the file the object was read from, empty for an object read from the cluster.
	Source string
}

// Path is a migration path, e.g. from the deprecated API fields to their replacements.
type Path interface {
	// Migrate rewrites obj.Root in place and returns a finding per field it migrated or couldn't
	// migrate. A returned error means the object couldn't be processed at all.
	Migrate(obj *Object) ([]Finding, error)
	// Fetch lists the objects the path migrates from the cluster.
	Fetch(ctx context.Context, cli client.Client, namespace string) ([]client.Object, error)
}

// Options configures a run.
type Options struct {
	// Files are the files or directories to read, or "-" for stdin.
	Files []string
	// FromCluster reads the objects from the cluster instead of files.
	FromCluster bool
	// Namespace limits the objects read from the cluster. Empty means all namespaces.
	Namespace string
	// Mode selects what to do with the rewritten objects.
	Mode Mode
}

// ExitError is returned when the run completed but the exit code must be nonzero.
type ExitError struct {
	Code    int
	Message string
}

func (e *ExitError) Error() string { return e.Message }

const (
	// ExitCodeWouldChange is returned in check-only mode when something would be migrated.
	ExitCodeWouldChange = 1
	// ExitCodeError is returned when at least one field is reported as an error.
	ExitCodeError = 2
)

// Validate checks that the options are consistent.
func (o *Options) Validate() error {
	switch o.Mode {
	case ModePrint, ModeInPlace, ModeCheckOnly:
	default:
		return fmt.Errorf("invalid --mode %q, must be one of %v", o.Mode, Modes)
	}
	if o.FromCluster == (len(o.Files) > 0) {
		return errors.New("exactly one of -f/--file or --from-cluster must be specified")
	}
	if o.Mode == ModeInPlace {
		if o.FromCluster {
			return errors.New("--mode in-place is not supported with --from-cluster yet")
		}
		for _, f := range o.Files {
			if f == "-" {
				return errors.New("--mode in-place can't be used with -f -, there is no file to write back to")
			}
		}
	}
	return nil
}

// Run migrates the objects selected by opts with path. Migrated YAML is written to out, the report
// to report. cli is only used with opts.FromCluster.
func Run(ctx context.Context, path Path, opts Options, cli client.Client, out, report io.Writer) error {
	if err := opts.Validate(); err != nil {
		return err
	}

	var (
		rep    = newReporter(report)
		inputs []*input
		err    error
	)
	if opts.FromCluster {
		inputs, err = loadCluster(ctx, path, cli, opts.Namespace)
	} else {
		inputs, err = loadFiles(opts.Files)
	}
	if err != nil {
		return err
	}

	for _, in := range inputs {
		for _, doc := range in.docs {
			if doc.obj == nil {
				continue
			}
			migrateDoc(path, doc, rep, make(map[*yaml.Node]bool))
		}
	}

	if err := rep.flush(); err != nil {
		return err
	}

	switch opts.Mode {
	case ModePrint:
		for i, in := range inputs {
			if err := in.write(out, i == 0); err != nil {
				return err
			}
		}
	case ModeInPlace:
		for _, in := range inputs {
			if !in.changed() {
				continue
			}
			if err := in.writeFile(); err != nil {
				return err
			}
		}
	case ModeCheckOnly:
	}

	switch {
	case rep.errors > 0:
		return &ExitError{Code: ExitCodeError, Message: fmt.Sprintf("%d field(s) can't be migrated automatically, see the report", rep.errors)}
	case opts.Mode == ModeCheckOnly && rep.migrated > 0:
		return &ExitError{Code: ExitCodeWouldChange, Message: fmt.Sprintf("%d field(s) would be migrated", rep.migrated)}
	}
	return nil
}

// migrateDoc runs the path on one document and records its findings. Managed objects are reverted
// and every finding on them is reported as an error.
func migrateDoc(path Path, doc *document, rep *reporter, seen map[*yaml.Node]bool) {
	obj := doc.obj
	root := Resolve(obj.Root)
	if seen[root] {
		return
	}
	seen[root] = true
	// kubectl can emit generic Lists or typed lists whose items omit their GVK.
	if obj.APIVersion != "" && strings.HasSuffix(obj.Kind, "List") {
		if items := Resolve(Get(root, "items")); items != nil && items.Kind == yaml.SequenceNode {
			for i, item := range items.Content {
				if Resolve(item).Kind != yaml.MappingNode {
					continue
				}
				child := newObject(item, obj.Source)
				if obj.Kind != "List" {
					if child.APIVersion == "" {
						child.APIVersion = obj.APIVersion
					}
					if child.Kind == "" {
						child.Kind = strings.TrimSuffix(obj.Kind, "List")
					}
				}
				d := &document{obj: child, managedBy: doc.managedBy}
				migrateDoc(path, d, rep, seen)
				items.Content[i] = child.Root
				doc.changed = doc.changed || d.changed
			}
			return
		}
	}
	original := Clone(obj.Root)
	findings, err := path.Migrate(obj)
	if err != nil {
		obj.Root = original
		rep.add(obj, Finding{Field: "-", Result: ResultError, Note: err.Error()})
		return
	}
	if doc.managedBy != "" {
		obj.Root = original
		for i := range findings {
			findings[i] = Finding{
				Field:  findings[i].Field,
				Result: ResultError,
				Note:   fmt.Sprintf("managed by %s, not changed: fix the source instead", doc.managedBy),
			}
		}
	}
	for _, f := range findings {
		if f.Result == ResultMigrated {
			doc.changed = true
		}
		rep.add(obj, f)
	}
}
