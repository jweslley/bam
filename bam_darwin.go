package main

func init() {
	configTemplates["config"] = defaultConfig
	configTemplates["firewall"] = `<?xml version="1.0" encoding="UTF-8"?>
<!DOCTYPE plist PUBLIC "-//Apple//DTD PLIST 1.0//EN" "http://www.apple.com/DTDs/PropertyList-1.0.dtd">
<plist version="1.0">
<dict>
	<key>Label</key>
	<string>bam.firewall</string>
	<key>ProgramArguments</key>
	<array>
		<string>/bin/sh</string>
		<string>-c</string>
		<string>
			sysctl -w net.inet.ip.forwarding=1;
			echo "rdr pass proto tcp from any to any port {80,{{.ProxyPort}}} -> 127.0.0.1 port {{.ProxyPort}}" | pfctl -a "com.apple/250.BamFirewall" -Ef -
		</string>
	</array>
	<key>RunAtLoad</key>
	<true/>
	<key>UserName</key>
	<string>root</string>
</dict>
</plist>
`
	configTemplates["help"] = `
Available generate options are 'config' and 'firewall'.

# bam

Generate configuration file with default values for bam.

Example:

    bam -generate config


# firewall

Generate plist file to set up firewall rules for forwarding incoming connections to bam's proxy.

Example:

    bam -generate firewall > bam.firewall.plist

IMPORTANT: If you are using a custom config file you also must use -config option.

Execute the following commands in order to apply these firewall rules using launchd:

    sudo cp bam.firewall.plist /Library/LaunchDaemons/bam.firewall.plist
    sudo launchctl bootstrap system /Library/LaunchDaemons/bam.firewall.plist 2>/dev/null
    sudo launchctl enable system/bam.firewall 2>/dev/null
    sudo launchctl kickstart -k system/bam.firewall 2>/dev/null


In order to get BAM! working properly you also will need to install localdns:

    https://github.com/jweslley/localdns
`
}
