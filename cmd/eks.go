// eks.go
package cmd

import (
	"encoding/json"
	"fmt"
	"log"
	"os"
	"os/exec"

	"github.com/spf13/cobra"
	"github.com/spf13/viper"
)

func init() {
	cobra.OnInitialize(initConfig)

	eksCmd.PersistentFlags().StringP("region", "r", "eu-central-1", "cluster region")
	eksCmd.PersistentFlags().StringP("cluster-name", "n", "", "cluster name")
	eksCmd.MarkPersistentFlagRequired("cluster-name")
	eksCmd.PersistentFlags().StringP("serial-number", "s", "", "arn user name")
	eksCmd.MarkPersistentFlagRequired("serial-number")
	eksCmd.PersistentFlags().StringP("token-code", "t", "", "two factor authentication code")
	eksCmd.MarkPersistentFlagRequired("token-code")
	eksCmd.PersistentFlags().StringP("profile", "p", "", "account profile")
	viper.BindPFlag("profile", eksCmd.PersistentFlags().Lookup("profile"))
	// eksCmd.MarkPersistentFlagRequired("profile")

	rootCmd.AddCommand(eksCmd)
}

func initConfig() {
	println("In initConfig")
	cfgFile := ".config.yml"
	if cfgFile != "" {
		// Use config file from the flag.
		viper.SetConfigFile(cfgFile)
	} //else {
	// 	// Find home directory.
	// 	home, err := homedir.Dir()
	// 	if err != nil {
	// 		er(err)
	// 	}

	// 	// Search config in home directory with name ".cobra" (without extension).
	// 	viper.AddConfigPath(home)
	// 	viper.SetConfigName(".cobra")
	// }

	viper.AutomaticEnv()

	if err := viper.ReadInConfig(); err == nil {
		fmt.Println("Using config file:", viper.ConfigFileUsed())
	} else {
		fmt.Printf("Error reading config file: %v\n", err)
	}
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
		sn, _ := cmd.Flags().GetString("serial-number")
		t, _ := cmd.Flags().GetString("token-code")
		p, _ := cmd.Flags().GetString("profile")
		if p == "" {
			p = viper.GetString("profile")
		}

		sessionToken_, err := exec.Command("aws", "sts", "get-session-token", "--serial-number", sn, "--token-code", t, "--profile", p).Output()

		if err != nil {
			log.Fatal("command failed ", err)
		}

		// read output json format
		var crd Identification
		json.Unmarshal([]byte(sessionToken_), &crd)

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
