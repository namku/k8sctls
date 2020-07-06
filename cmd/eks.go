// eks.go
package cmd

import (
	"encoding/json"
	"log"
	"os"
	"os/exec"

	"github.com/mitchellh/go-homedir"
	"github.com/namku/k8sctls/cmd/dialog"
	"github.com/spf13/cobra"
	"github.com/spf13/viper"
)

func init() {
	cobra.OnInitialize(initConfig)

	eksCmd.PersistentFlags().StringP("region", "r", "", "cluster region")
	eksCmd.PersistentFlags().StringP("cluster-name", "n", "", "cluster name")
	eksCmd.PersistentFlags().StringP("serial-number", "s", "", "arn user name")
	eksCmd.PersistentFlags().StringP("token-code", "t", "", "two factor authentication code (MFA)")
	eksCmd.PersistentFlags().StringP("profile", "p", "", "account profile")
	viper.BindPFlag("region", eksCmd.PersistentFlags().Lookup("region"))
	viper.BindPFlag("cluster-name", eksCmd.PersistentFlags().Lookup("cluster-name"))
	viper.BindPFlag("serial-number", eksCmd.PersistentFlags().Lookup("serial-number"))
	viper.BindPFlag("profile", eksCmd.PersistentFlags().Lookup("profile"))

	rootCmd.AddCommand(eksCmd)
}

func initConfig() {
	// Find home directory.
	home, err := homedir.Dir()
	if err != nil {
		log.Fatalln(err)
	}

	// Search config in ~/.k8sctls directory with name "config.json".
	viper.AddConfigPath(home + "/.k8sctls")
	viper.SetConfigType("json")
	viper.SetConfigName("config")

	viper.AutomaticEnv()

	if err := viper.ReadInConfig(); err == nil {
		configFile := "Config file found [ " + viper.ConfigFileUsed() + " ]"
		dialog.Log("Info", configFile, nil)
	}
}

type cluster struct {
	Region       string
	Serialnumber string
	Profile      string
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
	Short: "Configure eks credentials",
	Long:  `Import configuration of eks cluster in .kube/config file.`,
	Run: func(cmd *cobra.Command, args []string) {
		// get flags or config file values, flag take precedence over config file.
		n, _ := cmd.Flags().GetString("cluster-name")
		s, _ := cmd.Flags().GetString("serial-number")
		r, _ := cmd.Flags().GetString("region")
		p, _ := cmd.Flags().GetString("profile")
		t, _ := cmd.Flags().GetString("token-code")

		if n != "" {
			n = viper.GetString("cluster-name")
		} else {
			dialog.Log("Error", "required flag(s) \"cluster-name\" not set.", cmd)
		}

		// Read config.json file
		var C cluster
		clusterTree := viper.Sub(n)
		if clusterTree != nil {
			err := clusterTree.Unmarshal(&C)
			if err != nil {
				log.Fatalf("Unable to decode into struct, %v", err)
			}
		} else {
			// check if cluster is not configured in the config.json
			if err := viper.ReadInConfig(); err == nil {
				dialog.Log("Warning", "Cluster ["+n+"] not found in "+viper.ConfigFileUsed(), cmd)
			}
		}

		// check inputs values in flags or config file, priority, first flag then config file.
		if s != "" {
			s = viper.GetString("serial-number")
		} else if C.Serialnumber != "" {
			s = C.Serialnumber
		} else {
			dialog.Log("Error", "required flag(s) \"serial-number\" not set.", cmd)
		}

		if r != "" {
			r = viper.GetString("region")
		} else if C.Region != "" {
			r = C.Region
		} else {
			dialog.Log("Error", "required flag(s) \"region\" not set.", cmd)
		}

		if p != "" {
			p = viper.GetString("profile")
		} else if C.Profile != "" {
			p = C.Profile
		} else {
			dialog.Log("Error", "required flag(s) \"profile\" not set.", cmd)
		}

		sessionToken_, err := exec.Command("aws", "sts", "get-session-token", "--serial-number", s, "--token-code", t, "--profile", p).Output()

		if err != nil {
			dialog.Log("Error", "problem getting session token, maybe MFA was wrong", cmd)
		}

		// read output json format
		var crd Identification
		json.Unmarshal([]byte(sessionToken_), &crd)

		// set environment vars
		os.Setenv("AWS_SECRET_ACCESS_KEY", crd.Credentials.SecretAccessKey)
		os.Setenv("AWS_SESSION_TOKEN", crd.Credentials.SessionToken)

		// create context configuration eks cluster
		awsContext := exec.Command("aws", "eks", "--region", r, "update-kubeconfig", "--name", n, "--profile", p)
		err = awsContext.Run()
		if err != nil {
			dialog.Log("Error", "problem creating new context, somthing was wrong", cmd)
		}
	},
}
