package app

import (
	"gopkg.in/alecthomas/kingpin.v2"
)

var (
	ControlPlaneHostname = ""
)

func DefineControlPlaneFlags(cmd *kingpin.CmdClause) {
	cmd.Flag("hostname", "Control plane hostname to check").
		Envar(configEnvName("CONTROL_PLANE_HOSTNAME")).
		Required().
		StringVar(&ControlPlaneHostname)
}
