ifeq ($(OS), Windows_NT)
	EXE:=.exe
else
	EXE:=
endif

all: signalControl$(EXE)

signalControl$(EXE): signalControl.go
	go build signalControl.go

.PHONY: clean
clean:
	-rm signalControl$(EXE)

.PHONY: deps
deps:
	go get github.com/mattn/go-sqlite3
	go get github.com/julienschmidt/httprouter
