// eks.go
package cmd

import (
	"encoding/json"
	"fmt"
	"log"
	"os"
	"os/exec"

	"github.com/mitchellh/go-homedir"
	"github.com/spf13/cobra"
	"github.com/spf13/viper"
)

func init() {
	cobra.OnInitialize(initConfig)

	eksCmd.PersistentFlags().StringP("region", "r", "", "cluster region")
	eksCmd.PersistentFlags().StringP("cluster-name", "n", "", "cluster name")
	eksCmd.PersistentFlags().StringP("serial-number", "s", "", "arn user name")
	eksCmd.PersistentFlags().StringP("token-code", "t", "", "two factor authentication code")
	eksCmd.MarkPersistentFlagRequired("token-code")
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
		fmt.Println("Using config file:", viper.ConfigFileUsed())
	} else {
		fmt.Printf("Error reading config file: %v\n", err)
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

		if n != "" {
			n = viper.GetString("cluster-name")
		} else {
			log.Fatalln("Error setting cluster-name flag")
		}

		t, _ := cmd.Flags().GetString("token-code")

		var C cluster
		clusterTree := viper.Sub(n)
		// check if cluster is not configured
		if clusterTree != nil {
			err := clusterTree.Unmarshal(&C)
			if err != nil {
				log.Fatalf("Unable to decode into struct, %v", err)
			}
		} else {
			log.Println("WARNING: cluster name don't configured in config.json")
		}

		if s != "" {
			s = viper.GetString("serial-number")
		} else if C.Serialnumber != "" {
			s = C.Serialnumber
		} else {
			// cmd.Help().Error()
			log.Fatalln("ERROR: serial-number is not set")
		}

		if r != "" {
			r = viper.GetString("region")
		} else if C.Region != "" {
			r = C.Region
		} else {
			log.Fatalln("ERROR: region is not set")
		}

		if p != "" {
			p = viper.GetString("profile")
		} else if C.Profile != "" {
			p = C.Profile
		} else {
			log.Fatalln("ERROR: profile is not set")
		}

		sessionToken_, err := exec.Command("aws", "sts", "get-session-token", "--serial-number", s, "--token-code", t, "--profile", p).Output()

		if err != nil {
			log.Fatal("command failed (aws sts get-session-token --serial-number " + s + " --token-code " + t + " --profile " + p + ")\n")
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
			log.Fatalf("ERROR: problem creating new context (aws eks --region "+r+" update-kubeconfig --name "+n+" --profile "+p+")%v", err)
		}
	},
}
