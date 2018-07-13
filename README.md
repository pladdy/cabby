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
Building debian package for a vagrant VM running ubuntu: `make debian-build`

### References
Example: https://fabianlee.org/2017/05/21/golang-running-a-go-binary-as-a-systemd-service-on-ubuntu-16-04/
Prod config for linux: https://serverfault.com/questions/413397/how-to-set-environment-variable-in-systemd-service#413408
Debian Policy Manual: https://www.debian.org/doc/debian-policy/#debian-policy-manual
Dependencies in Debian: https://www.debian.org/doc/debian-policy/#s-binarydeps
How to use systemctl: https://www.digitalocean.com/community/tutorials/how-to-use-systemctl-to-manage-systemd-services-and-units
Understanding systemd: https://www.digitalocean.com/community/tutorials/understanding-systemd-units-and-unit-files

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
curl -isk -basic -u test@cabby.com:test -H 'Accept: application/vnd.oasis.taxii+json' 'https://localhost:1234/taxii/' && echo
# parsed json
curl -sk -basic -u test@cabby.com:test -H 'Accept: application/vnd.oasis.taxii+json' 'https://localhost:1234/taxii/' | jq .
# without a trailing slash
curl -sk --location-trusted -basic -u test@cabby.com:test -H 'Accept: application/vnd.oasis.taxii+json' 'https://localhost:1234/taxii' | jq .
```

#### Delete TAXII Discovery (Admin Functionality)
```sh
curl -sk -basic -u test@cabby.com:test -H 'Accept: application/vnd.oasis.taxii+json' -X DELETE 'https://localhost:1234/admin/discovery/' | jq .
```

Check it:
```sh
# with headers
curl -isk -basic -u test@cabby.com:test -H 'Accept: application/vnd.oasis.taxii+json' 'https://localhost:1234/taxii/' && echo
# parsed json
curl -sk -basic -u test@cabby.com:test -H 'Accept: application/vnd.oasis.taxii+json' 'https://localhost:1234/taxii/' | jq .
```

#### Create TAXII Discovery (Admin Functionality)
```sh
curl -sk -basic -u test@cabby.com:test -H 'Accept: application/vnd.oasis.taxii+json' -X POST 'https://localhost:1234/admin/discovery/' -d '{
  "title": "a local taxii 2 server",
  "description": "this is a test taxii2 server written in golang",
  "contact": "github.com/pladdy",
  "default": "https://localhost:1234/taxii/"
}' | jq .
```

Check it:
```sh
# with headers
curl -isk -basic -u test@cabby.com:test -H 'Accept: application/vnd.oasis.taxii+json' 'https://localhost:1234/taxii/' && echo
# parsed json
curl -sk -basic -u test@cabby.com:test -H 'Accept: application/vnd.oasis.taxii+json' 'https://localhost:1234/taxii/' | jq .
```

#### Update TAXII Discovery (Admin Functionality)
```sh
curl -sk -basic -u test@cabby.com:test -H 'Accept: application/vnd.oasis.taxii+json' -X PUT 'https://localhost:1234/admin/discovery/' -d '{
  "title": "a local taxii 2 server, updated",
  "description": "this is a test taxii2 server written in golang, updated",
  "contact": "github.com/pladdy",
  "default": "https://localhost:1234/taxii/"
}' | jq .
```

Check it:
```sh
# with headers
curl -isk -basic -u test@cabby.com:test -H 'Accept: application/vnd.oasis.taxii+json' 'https://localhost:1234/taxii/' && echo
# parsed json
curl -sk -basic -u test@cabby.com:test -H 'Accept: application/vnd.oasis.taxii+json' 'https://localhost:1234/taxii/' | jq .
```

#### View API Root
```sh
# with headers
curl -isk -basic -u test@cabby.com:test -H 'Accept: application/vnd.oasis.taxii+json' 'https://localhost:1234/cabby_test_root/' && echo
# parsed json
curl -sk -basic -u test@cabby.com:test -H 'Accept: application/vnd.oasis.taxii+json' 'https://localhost:1234/cabby_test_root/' | jq .
```

#### Create API Root (Admin functionality)
```sh
curl -sk -basic -u test@cabby.com:test -H 'Accept: application/vnd.oasis.taxii+json' -X POST 'https://localhost:1234/admin/api_root/' -d '{
  "path": "my_api_root", "title": "my api root"
}' | jq .
```

Check it:
```sh
# with headers
curl -isk -basic -u test@cabby.com:test -H 'Accept: application/vnd.oasis.taxii+json' 'https://localhost:1234/my_api_root/' && echo
# parsed json
curl -sk -basic -u test@cabby.com:test -H 'Accept: application/vnd.oasis.taxii+json' 'https://localhost:1234/my_api_root/' | jq .
```

#### Update API Root (Admin functionality)
```sh
curl -sk -basic -u test@cabby.com:test -H 'Accept: application/vnd.oasis.taxii+json' -X PUT 'https://localhost:1234/admin/api_root/' -d '{
  "path": "my_api_root", "title": "my api root", "max_content_length": 8388608
}' | jq .
```

Check it:
```sh
# with headers
curl -isk -basic -u test@cabby.com:test -H 'Accept: application/vnd.oasis.taxii+json' 'https://localhost:1234/my_api_root/' && echo
# parsed json
curl -sk -basic -u test@cabby.com:test -H 'Accept: application/vnd.oasis.taxii+json' 'https://localhost:1234/my_api_root/' | jq .
```

#### Delete API Root (Admin functionality)
```sh
curl -sk -basic -u test@cabby.com:test -H 'Accept: application/vnd.oasis.taxii+json' -X DELETE 'https://localhost:1234/admin/api_root/' -d '{
  "path": "my_api_root", "title": "my api root", "max_content_length": 8388608
}' | jq .
```

Check it:
```sh
# with headers
curl -isk -basic -u test@cabby.com:test -H 'Accept: application/vnd.oasis.taxii+json' 'https://localhost:1234/my_api_root/' && echo
# parsed json
curl -sk -basic -u test@cabby.com:test -H 'Accept: application/vnd.oasis.taxii+json' 'https://localhost:1234/my_api_root/' | jq .
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

