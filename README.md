# BAM!

[![Travis](https://api.travis-ci.org/jweslley/bam.png)](http://travis-ci.org/jweslley/bam)

BAM! is a web-server for developers.

Heavily inspired by [pow](http://pow.cx), BAM! goes further by supporting not only ruby/rack applications, but any [Procfile-based application][procfile]. Also, BAM! is planned to support both Linux and Mac OS X.


## Installing

[Download][] and put the binary somewhere in your path.

### Building from source

    git clone http://github.com/jweslley/bam
    make build

## How it works

BAM! is composed by a [reverse-proxy](https://github.com/jweslley/bam/blob/master/proxy.go) and an [application manager](https://github.com/jweslley/bam/blob/master/command_center.go). On startup, the application manager search for Procfile-based applications in a given directory (customized by configuration file) and list all applications found at address http://bam.dev . From this listing, the user can start/stop applications.

However, in order to work properly, BAM! requires a couple of firewall rules and some tricks to resolve domain names. Thus, when the user access some application, like http://myblog.dev, the host OS must be configured to resolve any request to the top-level-domain `.dev` to `localhost`. After reaching port 80, the firewall must forward the request to reverse-proxy's port (defaults to 42042). Once a request reach the proxy, it will be forwarded to the target application (`myblog` in this example).

Since BAM! manages start/stop of the applications, it's important the web process in the Procfile be declared with a PORT environment variable, which will be informed by BAM! during application's startup. Something like:

    web: rails server -p $PORT

Please checkout the `examples` directory.

> The applications's directory, the top-level domain and proxy's port can be customized by configuration file.

## Features

#### Application management

In BAM's terminology, an application is a directory in the applications's directory containing either a Procfile or a `index.html` file. Primarily, BAM! deals with Procfile-based applications (a directory with a Procfile). However, if a directory contains a `index.html` file, BAM! will serve all files inside this directory as an static server.

During application's start, BAM! will pick an unused port and start either a couple of external processes depending on Procfile or a static web server. The application will be accessible at the address: `http://<application-name>.dev` For example, the `myblog` application will be accessible at http://myblog.dev

#### Subdomains

Once a application is started, it's also automatically accessible from all subdomains.

* http://www.myblog.dev/
* http://assets.myblog.dev/

#### Port aliases

BAM! lets you access others applications running in your machine using better names. For example, I like to use [btsync](https://www.getsync.com/) to synchronize files between my computers, by default btsync start at system's boot at port 8888, using port aliases I can access btsync by typing http://btsync.dev instead of http://localhost:8888, it's easier to remember.

#### Accessing your applications from other computers

Sometimes you need to access your applications from another computer on your local network, but the .dev domain will only work on your local computer. In this case, you can use the special [.xip.io domain](http://xip.io) to remotely access your applications.

* http://myblog.192.168.1.15.xip.io
* http://assets.myblog.192.168.1.15.xip.io

> 192.168.1.15 is my current IP address in the local network!

#### Sharing applications to the Internet

xip.io is great, but works only on your local network. Nowadays, remote work is common and to show your application to a remote coworker or even a client your need to configure a VPS containing a configured environment. To simplify this task, BAM! lets you share/unshare your application to the Internet through [localtunnel](http://localtunnel.me/).

#### Storing config in the environment

During application's start, BAM! will loads `.env` file (if available) in the application's directory and pass all environment variables to the applications's processes.

#### Command center

The command center is the application manager. A web application accessible at http://bam.dev from where you will list, start, stop, share and unshare your applications.


## Configuring BAM!

As stated before, BAM! requires a couple of firewall rules and some tricks to resolve domain names. Both of these requirements vary depending on your machine's operating system. Currently I tested BAM! in Archlinux (my beloved OS), but it should work in other Linux distributions too. Mac OSX support also must work, but not tested, configuration procedures were stolen from Pow. Feedback is welcome! ^^

### Linux

First at all, we need a way to resolve domain names in top-level domain `.dev` to localhost. In Linux, we could run a custom DNS server (like [localdns][]) for accomplish this, but there is a better way than running an extra process. For this, we will use [localtld][], a custom NSSwitch plugin to resolve domains for local applications. It's very straightforward to install, thus visit the [project page][localtld] and follow installation's instructions.

Since your machine is already resolving domain names to localhost, now we need forward all incoming connections from port 80 to BAM's reverse-proxy (port 42042). In Linux, the easier way is using iptables. BAM has a command to generate the iptables rules required. Thus, run the following commands:

    bam -generate iptables > iptables.rules
    sudo iptables-restore < iptables.rules

> Depending your Linux distribution, you will need extra configuration to load these rules at system's boot.

It's done. Now, start BAM! and have fun.


### MAC OS X

In Mac OSX, we need [localdns][] to [resolve][darwin-resolver] domain names in top-level domain `.dev` to localhost. Thus, visit the [project page][localdns] and follow installation's instructions.

Since your machine is already resolving domain names to localhost, now we need forward all incoming connections from port 80 to BAM's reverse-proxy (port 42042). In Mac OSX, we will use [plist][] to setup the firewall rules for us. BAM has a command to generate the plist file. Thus, run the following commands:

    bam -generate firewall > bam.firewall.plist
    sudo cp bam.firewall.plist /Library/LaunchDaemons/bam.firewall.plist
    sudo launchctl bootstrap system /Library/LaunchDaemons/bam.firewall.plist 2>/dev/null
    sudo launchctl enable system/bam.firewall 2>/dev/null
    sudo launchctl kickstart -k system/bam.firewall 2>/dev/null

It's done. Now, start BAM! and have fun.

### Windows

Yeah, Windows! Why not?! Theorically, with some windows expertise, is possible to configure both a DNS server to resolve top-level domain to localhost and the required firewall rules. I tried to configure [localdns][] as a secondary DNS on Windows, but I discovered that the Windows DNS's client [does not query the secondary DNS][dns-windows] before a [15 minute timeout][dns-timeout]. :(

If you are experienced in Windows and want to help us, please take a look [this issue](https://github.com/jweslley/bam/issues/2).


## Running BAM!

After [download][] and put the binary somewhere in your path. Lets run BAM!, for this go to the directory containing your applications and execute it:

    cd /path/to/my/apps
    bam

It's it! By default, BAM! will look for applications in the current directory. To change this or some other configuration generate a configuration file and customize it.

    bam -generate config > ~/.bam.conf

After edit the configuration file, run BAM! using it:

    bam -config ~/.bam.conf


## Upcoming features

Please take a look at issues tracker for [upcoming features](https://github.com/jweslley/bam/labels/feature).


## Bugs and Feedback

If you discover any bugs or have some idea, feel free to create an issue on GitHub:

    http://github.com/jweslley/bam/issues


## License

MIT license. Copyright (c) 2013-2015 Jonhnny Weslley <http://jonhnnyweslley.net>

See the LICENSE file provided with the source distribution for full details.

[procfile]: https://devcenter.heroku.com/articles/procfile
[download]: https://github.com/jweslley/bam/releases
[localtld]: https://github.com/jweslley/localtld
[localdns]: https://github.com/jweslley/localdns
[darwin-resolver]: https://developer.apple.com/library/mac/documentation/Darwin/Reference/ManPages/man5/resolver.5.html
[plist]: https://developer.apple.com/library/mac/documentation/Darwin/Reference/ManPages/man5/plist.5.html
[dns-windows]: https://groups.google.com/forum/#!topic/microsoft.public.windows.server.active_directory/wcNs42YNKeo
[dns-timeout]: https://support.microsoft.com/en-us/kb/320760/en-us?p=1
