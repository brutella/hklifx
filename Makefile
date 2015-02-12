GO ?= go

bbb:
	GOOS=linux GOARCH=arm GOARM=7 $(GO) build daemon/hklifxd.go

rpi:
	GOOS=linux GOARCH=arm GOARM=6 $(GO) build daemon/hklifxd.go