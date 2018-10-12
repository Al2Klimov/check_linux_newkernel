## About

The check plugin **check\_linux\_newkernel** monitors
whether a new Linux kernel has been installed since boot.

To check whether a host should be rebooted, use this plugin
together with [check_systemd_needrestart].

## Demonstration

1. `$ docker run -itp 8080:80 grandmaster/check_linux_newkernel`
2. Open http://localhost:8080 and navigate to the (only) service

## Usage

The [plug-and-play Linux binaries]
don't take any CLI arguments or environment variables.

### Legal info

To print the legal info, execute the plugin in a terminal:

```
$ ./check_linux_newkernel
```

In this case the program will always terminate with exit status 3 ("unknown")
without actually checking anything.

### Testing

If you want to actually execute a check inside a terminal,
you have to connect the standard output of the plugin to anything
other than a terminal – e.g. the standard input of another process:

```
$ ./check_linux_newkernel |cat
```

In this case the exit code is likely to be the cat's one.
This can be worked around like this:

```
bash $ set -o pipefail
bash $ ./check_linux_newkernel |cat
```

### Actual monitoring

Just integrate the plugin into the monitoring tool of your choice
like any other check plugin. (Consult that tool's manual on how to do that.)
It should work with any monitoring tool
supporting the [Nagio$ check plugin API].

The only limitation: check\_linux\_newkernel must be run on the host
to be checked for any new kernel –
either with an agent of your monitoring tool or by SSH.
Otherwise it will check the host
your monitoring tool runs on for any new kernel.

#### Icinga 2

This repository ships the [check command definition]
as well as a [service template] and [host example] for [Icinga 2].

The service definition will work in both correctly set up [Icinga 2 clusters]
and Icinga 2 instances not being part of any cluster
as long as the [hosts] are named after the [endpoints].

[check_systemd_needrestart]: https://github.com/Al2Klimov/check_systemd_needrestart
[plug-and-play Linux binaries]: https://github.com/Al2Klimov/check_linux_newkernel/releases
[Nagio$ check plugin API]: https://nagios-plugins.org/doc/guidelines.html#AEN78
[check command definition]: ./icinga2/check_linux_newkernel.conf
[service template]: ./icinga2/check_linux_newkernel-service.conf
[host example]: ./icinga2/check_linux_newkernel-host.conf
[Icinga 2]: https://www.icinga.com/docs/icinga2/latest/doc/01-about/
[Icinga 2 clusters]: https://www.icinga.com/docs/icinga2/latest/doc/06-distributed-monitoring/
[hosts]: https://www.icinga.com/docs/icinga2/latest/doc/09-object-types/#host
[endpoints]: https://www.icinga.com/docs/icinga2/latest/doc/09-object-types/#endpoint
