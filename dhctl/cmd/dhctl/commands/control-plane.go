package commands

import (
	"fmt"

	"gopkg.in/alecthomas/kingpin.v2"

	"github.com/deckhouse/deckhouse/dhctl/pkg/app"
	"github.com/deckhouse/deckhouse/dhctl/pkg/kubernetes/client"
	"github.com/deckhouse/deckhouse/dhctl/pkg/log"
	"github.com/deckhouse/deckhouse/dhctl/pkg/operations/ready/control_plane"
	"github.com/deckhouse/deckhouse/dhctl/pkg/system/ssh"
)

func DefineTestControlPlaneManagerReadyCommand(parent *kingpin.CmdClause) *kingpin.CmdClause {
	cmd := parent.Command("manager", "Test control plane manager is ready.")
	app.DefineSSHFlags(cmd)
	app.DefineBecomeFlags(cmd)
	app.DefineKubeFlags(cmd)
	app.DefineControlPlaneFlags(cmd)

	cmd.Action(func(c *kingpin.ParseContext) error {
		sshClient, err := ssh.NewInitClientFromFlags(true)
		if err != nil {
			return err
		}

		kubeCl := client.NewKubernetesClient().WithSSHClient(sshClient)
		// auto init
		err = kubeCl.Init(client.AppKubernetesInitParams())
		if err != nil {
			return fmt.Errorf("open kubernetes connection: %v", err)
		}

		checker := control_plane.NewControlPlaneManagerReadinessChecker(app.ControlPlaneHostname, kubeCl)
		ready, err := checker.IsReady()
		if err != nil {
			return fmt.Errorf("Control plane manager is not ready: %s", err)
		}

		if ready {
			log.InfoLn("Control plane manager is ready")
		} else {
			log.WarnLn("Control plane manager is not ready")
		}

		return nil
	})
	return cmd
}
