# TODO add documentation generation


setup:
	go get github.com/axw/gocov/gocov


install:
	go install


test:
	go test -v


coverage:
	gocov test | gocov annotate -
