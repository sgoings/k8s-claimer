package commands

import (
	"bytes"
	"fmt"
	"io"
	"log"
	"os"
	"strings"

	"github.com/codegangsta/cli"
	"github.com/deis/k8s-claimer/client"
)

const (
	ipEnvVarName          = "IP"
	tokenEnvVarName       = "TOKEN"
	clusterNameEnvVarName = "CLUSTER_NAME"
)

// CreateLease is a cli.Command action for creating a lease
func CreateLease(c *cli.Context) error {
	// inspect env for auth env var
	authToken := os.Getenv("AUTH_TOKEN")
	if authToken == "" {
		log.Fatalf("An authorization token is required in the form of an env var AUTH_TOKEN")
	}
	server := c.GlobalString("server")
	if server == "" {
		log.Fatalf("Server missing")
	}
	durationSec := c.Int("duration")
	if durationSec <= 0 {
		log.Fatalf("Invalid duration %d", durationSec)
	}
	envPrefix := c.String("env-prefix")
	clusterRegex := c.String("cluster-regex")
	clusterVersion := c.String("cluster-version")
	kcfgFile := c.String("kubeconfig-file")
	if len(kcfgFile) < 1 {
		log.Fatalf("Missing kubeconfig-file")
	}

	fd, err := os.Create(kcfgFile)
	if err != nil {
		log.Fatalf("Error opening %s (%s)", kcfgFile, err)
	}
	defer fd.Close()

	resp, err := client.CreateLease(server, authToken, clusterVersion, clusterRegex, durationSec)
	if err != nil {
		log.Fatalf("Error creating lease (%s)", err)
	}

	kcfg, err := resp.KubeConfigBytes()
	if err != nil {
		log.Fatalf("Error decoding kubeconfig (%s)", err)
	}
	fmt.Println(exportVar(envPrefix, ipEnvVarName, resp.IP))
	fmt.Println(exportVar(envPrefix, tokenEnvVarName, resp.Token))
	fmt.Println(exportVar(envPrefix, clusterNameEnvVarName, resp.ClusterName))

	if _, err := io.Copy(fd, bytes.NewBuffer(kcfg)); err != nil {
		log.Fatalf("Error writing new Kubeconfig file to %s (%s)", kcfgFile, err)
	}
	return nil
}

func exportVar(prefix, envVarName, val string) string {
	if prefix != "" {
		envVarName = fmt.Sprintf("%s_%s", prefix, envVarName)
	}
	val = strings.Replace(val, `"`, `\"`, -1)
	return fmt.Sprintf(`export %s="%s"`, envVarName, val)
}
