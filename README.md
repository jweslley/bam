# BAM!

[![Travis](https://api.travis-ci.org/jweslley/bam.png)](http://travis-ci.org/jweslley/bam)

BAM! is a web-server for developers.

Heavily inspired by [pow](http://pow.cx), BAM! goes further by supporting not only ruby/rack applications, but any [Procfile-based application](https://devcenter.heroku.com/articles/procfile). Also, BAM! is planned to support both Linux and Mac OS X.


## Getting started

### Requirements

On Linux, BAM! requires iptables and [localtld](https://github.com/jweslley/localtld).


### Building from source

    git clone http://github.com/jweslley/bam
    make build


### Linux

    ./bam -generate iptables > iptables.rules
    sudo iptables-restore < iptables.rules
    ./bam


## How it works

BAM! is composed by a [reverse-proxy](https://github.com/jweslley/bam/blob/master/proxy.go) and an [application manager](https://github.com/jweslley/bam/blob/master/command_center.go). On startup, the application manager search for Procfile-based applications in a given directory (customized by configuration file) and list all applications found at address http://bam.app . From this listing, the user can start/stop applications.

However, in order to work properly, BAM! requires a couple of firewall rules and some tricks to resolve domain names. Thus, when the user access some application, like http://myblog.app, the DNS must be configured to resolve any request to the top-level-domain `.app` to `localhost`. After reaching port 80, the firewall must forward the request to reverse-proxy's port (defaults to 42042). Once inside the proxy, the request will be forwarded to the target application (`myblog` in this example).


### Linux

* Firewall rules are managed by [iptables](https://en.wikipedia.org/wiki/Iptables).
* DNS resolution relies on [localtld](https://github.com/jweslley/localtld).

### MAC OS X

* Uses plists for launching and managing the firewall.
* DNS resolution relies on `/etc/resolver`. [More details here](https://news.ycombinator.com/item?id=2421186).


## Roadmap

* OSX support
* SSL support
* Start scripts
* Package installers
* Better UI


## Bugs and Feedback

If you discover any bugs or have some idea, feel free to create an issue on GitHub:

    http://github.com/jweslley/bam/issues


## License

MIT license. Copyright (c) 2013-2015 Jonhnny Weslley <http://jonhnnyweslley.net>

See the LICENSE file provided with the source distribution for full details.
