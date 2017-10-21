package main

import (
	"fmt"
	"net"
	"os"

	"github.com/mitchellh/cli"
)

// MayaCtlName will be fiiled in by the GNUMakefile using an EVN variable
//  MAYACTL
var MayaCtlName = "maya"

func init() {

	mapiaddr := os.Getenv("MAPI_ADDR")
	if mapiaddr == "" {
		mapiaddr = getEnvOrDefault(mapiaddr)

		os.Setenv("MAPI_ADDR", mapiaddr)

	}

}

func main() {
	os.Exit(Run(os.Args[1:]))
}

// Run executes the command with passed args
func Run(args []string) int {
	return RunCustom(args, Commands(nil))
}

// RunCustom executes the command with passed args and custom commands
func RunCustom(args []string, commands map[string]cli.CommandFactory) int {
	// Get the command line args. We shortcut "--version" and "-v" to
	// just show the version.
	for _, arg := range args {
		if arg == "-v" || arg == "-version" || arg == "--version" {
			newArgs := make([]string, len(args)+1)
			newArgs[0] = "version"
			copy(newArgs[1:], args)
			args = newArgs
			break
		}
	}

	// Extract the commands to include in the help
	commandNames := make([]string, 0, len(commands))
	for k := range commands {
		commandNames = append(commandNames, k)
	}

	cli := &cli.CLI{
		Args:     args,
		Commands: commands,
		HelpFunc: cli.FilteredHelpFunc(commandNames, cli.BasicHelpFunc(MayaCtlName)),
	}

	exitCode, err := cli.Run()
	if err != nil {
		fmt.Fprintf(os.Stderr, "Error executing CLI: %s\n", err.Error())
		return 1
	}

	return exitCode
}

func getEnvOrDefault(env string) string {
	if env == "" {
		host, _ := os.Hostname()
		addrs, _ := net.LookupIP(host)
		for _, addr := range addrs {
			if ipv4 := addr.To4(); ipv4 != nil {
				env = ipv4.String()
				if env == "127.0.0.1" {
					continue
				}
				break
			}
		}
	}
	return "http://" + env + ":5656"
}
