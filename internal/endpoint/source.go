// Copyright Envoy Gateway Authors
// SPDX-License-Identifier: Apache-2.0
// The full text of the Apache license is available in the LICENSE file at
// the root of the repo.

// Package endpoint shares immutable, versioned Kubernetes endpoint inputs between
// full translation and endpoint publication.
package endpoint

import (
	"hash/fnv"
	"reflect"
	"sort"
	"sync"

	corev1 "k8s.io/api/core/v1"
	discoveryv1 "k8s.io/api/discovery/v1"

	"github.com/envoyproxy/gateway/internal/ir"
)

type Key struct {
	Kind, Namespace, Name string
}

func BackendKey(backend ir.EndpointBackend) Key {
	return Key{Kind: backend.Kind, Namespace: backend.Namespace, Name: backend.Name}
}

type Port struct {
	Name     string
	Protocol corev1.Protocol
}

// Snapshot is immutable after collection. Slices and derived endpoints always
// describe the same revision; consumers may share them but must not mutate them.
type Snapshot struct {
	Backend     ir.EndpointBackend
	Revision    uint64
	Slices      []*discoveryv1.EndpointSlice
	AddressType *ir.DestinationAddressType
	Ports       map[Port][]*ir.DestinationEndpoint
}

// DeepCopy shares immutable data across the provider resource boundary.
func (s *Snapshot) DeepCopy() *Snapshot { return s }

func (s *Snapshot) DeepCopyInto(out *Snapshot) { *out = *s }

func (s *Snapshot) Reference(portName string, protocol corev1.Protocol) *ir.EndpointSource {
	return &ir.EndpointSource{Backend: s.Backend, Revision: s.Revision, PortName: portName, Protocol: string(protocol)}
}

// Source serializes collection of the same backend and coalesces notifications.
// Its zero value is ready to use. The xDS runner is the sole notification consumer.
type Source struct {
	once    sync.Once
	mu      sync.Mutex
	collect [64]sync.Mutex
	next    uint64
	entries map[Key]*Snapshot
	owners  map[string]map[Key]struct{}
	refs    map[Key]int
	pending map[Key]*Snapshot
	wake    chan struct{}
}

func (s *Source) init() {
	s.once.Do(func() {
		s.entries = make(map[Key]*Snapshot)
		s.owners = make(map[string]map[Key]struct{})
		s.refs = make(map[Key]int)
		s.pending = make(map[Key]*Snapshot)
		s.wake = make(chan struct{}, 1)
	})
}

// Refresh holds a collection lock across the loader so an older capture cannot
// receive a newer revision. Errors leave the previous source entry untouched.
func (s *Source) Refresh(key Key, load func() (ir.EndpointBackend, []*discoveryv1.EndpointSlice, error)) (*Snapshot, bool, error) {
	s.init()
	hash := fnv.New32a()
	_, _ = hash.Write([]byte(key.Kind + "/" + key.Namespace + "/" + key.Name))
	lock := &s.collect[hash.Sum32()%uint32(len(s.collect))]
	lock.Lock()
	defer lock.Unlock()
	backend, slices, err := load()
	if err != nil {
		return nil, false, err
	}
	current := NewSnapshot(backend, slices)
	s.mu.Lock()
	defer s.mu.Unlock()
	previous := s.entries[key]
	changed := previous == nil || previous.Backend != current.Backend ||
		!reflect.DeepEqual(previous.AddressType, current.AddressType) || !reflect.DeepEqual(previous.Ports, current.Ports)
	if changed {
		s.next++
		current.Revision = s.next
	} else {
		current.Revision = previous.Revision
	}
	if s.tracked(key) {
		s.entries[key] = current
	}
	return current, changed, nil
}

// Notify is called only after the provider has queued full reconciliation.
func (s *Source) Notify(snapshot *Snapshot) {
	s.init()
	s.mu.Lock()
	defer s.mu.Unlock()
	key := BackendKey(snapshot.Backend)
	if !s.tracked(key) {
		return
	}
	if previous := s.pending[key]; previous != nil && previous.Revision >= snapshot.Revision {
		return
	}
	s.pending[key] = snapshot
	select {
	case s.wake <- struct{}{}:
	default:
	}
}

func (s *Source) Updates() <-chan struct{} {
	s.init()
	return s.wake
}

func (s *Source) Drain() map[Key]*Snapshot {
	s.init()
	s.mu.Lock()
	defer s.mu.Unlock()
	updates := s.pending
	s.pending = make(map[Key]*Snapshot)
	return updates
}

