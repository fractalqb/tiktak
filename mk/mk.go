package main

import (
	"flag"
	"os"

	"git.fractalqb.de/fractalqb/gomk"
	"git.fractalqb.de/fractalqb/qblog"
)

const version = "0.4.1"

var (
	cmds = []string{
		"cmd/tikflt/tikflt",
		"cmd/tikmig/tikmig",
		"cmd/tiktak/tiktak",
	}

	goBuild = gomk.GoBuild{
		TrimPath: true,
		LDFlags:  []string{"-s", "-w"},
		SetVars: []string{
			"git.fractalqb.de/fractalqb/tiktak/cmd.Version=" + version,
		},
	}

	log      = qblog.New(&qblog.DefaultConfig).Logger
	writeDot bool
)

func flags() {
	fLog := flag.String("log", "", "Set log level")
	fInstall := flag.Bool("install", false, "Install commands")
	flag.BoolVar(&writeDot, "dot", false, "Write project as graphviz dot file")
	flag.Parse()
	goBuild.Install = *fInstall
	if *fLog != "" {
		qblog.DefaultConfig.ParseFlag(*fLog)
	}
}

func main() {
	flags()

	prj := gomk.NewProject(".")

	gVulnchk := prj.Goal(gomk.Abstract("vulncheck")).
		By(&gomk.GoVulncheck{Patterns: []string{"./..."}})

	gTest := prj.Goal(gomk.Abstract("test")).
		By(&gomk.GoTest{Pkgs: []string{"./..."}}, gVulnchk)

	gCmds := prj.Goal(gomk.DirContent("cmds"))

	for _, cmd := range cmds {
		g := prj.Goal(gomk.File(cmd)).By(&goBuild, gTest)
		gCmds.ImpliedBy(g)
	}

	if writeDot {
		prj.WriteDot(os.Stdout)
		return
	}

	builder := gomk.Builder{Env: gomk.DefaultEnv()}
	builder.Env.Log = log
	if err := builder.Project(prj); err != nil {
		log.Error(err.Error())
	}
}
