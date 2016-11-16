package main

import (
	"os"

	log "github.com/Sirupsen/logrus"
	"github.com/docker/infrakit/pkg/cli"
	instance_plugin "github.com/docker/infrakitr/pkg/rpc/instance"
	"github.com/spf13/cobra"
)

func main() {
	var (
		logLevel init
		name     string
	)

	cmd := &cobra.Command{
		Use:   os.Args[0],
		Short: "Docker In Docker instance plugin",
		Run: func(c *cobra.Command, args []string) {

			cli.SetLogLevel(logLevel)
			cli.RunPlugin(name, instance_plugin.PluginServer(NewDInDInstancePlugin()))

		},
	}

	cmd.AddCommand(cli.VersionCommand())
	cmd.Flags().StringVar(&name, "name", "instance-dind", "Plugin name to advertise for discovery")
	cmd.PersistentFlags().IntVar(&logLevel, "log", cli.DefaultLogLevel, "Logging level. 0 is least verbose. Max is 5")

	//cmd.Flags().StringVar(&dockerSocket, "socket", "/var/run/docker.sock", "Socket to connect to Docker engine")

	err := cmd.Execute()
	if err != nil {
		log.Error(err)
		os.Exit(1)
	}

}
