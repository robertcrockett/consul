// Copyright (c) HashiCorp, Inc.
// SPDX-License-Identifier: MPL-2.0

package catalogv2beta1

import pbresource "github.com/hashicorp/consul/proto-public/pbresource/v1"

// GetUnderlyingDestinations will collect FailoverDestinations from all
// internal fields and bundle them up in one slice.
//
// NOTE: no deduplication occurs.
func (x *ComputedFailoverPolicy) GetUnderlyingDestinations() []*FailoverDestination {
	if x == nil {
		return nil
	}

	estimate := 0
	for _, pc := range x.PortConfigs {
		estimate += len(pc.Destinations)
	}

	out := make([]*FailoverDestination, 0, estimate)
	for _, pc := range x.PortConfigs {
		out = append(out, pc.Destinations...)
	}
	return out
}

// GetUnderlyingDestinationRefs is like GetUnderlyingDestinations except it
// returns a slice of References.
//
// NOTE: no deduplication occurs.
func (x *ComputedFailoverPolicy) GetUnderlyingDestinationRefs() []*pbresource.Reference {
	if x == nil {
		return nil
	}

	dests := x.GetUnderlyingDestinations()

	out := make([]*pbresource.Reference, 0, len(dests))
	for _, dest := range dests {
		if dest.Ref != nil {
			out = append(out, dest.Ref)
		}
	}

	return out
}
