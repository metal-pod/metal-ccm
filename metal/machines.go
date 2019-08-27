package metal

import (
	"context"
	"fmt"
	"k8s.io/apimachinery/pkg/util/runtime"
	"log"
	"strings"

	"k8s.io/component-base/logs"

	"github.com/metal-pod/metal-go"
	"github.com/metal-pod/metal-go/api/models"

	"github.com/pkg/errors"

	"k8s.io/api/core/v1"
	"k8s.io/apimachinery/pkg/types"
	"k8s.io/cloud-provider"
)

type machines struct {
	client *metalgo.Driver
	logger *log.Logger
}

func newMachines(client *metalgo.Driver) cloudprovider.Instances {
	logs.InitLogs()
	logger := logs.NewLogger("metal-ccm machines | ")

	return &machines{
		client: client,
		logger: logger,
	}
}

func (m *machines) getMachines(nodes []*v1.Node) []*models.V1MachineResponse {
	var mm []*models.V1MachineResponse
	for _, n := range nodes {
		m, err := machineByHostname(m.client, types.NodeName(n.Name))
		if err != nil {
			runtime.HandleError(err)
			continue
		}
		mm = append(mm, m.Machine)
	}

	return mm
}

// NodeAddresses returns the addresses of the specified instance.
func (m *machines) NodeAddresses(_ context.Context, name types.NodeName) ([]v1.NodeAddress, error) {
	m.logger.Printf("NodeAddresses: nodeName %q", name)
	machine, err := machineByHostname(m.client, name)
	if err != nil {
		return nil, err
	}

	return nodeAddresses(machine)
}

// NodeAddressesByProviderID returns the addresses of the specified instance.
// The instance is specified using the providerID of the node. The
// ProviderID is a unique identifier of the node. This will not be called
// from the node whose node addresses are being queried. m.e. local metadata
// services cannot be used in this method to obtain node addresses.
func (m *machines) NodeAddressesByProviderID(_ context.Context, providerID string) ([]v1.NodeAddress, error) {
	m.logger.Printf("NodeAddressesByProviderID: providerID %q", providerID)
	machine, err := m.machineFromProviderID(providerID)
	if err != nil {
		return nil, err
	}

	return nodeAddresses(machine)
}

func nodeAddresses(machine *metalgo.MachineGetResponse) ([]v1.NodeAddress, error) {
	var addresses []v1.NodeAddress
	for _, nw := range machine.Machine.Allocation.Networks {
		if nw == nil || (nw.Private != nil && *nw.Private) {
			addresses = append(addresses, v1.NodeAddress{Type: v1.NodeHostName, Address: nw.Ips[0]})
			addresses = append(addresses, v1.NodeAddress{Type: v1.NodeInternalIP, Address: nw.Ips[0]})
		} else {
			for _, ip := range nw.Ips {
				addresses = append(addresses, v1.NodeAddress{Type: v1.NodeExternalIP, Address: ip})
			}
		}
	}
	return addresses, nil
}

// InstanceID returns the cloud provider ID of the node with the specified NodeName.
// Note that if the instance does not exist or is no longer running, we must return ("", cloudprovider.InstanceNotFound).
func (m *machines) InstanceID(_ context.Context, nodeName types.NodeName) (string, error) {
	m.logger.Printf("InstanceID: nodeName %q", nodeName)
	machine, err := machineByHostname(m.client, nodeName)
	if err != nil {
		return "", err
	}

	return *machine.Machine.ID, nil
}

// InstanceType returns the type of the specified instance.
func (m *machines) InstanceType(_ context.Context, nodeName types.NodeName) (string, error) {
	m.logger.Printf("InstanceType: nodeName %q", nodeName)
	machine, err := machineByHostname(m.client, nodeName)
	if err != nil {
		return "", err
	}

	return *machine.Machine.Size.ID, nil
}

// InstanceTypeByProviderID returns the type of the specified instance.
func (m *machines) InstanceTypeByProviderID(_ context.Context, providerID string) (string, error) {
	m.logger.Printf("InstanceTypeByProviderID: providerID %q", providerID)
	machine, err := m.machineFromProviderID(providerID)
	if err != nil {
		return "", err
	}

	return *machine.Machine.Size.ID, nil
}

// AddSSHKeyToAllInstances adds an SSH public key as a legal identity for all machines.
// Expected format for the key is standard ssh-keygen format: <protocol> <blob>.
func (m *machines) AddSSHKeyToAllInstances(_ context.Context, user string, keyData []byte) error {
	return cloudprovider.NotImplemented
}

