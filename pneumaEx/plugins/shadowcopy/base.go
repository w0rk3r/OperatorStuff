//+build windows

package main

import (
	"flag"
	"os"
	"os/exec"
	"path/filepath"

	"github.com/hashicorp/go-hclog"
	"github.com/hashicorp/go-plugin"
	"github.com/preludeorg/pneuma/util"
)

var key = "JWHQZM9Z4HQOYICDHW4OCJAXPPNHBA"

type PneumaModule struct {
	logger hclog.Logger
}

func (pm *PneumaModule) Entry() []string {
	var funcs []string
	for k := range Functions {
		funcs = append(funcs, ModuleName+"."+k)
	}
	return funcs
}

func (pm *PneumaModule) RunTask(function string, args []string) ([]byte, int, int) {
	results, status := Functions[function](args)
	return results, status, os.Getpid()
}

var handshakeConfig = plugin.HandshakeConfig{
	ProtocolVersion:  1,
	MagicCookieKey:   "PNEUMA_PLUGIN",
	MagicCookieValue: "7010422c-d246-44db-bd3f-1387f3eebda7",
}

func RunStandalone(function string, args string) {
	currExecutable, _ := filepath.Abs(os.Args[0])
	cmd := exec.Command(currExecutable, "-mode", "standalone", "-function", function, "-args", args)
	cmd.Start()
}

func main() {
	mode := flag.String("mode", "module", "Run as a module or standalone binary")
	function := flag.String("function", "", "Function to call")
	args := flag.String("args", "", "Function arguments")
	flag.Parse()
	if *mode != "module" {
		ExecFunctions[*function](*args)
	} else {
		logger := hclog.New(&hclog.LoggerOptions{
			Level:      hclog.Trace,
			Output:     os.Stderr,
			JSONFormat: true,
		})

		module := &PneumaModule{
			logger: logger,
		}
		// pluginMap is the map of plugins we can dispense.
		var pluginMap = map[string]plugin.Plugin{
			ModuleName: &util.ModulePluginRPC{Impl: module},
		}

		plugin.Serve(&plugin.ServeConfig{
			HandshakeConfig: handshakeConfig,
			Plugins:         pluginMap,
		})
	}
}
