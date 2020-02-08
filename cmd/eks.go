// eks.go
package cmd

import (
	"encoding/json"
	"fmt"
	"log"
	"os"
	"os/exec"

	"github.com/spf13/cobra"
)

func init() {
	rootCmd.AddCommand(eksCmd)

	eksCmd.PersistentFlags().String("duration-seconds", "86400", "Time in seconds until reset credentials")
	eksCmd.PersistentFlags().String("serial-number", "", "arn user name")
	eksCmd.MarkPersistentFlagRequired("serial-number")
	eksCmd.PersistentFlags().StringP("token-code", "t", "", "two factor authentication code")
	eksCmd.MarkPersistentFlagRequired("token-code")
	eksCmd.PersistentFlags().String("profile", "", "account profile")
	eksCmd.MarkPersistentFlagRequired("profile")
}

// struct credentials
type Credentials struct {
	AcessKeyId      string
	SecretAccessKey string
	SessionToken    string
	Expiration      string
}

// struct identification
type Identification struct {
	Credentials Credentials
}

var eksCmd = &cobra.Command{
	Use:   "eks",
	Short: "Configure credentials of eks cluster",
	Long:  `Import configuration of eks cluster in .kube/config file.`,
	Run: func(cmd *cobra.Command, args []string) {

		ds, _ := cmd.Flags().GetString("duration-seconds")
		sn, _ := cmd.Flags().GetString("serial-number")
		tc, _ := cmd.Flags().GetString("token-code")
		pf, _ := cmd.Flags().GetString("profile")

		fmt.Println("aws sts get-session-token --duration-seconds " + ds + " --serial-number " + sn + " --token-code " + tc + " --profile " + pf)
		secretAccessKey, err := exec.Command("aws", "sts", "get-session-token", "--duration-seconds", ds, "--serial-number", sn, "--token-code", tc, "--profile", pf).Output()
		if err != nil {
			log.Fatal("command failed", err)
		}

		// read output json format
		var crd Identification
		json.Unmarshal([]byte(secretAccessKey), &crd)

		fmt.Printf("%+v", crd.Credentials.SecretAccessKey)
		fmt.Printf("%+v", crd.Credentials.SessionToken)

		// set environment vars
		os.Setenv("AWS_SECRET_ACCESS_KEY", crd.Credentials.SecretAccessKey)
		os.Setenv("AWS_SECRET_ACCESS_KEY", crd.Credentials.SessionToken)

		// create context configuration eks cluster
		awsContext := exec.Command("aws", "eks", "--region", "eu-central-1", "update-kubeconfig", "--name", "koble-stg-eks", "--profile", "nexus")
		err = awsContext.Run()
		if err != nil {
			log.Fatal(err)
		}
	},
}