// CurrentNodeName returns the name of the node we are currently running on.
// On most clouds (e.g. GCE) this is the hostname, so we provide the hostname.
func (m *machines) CurrentNodeName(_ context.Context, nodeName string) (types.NodeName, error) {
	m.logger.Printf("CurrentNodeName: nodeName %q", nodeName)
	return types.NodeName(nodeName), nil
}

// InstanceExistsByProviderID returns true if the instance for the given provider exists.
// If false is returned with no error, the instance will be immediately deleted by the cloud controller manager.
// This method should still return true for machines that exist but are stopped/sleeping.
func (m *machines) InstanceExistsByProviderID(_ context.Context, providerID string) (bool, error) {
	m.logger.Printf("InstanceExistsByProviderID: providerID %q", providerID)
	machine, err := m.machineFromProviderID(providerID)
	if err != nil {
		return false, err
	}

	return machine.Machine.Allocation != nil, nil
}

// InstanceShutdownByProviderID returns true if the instance is shutdown in cloudprovider.
func (m *machines) InstanceShutdownByProviderID(_ context.Context, providerID string) (bool, error) {
	m.logger.Printf("InstanceShutdownByProviderID: providerID %q", providerID)
	if true { //TODO Remove
		return true, nil
	}
	machine, err := m.machineFromProviderID(providerID)
	if err != nil || machine.Machine.Allocation == nil {
		return true, err
	}
	if machine.Machine.Events != nil && len(machine.Machine.Events.Log) > 0 {
		lastEvent := machine.Machine.Events.Log[0].Event
		return *lastEvent != "Phoned Home", nil
	}
	return true, nil
}

func machineByID(client *metalgo.Driver, id string) (*metalgo.MachineGetResponse, error) {
	if true { //TODO Remove
		return createDummyMachine(id), nil
	}
	machine, err := client.MachineGet(id)
	if err != nil {
		return nil, err
	}

	return machine, nil
}

// machineByHostname returns an instance where hostname matches the kubernetes node.Name.
func machineByHostname(client *metalgo.Driver, nodeName types.NodeName) (*metalgo.MachineGetResponse, error) {
	machineHostname := string(nodeName)
	if true { //TODO Remove
		if strings.HasPrefix(machineHostname, "kind-worker") {
			return createDummyMachine(machineHostname), nil
		}
	}
	mfr := &metalgo.MachineFindRequest{
		AllocationHostname: &machineHostname,
	}
	machines, err := client.MachineFind(mfr)
	if err != nil {
		return nil, err
	}
	if len(machines.Machines) == 0 {
		return nil, fmt.Errorf("no machine with name:%s found", nodeName)
	}
	if len(machines.Machines) > 1 {
		return nil, fmt.Errorf("more than one (%d) machine with name:%s found", len(machines.Machines), nodeName)
	}

	return &metalgo.MachineGetResponse{
		Machine: machines.Machines[0],
	}, nil
}

// machineIDFromProviderID returns a machine's ID from providerID.
//
// The providerID spec should be retrievable from the Kubernetes
// node object. The expected format is: metal://machine-id.
func machineIDFromProviderID(providerID string) (string, error) {
	if providerID == "" {
		return "", errors.New("providerID cannot be empty string")
	}

	split := strings.Split(providerID, "://")
	if len(split) != 2 {
		return "", errors.Errorf("unexpected providerID format: %s, format should be: metal://machine-id", providerID)
	}

	if split[0] != providerName {
		return "", errors.Errorf("provider name from providerID should be metal: %s", providerID)
	}

	return split[1], nil
}

// machineFromProviderID uses providerID to get the machine id and return the machine.
func (m *machines) machineFromProviderID(providerID string) (*metalgo.MachineGetResponse, error) {
	id, err := machineIDFromProviderID(providerID)
	if err != nil {
		return nil, err
	}

	return machineByID(m.client, id)
}

func createDummyMachine(machineHostname string) *metalgo.MachineGetResponse { //TODO Remove
	network := "internet-nbg-w8101"
	asn := int64(4711)
	ip := "10.10.0.1"
	if strings.HasSuffix(machineHostname, "2") {
		ip = "10.10.0.2"
	}
	sizeID := "test"
	return &metalgo.MachineGetResponse{
		Machine: &models.V1MachineResponse{
			ID: &machineHostname,
			Size: &models.V1SizeResponse{
				ID: &sizeID,
			},
			Allocation: &models.V1MachineAllocation{
				Hostname: &machineHostname,
				Name:     &machineHostname,
				Networks: []*models.V1MachineNetwork{
					{
						Networkid: &network,
						Ips: []string{
							ip,
						},
						Asn: &asn,
					},
				},
			},
		},
	}
}
