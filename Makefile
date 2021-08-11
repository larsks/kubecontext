bindir = $(HOME)/bin
exename = kubecontext

SOURCES = \
	  main.go \
	  config.go

INSTALL = install
INSTALL_EXE = $(INSTALL) -m 755

all: kubecontext

kubecontext: $(SOURCES)
	go build

install: all
	$(INSTALL_EXE) kubecontext $(DESTDIR)$(bindir)/$(exename)
