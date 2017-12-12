ifeq ($(OS), Windows_NT)
	EXE=.exe
else
	EXE=
endif

all: signalControl$(EXE)

signalControl$(EXE): signalControl.go
	go build signalControl.go

.PHONY: run
run: signalControl$(EXE)
	./signalControl
