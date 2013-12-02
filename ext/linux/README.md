# NSSwitch hosts resolver plugin

Allows software that rely on NSSwitch to resolve domains for local applications (like myapp.app). Command line software like `host` and `dig` won't resolve the address, but `getent` will.

## Install

Compile, copy or link the shared object to `/lib` and add `localtld` to the `hosts` line of `/etc/nsswitch.conf`, and eventually restart your browser (this is required).

    $ cd ext/linux/
    $ make
    $ sudo make install
    $ sudo vi /etc/nsswitch.conf


    hosts: files mdns4_minimal [NOTFOUND=return] dns mdns4 localtld

Note: do not change the whole line to look like this, just add `localtld` at the end of it or before `dns`.


## Credits

Code is from [prax](https://github.com/ysbaddaden/prax) by ysbaddaden and collaborators.
