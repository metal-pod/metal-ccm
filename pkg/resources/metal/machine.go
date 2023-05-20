package metal

import (
	"context"
	"errors"
	"fmt"
	"strings"
	"time"

	"github.com/metal-stack/metal-ccm/pkg/resources/constants"
	clientset "k8s.io/client-go/kubernetes"

	metalgo "github.com/metal-stack/metal-go"
	"github.com/metal-stack/metal-go/api/client/machine"
	"github.com/metal-stack/metal-go/api/models"

	"github.com/metal-stack/metal-lib/pkg/cache"
	"k8s.io/apimachinery/pkg/types"

	v1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

type MetalService struct {
	client             metalgo.Client
	k8sclient          clientset.Interface
	machineByUUIDCache *cache.Cache[string, *models.V1MachineResponse]
}

func New(client metalgo.Client) *MetalService {
	machineByUUIDCache := cache.New(time.Minute, func(ctx context.Context, id string) (*models.V1MachineResponse, error) {
		machine, err := client.Machine().FindMachine(machine.NewFindMachineParams().WithContext(ctx).WithID(id), nil)
		if err != nil {
			return nil, err
		}

		return machine.Payload, nil
	})
	ms := &MetalService{
		client:             client,
		machineByUUIDCache: machineByUUIDCache,
	}
	return ms
}

// GetMachinesFromNodes gets metal machines from K8s nodes.
func (ms *MetalService) GetMachinesFromNodes(ctx context.Context, nodes []v1.Node) ([]*models.V1MachineResponse, error) {
	var mm []*models.V1MachineResponse
	for _, n := range nodes {
		m, err := ms.GetMachineFromProviderID(ctx, n.Spec.ProviderID)
		if err != nil {
			return nil, err
		}
		mm = append(mm, m)
	}

	return mm, nil
}

// GetMachineFromNode returns a machine where hostname matches the kubernetes node.Name.
func (ms *MetalService) GetMachineFromNode(ctx context.Context, nodeName types.NodeName) (*models.V1MachineResponse, error) {
	node, err := ms.k8sclient.CoreV1().Nodes().Get(ctx, string(nodeName), metav1.GetOptions{})
	if err != nil {
		return nil, err
	}

	id, err := decodeMachineIDFromProviderID(node.Spec.ProviderID)
	if err != nil {
		return nil, err
	}
	return ms.GetMachineFromProviderID(ctx, id)
}

// GetMachineFromProviderID uses providerID to get the machine id and returns the machine.
func (ms *MetalService) GetMachineFromProviderID(ctx context.Context, providerID string) (*models.V1MachineResponse, error) {
	id, err := decodeMachineIDFromProviderID(providerID)
	if err != nil {
		return nil, err
	}

	machine, err := ms.machineByUUIDCache.Get(ctx, id)
	if err != nil {
		return nil, err
	}

	return machine, nil
}

// machineIDFromProviderID returns a machine's ID from providerID.
//
// The providerID spec should be retrievable from the Kubernetes
// node object. The expected format is: metal://partition-id/machine-id.
func decodeMachineIDFromProviderID(providerID string) (string, error) {
	if providerID == "" {
		return "", errors.New("providerID cannot be empty")
	}

	if !strings.HasPrefix(providerID, constants.ProviderName+"://") {
		return "", fmt.Errorf("unexpected providerID format %q, format should be %q or %q", providerID, constants.ProviderName+"://<machine-id>", constants.ProviderName+"://<partition-id>/<machine-id>")
	}

	idparts := strings.Split(providerID, "/")
	return idparts[len(idparts)-1], nil
}
