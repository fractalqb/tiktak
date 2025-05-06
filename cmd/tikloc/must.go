package main

import "log"

func must(err error) {
	if err != nil {
		log.Fatal(err)
	}
}

func mustRet[T any](v T, err error) T {
	must(err)
	return v
}
