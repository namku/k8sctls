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

	eksCmd.PersistentFlags().StringP("region", "r", "eu-central-1", "cluster region")
	eksCmd.PersistentFlags().StringP("cluster-name", "n", "", "cluster name")
	eksCmd.MarkPersistentFlagRequired("cluster-name")
	eksCmd.PersistentFlags().String("duration-seconds", "86400", "time in seconds until reset credentials")
	eksCmd.PersistentFlags().StringP("serial-number", "s", "", "arn user name")
	eksCmd.MarkPersistentFlagRequired("serial-number")
	eksCmd.PersistentFlags().StringP("token-code", "t", "", "two factor authentication code")
	eksCmd.MarkPersistentFlagRequired("token-code")
	eksCmd.PersistentFlags().StringP("profile", "p", "", "account profile")
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

		// get flags values
		n, _ := cmd.Flags().GetString("cluster-name")
		r, _ := cmd.Flags().GetString("region")
		ds, _ := cmd.Flags().GetString("duration-seconds")
		sn, _ := cmd.Flags().GetString("serial-number")
		t, _ := cmd.Flags().GetString("token-code")
		p, _ := cmd.Flags().GetString("profile")

		fmt.Println("aws sts get-session-token --duration-seconds " + ds + " --serial-number " + sn + " --token-code " + t + " --profile " + p)
		secretAccessKey, err := exec.Command("aws", "sts", "get-session-token", "--duration-seconds", ds, "--serial-number", sn, "--token-code", t, "--profile", p).Output()
		if err != nil {
			log.Fatal("command failed ", err)
		}

		// read output json format
		var crd Identification
		json.Unmarshal([]byte(secretAccessKey), &crd)

		// set environment vars
		os.Setenv("AWS_SECRET_ACCESS_KEY", crd.Credentials.SecretAccessKey)
		os.Setenv("AWS_SECRET_ACCESS_KEY", crd.Credentials.SessionToken)

		// create context configuration eks cluster
		awsContext := exec.Command("aws", "eks", "--region", r, "update-kubeconfig", "--name", n, "--profile", p)
		err = awsContext.Run()
		if err != nil {
			log.Fatalf("ERROR: problem creating new context %v", err)
		}
	},
}
