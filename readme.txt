What
====

Streaming obfuscator of sensitive data in PostgreSQL dumps (pg_dump).

    $ git clone https://github.com/ostrovok-team/pgdump-obfuscator
    $ cd pgdump-obfuscator
    $ go test
    $ go install
    $ pg_dump [...] |pgdump-obfuscator |less  # inspect


TODO
====

* Config file. Currently obfuscation rules are hardcoded in config.go, so you have to recompile to try new rules. Easy with `go run`.


Docker
====

```
docker run -v /path/to/destination:/data ostrovok-team/pgdump-obfuscator zcat /data/dump.gz | pgdump-obfuscator | gzip > /data/dump.obfuscated.gz
```

Who
===

Idea: Denis Orlikhin https://github.com/overplumbum
Initial implementation: Sergey Shepelev https://github.com/temoto
pgdump-obfuscator was implemented during a hackathon at http://ostrovok.ru/ big thanks to Evgeny Kuryshev and everyone involved in organization of this event.
