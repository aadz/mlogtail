# About mlogtail
The main purpose of the program is monitoring of mail service (MTA) by reading new data appearing in log file and counting the values of some parameters characterizing the operation of a mail server. Currently only Postfix logs are supported.

The program has two main usage modes. In the first case (`tail` command), the program reads new data from the log file in background and maintains several counters.

`mlogtail` monitors state of the log file it woking with, so there is no needs to do anything at time of normal logs rotating.

In the second mode, `mlogtail` is used to call to a log reading process and get (and/or reset) current values of the counters.

## Usage

```none
# mlogtail -h
Usage:

  mlogtail [OPTIONS] tail
  mlogtail [OPTIONS] "stats | stats_reset | reset"
  mlogtail [OPTIONS] <COUNTER_NAME>
  mlogtail -f <LOG_FILE_NAME>

Options:
  -f string
        Mail log file path, if the path is "-" then read from STDIN (default "/var/log/mail.log")
  -h    Show this help
  -l string
        Log reader process is listening for commands on a socket file, or IPv4:PORT,
        or [IPv6]:PORT (default "unix:/var/run/mlogtail.sock")
  -o string
        Set a socket OWNER[:GROUP] while listening on a socket file
  -p int
        Set a socket access permissions while listening on a socket file (default 666)
  -t string
        Mail log type. It is "postfix" only allowed for now (default "postfix")
  -v    Show version information and exit
```

### Log tailing mode

Unfortunately, in Go, the process has no good ways to become a demon, so we launch the "reader" just in background:

```none
# mlogtail tail &
```

or by `systemctl`. If a log reading process have to listen to a socket then it is required to specify a netwoirking type. For example: `unix:/tmp/some.sock`.

### Getting counter values

```none
# mlogtail stats
bytes-received  1059498852
bytes-delivered 1039967394
received        2733
delivered       2944
forwarded       4
deferred        121
bounced         105
rejected        4
held            0
discarded       0
```

It should be noted that if the "reader" is started with the `-l` option, setting the socket or IP address and port on which the process is listening for requests, then the same command line parameters should be used for getting counter values.

Probably a more frequent case of addressing the counters is to get the current value of one of them, for example:

```none
# mlogtail bytes-received
1059498852
```
```none
# mlogtail rejected
4
```

### Log file statistics

In addition to working in real time, mlogtail can be used with a mail log file:

```none
# mlogtail -f /var/log/mail.log
```
or STDIN:
```none
# grep '^Apr  1' /var/log/mail.log | mlogtail -f -
```
for example, to get counter for some defined date.

## Installation

```none
go get -u github.com/hpcloud/tail &&
  go build && strip mlogtail &&
  cp mlogtail /usr/local/sbin &&
  chown root:bin /usr/local/sbin/mlogtail &&
  chmod 0711 /usr/local/sbin/mlogtail
```
