## Configuring firewall in OSX

Copy the file `bam.firewall.plist` to directory `/Library/LaunchDaemons/`:

    sudo cp bam.firewall.plist /Library/LaunchDaemons/bam.firewall.plist

Execute the following commands in order to apply these firewall rules using launchd:

    sudo launchctl bootstrap system /Library/LaunchDaemons/bam.firewall.plist 2>/dev/null
    sudo launchctl enable system/bam.firewall 2>/dev/null
    sudo launchctl kickstart -k system/bam.firewall 2>/dev/null

Note: The `bam.firewall.plist` file assumes BAM will run at port **42042**, if you will execute it in a diferent port just edit the file before following the instructions above.
