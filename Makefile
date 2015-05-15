GO ?= go

bbb:
	GOOS=linux GOARCH=arm GOARM=7 $(GO) build hklifxd.go

rpi:
	GOOS=linux GOARCH=arm GOARM=6 $(GO) build hklifxd.go