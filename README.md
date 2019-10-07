[![Go Report Card](https://goreportcard.com/badge/github.com/pladdy/cabby)](https://goreportcard.com/report/github.com/pladdy/cabby)
[![Code Coverage](https://codecov.io/gh/pladdy/cabby/branch/master/graph/badge.svg)](https://codecov.io/gh/pladdy/cabby)
[![Build Status](https://travis-ci.com/pladdy/cabby.svg?branch=master)](https://travis-ci.com/pladdy/cabby)
[![Go Doc](https://img.shields.io/badge/godoc-reference-blue.svg?style=flat-square)](http://godoc.org/github.com/pladdy/cabby)
[![Go City](https://img.shields.io/badge/go--city-view-blue.svg)](https://go-city.github.io/#/github.com/pladdy/cabby)
[![Release](https://img.shields.io/github/release/golang-standards/project-layout.svg?style=flat-square)](https://github.com/pladdy/cabby/releases/latest)

# cabby
TAXII 2.0 server in Golang.

## Dependencies
- Golang 1.10 or 1.11
- SQLite

## Setup
`make`

## Testing
To run all tests: `make test`

"Helper" functions are in `test_helper_test.go`.  The goal with this file was to put repetitive code that make the
tests verbose into a DRY'er format.

### GoSec
Security checking tool: https://github.com/securego/gosec
`make sec` to run security tests
`make sec pkg=<package>` to run against a specific package.  Example: `make sec pkg=sqlite` to run it against sqlite package

## Building
Cabby uses sqlite3 as its data store.  The library in golang being used requires C extensions/bindings and as such
doesn't build on a Mac for a Ubuntu OS.  Therefore vagrant vm is used to build a debian for ubuntu (i know...gross).

Building debian package for a vagrant VM running ubuntu: `make build-debian`

### Building: Troubleshooting on the VM
Reference: https://www.digitalocean.com/community/tutorials/how-to-use-journalctl-to-view-and-manipulate-systemd-logs
```sh
# tail the log for cabby
sudo journalctl -u cabby -f

# pipe it to less
sudo journalctl -u cabby | less
```

### Building: Metrics/Logs on the VM (TICK Stack)
I used this project to also explore the TICK stack: https://www.influxdata.com/time-series-platform/

To use/play with it, run `make vagrant`

Syslog input reference (the VM uses syslog to as an input to influx-db):
- https://github.com/influxdata/telegraf/blob/release-1.8/plugins/inputs/syslog/README.md

### Building: References
- Example: https://fabianlee.org/2017/05/21/golang-running-a-go-binary-as-a-systemd-service-on-ubuntu-16-04/
- Prod config for linux: https://serverfault.com/questions/413397/how-to-set-environment-variable-in-systemd-service#413408
- Debian Policy Manual: https://www.debian.org/doc/debian-policy/#debian-policy-manual
- Dependencies in Debian: https://www.debian.org/doc/debian-policy/#s-binarydeps
- Depend Differences: https://askubuntu.com/questions/83553/what-is-the-difference-between-dependencies-and-pre-depends#83559
- How to use systemctl: https://www.digitalocean.com/community/tutorials/how-to-use-systemctl-to-manage-systemd-services-and-units
- Understanding systemd: https://www.digitalocean.com/community/tutorials/understanding-systemd-units-and-unit-files

## Versioning

Release branches are associated to a TAXII spec version and a [SemVer](https://semver.org/) version down to the minor revision.

Patches are annotated by creating a tag off of the release branch.

Examples:

1. First major release of a TAXII 2.0 server:
  1. branch: `release/2.0/1.0`
  1. tag: `release/2.0/1.0.0`
1. A version of TAXII 2.0 spec server that has backward incompatible changes BUT is still a TAXII 2.0 server
  1. branch: `release/2.0/2.0`
  1. tag: `release/2.0/2.0.0`
1. Initial branch for a 2.1 spec server:
  1. branch: `release/2.1/0.0`
  1. tag: `release/2.1/0.0.1` (once a chunk of code worth tagging is ready)

## Configuration
The `make` task will generate certs and a default config file.  Edit the `config/cabby.json` file to adjust things like
- port
- data store file path
- cert paths

## DB Setup
Using Sqlite as a light-weight data store to run this in development mode.  Goal is to move to some kind of JSON store
(rethinkdb or elasticsearch) in the future.  See below API examples for setup instructions.

## API Examples with a test user
The examples below require
- jq
- sqlite

On a mac you can install via `brew`:
```sh
brew install sqlite
brew install jq
```

Set up the DB for dev/test:
`make dev-db`
The user set up with `make dev-db` is an admin (so it can do admin things like create/update/delete certain resources).

In another terminal, run a server:
`make run`

#### View TAXII Discovery
```sh
# with headers
curl -isk -basic -u test@cabby.com:test-password -H 'Accept: application/vnd.oasis.taxii+json' 'https://localhost:1234/taxii2/' && echo
# parsed json
curl -sk -basic -u test@cabby.com:test-password -H 'Accept: application/vnd.oasis.taxii+json' 'https://localhost:1234/taxii2/' | jq .
# without a trailing slash
curl -sk --location-trusted -basic -u test@cabby.com:test-password -H 'Accept: application/vnd.oasis.taxii+json' 'https://localhost:1234/taxii2' | jq .
```

#### View API Root
```sh
# with headers
curl -isk -basic -u test@cabby.com:test-password -H 'Accept: application/vnd.oasis.taxii+json' 'https://localhost:1234/cabby_test_root/' && echo
# parsed json
curl -sk -basic -u test@cabby.com:test-password -H 'Accept: application/vnd.oasis.taxii+json' 'https://localhost:1234/cabby_test_root/' | jq .
```

#### View Collections
```sh
# with headers
curl -isk -basic -u test@cabby.com:test-password -H 'Accept: application/vnd.oasis.taxii+json' 'https://localhost:1234/cabby_test_root/collections/' && echo
# parsed json
curl -sk -basic -u test@cabby.com:test-password -H 'Accept: application/vnd.oasis.taxii+json' 'https://localhost:1234/cabby_test_root/collections/' | jq .
# view 1 of N with headers
curl -isk -basic -u test@cabby.com:test-password -H 'Accept: application/vnd.oasis.taxii+json' -H 'Range: items 0-0' 'https://localhost:1234/cabby_test_root/collections/' && echo
# view 1 0f N parsed json
curl -sk -basic -u test@cabby.com:test-password -H 'Accept: application/vnd.oasis.taxii+json' -H 'Range: items 0-0' 'https://localhost:1234/cabby_test_root/collections/' | jq .
# view 2nd of N parsed json
curl -sk -basic -u test@cabby.com:test-password -H 'Accept: application/vnd.oasis.taxii+json' -H 'Range: items 1-1' 'https://localhost:1234/cabby_test_root/collections/' | jq .
```

#### View Collection
```sh
# with headers
curl -isk -basic -u test@cabby.com:test-password -H 'Accept: application/vnd.oasis.taxii+json' 'https://localhost:1234/cabby_test_root/collections/352abc04-a474-4e22-9f4d-944ca508e68c/' && echo
# parsed json
curl -sk -basic -u test@cabby.com:test-password -H 'Accept: application/vnd.oasis.taxii+json' 'https://localhost:1234/cabby_test_root/collections/352abc04-a474-4e22-9f4d-944ca508e68c/' | jq .
```

#### Add Objects
In the above example, new collections were added.  Kill the server (CTRL+C) and `make run` again.  The logs will show new routes are added.

Now post a bundle of STIX 2.0 data:
```sh
curl -sk -basic -u test@cabby.com:test-password -H 'Accept: application/vnd.oasis.taxii+json' -H 'Content-Type: application/vnd.oasis.stix+json' -X POST 'https://localhost:1234/cabby_test_root/collections/352abc04-a474-4e22-9f4d-944ca508e68c/objects/' -d @sqlite/testdata/malware_bundle.json | jq .
```

#### Check status
From the above POST, you get a status object.  You can query it from the server
```sh
curl -sk -basic -u test@cabby.com:test-password -H 'Accept: application/vnd.oasis.taxii+json' "https://localhost:1234/cabby_test_root/status/<your id here>/" | jq .
```

#### View Objects
```sh
# with headers
curl -isk -basic -u test@cabby.com:test-password -H 'Accept: application/vnd.oasis.stix+json' 'https://localhost:1234/cabby_test_root/collections/352abc04-a474-4e22-9f4d-944ca508e68c/objects/' && echo
# parsed json
curl -sk -basic -u test@cabby.com:test-password -H 'Accept: application/vnd.oasis.stix+json' 'https://localhost:1234/cabby_test_root/collections/352abc04-a474-4e22-9f4d-944ca508e68c/objects/' | jq .

# view 1 of N with headers
curl -isk -basic -u test@cabby.com:test-password -H 'Accept: application/vnd.oasis.stix+json' -H 'Range: items 0-0' 'https://localhost:1234/cabby_test_root/collections/352abc04-a474-4e22-9f4d-944ca508e68c/objects/' && echo
# view 1 0f N parsed json
curl -sk -basic -u test@cabby.com:test-password -H 'Accept: application/vnd.oasis.stix+json' -H 'Range: items 0-0' 'https://localhost:1234/cabby_test_root/collections/352abc04-a474-4e22-9f4d-944ca508e68c/objects/' | jq .
```

#### View Manifest
```sh
# with headers
curl -isk -basic -u test@cabby.com:test-password -H 'Accept: application/vnd.oasis.taxii+json' 'https://localhost:1234/cabby_test_root/collections/352abc04-a474-4e22-9f4d-944ca508e68c/manifest/' && echo
# parsed json
curl -sk -basic -u test@cabby.com:test-password -H 'Accept: application/vnd.oasis.taxii+json' 'https://localhost:1234/cabby_test_root/collections/352abc04-a474-4e22-9f4d-944ca508e68c/manifest/' | jq .
```

#### Filter objects
```sh
# filter on types
curl -sk -basic -u test@cabby.com:test-password -H 'Accept: application/vnd.oasis.stix+json' 'https://localhost:1234/cabby_test_root/collections/352abc04-a474-4e22-9f4d-944ca508e68c/objects/?match\[type\]=indicator,malware' | jq .
# filter on id
curl -sk -basic -u test@cabby.com:test-password -H 'Accept: application/vnd.oasis.stix+json' 'https://localhost:1234/cabby_test_root/collections/352abc04-a474-4e22-9f4d-944ca508e68c/objects/?match\[id\]=indicator--8e2e2d2b-17d4-4cbf-938f-98ee46b3cd3f' | jq .

# add objects to filter on versions
# the below bundle has objects that already exist; status will have 3 failures
curl -sk -basic -u test@cabby.com:test-password -H 'Accept: application/vnd.oasis.stix+json' -X POST 'https://localhost:1234/cabby_test_root/collections/352abc04-a474-4e22-9f4d-944ca508e68c/objects/' -d @sqlite/testdata/versions_bundle.json | jq .

# check status to confirm
curl -sk -basic -u test@cabby.com:test-password -H 'Accept: application/vnd.oasis.taxii+json' "https://localhost:1234/cabby_test_root/status/<your id here>/" | jq .

# filter on latest versions (indicator will be 2018)
curl -sk -basic -u test@cabby.com:test-password -H 'Accept: application/vnd.oasis.stix+json' 'https://localhost:1234/cabby_test_root/collections/352abc04-a474-4e22-9f4d-944ca508e68c/objects/?match\[version\]=last' | jq .
# filter on oldest versions (indicator will be 2016)
curl -sk -basic -u test@cabby.com:test-password -H 'Accept: application/vnd.oasis.stix+json' 'https://localhost:1234/cabby_test_root/collections/352abc04-a474-4e22-9f4d-944ca508e68c/objects/?match\[version\]=first' | jq .
# filter on specific versions (indicator will be 2017)
curl -sk -basic -u test@cabby.com:test-password -H 'Accept: application/vnd.oasis.stix+json' 'https://localhost:1234/cabby_test_root/collections/352abc04-a474-4e22-9f4d-944ca508e68c/objects/?match\[version\]=2017-01-01T12:15:12.123Z' | jq .
```

## Resources
- [OASIS Resources](https://oasis-open.github.io/cti-documentation/resources)
  - [TAXII 2.1 Spec](https://docs.google.com/document/d/1EsiWY7TGqt9yH6QUXv4c-opXSr3wR0TDMt8Q0yJjpoo)
  - [STIX 2.0 Spec](https://docs.oasis-open.org/cti/stix/v2.0/stix-v2.0-part1-stix-core.html)
  - [STIX/TAXII Graphics](https://freetaxii.github.io/)
- [TLS in Golang Examples](https://gist.github.com/denji/12b3a568f092ab951456)
  - [Perfect SSL Labs Score Article](https://blog.bracebin.com/achieving-perfect-ssl-labs-score-with-go)
