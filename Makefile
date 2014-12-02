host := $(shell go env GOHOSTOS)
ifeq ($(host),darwin)
	ldflags = -lc++
else
	ldflags = -lm -lstdc++
endif

# We use pkg-config so that we can use absolute paths to the libsass directory
# that don't need to be hard coded in the scss.go file. It would be nice if 
# the cgo macros in go had some sort of current directory variable. This would
# eliminiate the need to do this.

define PKG_CONFIG_BODY
Name: scss.go
Version: 0.0.1
Description: scss.go
Cflags: -g -I$(PWD)/libsass
Libs:  $(ldflags) $(PWD)/libsass/lib/libsass.a
endef
# why am i exporting this variable? see
# http://stackoverflow.com/questions/649246/is-it-possible-to-create-a-multi-line-string-variable-in-a-makefile
export PKG_CONFIG_BODY


install: libsass/lib/libsass.a *.go scss.pc
	go install

scss.pc:
	echo "$$PKG_CONFIG_BODY" > scss.pc
	
libsass/lib/libsass.a: libsass/*.cpp libsass/*.hpp libsass/*.h
	$(MAKE) -C ./libsass 

scss.go.test: libsass/lib/libsass.a *.go *.c scss.pc
	go test -c

test: scss.go.test
	./scss.go.test

clean:
	$(MAKE) -C ./libsass clean
	rm -f scss.go.test scss.pc

.PHONY: install test clean
