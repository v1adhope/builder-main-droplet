.SILENT:

build:
	go build -o .bin/app cmd/main.go

run: build
	sudo ./.bin/app $(kpc)
