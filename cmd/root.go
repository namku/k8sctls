// root.go
package cmd

import (
	"github.com/spf13/cobra"
)

var rootCmd = &cobra.Command{
	Use:   "k8sctls",
	Short: "k8sctls import credentials in ~/.kube/config",
	Long: `Configure k8s credentials to use kubectl
	Complete documentation is available at http://github.com/namku/k8sctls`,
	// Run: func(cmd *cobra.Command, args []string) {
	// 	// Do Stuff Here
	// 	fmt.Println("cli for import credentials of k8s")
	// },
}

func Execute() error {
	return rootCmd.Execute()
}
