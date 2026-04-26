// Copyright Envoy Gateway Authors
// SPDX-License-Identifier: Apache-2.0
// The full text of the Apache license is available in the LICENSE file at
// the root of the repo.

// Package main implements a minimal dynamic module load balancer for e2e testing.
// It always selects the first healthy host at priority 0, which validates that
// the dynamic module LB policy pipeline works end-to-end.
package main

/*
#include <stddef.h>
#include <stdint.h>
#include <stdbool.h>

// Matches envoy_dynamic_module_type_envoy_buffer from abi.h.
typedef struct {
	uintptr_t ptr;
	size_t length;
} envoy_dynamic_module_type_envoy_buffer;

// Matches the return type expected by Envoy for the program init callback.
typedef const void* envoy_dynamic_module_type_abi_version_module_ptr;
*/
import "C"
import "unsafe"

// abiVersion must match the Envoy dynamic modules ABI version.
var abiVersion = "v0.1.0\x00"

//export envoy_dynamic_module_on_program_init
func envoy_dynamic_module_on_program_init() C.envoy_dynamic_module_type_abi_version_module_ptr {
	return C.envoy_dynamic_module_type_abi_version_module_ptr(unsafe.Pointer(unsafe.StringData(abiVersion)))
}

//export envoy_dynamic_module_on_lb_config_new
func envoy_dynamic_module_on_lb_config_new(
	lbConfigEnvoyPtr unsafe.Pointer,
	name C.envoy_dynamic_module_type_envoy_buffer,
	config C.envoy_dynamic_module_type_envoy_buffer,
) unsafe.Pointer {
	// Return a non-null sentinel to indicate success.
	return unsafe.Pointer(uintptr(1))
}

//export envoy_dynamic_module_on_lb_config_destroy
func envoy_dynamic_module_on_lb_config_destroy(configModulePtr unsafe.Pointer) {}

//export envoy_dynamic_module_on_lb_new
func envoy_dynamic_module_on_lb_new(
	configModulePtr unsafe.Pointer,
	lbEnvoyPtr unsafe.Pointer,
) unsafe.Pointer {
	// Return a non-null sentinel to indicate success.
	return unsafe.Pointer(uintptr(1))
}

// envoy_dynamic_module_on_lb_choose_host selects a host for an upstream request.
// This simple implementation always picks the first healthy host at priority 0.
//
//export envoy_dynamic_module_on_lb_choose_host
func envoy_dynamic_module_on_lb_choose_host(
	lbEnvoyPtr unsafe.Pointer,
	lbModulePtr unsafe.Pointer,
	contextEnvoyPtr unsafe.Pointer,
	resultPriority *C.uint32_t,
	resultIndex *C.uint32_t,
) C.bool {
	*resultPriority = 0
	*resultIndex = 0
	return true
}

//export envoy_dynamic_module_on_lb_on_host_membership_update
func envoy_dynamic_module_on_lb_on_host_membership_update(
	lbEnvoyPtr unsafe.Pointer,
	lbModulePtr unsafe.Pointer,
	numHostsAdded C.size_t,
	numHostsRemoved C.size_t,
) {
}

//export envoy_dynamic_module_on_lb_destroy
func envoy_dynamic_module_on_lb_destroy(lbModulePtr unsafe.Pointer) {}

func main() {}
