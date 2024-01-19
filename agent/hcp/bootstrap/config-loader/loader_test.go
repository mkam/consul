// Copyright (c) HashiCorp, Inc.
// SPDX-License-Identifier: BUSL-1.1

package loader

import (
	"os"
	"path/filepath"
	"testing"

	"github.com/hashicorp/consul/agent/config"
	"github.com/hashicorp/consul/agent/hcp/bootstrap"
	"github.com/hashicorp/consul/lib"
	"github.com/stretchr/testify/require"
)

func TestBootstrapConfigLoader(t *testing.T) {
	baseLoader := func(source config.Source) (config.LoadResult, error) {
		return config.Load(config.LoadOpts{
			DefaultConfig: source,
			HCL: []string{
				`server = true`,
				`bind_addr = "127.0.0.1"`,
				`data_dir = "/tmp/consul-data"`,
			},
		})
	}

	bootstrapLoader := func(source config.Source) (config.LoadResult, error) {
		return bootstrapConfigLoader(baseLoader, &bootstrap.RawBootstrapConfig{
			ConfigJSON:      `{"bootstrap_expect": 8}`,
			ManagementToken: "test-token",
		})(source)
	}

	result, err := bootstrapLoader(nil)
	require.NoError(t, err)

	// bootstrap_expect and management token are injected from bootstrap config received from HCP.
	require.Equal(t, 8, result.RuntimeConfig.BootstrapExpect)
	require.Equal(t, "test-token", result.RuntimeConfig.Cloud.ManagementToken)

	// Response header is always injected from a constant.
	require.Equal(t, "x-consul-default-acl-policy", result.RuntimeConfig.HTTPResponseHeaders[accessControlHeaderName])
}

func Test_finalizeRuntimeConfig(t *testing.T) {
	type testCase struct {
		rc       *config.RuntimeConfig
		cfg      *bootstrap.RawBootstrapConfig
		verifyFn func(t *testing.T, rc *config.RuntimeConfig)
	}
	run := func(t *testing.T, tc testCase) {
		finalizeRuntimeConfig(tc.rc, tc.cfg)
		tc.verifyFn(t, tc.rc)
	}

	tt := map[string]testCase{
		"set header if not present": {
			rc: &config.RuntimeConfig{},
			cfg: &bootstrap.RawBootstrapConfig{
				ManagementToken: "test-token",
			},
			verifyFn: func(t *testing.T, rc *config.RuntimeConfig) {
				require.Equal(t, "test-token", rc.Cloud.ManagementToken)
				require.Equal(t, "x-consul-default-acl-policy", rc.HTTPResponseHeaders[accessControlHeaderName])
			},
		},
		"append to header if present": {
			rc: &config.RuntimeConfig{
				HTTPResponseHeaders: map[string]string{
					accessControlHeaderName: "Content-Encoding",
				},
			},
			cfg: &bootstrap.RawBootstrapConfig{
				ManagementToken: "test-token",
			},
			verifyFn: func(t *testing.T, rc *config.RuntimeConfig) {
				require.Equal(t, "test-token", rc.Cloud.ManagementToken)
				require.Equal(t, "Content-Encoding,x-consul-default-acl-policy", rc.HTTPResponseHeaders[accessControlHeaderName])
			},
		},
	}

	for name, tc := range tt {
		t.Run(name, func(t *testing.T) {
			run(t, tc)
		})
	}
}

func TestValidatePersistedConfig(t *testing.T) {
	type testCase struct {
		configContents string
		expectErr      string
	}

	run := func(t *testing.T, tc testCase) {
		dataDir, err := os.MkdirTemp(os.TempDir(), "load-bootstrap-test-")
		require.NoError(t, err)
		t.Cleanup(func() { os.RemoveAll(dataDir) })

		dir := filepath.Join(dataDir, bootstrap.SubDir)
		require.NoError(t, lib.EnsurePath(dir, true))

		if tc.configContents != "" {
			name := filepath.Join(dir, bootstrap.ConfigFileName)
			require.NoError(t, os.WriteFile(name, []byte(tc.configContents), 0600))
		}

		err = validatePersistedConfig(dataDir)
		if tc.expectErr != "" {
			require.Error(t, err)
			require.Contains(t, err.Error(), tc.expectErr)
		} else {
			require.NoError(t, err)
		}
	}

	tt := map[string]testCase{
		"valid": {
			configContents: `{"bootstrap_expect": 1, "cloud": {"resource_id": "id"}}`,
		},
		"invalid config key": {
			configContents: `{"not_a_consul_agent_config_field": "zap"}`,
			expectErr:      "invalid config key not_a_consul_agent_config_field",
		},
		"invalid format": {
			configContents: `{"not_json" = "invalid"}`,
			expectErr:      "invalid character '=' after object key",
		},
		"missing configuration file": {
			expectErr: "no such file or directory",
		},
	}

	for name, tc := range tt {
		t.Run(name, func(t *testing.T) {
			run(t, tc)
		})
	}
}
