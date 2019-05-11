# О программе

Основное назначение программы - могниториг почтового сервиса (MTA) путем чтения новых данных, появляющихся в лог-файле, и подсчета значений некоторых параметров, характеризующих работу почтового сервера. В настоящее время поддерживаются только логи Postfix.

У программы два основных режима использования. В первом случае (команда tail) программа в фоновом режиме читает новые данные из лог-файла и ведет несколько счетчиков. mlogtail самостоятельно ослеживает состояние лог-файла, с которым он работает, поэтому при ротации логов не нужно ничего предпринимать.

Во втором - mlogtail используется для обращения в процессу, читающему лог, и получения (и/или обнуления) текущих значений счетчиков.

## Как пользоваться

```none
# mlogtail -h
Usage:

  mlogtail [OPTIONS] tail
  mlogtail [OPTIONS] "stats | stats_reset | reset"
  mlogtail [OPTIONS] <COUNTER_NAME>
  mlogtail -f <LOG_FILE_NAME>

Options:
  -f string
        Mail log file path, if path is "-" then read from STDIN (default "/var/log/mail.log")
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

### Запуск в режиме чтения лога

К сожалению, в Go у процесса нет хороших способов стать демоном, поэтому запускаем "читателя" просто в фоновом режиме:

```none
# mlogtail tail &
```

или при помощи systemctl. Если процесс, читающий лог, должен слушать сокет, то указание типа `unix` обязательно, например `unix:/tmp/some.sock`.

### Получение значений счетчиков

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

Необходимо обратить внимание, что если "читатель" запущен с опцией `-l`, с указанием сокета или IP-адреса и порта, на котором процесс ждет запросов, то и с командой получения значений счетчиков должен использоваться тот же парамер командной строки.

Вероятно, более частый случай обращения к счетчикам - это получение текущего значения одного из них, например:

```none
# mlogtail bytes-received
1059498852
```
```none
# mlogtail rejected
4
```

### Статистика по лог-файлу

Кроме работы в реальном времени mlogtail может использоваться и с лог-файлом:
```none
# mlogtail -f /var/log/mail.log
```
или STDIN:
```none
# grep '^Apr  1' /var/log/mail.log | mlogtail -f -
```
например, для получения данных на определенное число.

## Установка

```none
go get -u github.com/hpcloud/tail &&
  go build && strip mlogtail &&
  cp mlogtail /usr/local/sbin &&
  chown root:bin /usr/local/sbin/mlogtail &&
  chmod 0711 /usr/local/sbin/mlogtail
```