// SetReferences retains sources needed by either current provider inputs or a
// published configuration. Revisions never reset when an entry is released.
func (s *Source) SetReferences(owner string, keys []Key) {
	s.init()
	s.mu.Lock()
	defer s.mu.Unlock()
	next := make(map[Key]struct{}, len(keys))
	for _, key := range keys {
		next[key] = struct{}{}
		if _, found := s.owners[owner][key]; !found {
			s.refs[key]++
		}
	}
	for key := range s.owners[owner] {
		if _, found := next[key]; found {
			continue
		}
		s.refs[key]--
		if s.refs[key] == 0 {
			delete(s.refs, key)
			delete(s.entries, key)
			delete(s.pending, key)
		}
	}
	if len(next) == 0 {
		delete(s.owners, owner)
	} else {
		s.owners[owner] = next
	}
}

func (s *Source) Tracked(key Key) bool {
	s.init()
	s.mu.Lock()
	defer s.mu.Unlock()
	return s.tracked(key)
}

func (s *Source) tracked(key Key) bool {
	return s.refs[key] > 0
}

func NewSnapshot(backend ir.EndpointBackend, slices []*discoveryv1.EndpointSlice) *Snapshot {
	slices = append([]*discoveryv1.EndpointSlice(nil), slices...)
	sort.SliceStable(slices, func(i, j int) bool {
		if !slices[i].CreationTimestamp.Equal(&slices[j].CreationTimestamp) {
			return slices[i].CreationTimestamp.Before(&slices[j].CreationTimestamp)
		}
		return slices[i].Namespace+"/"+slices[i].Name < slices[j].Namespace+"/"+slices[j].Name
	})
	snapshot := &Snapshot{Backend: backend, Slices: slices, Ports: make(map[Port][]*ir.DestinationEndpoint)}
	for _, slice := range slices {
		for _, port := range slice.Ports {
			key := Port{Name: portName(port), Protocol: portProtocol(port)}
			if _, found := snapshot.Ports[key]; !found {
				snapshot.Ports[key], _ = FromSlices(slices, key.Name, key.Protocol)
			}
		}
	}
	snapshot.AddressType = addressType(slices)
	return snapshot
}

func portName(port discoveryv1.EndpointPort) string {
	if port.Name == nil {
		return ""
	}
	return *port.Name
}

func portProtocol(port discoveryv1.EndpointPort) corev1.Protocol {
	if port.Protocol == nil {
		return corev1.ProtocolTCP
	}
	return *port.Protocol
}

// FromSlices is the shared full/fast endpoint conversion. Callers supply slices
// in provider order; nil ports cannot identify a usable endpoint.
func FromSlices(slices []*discoveryv1.EndpointSlice, name string, protocol corev1.Protocol) ([]*ir.DestinationEndpoint, *ir.DestinationAddressType) {
	var capacity int
	for _, slice := range slices {
		capacity += len(slice.Endpoints)
	}
	endpoints := make([]*ir.DestinationEndpoint, 0, capacity)
	for _, slice := range slices {
		endpoints = append(endpoints, FromSlice(slice, name, protocol)...)
	}
	if len(endpoints) == 0 {
		endpoints = nil
	}
	return endpoints, addressType(slices)
}

func addressType(slices []*discoveryv1.EndpointSlice) *ir.DestinationAddressType {
	var addressType *ir.DestinationAddressType
	for _, slice := range slices {
		kind := ir.IP
		if slice.AddressType == discoveryv1.AddressTypeFQDN {
			kind = ir.FQDN
		}
		if addressType == nil {
			addressType = new(kind)
		} else if *addressType != kind {
			addressType = new(ir.MIXED)
		}
	}
	return addressType
}

func FromSlice(slice *discoveryv1.EndpointSlice, name string, protocol corev1.Protocol) []*ir.DestinationEndpoint {
	var endpoints []*ir.DestinationEndpoint
	for _, endpoint := range slice.Endpoints {
		for _, port := range slice.Ports {
			if portName(port) != name || portProtocol(port) != protocol || port.Port == nil {
				continue
			}
			conditions := endpoint.Conditions
			draining := conditions.Ready != nil && !*conditions.Ready
			if conditions.Serving != nil && conditions.Terminating != nil {
				draining = *conditions.Terminating || !*conditions.Serving
			}
			for _, address := range endpoint.Addresses {
				endpoints = append(endpoints, ir.NewDestEndpoint(nil, address, uint32(*port.Port), draining, endpoint.Zone))
			}
		}
	}
	return endpoints
}
