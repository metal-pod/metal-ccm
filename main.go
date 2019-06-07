package main

import (
	goflag "flag"
	"fmt"
	"math/rand"
	"os"
	"time"

	utilflag "k8s.io/apiserver/pkg/util/flag"
	"k8s.io/apiserver/pkg/util/logs"
	"k8s.io/kubernetes/cmd/cloud-controller-manager/app"
	_ "k8s.io/kubernetes/pkg/client/metrics/prometheus" // for client metric registration
	_ "k8s.io/kubernetes/pkg/version/prometheus"        // for version metric registration

	_ "github.com/metal-pod/metal-ccm/metal"
	"github.com/metal-pod/v"
	"github.com/spf13/pflag"
)

func main() {
	rand.Seed(time.Now().UTC().UnixNano())

	command := app.NewCloudControllerManagerCommand()

	pflag.CommandLine.SetNormalizeFunc(utilflag.WordSepNormalizeFunc)
	pflag.CommandLine.AddGoFlagSet(goflag.CommandLine)

	logs.InitLogs()
	defer logs.FlushLogs()
	logger := logs.NewLogger("metal-ccm")
	logger.Printf(" starting version:%s", v.V)

	if err := command.Execute(); err != nil {
		fmt.Fprintf(os.Stderr, "error: %v\n", err)
		os.Exit(1)
	}
}
