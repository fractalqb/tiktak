package main

import (
	"flag"
	"fmt"
	"os"

	"git.fractalqb.de/fractalqb/gomk"
)

const (
	version = "0.4.1"

	cmdsName = "commands"
)

var (
	must     = gomk.Must
	commands = []string{"tiktak", "tikflt", "tikmig"}

	goBuild = gomk.CommandDef{
		Name: "go",
		Args: []string{"build",
			"-trimpath",
			"-ldflags",
			fmt.Sprintf("-s -w -X git.fractalqb.de/fractalqb/tiktak/cmd.Version=%s", version),
		},
	}
)

func flags() {
	inst := flag.Bool("install", false, "Install commands after building")
	flag.Parse()
	if *inst {
		goBuild.Args[0] = "install"
	}
}

func main() {
	flags()

	prj := gomk.NewProject(must, &gomk.Config{Env: os.Environ()})

	tCmds := gomk.NewNopTask(must, prj, cmdsName)

	for _, cmd := range commands {
		t := gomk.NewCmdDefTask(must, prj, "cmd:"+cmd, goBuild).
			WorkDir("cmd", cmd)
		tCmds.DependOn(t.Name())
	}

	if flag.NArg() == 0 {
		gomk.Build(prj, cmdsName)
	} else {
		for _, arg := range flag.Args() {
			gomk.Build(prj, arg)
		}
	}
}
