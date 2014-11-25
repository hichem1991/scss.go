install: libsass/lib/libsass.a *.go
	go install

	
libsass/lib/libsass.a: libsass/*.cpp libsass/*.hpp libsass/*.h
	$(MAKE) -C ./libsass 

scss.go.test: libsass/lib/libsass.a *.go *.c
	go test -c

test: scss.go.test
	./scss.go.test

clean:
	$(MAKE) -C ./libsass clean
	rm -f scss.go.test

.PHONY: install test clean
