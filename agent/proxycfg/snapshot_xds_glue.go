package proxycfg

import (
	"github.com/hashicorp/consul/acl"
	"github.com/hashicorp/consul/agent/structs"
	"github.com/hashicorp/consul/logging"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
)

// This file contains the implementation of the xds.ConfigSnapshot interface for proxycfg.ConfigSnapshot

type ProxySnapshot interface {
	AllowEmptyListeners() bool
	AllowEmptyRoutes() bool
	AllowEmptyClusters() bool
	Authorize(authz acl.Authorizer) error
	LoggerName() string
}

func (s *ConfigSnapshot) AllowEmptyListeners() bool {
	// Ingress and API gateways are allowed to inform LDS of no listeners.
	return s.Kind == structs.ServiceKindIngressGateway ||
		s.Kind == structs.ServiceKindAPIGateway
}

func (s *ConfigSnapshot) AllowEmptyRoutes() bool {
	// Ingress and API gateways are allowed to inform RDS of no routes.
	return s.Kind == structs.ServiceKindIngressGateway ||
		s.Kind == structs.ServiceKindAPIGateway
}

func (s *ConfigSnapshot) AllowEmptyClusters() bool {
	// Mesh, Ingress, API and Terminating gateways are allowed to inform CDS of no clusters.
	return s.Kind == structs.ServiceKindMeshGateway ||
		s.Kind == structs.ServiceKindTerminatingGateway ||
		s.Kind == structs.ServiceKindIngressGateway ||
		s.Kind == structs.ServiceKindAPIGateway
}

func (s *ConfigSnapshot) Authorize(authz acl.Authorizer) error {
	var authzContext acl.AuthorizerContext
	switch s.Kind {
	case structs.ServiceKindConnectProxy:
		s.ProxyID.EnterpriseMeta.FillAuthzContext(&authzContext)
		if err := authz.ToAllowAuthorizer().ServiceWriteAllowed(s.Proxy.DestinationServiceName, &authzContext); err != nil {
			return status.Errorf(codes.PermissionDenied, err.Error())
		}
	case structs.ServiceKindMeshGateway, structs.ServiceKindTerminatingGateway, structs.ServiceKindIngressGateway, structs.ServiceKindAPIGateway:
		s.ProxyID.EnterpriseMeta.FillAuthzContext(&authzContext)
		if err := authz.ToAllowAuthorizer().ServiceWriteAllowed(s.Service, &authzContext); err != nil {
			return status.Errorf(codes.PermissionDenied, err.Error())
		}
	default:
		return status.Errorf(codes.Internal, "Invalid service kind")
	}

	// Authed OK!
	return nil
}

// todo (ishustava): do we need to implement it for IR?
func (s *ConfigSnapshot) LoggerName() string {
	switch s.Kind {
	case structs.ServiceKindConnectProxy:
	case structs.ServiceKindTerminatingGateway:
		return logging.TerminatingGateway
	case structs.ServiceKindMeshGateway:
		return logging.MeshGateway
	case structs.ServiceKindIngressGateway:
		return logging.IngressGateway
	}

	return ""
}
