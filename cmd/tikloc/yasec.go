package main

import (
	"encoding/hex"
	"log"

	"git.fractalqb.de/fractalqb/yacfg/yasec"
)

func yasecInit(notTerm bool) {
	const appName = "tikloc"
	rawSalt := mustRet(hex.DecodeString(yaecSalt))
	yasec.DefaultConfig.Salt = rawSalt
	if notTerm {
		log.Println("Waiting for " + appName + ".yasec")
		err := yasec.DefaultConfig.SetFromUnixSocket(appName+".yasec", rawSalt)
		if err != nil {
			log.Fatal(err.Error())
		}
		log.Println("yasec passphrase received")
	} else {
		yasec.DefaultConfig.SetFromPrompt("yasec passphrase:", rawSalt)
	}
}
