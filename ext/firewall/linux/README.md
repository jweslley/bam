## Configuring firewall in Linux

By default, BAM run at port 80, a privileged port. But, if you wanna use without a privileged user, you need to run using high ports. Thus, you will need some firewalls tricks in order to transfer the trafic from port 80 to your preferred high port. In the file

Apply these rules to iptables using the following command:

    sudo iptables-restore < bam.iptables.rules

Check the rules applied to iptables by executing:

    sudo iptables -t nat -L -n

These rules will not survive after system's restart. Depending your distro, there are distinct ways to make these rules permanent.

Note: The `bam.iptables.rules` file assumes BAM will run at port **42042**, if you will execute it in a diferent port just edit the file before following the instructions above.
