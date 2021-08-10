bindir = $(HOME)/bin
exename = kubecontext

INSTALL = install
INSTALL_EXE = $(INSTALL) -m 755

all: kubecontext

kubecontext: main.go
	go build

install: all
	$(INSTALL_EXE) kubecontext $(DESTDIR)$(bindir)/$(exename)
