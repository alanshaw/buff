.PHONY: buff
buff:
	go build .
install: buff
	go install .