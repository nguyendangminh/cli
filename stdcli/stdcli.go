package stdcli

import (
	"fmt"
	"io/ioutil"
	"os"
	"os/exec"
	"path"
	"path/filepath"
	"time"

	"github.com/convox/cli/Godeps/_workspace/src/github.com/briandowns/spinner"
	"github.com/convox/cli/Godeps/_workspace/src/github.com/codegangsta/cli"
)

var (
	Binary   string
	Commands []cli.Command
	Exiter   func(code int)
	Runner   func(bin string, args ...string) error
	Querier  func(bin string, args ...string) ([]byte, error)
	Spinner  *spinner.Spinner
	Tagger   func() string
	Writer   func(filename string, data []byte, perm os.FileMode) error
)

func init() {
	Binary = filepath.Base(os.Args[0])
	Exiter = os.Exit
	Querier = queryExecCommand
	Runner = runExecCommand
	Spinner = spinner.New(spinner.CharSets[9], 100*time.Millisecond)
	Tagger = tagTimeUnix
	Writer = ioutil.WriteFile

	cli.AppHelpTemplate = `{{.Name}}: {{.Usage}}

Usage:
  {{.Name}} <command> [args...]

Subcommands: ({{.Name}} help <subcommand>)
  {{range .Commands}}{{join .Names ", "}}{{ "\t" }}{{.Description}}
  {{end}}{{if .Flags}}
Options:
  {{range .Flags}}{{.}}
  {{end}}{{end}}
`

	cli.CommandHelpTemplate = fmt.Sprintf(`%s {{.FullName}}: {{.Description}}

Usage:
  %s {{.FullName}} {{.Usage}}
{{if .Subcommands}}
Subcommands: (%s {{.FullName}} help <subcommand>)
  {{range .Subcommands}}{{join .Names ", "}}{{ "\t" }}{{.Description}}
  {{end}}{{end}}{{if .Flags}}
Options:
   {{range .Flags}}{{.}}
   {{end}}{{ end }}
`, Binary, Binary, Binary)

}

func New() *cli.App {
	app := cli.NewApp()

	app.Name = Binary
	app.Commands = Commands

	app.CommandNotFound = func(c *cli.Context, cmd string) {
		fmt.Fprintf(os.Stderr, "No such command \"%s\". Try `%s help`\n", cmd, Binary)
	}

	return app
}

func Debug() bool {
	if debug := os.Getenv("DEBUG"); debug != "" {
		return true
	}
	return false
}

func DirApp(c *cli.Context, wd string) (string, string, error) {
	abs, err := filepath.Abs(wd)

	if err != nil {
		return "", "", err
	}

	app := c.String("app")

	if app == "" {
		app = path.Base(abs)
	}

	return abs, app, nil
}

func RegisterCommand(cmd cli.Command) {
	Commands = append(Commands, cmd)
}

func Run(bin string, args ...string) error {
	return Runner(bin, args...)
}

func Query(bin string, args ...string) ([]byte, error) {
	return Querier(bin, args...)
}

func Tag() string {
	return Tagger()
}

func VersionPrinter(printer func(*cli.Context)) {
	cli.VersionPrinter = printer
}

func Error(err error) {
	fmt.Fprintf(os.Stderr, "ERROR: %s\n", err)
	Exiter(1)
}

func Usage(c *cli.Context, name string) {
	cli.ShowCommandHelp(c, name)
	Exiter(129)
}

func runExecCommand(bin string, args ...string) error {
	cmd := exec.Command(bin, args...)
	cmd.Stdin = os.Stdin
	cmd.Stdout = os.Stdout
	cmd.Stderr = os.Stderr
	err := cmd.Run()

	if Debug() {
		fmt.Fprintf(os.Stderr, "DEBUG: exec: '%v', '%v', '%v'\n", bin, args, err)
	}

	return err
}

func queryExecCommand(bin string, args ...string) ([]byte, error) {
	return exec.Command(bin, args...).CombinedOutput()
}

func tagTimeUnix() string {
	return fmt.Sprintf("%v", time.Now().Unix())
}
