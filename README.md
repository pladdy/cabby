[![Go Report Card](https://goreportcard.com/badge/github.com/pladdy/cabby)](https://goreportcard.com/report/github.com/pladdy/cabby)
[![Code Coverage](https://codecov.io/gh/pladdy/cabby/branch/master/graph/badge.svg)](https://codecov.io/gh/pladdy/cabby)
[![Build Status](https://travis-ci.org/pladdy/cabby.svg?branch=master)](https://travis-ci.org/pladdy/cabby)

# cabby
TAXII 2.0 server in Golang.

## Dependencies
- Golang 1.9.x
- SQLite

## Setup
`make`

## Testing
To run all tests: `make test`

"Helper" functions are in `test_helper_test.go`.  The goal with this file was to put repetitive code that make the
tests verbose into a DRY'er format.

## Building
Prod config for linux: https://serverfault.com/questions/413397/how-to-set-environment-variable-in-systemd-service#413408
Debian Policy Manual: https://www.debian.org/doc/debian-policy/#debian-policy-manual
Dependencies in Debian: https://www.debian.org/doc/debian-policy/#s-binarydeps

## Configuration
The `make` task will generate certs and a default config file.  Edit the `config/cabby.json` file to adjust things like
- port
- data store file path
- cert paths

## DB Setup
Using Sqlite as a light-weight data store to run this in development mode.  Goal is to move to some kind of JSON store
(rethinkdb or elasticsearch) in the future.

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

In another terminal, run a server:
`make run`

#### View TAXII Root
```sh
# with headers
curl -isk -basic -u test@cabby.com:test -H 'Accept: application/vnd.oasis.taxii+json' 'https://localhost:1234/taxii/' && echo
# parsed json
curl -sk -basic -u test@cabby.com:test -H 'Accept: application/vnd.oasis.taxii+json' 'https://localhost:1234/taxii/' | jq .
# without a trailing slash
curl -sk --location-trusted -basic -u test@cabby.com:test -H 'Accept: application/vnd.oasis.taxii+json' 'https://localhost:1234/taxii' | jq .
```

#### View API Root
```sh
# with headers
curl -isk -basic -u test@cabby.com:test -H 'Accept: application/vnd.oasis.taxii+json' 'https://localhost:1234/cabby_test_root/' && echo
# parsed json
curl -sk -basic -u test@cabby.com:test -H 'Accept: application/vnd.oasis.taxii+json' 'https://localhost:1234/cabby_test_root/' | jq .
```

#### View Collections
```sh
# with headers
curl -isk -basic -u test@cabby.com:test -H 'Accept: application/vnd.oasis.taxii+json' 'https://localhost:1234/cabby_test_root/collections/' && echo
# parsed json
curl -sk -basic -u test@cabby.com:test -H 'Accept: application/vnd.oasis.taxii+json' 'https://localhost:1234/cabby_test_root/collections/' | jq .
# view 1 of N with headers
curl -isk -basic -u test@cabby.com:test -H 'Accept: application/vnd.oasis.taxii+json' -H 'Range: items 0-0' 'https://localhost:1234/cabby_test_root/collections/' && echo
# view 1 0f N parsed json
curl -sk -basic -u test@cabby.com:test -H 'Accept: application/vnd.oasis.taxii+json' -H 'Range: items 0-0' 'https://localhost:1234/cabby_test_root/collections/' | jq .
# view 2nd of N parsed json
curl -sk -basic -u test@cabby.com:test -H 'Accept: application/vnd.oasis.taxii+json' -H 'Range: items 1-1' 'https://localhost:1234/cabby_test_root/collections/' | jq .
```

#### Create a collection in API Root
Let the server assign an ID:
```sh
curl -sk -basic -u test@cabby.com:test -H 'Accept: application/vnd.oasis.taxii+json' -X POST 'https://localhost:1234/cabby_test_root/collections/' -d '{
  "title": "a collection"
}' | jq .
```

Check it:
```sh
# with headers
curl -isk -basic -u test@cabby.com:test -H 'Accept: application/vnd.oasis.taxii+json' 'https://localhost:1234/cabby_test_root/collections/' && echo
# parsed json
curl -sk -basic -u test@cabby.com:test -H 'Accept: application/vnd.oasis.taxii+json' 'https://localhost:1234/cabby_test_root/collections/' | jq .
```

#### Create a collection with an ID in API Root
```sh
curl -sk -basic -u test@cabby.com:test -H 'Accept: application/vnd.oasis.taxii+json' -X POST 'https://localhost:1234/cabby_test_root/collections/' -d '{
  "title": "another collection",
  "id": "411abc04-a474-4e22-9f4d-944ca508e68c"
}' | jq .
```

Check it:
```sh
# with headers
curl -isk -basic -u test@cabby.com:test -H 'Accept: application/vnd.oasis.taxii+json' 'https://localhost:1234/cabby_test_root/collections/352abc04-a474-4e22-9f4d-944ca508e68c/' && echo
# parsed json
curl -sk -basic -u test@cabby.com:test -H 'Accept: application/vnd.oasis.taxii+json' 'https://localhost:1234/cabby_test_root/collections/352abc04-a474-4e22-9f4d-944ca508e68c/' | jq .
```

#### Add Objects
In the above example, new collections were added.  Kill the server (CTRL+C) and `make run` again.  The logs will show new routes are added.

Now post a bundle of STIX 2.0 data:
```sh
curl -sk -basic -u test@cabby.com:test -H 'Accept: application/vnd.oasis.stix+json' -X POST 'https://localhost:1234/cabby_test_root/collections/352abc04-a474-4e22-9f4d-944ca508e68c/objects/' -d @testdata/malware_bundle.json | jq .
```

#### Check status
From the above POST, you get a status object.  You can query it from the server
```sh
curl -sk -basic -u test@cabby.com:test -H 'Accept: application/vnd.oasis.taxii+json' -X POST 'https://localhost:1234/cabby_test_root/status/191cd890-2a12-4672-9a87-7c846d837119/' | jq .
```

#### View Objects
```sh
# with headers
curl -isk -basic -u test@cabby.com:test -H 'Accept: application/vnd.oasis.stix+json' 'https://localhost:1234/cabby_test_root/collections/352abc04-a474-4e22-9f4d-944ca508e68c/objects/' && echo
# parsed json
curl -sk -basic -u test@cabby.com:test -H 'Accept: application/vnd.oasis.stix+json' 'https://localhost:1234/cabby_test_root/collections/352abc04-a474-4e22-9f4d-944ca508e68c/objects/' | jq .

# view 1 of N with headers
curl -isk -basic -u test@cabby.com:test -H 'Accept: application/vnd.oasis.stix+json' -H 'Range: items 0-0' 'https://localhost:1234/cabby_test_root/collections/352abc04-a474-4e22-9f4d-944ca508e68c/objects/' && echo
# view 1 0f N parsed json
curl -sk -basic -u test@cabby.com:test -H 'Accept: application/vnd.oasis.stix+json' -H 'Range: items 0-0' 'https://localhost:1234/cabby_test_root/collections/352abc04-a474-4e22-9f4d-944ca508e68c/objects/' | jq .
```

#### View Manifest
```sh
# with headers
curl -isk -basic -u test@cabby.com:test -H 'Accept: application/vnd.oasis.taxii+json' 'https://localhost:1234/cabby_test_root/collections/352abc04-a474-4e22-9f4d-944ca508e68c/manifest/' && echo
# parsed json
curl -sk -basic -u test@cabby.com:test -H 'Accept: application/vnd.oasis.taxii+json' 'https://localhost:1234/cabby_test_root/collections/352abc04-a474-4e22-9f4d-944ca508e68c/manifest/' | jq .
```

#### Filter objects
Filters on 'match' requires square brackers that need to be escaped:
```sh
curl -sk -basic -u test@cabby.com:test -H 'Accept: application/vnd.oasis.stix+json' 'https://localhost:1234/cabby_test_root/collections/352abc04-a474-4e22-9f4d-944ca508e68c/objects/?match\[type\]=indicator' | jq .
```

## Resources
- OASIS Doc: https://oasis-open.github.io/cti-documentation/resources
- TAXII 2.0 Spec: https://docs.google.com/document/d/1Jv9ICjUNZrOnwUXtenB1QcnBLO35RnjQcJLsa1mGSkI
- STIX 2.0 Spec: https://docs.oasis-open.org/cti/stix/v2.0/stix-v2.0-part1-stix-core.html
- TLS in Golang Examples: https://gist.github.com/denji/12b3a568f092ab951456
- Perfect SSL Labs Score Article: https://blog.bracebin.com/achieving-perfect-ssl-labs-score-with-go
