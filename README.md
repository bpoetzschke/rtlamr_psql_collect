# Purpose

`rtlamr-psql-collect` is a data collection client for [rtlamr](https://github.com/bemasher/rtlamr). This tool is storing
the data in a PostgreSQL database.

# Requirements

The following requirements have to be fulfilled in order to run `rtlamr-psql-collect`:

- Golang >= 1.16
- PostgreSQL >= 10
- `rtlamr`

To install `rtlamr` run:
```bash
go get github.com/bemasher/rtlamr
```

# Installation
Install `rtlamr-psql-collect` by running:
```bash
go get github.com/bpoetzschke/rtlamr-psql-collect
```

# Usage
`rtlamr-psql-collect` is entirely configured through environment variables.

- `DEBUG` - enable debug log
- `DB_HOST` - Database host; required
- `DB_PORT` - Database port; required
- `DB_USER` - Database user; required
- `DB_PASSWORD` - Database password; required
- `DB_DATABASE` - Database to use; required
- `RTLAMR_SERVER` - IP address and port of rtl_tcp instance; format: IP:PORT
- `RTLAMR_FILTERID` - List your meter id's here separated by commas. Not required but it is suggested to use it.