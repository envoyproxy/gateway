// Copyright Envoy Gateway Authors
// SPDX-License-Identifier: Apache-2.0
// The full text of the Apache license is available in the LICENSE file at
// the root of the repo.

package message

import "github.com/envoyproxy/gateway/internal/ir"

// copyXdsIR preserves sharing of HTTP destination settings across matches and
// listeners. The copy owns its data independently of the input; the xDS
// translator only reads these settings. Other fields use generated deep copies.
func copyXdsIR(in *ir.Xds) *ir.Xds {
	settings := sharedHTTPDestinations(in)
	if settings == nil {
		return in.DeepCopy()
	}
	base := *in
	base.HTTP = nil
	out := base.DeepCopy()
	out.HTTP = make([]*ir.HTTPListener, len(in.HTTP))
	for i, listener := range in.HTTP {
		if listener == nil {
			continue
		}
		baseListener := *listener
		baseListener.Routes = nil
		out.HTTP[i] = baseListener.DeepCopy()
		if listener.Routes == nil {
			continue
		}
		out.HTTP[i].Routes = make([]*ir.HTTPRoute, len(listener.Routes))
		for j, route := range listener.Routes {
			out.HTTP[i].Routes[j] = copyHTTPRoute(route, settings)
		}
	}
	return out
}

// Return a reusable memo only when a destination setting occurs more than once.
func sharedHTTPDestinations(in *ir.Xds) map[*ir.DestinationSetting]*ir.DestinationSetting {
	var capacity int
	for _, listener := range in.HTTP {
		if listener != nil {
			capacity = len(listener.Routes)
			break
		}
	}
	var settings map[*ir.DestinationSetting]*ir.DestinationSetting
	for _, listener := range in.HTTP {
		if listener == nil {
			continue
		}
		for _, route := range listener.Routes {
			if route == nil || route.Destination == nil {
				continue
			}
			for _, setting := range route.Destination.Settings {
				if setting == nil {
					continue
				}
				if settings == nil {
					settings = make(map[*ir.DestinationSetting]*ir.DestinationSetting, capacity)
				}
				if settings[setting] != nil {
					clear(settings)
					return settings
				}
				settings[setting] = setting
			}
		}
	}
	return nil
}

func copyHTTPRoute(in *ir.HTTPRoute, settings map[*ir.DestinationSetting]*ir.DestinationSetting) *ir.HTTPRoute {
	if in == nil || in.Destination == nil {
		return in.DeepCopy()
	}
	base := *in
	base.Destination = nil
	out := base.DeepCopy()
	destination := *in.Destination
	destination.Settings = nil
	out.Destination = destination.DeepCopy()
	if in.Destination.Settings == nil {
		return out
	}
	out.Destination.Settings = make([]*ir.DestinationSetting, len(in.Destination.Settings))
	for i, setting := range in.Destination.Settings {
		if setting == nil {
			continue
		}
		copied := settings[setting]
		if copied == nil {
			copied = setting.DeepCopy()
			settings[setting] = copied
		}
		out.Destination.Settings[i] = copied
	}
	return out
}
