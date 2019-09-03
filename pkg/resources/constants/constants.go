package constants

const (
	MetalAPIUrlEnvVar      = "METAL_API_URL"
	MetalAuthTokenEnvVar   = "METAL_AUTH_TOKEN"
	MetalAuthHMACEnvVar    = "METAL_AUTH_HMAC"
	MetalProjectIDEnvVar   = "METAL_PROJECT_ID"
	MetalPartitionIDEnvVar = "METAL_PARTITION_ID"
	MetalNetworkIDEnvVar   = "METAL_NETWORK_ID"

	ProviderName = "metal"

	ASNNodeLabel = "machine.metal-pod.io/network.primary.asn"

	CalicoIPTunnelAddr = "projectcalico.org/IPv4IPIPTunnelAddr"

	MetalLBSpecificAddressPool = "metallb.universe.tf/address-pool"

	IPPrefix = "metallb-"
)
