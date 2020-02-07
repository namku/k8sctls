// eks.go
package cmd

import (
	"fmt"
	"log"
	"os/exec"

	"github.com/spf13/cobra"
)

func init() {
	rootCmd.AddCommand(eksCmd)
}

var eksCmd = &cobra.Command{
	Use:   "eks",
	Short: "Configure credentials of eks cluster",
	Long:  `Import configuration of eks cluster in .kube/config file.`,
	Run: func(cmd *cobra.Command, args []string) {
		fmt.Println("Importing credentials..")
		// aws sts get-session-token --duration-seconds $(durationSeconds) --serial-number $(serialNumber)  --token-code $(tokenCode) --profile $(profile)
		list, err := exec.Command("aws", "sts", "get-session-token", "--duration-seconds", "86400").Output()
		// err := list.Run()
		if err != nil {
			log.Fatal("command failed", err)
		}
		fmt.Printf("%s", list)
	},
}
