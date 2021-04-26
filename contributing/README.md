# Contributing to Consul

See [our contributing guide](../.github/CONTRIBUTING.md) to get started.

This directory contains documentation intended for anyone interested in
understanding, and contributing changes to, the Consul codebase.

## Contents

1. [Overview](./INTERNALS.md)
2. [Configuration](./checklist-adding-config-fields.md)
3. [Streaming](./streaming)
4. Persistence
   - raft
   - 
5. RPC
6. Connect - Service Mesh
   - call out: envoy/proxy is the data plane, Consul is the control plane
   - agent/xds - gRPC service that implements
     [xDS](https://www.envoyproxy.io/docs/envoy/latest/api-docs/xds_protocol)
   - [agent/proxycfg](https://github.com/hashicorp/consul/blob/master/agent/proxycfg/proxycfg.go)
   - CA Manager - certificate authority
   - command/connect/envoy - bootstrapping and running envoy
   - command/connect/proxy - built-in proxy that is dev-only and not supported 
     for production.
   - `connect/` - "Native" service mesh
7. Cluster membership
