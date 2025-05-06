package main

import (
	"encoding/hex"
	"log"
	"os"

	"git.fractalqb.de/fractalqb/yacfg/yasec"
	"golang.org/x/term"
)

func yasecInit() {
	const appName = "tikloc"
	rawSalt := mustRet(hex.DecodeString(yaecSalt))
	yasec.DefaultConfig.Salt = rawSalt

	if term.IsTerminal(int(os.Stdin.Fd())) {
		yasec.DefaultConfig.SetFromPrompt("yasec passphrase:", rawSalt)
		return
	} else {
		log.Println("Waiting for " + appName + ".yasec")
		err := yasec.DefaultConfig.SetFromUnixSocket(appName+".yasec", rawSalt)
		if err != nil {
			log.Fatal(err.Error())
		}
		log.Println("yasec passphrase received")
	}
}
