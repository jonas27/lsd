package cmd

import (
	"context"
	"fmt"
	"io"
	"log/slog"
	"os"
	"os/exec"

	"github.com/spf13/cobra"
)


var inputFile *string

func init() {
	inputFile = rootCmd.Flags().StringP("file", "f", "-", "file path to neat, or - to read from stdin")
	rootCmd.SetOut(os.Stdout)
	rootCmd.SetErr(os.Stderr)
	rootCmd.MarkFlagFilename("file")
	rootCmd.AddCommand(getCmd)
}

// Execute is the entry point for the command package
func Execute(ctx context.Context, log *slog.Logger) error {
	return rootCmd.Execute()
}

var rootCmd = &cobra.Command{
	Use: "lsd",
	Example: `kubectl get secret my-secret -o yaml | kubectl lsd
kubectl get secret mysecret -oyaml | kubectl lsd
kubectl neat -f - <./my-secret.json
kubectl neat -f ./my-secret.json
kubectl neat -f ./my-secret.json`,
	RunE: func(cmd *cobra.Command, args []string) error {
		var in []byte
		var err error
		if *inputFile == "-" {
			stdin := cmd.InOrStdin()
			if in, err = io.ReadAll(stdin); err != nil {
				return fmt.Errorf("error reading from stdin: %w", err)
			}
		} else {
			in, err = os.ReadFile(*inputFile)
			if err != nil {
				return fmt.Errorf("error reading file %s: %w", *inputFile, err)
			}
		}
    out, err := Lsd(in)
    if err != nil {
			return fmt.Errorf("error running Lsd: %w", err)
		}
		cmd.Print(out)
		return nil
	},
}

var kubectl string = "kubectl"

var getCmd = &cobra.Command{
	Use: "get",
	Example: `kubectl neat get -- pod mypod -oyaml
kubectl neat get -- svc -n default myservice --output json`,
	FParseErrWhitelist: cobra.FParseErrWhitelist{UnknownFlags: true}, // don't try to validate kubectl get's flags
	RunE: func(cmd *cobra.Command, args []string) error {
		var err error

		kubectlCmd := exec.Command(kubectl, args...)
		kres, err := kubectlCmd.CombinedOutput()
		if err != nil {
			return fmt.Errorf("error invoking kubectl as %v: %w", kubectlCmd.Args, err)
		}

    out, err := Lsd(kres)
		if err != nil {
			return err
		}
		cmd.Println(out)
		return nil
	},
}
