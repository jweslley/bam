CC=gcc
CFLAGS=-O2

OBJ=nss_localtld.o

%.o: %.c
	$(CC) -Wall -fPIC -c -o $@ $< $(CFLAGS)

libnss_localtld.so.2: $(OBJ)
	$(CC) -shared -o $@ $^ -Wl,-soname,libnss_localtld.so.2 $(CFLAGS)

clean:
	rm -f *.o *~ libnss_localtld.so.2

install: libnss_localtld.so.2
	ln -sf $(CURDIR)/libnss_localtld.so.2 /lib

uninstall:
	rm /lib/libnss_localtld.so.2

