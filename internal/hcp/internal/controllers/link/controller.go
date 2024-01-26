// Copyright (c) HashiCorp, Inc.
// SPDX-License-Identifier: BUSL-1.1

package link

import (
	"context"
	"strings"

	gnmmod "github.com/hashicorp/hcp-sdk-go/clients/cloud-global-network-manager-service/preview/2022-02-15/models"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
	"google.golang.org/protobuf/types/known/anypb"

	"github.com/hashicorp/consul/agent/hcp/bootstrap"
	hcpclient "github.com/hashicorp/consul/agent/hcp/client"
	"github.com/hashicorp/consul/agent/hcp/config"
	"github.com/hashicorp/consul/internal/controller"
	"github.com/hashicorp/consul/internal/hcp/internal/types"
	"github.com/hashicorp/consul/internal/resource"
	"github.com/hashicorp/consul/internal/storage"
	pbhcp "github.com/hashicorp/consul/proto-public/pbhcp/v2"
	"github.com/hashicorp/consul/proto-public/pbresource"
)

// HCPClientFn is a function that can be used to create an HCP client from a Link object.
// This function type should be passed to a LinkController in order to tell it how to make a client from
// a Link. For normal use, DefaultHCPClientFn should be used, but tests can substitute in a function that creates a
// mock client.
type HCPClientFn func(link *pbhcp.Link) (hcpclient.Client, error)

var DefaultHCPClientFn HCPClientFn = func(link *pbhcp.Link) (hcpclient.Client, error) {
	hcpClient, err := hcpclient.NewClient(config.CloudConfig{
		ResourceID:   link.ResourceId,
		ClientID:     link.ClientId,
		ClientSecret: link.ClientSecret,
	})
	if err != nil {
		return nil, err
	}
	return hcpClient, nil
}

func LinkController(
	resourceApisEnabled bool,
	hcpAllowV2ResourceApis bool,
	hcpClientFn HCPClientFn,
	cfg config.CloudConfig,
	dataDir string,
) *controller.Controller {
	return controller.NewController("link", pbhcp.LinkType).
		WithInitializer(&linkInitializer{
			cloudConfig: cfg,
		}).
		WithReconciler(&linkReconciler{
			resourceApisEnabled:    resourceApisEnabled,
			hcpAllowV2ResourceApis: hcpAllowV2ResourceApis,
			hcpClientFn:            hcpClientFn,
			dataDir:                dataDir,
		})
}

type linkReconciler struct {
	resourceApisEnabled    bool
	hcpAllowV2ResourceApis bool
	hcpClientFn            HCPClientFn
	dataDir                string
}

func hcpAccessLevelToConsul(level *gnmmod.HashicorpCloudGlobalNetworkManager20220215ClusterConsulAccessLevel) pbhcp.AccessLevel {
	if level == nil {
		return pbhcp.AccessLevel_ACCESS_LEVEL_UNSPECIFIED
	}

	switch *level {
	case gnmmod.HashicorpCloudGlobalNetworkManager20220215ClusterConsulAccessLevelCONSULACCESSLEVELUNSPECIFIED:
		return pbhcp.AccessLevel_ACCESS_LEVEL_UNSPECIFIED
	case gnmmod.HashicorpCloudGlobalNetworkManager20220215ClusterConsulAccessLevelCONSULACCESSLEVELGLOBALREADWRITE:
		return pbhcp.AccessLevel_ACCESS_LEVEL_GLOBAL_READ_WRITE
	case gnmmod.HashicorpCloudGlobalNetworkManager20220215ClusterConsulAccessLevelCONSULACCESSLEVELGLOBALREADONLY:
		return pbhcp.AccessLevel_ACCESS_LEVEL_GLOBAL_READ_ONLY
	default:
		return pbhcp.AccessLevel_ACCESS_LEVEL_UNSPECIFIED
	}
}

