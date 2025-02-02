package config

import (
	"fmt"
	"testing"

	"github.com/google/go-cmp/cmp"
	"github.com/google/go-cmp/cmp/cmpopts"
	"github.com/metal-stack/metal-go/api/models"
	"github.com/metal-stack/metal-lib/pkg/pointer"
	"github.com/metal-stack/metal-lib/pkg/tag"
	v1 "k8s.io/api/core/v1"
	"k8s.io/apimachinery/pkg/util/sets"
)

func TestCiliumConfig(t *testing.T) {
	tests := []struct {
		name    string
		nws     sets.Set[string]
		ips     []*models.V1IPResponse
		nodes   []v1.Node
		wantErr error
		want    *ciliumConfig
	}{
		{
			name: "one ip acquired, no nodes",
			nws:  testNetworks,
			ips: []*models.V1IPResponse{
				{
					Ipaddress: pointer.Pointer("84.1.1.1"),
					Name:      "acquired-before",
					Networkid: pointer.Pointer("internet"),
					Projectid: pointer.Pointer("project-a"),
					Tags: []string{
						fmt.Sprintf("%s=%s", tag.ClusterID, "this-cluster"),
					},
					Type: pointer.Pointer("ephemeral"),
				},
			},
			nodes:   []v1.Node{},
			wantErr: nil,
			want: &ciliumConfig{
				base: &baseConfig{
					AddressPools: addressPools{
						"internet-ephemeral": {
							Name:       "internet-ephemeral",
							Protocol:   "bgp",
							AutoAssign: pointer.Pointer(false),
							CIDRs: []string{
								"84.1.1.1/32",
							},
						},
					},
					Peers: []*peer{},
				},
			},
		},
		{
			name: "two ips acquired, no nodes",
			nws:  testNetworks,
			ips: []*models.V1IPResponse{
				{
					Ipaddress: pointer.Pointer("84.1.1.1"),
					Name:      "acquired-before",
					Networkid: pointer.Pointer("internet"),
					Projectid: pointer.Pointer("project-a"),
					Tags: []string{
						fmt.Sprintf("%s=%s", tag.ClusterID, "this-cluster"),
					},
					Type: pointer.Pointer("ephemeral"),
				},
				{
					Ipaddress: pointer.Pointer("84.1.1.2"),
					Name:      "acquired-before-2",
					Networkid: pointer.Pointer("internet"),
					Projectid: pointer.Pointer("project-a"),
					Tags: []string{
						fmt.Sprintf("%s=%s", tag.ClusterID, "this-cluster"),
					},
					Type: pointer.Pointer("ephemeral"),
				},
			},
			nodes:   []v1.Node{},
			wantErr: nil,
			want: &ciliumConfig{
				base: &baseConfig{
					AddressPools: addressPools{
						"internet-ephemeral": {
							Name:       "internet-ephemeral",
							Protocol:   "bgp",
							AutoAssign: pointer.Pointer(false),
							CIDRs: []string{
								"84.1.1.1/32",
								"84.1.1.2/32",
							},
						},
					},
					Peers: []*peer{},
				},
			},
		},
		{
			name: "two ips acquired, one static ip, no nodes",
			nws:  testNetworks,
			ips: []*models.V1IPResponse{
				{
					Ipaddress: pointer.Pointer("84.1.1.1"),
					Name:      "acquired-before",
					Networkid: pointer.Pointer("internet"),
					Projectid: pointer.Pointer("project-a"),
					Tags: []string{
						fmt.Sprintf("%s=%s", tag.ClusterID, "this-cluster"),
					},
					Type: pointer.Pointer("ephemeral"),
				},
				{
					Ipaddress: pointer.Pointer("84.1.1.2"),
					Name:      "acquired-before-2",
					Networkid: pointer.Pointer("internet"),
					Projectid: pointer.Pointer("project-a"),
					Tags: []string{
						fmt.Sprintf("%s=%s", tag.ClusterID, "this-cluster"),
					},
					Type: pointer.Pointer("ephemeral"),
				},
				{
					Ipaddress: pointer.Pointer("84.1.1.3"),
					Name:      "static-ip",
					Networkid: pointer.Pointer("internet"),
					Projectid: pointer.Pointer("project-a"),
					Tags: []string{
						fmt.Sprintf("%s=%s", tag.ClusterID, "this-cluster"),
					},
					Type: pointer.Pointer("static"),
				},
			},
			nodes:   []v1.Node{},
			wantErr: nil,
			want: &ciliumConfig{
				base: &baseConfig{
					AddressPools: addressPools{
						"internet-ephemeral": {
							Name:       "internet-ephemeral",
							Protocol:   "bgp",
							AutoAssign: pointer.Pointer(false),
							CIDRs: []string{
								"84.1.1.1/32",
								"84.1.1.2/32",
							},
						},
						"internet-static": {
							Name:       "internet-static",
							Protocol:   "bgp",
							AutoAssign: pointer.Pointer(false),
							CIDRs: []string{
								"84.1.1.3/32",
							},
						},
					},
					Peers: []*peer{},
				},
			},
		},
		{
			name: "connected to internet,storage,dmz and mpls, two ips acquired, one static ip, no nodes",
			nws:  testNetworks,
			ips: []*models.V1IPResponse{
				{
					Ipaddress: pointer.Pointer("84.1.1.1"),
					Name:      "acquired-before",
					Networkid: pointer.Pointer("internet"),
					Projectid: pointer.Pointer("project-a"),
					Tags: []string{
						fmt.Sprintf("%s=%s", tag.ClusterID, "this-cluster"),
					},
					Type: pointer.Pointer("ephemeral"),
				},
				{
					Ipaddress: pointer.Pointer("84.1.1.2"),
					Name:      "acquired-before-2",
					Networkid: pointer.Pointer("internet"),
					Projectid: pointer.Pointer("project-a"),
					Tags: []string{
						fmt.Sprintf("%s=%s", tag.ClusterID, "this-cluster"),
					},
					Type: pointer.Pointer("ephemeral"),
				},
				{
					Ipaddress: pointer.Pointer("84.1.1.3"),
					Name:      "static-ip",
					Networkid: pointer.Pointer("internet"),
					Projectid: pointer.Pointer("project-a"),
					Tags: []string{
						fmt.Sprintf("%s=%s", tag.ClusterID, "this-cluster"),
					},
					Type: pointer.Pointer("static"),
				},
				{
					Ipaddress: pointer.Pointer("10.131.44.2"),
					Name:      "static-ip",
					Networkid: pointer.Pointer("shared-storage-network"),
					Projectid: pointer.Pointer("project-a"),
					Tags: []string{
						fmt.Sprintf("%s=%s", tag.ClusterID, "this-cluster"),
					},
					Type: pointer.Pointer("static"),
				},
				{
					Ipaddress: pointer.Pointer("100.127.130.2"),
					Name:      "static-ip",
					Networkid: pointer.Pointer("mpls-network"),
					Projectid: pointer.Pointer("project-a"),
					Tags: []string{
						fmt.Sprintf("%s=%s", tag.ClusterID, "this-cluster"),
					},
					Type: pointer.Pointer("static"),
				},
				{
					Ipaddress: pointer.Pointer("100.127.130.3"),
					Name:      "ephemeral-mpls-ip",
					Networkid: pointer.Pointer("mpls-network"),
					Projectid: pointer.Pointer("project-a"),
					Tags: []string{
						fmt.Sprintf("%s=%s", tag.ClusterID, "this-cluster"),
					},
					Type: pointer.Pointer("ephemeral"),
				},
				{
					Ipaddress: pointer.Pointer("10.129.172.2"),
					Name:      "static-ip",
					Networkid: pointer.Pointer("dmz-network"),
					Projectid: pointer.Pointer("project-a"),
					Tags: []string{
						fmt.Sprintf("%s=%s", tag.ClusterID, "this-cluster"),
					},
					Type: pointer.Pointer("static"),
				},
			},
			nodes:   []v1.Node{},
			wantErr: nil,
			want: &ciliumConfig{
				base: &baseConfig{
					AddressPools: addressPools{
						"internet-ephemeral": {
							Name:       "internet-ephemeral",
							Protocol:   "bgp",
							AutoAssign: pointer.Pointer(false),
							CIDRs: []string{
								"84.1.1.1/32",
								"84.1.1.2/32",
							},
						},
						"internet-static": {
							Name:       "internet-static",
							Protocol:   "bgp",
							AutoAssign: pointer.Pointer(false),
							CIDRs: []string{
								"84.1.1.3/32",
							},
						},
						"shared-storage-network-static": {
							Name:       "shared-storage-network-static",
							Protocol:   "bgp",
							AutoAssign: pointer.Pointer(false),
							CIDRs: []string{
								"10.131.44.2/32",
							},
						},
						"mpls-network-static": {
							Name:       "mpls-network-static",
							Protocol:   "bgp",
							AutoAssign: pointer.Pointer(false),
							CIDRs: []string{
								"100.127.130.2/32",
							},
						},
						"mpls-network-ephemeral": {
							Name:       "mpls-network-ephemeral",
							Protocol:   "bgp",
							AutoAssign: pointer.Pointer(false),
							CIDRs: []string{
								"100.127.130.3/32",
							},
						},
						"dmz-network-static": {
							Name:       "dmz-network-static",
							Protocol:   "bgp",
							AutoAssign: pointer.Pointer(false),
							CIDRs: []string{
								"10.129.172.2/32",
							},
						},
					},
					Peers: []*peer{},
				},
			}},
	}
	for _, tt := range tests {
		tt := tt
		t.Run(tt.name, func(t *testing.T) {
			cfg, err := New(LoadBalancerTypeCilium, tt.ips, tt.nws, tt.nodes, nil, nil)
			if diff := cmp.Diff(err, tt.wantErr); diff != "" {
				t.Errorf("error = %v", diff)
				return
			}

			if diff := cmp.Diff(cfg, tt.want, cmpopts.IgnoreUnexported(ciliumConfig{})); diff != "" {
				t.Errorf("diff = %v", diff)
			}
		})
	}
}
