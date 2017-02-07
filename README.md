What
====

Streaming obfuscator of sensitive data in PostgreSQL dumps (pg_dump).

    $ git clone https://github.com/ostrovok-team/pgdump-obfuscator
    $ cd pgdump-obfuscator
    $ go test
    $ go install
    $ pg_dump [...] |pgdump-obfuscator |less  # inspect


Configuration
====

```
pgdump-obfuscator -h

Usage of ./pgdump-obfuscator:
  -c value
    	Configs, example: auth_user:email:email, auth_user:password:bytes
  ...
```

That mean you can provide as many "-c option" as you need.

Example:

```
pgdump-obfuscator -c auth_user:email:email -c auth_user:password:bytes -c address_useraddress:phone_number:digits -c address_useraddress:land_line:digits
```


Who
===

Idea: Denis Orlikhin https://github.com/overplumbum
Initial implementation: Sergey Shepelev https://github.com/temoto
pgdump-obfuscator was implemented during a hackathon at http://ostrovok.ru/ big thanks to Evgeny Kuryshev and everyone involved in organization of this event.
