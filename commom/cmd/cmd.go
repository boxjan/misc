package cmd

import (
	"flag"
	"fmt"
	"github.com/spf13/cobra"
	"k8s.io/klog/v2"
	"runtime"
)

var (
	buildTime = "just now"
	version   = "dev"

	ConfigPath string
	NewConfig  bool
)

func QuickCobraRun(name string, run func(cmd *cobra.Command, args []string)) *cobra.Command {
	klogFlags := flag.NewFlagSet("klog", flag.PanicOnError)
	klog.InitFlags(klogFlags)

	cmd := &cobra.Command{
		Use:     name,
		Version: fmt.Sprintf("%s@[%s]%s", version, runtime.Version(), buildTime),
		Run:     run,
	}

	flags := cmd.PersistentFlags()

	flags.AddGoFlagSet(klogFlags)
	flags.StringVarP(&ConfigPath, "config", "c", "", "config path")
	flags.BoolVarP(&NewConfig, "new-config", "", false, "generate empty config in given path")

	return cmd
}
