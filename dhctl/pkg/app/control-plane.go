package app

import (
	"gopkg.in/alecthomas/kingpin.v2"
)

var (
	ControlPlaneHostname = ""
	ControlPlaneIP       = ""
)

func DefineControlPlaneFlags(cmd *kingpin.CmdClause, ipRequired bool) {
	cmd.Flag("control-plane-node-hostname", "Control plane node hostname to check").
		Envar(configEnvName("CONTROL_PLANE_NODE_HOSTNAME")).
		Required().
		StringVar(&ControlPlaneHostname)

	ipFlag := cmd.Flag("control-plane-node-ip", "Control plane node ip to check").
		Envar(configEnvName("CONTROL_PLANE_NODE_IP"))

	if ipRequired {
		ipFlag.Required()
	}

	ipFlag.StringVar(&ControlPlaneIP)
}