func (r *linkReconciler) Reconcile(ctx context.Context, rt controller.Runtime, req controller.Request) error {
	// The runtime is passed by value so replacing it here for the remainder of this
	// reconciliation request processing will not affect future invocations.
	rt.Logger = rt.Logger.With("resource-id", req.ID, "controller", StatusKey)

	rt.Logger.Trace("reconciling link")

	rsp, err := rt.Client.Read(ctx, &pbresource.ReadRequest{Id: req.ID})
	switch {
	case status.Code(err) == codes.NotFound:
		rt.Logger.Trace("link has been deleted")
		return nil
	case err != nil:
		rt.Logger.Error("the resource service has returned an unexpected error", "error", err)
		return err
	}

	res := rsp.Resource
	var link pbhcp.Link
	if err := res.Data.UnmarshalTo(&link); err != nil {
		rt.Logger.Error("error unmarshalling link data", "error", err)
		return err
	}

	if err = addFinalizer(ctx, rt, res); err != nil {
		rt.Logger.Error("error adding finalizer to link resource", "error", err)
		return err
	}

	if resource.IsMarkedForDeletion(res) {
		if err = cleanup(ctx, rt, res, r.dataDir); err != nil {
			rt.Logger.Error("error cleaning up link resource", "error", err)
			return err
		}
		return nil
	}

	// Validation - Ensure V2 Resource APIs are not enabled unless hcpAllowV2ResourceApis flag is set
	var newStatus *pbresource.Status
	if r.resourceApisEnabled && !r.hcpAllowV2ResourceApis {
		newStatus = &pbresource.Status{
			ObservedGeneration: res.Generation,
			Conditions:         []*pbresource.Condition{ConditionDisabled},
		}
		return writeStatusIfNotEqual(ctx, rt, res, newStatus)
	}

	hcpClient, err := r.hcpClientFn(&link)
	if err != nil {
		rt.Logger.Error("error creating HCP client", "error", err)
		return err
	}

	// Sync cluster data from HCP
	cluster, err := hcpClient.GetCluster(ctx)
	if err != nil {
		rt.Logger.Error("error querying HCP for cluster", "error", err)
		linkingFailed(ctx, rt, res, err)
		return err
	}
	accessLevel := hcpAccessLevelToConsul(cluster.AccessLevel)

	if link.HcpClusterUrl != cluster.HCPPortalURL ||
		link.AccessLevel != accessLevel {

		link.HcpClusterUrl = cluster.HCPPortalURL
		link.AccessLevel = accessLevel

		updatedData, err := anypb.New(&link)
		if err != nil {
			rt.Logger.Error("error marshalling link data", "error", err)
			return err
		}
		_, err = rt.Client.Write(ctx, &pbresource.WriteRequest{Resource: &pbresource.Resource{
			Id: &pbresource.ID{
				Name: "global",
				Type: pbhcp.LinkType,
			},
			Metadata: res.Metadata,
			Data:     updatedData,
		}})
		if err != nil {
			rt.Logger.Error("error updating link", "error", err)
			return err
		}
	}

	// Load the management token if access is not set to read-only. Read-only clusters
	// will not have a management token provided by HCP.
	if accessLevel != pbhcp.AccessLevel_ACCESS_LEVEL_GLOBAL_READ_ONLY {
		_, err = bootstrap.LoadManagementToken(ctx, rt.Logger, hcpClient, r.dataDir)
		if err != nil {
			linkingFailed(ctx, rt, res, err)
			return err
		}
		// TODO: Update the HCP manager with the loaded management token as part of CC-7044
	}

	newStatus = &pbresource.Status{
		ObservedGeneration: res.Generation,
		Conditions:         []*pbresource.Condition{ConditionLinked(link.ResourceId)},
	}

	return writeStatusIfNotEqual(ctx, rt, res, newStatus)
}

type linkInitializer struct {
	cloudConfig config.CloudConfig
}

func (i *linkInitializer) Initialize(ctx context.Context, rt controller.Runtime) error {
	if !i.cloudConfig.IsConfigured() {
		return nil
	}

	// Construct a link resource to reflect the configuration
	data, err := anypb.New(&pbhcp.Link{
		ResourceId:   i.cloudConfig.ResourceID,
		ClientId:     i.cloudConfig.ClientID,
		ClientSecret: i.cloudConfig.ClientSecret,
	})
	if err != nil {
		return err
	}

	// Create the link resource for a configuration-based link
	_, err = rt.Client.Write(ctx,
		&pbresource.WriteRequest{
			Resource: &pbresource.Resource{
				Id: &pbresource.ID{
					Name: types.LinkName,
					Type: pbhcp.LinkType,
				},
				Metadata: map[string]string{
					types.MetadataSourceKey: types.MetadataSourceConfig,
				},
				Data: data,
			},
		},
	)
	if err != nil {
		if strings.Contains(err.Error(), storage.ErrWrongUid.Error()) ||
			strings.Contains(err.Error(), "leader unknown") {
			// If the  error is likely ignorable and could eventually resolve itself,
			// log it as TRACE rather than ERROR.
			rt.Logger.Trace("error initializing controller", "error", err)
		} else {
			rt.Logger.Error("error initializing controller", "error", err)
		}
		return err
	}

	return nil
}
