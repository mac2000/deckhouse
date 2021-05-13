package app

import (
	"time"

	"gopkg.in/alecthomas/kingpin.v2"
)

var (
	AppName        = "deckhouse"
	AppDescription = ""
)

var (
	PodName       = ""
	ContainerName = "deckhouse"
)

var (
	FeatureWatchRegistry               = "yes"
	InsecureRegistry                   = "no"
	SkipTLSVerifyRegistry              = "no"
	RegistrySecretPath                 = "/etc/registrysecret"
	RegistryErrorsMaxTimeBeforeRestart = time.Hour
)

const (
	DeckhouseLogTypeDefault         = "json"
	DeckhouseKubeClientQPSDefault   = "20"
	DeckhouseKubeClientBurstDefault = "40"

	DeckhouseHookMetricsListenPort = "9651"
)

func DefineStartCommandFlags(cmd *kingpin.CmdClause) {
	cmd.Flag("pod-name", "Pod name to get image digest.").
		Envar("DECKHOUSE_POD").
		Required().
		StringVar(&PodName)
	cmd.Flag("feature-watch-registry", "Enable docker registry watcher (yes|no).").
		Envar("DECKHOUSE_WATCH_REGISTRY").
		Default(FeatureWatchRegistry).
		StringVar(&FeatureWatchRegistry)
	cmd.Flag("insecure-registry", "Use http to access registry (yes|no).").
		Envar("DECKHOUSE_INSECURE_REGISTRY").
		Default(InsecureRegistry).
		StringVar(&InsecureRegistry)
	cmd.Flag("skip-tls-verify-registry", "Trust self signed certificate of registry (yes|no).").
		Envar("DECKHOUSE_SKIP_TLS_VERIFY_REGISTRY").
		Default(SkipTLSVerifyRegistry).
		StringVar(&SkipTLSVerifyRegistry)
}
