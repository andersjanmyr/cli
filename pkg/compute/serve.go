package compute

import (
	"io"
	"log"
	"os"
	"os/exec"
	"os/signal"
	"syscall"

	"github.com/fastly/cli/pkg/common"
	"github.com/fastly/cli/pkg/compute/manifest"
	"github.com/fastly/cli/pkg/config"
	"github.com/fastly/cli/pkg/text"
)

// ServeCommand produces and runs an artifact from files on the local disk.
type ServeCommand struct {
	common.Base
	manifest manifest.Data
	build    *BuildCommand

	// Build fields
	name       common.OptionalString
	lang       common.OptionalString
	includeSrc common.OptionalBool
	force      common.OptionalBool
}

// NewServeCommand returns a usable command registered under the parent.
func NewServeCommand(parent common.Registerer, globals *config.Data, build *BuildCommand) *ServeCommand {
	var c ServeCommand

	c.build = build
	c.Globals = globals
	c.CmdClause = parent.Command("serve", "Build and run a Compute@Edge package locally")

	c.manifest.File.SetOutput(c.Globals.Output)
	c.manifest.File.Read(manifest.Filename)

	// Build flags
	c.CmdClause.Flag("name", "Package name").Action(c.name.Set).StringVar(&c.name.Value)
	c.CmdClause.Flag("language", "Language type").Action(c.lang.Set).StringVar(&c.lang.Value)
	c.CmdClause.Flag("include-source", "Include source code in built package").Action(c.includeSrc.Set).BoolVar(&c.includeSrc.Value)
	c.CmdClause.Flag("force", "Skip verification steps and force build").Action(c.force.Set).BoolVar(&c.force.Value)

	return &c
}

// Exec implements the command interface.
func (c *ServeCommand) Exec(in io.Reader, out io.Writer) (err error) {
	// Reset the fields on the BuildCommand based on ServeCommand values.
	if c.name.WasSet {
		c.build.PackageName = c.name.Value
	}
	if c.lang.WasSet {
		c.build.Lang = c.lang.Value
	}
	if c.includeSrc.WasSet {
		c.build.IncludeSrc = c.includeSrc.Value
	}
	if c.force.WasSet {
		c.build.Force = c.force.Value
	}

	err = c.build.Exec(in, out)
	if err != nil {
		return err
	}

	text.Break(out)

	err = c.Local(out)
	if err != nil {
		return err
	}

	return nil
}

func (c *ServeCommand) Local(out io.Writer) error {
	sig := make(chan os.Signal, 1)

	signals := []os.Signal{
		syscall.SIGINT,
		syscall.SIGTERM,
	}

	signal.Notify(sig, signals...)

	cmd := exec.Command("viceroy", "bin/main.wasm", "-C", "fastly.toml")
	cmd.Stdout = out
	cmd.Stderr = out

	go func(sig chan os.Signal, cmd *exec.Cmd) {
		<-sig
		signal.Stop(sig)

		err := cmd.Process.Signal(os.Kill)
		if err != nil {
			log.Fatal(err)
		}
	}(sig, cmd)

	err := cmd.Start()
	if err != nil {
		return err
	}

	cmd.Wait()

	return nil
}
