VERSION=0.2.0

all: tiktak-cmd tikflt-cmd tikmig-cmd

install: tiktak-inst tikflt-inst tikmig-inst

%-inst:
	cd cmd/$*; go install --trimpath \
	-ldflags "-s -w -X git.fractalqb.de/fractalqb/tiktak/cmd.Version=$(VERSION)"

%-cmd:
	cd cmd/$*; go build --trimpath \
	-ldflags "-s -w -X git.fractalqb.de/fractalqb/tiktak/cmd.Version=$(VERSION)"
