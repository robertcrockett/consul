// Code generated by protoc-gen-resource-types. DO NOT EDIT.

package authv2beta1

import (
	pbresource "github.com/hashicorp/consul/proto-public/pbresource/v1"
)

const (
	GroupName = "auth"
	Version   = "v2beta1"

	ComputedTrafficPermissionsKind  = "ComputedTrafficPermissions"
	NamespaceTrafficPermissionsKind = "NamespaceTrafficPermissions"
	PartitionTrafficPermissionsKind = "PartitionTrafficPermissions"
	TrafficPermissionsKind          = "TrafficPermissions"
	WorkloadIdentityKind            = "WorkloadIdentity"
)

var (
	ComputedTrafficPermissionsType = &pbresource.Type{
		Group:        GroupName,
		GroupVersion: Version,
		Kind:         ComputedTrafficPermissionsKind,
	}

	NamespaceTrafficPermissionsType = &pbresource.Type{
		Group:        GroupName,
		GroupVersion: Version,
		Kind:         NamespaceTrafficPermissionsKind,
	}

	PartitionTrafficPermissionsType = &pbresource.Type{
		Group:        GroupName,
		GroupVersion: Version,
		Kind:         PartitionTrafficPermissionsKind,
	}

	TrafficPermissionsType = &pbresource.Type{
		Group:        GroupName,
		GroupVersion: Version,
		Kind:         TrafficPermissionsKind,
	}

	WorkloadIdentityType = &pbresource.Type{
		Group:        GroupName,
		GroupVersion: Version,
		Kind:         WorkloadIdentityKind,
	}
)
