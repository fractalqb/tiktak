package main

import (
	"fmt"
	"log"
	"path/filepath"

	"codeberg.org/fractalqb/gomklib"
	"git.fractalqb.de/fractalqb/gomk"
	"git.fractalqb.de/fractalqb/gomk/gomkore"
	"git.fractalqb.de/fractalqb/gomk/mkfs"
)

const version = "0.5.0"

var cmds = []string{
	"cmd/tikflt/tikflt",
	"cmd/tikmig/tikmig",
	"cmd/tiktak/tiktak",
	"cmd/tikloc/tikloc",
}

func main() {
	build := gomklib.GoModule{Env: gomkore.DefaultEnv(nil)}
	build.Env.Out = nil
	build.Flags()
	build.GoBuild().SetVars = []string{
		"git.fractalqb.de/fractalqb/tiktak/cmd.Version=" + version,
	}

	prj := gomkore.NewProject("")

	err := gomk.Edit(prj, func(prj gomk.ProjectEd) {
		goAls := build.DefaultGolas(prj)
		for _, cmd := range cmds {
			prj.Goal(mkfs.File(cmd)).
				By(build.GoBuild(),
					goAls.Test,
					prj.Goal(mkfs.DirList{Dir: filepath.Dir(cmd)}),
				)
		}
	})
	if err != nil {
		log.Fatal(err)
	}

	// Now, go for it…
	err = build.Make(prj)
	if err != nil {
		log.Fatal(err)
	}

	fmt.Println("DONE") // just for Go test Example
}
