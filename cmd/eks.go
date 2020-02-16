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
	println("In initConfig")
	cfgFile := ".config.json"
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
	Short: "Configure credentials of eks cluster",
	Long:  `Import configuration of eks cluster in .kube/config file.`,
	Run: func(cmd *cobra.Command, args []string) {

		var n string

		// get flags or config file values, flag take precedence over config file.
		if viper.IsSet("cluster-name") {
			n = viper.GetString("cluster-name")
		} else {
			log.Fatalln("Error setting cluster-name flag")
		}

		t, _ := cmd.Flags().GetString("token-code")

		clusterTree := viper.Sub(n)
		var c cluster
		err := clusterTree.Unmarshal(&c)
		if err != nil {
			log.Fatalf("Unable to decode into struct, %v", err)
		}

		// debug
		fmt.Println(n)
		fmt.Println(c.Serialnumber)
		fmt.Println(c.Profile)
		fmt.Println(c.Region)

		sessionToken_, err := exec.Command("aws", "sts", "get-session-token", "--serial-number", c.Serialnumber, "--token-code", t, "--profile", c.Profile).Output()

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
		awsContext := exec.Command("aws", "eks", "--region", c.Region, "update-kubeconfig", "--name", n, "--profile", c.Profile)
		err = awsContext.Run()
		if err != nil {
			log.Fatalf("ERROR: problem creating new context %v", err)
		}
	},
}