#### Create a collection in API Root (Admin functionality)
Let the server assign an ID:
```sh
curl -sk -basic -u test@cabby.com:test -H 'Accept: application/vnd.oasis.taxii+json' -X POST 'https://localhost:1234/admin/collections/' -d '{
  "api_root_path": "cabby_test_root",
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

#### Create a collection with an ID (Admin functionality)
```sh
curl -sk -basic -u test@cabby.com:test -H 'Accept: application/vnd.oasis.taxii+json' -X POST 'https://localhost:1234/admin/collections/' -d '{
  "api_root_path": "cabby_test_root",
  "title": "another collection",
  "id": "411abc04-a474-4e22-9f4d-944ca508e68c"
}' | jq .
```

Check it:
```sh
# with headers
curl -isk -basic -u test@cabby.com:test -H 'Accept: application/vnd.oasis.taxii+json' 'https://localhost:1234/cabby_test_root/collections/411abc04-a474-4e22-9f4d-944ca508e68c/' && echo
# parsed json
curl -sk -basic -u test@cabby.com:test -H 'Accept: application/vnd.oasis.taxii+json' 'https://localhost:1234/cabby_test_root/collections/411abc04-a474-4e22-9f4d-944ca508e68c/' | jq .
```

#### Update Collection (Admin functionality)
```sh
curl -sk -basic -u test@cabby.com:test -H 'Accept: application/vnd.oasis.taxii+json' -X PUT 'https://localhost:1234/admin/collections/' -d '{
  "api_root_path": "cabby_test_root",
  "title": "a better titled collection",
  "id": "411abc04-a474-4e22-9f4d-944ca508e68c"
}' | jq .
```

Check it:
```sh
# with headers
curl -isk -basic -u test@cabby.com:test -H 'Accept: application/vnd.oasis.taxii+json' 'https://localhost:1234/cabby_test_root/collections/411abc04-a474-4e22-9f4d-944ca508e68c/' && echo
# parsed json
curl -sk -basic -u test@cabby.com:test -H 'Accept: application/vnd.oasis.taxii+json' 'https://localhost:1234/cabby_test_root/collections/411abc04-a474-4e22-9f4d-944ca508e68c/' | jq .
```

#### Delete Collection (Admin functionality)
```sh
curl -sk -basic -u test@cabby.com:test -H 'Accept: application/vnd.oasis.taxii+json' -X DELETE 'https://localhost:1234/admin/collections/' -d '{
  "id": "411abc04-a474-4e22-9f4d-944ca508e68c"
}' | jq .
```

Check it:
```sh
# with headers
curl -isk -basic -u test@cabby.com:test -H 'Accept: application/vnd.oasis.taxii+json' 'https://localhost:1234/cabby_test_root/collections/411abc04-a474-4e22-9f4d-944ca508e68c/' && echo
# parsed json
curl -sk -basic -u test@cabby.com:test -H 'Accept: application/vnd.oasis.taxii+json' 'https://localhost:1234/cabby_test_root/collections/411abc04-a474-4e22-9f4d-944ca508e68c/' | jq .
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
ID=<id here> curl -sk -basic -u test@cabby.com:test -H 'Accept: application/vnd.oasis.taxii+json' -X POST "https://localhost:1234/cabby_test_root/status/${ID}/" | jq .
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
