[![Go Report Card](https://goreportcard.com/badge/github.com/pladdy/cabby2)](https://goreportcard.com/report/github.com/pladdy/cabby2)
[![Code Coverage](https://codecov.io/gh/pladdy/cabby2/branch/master/graph/badge.svg)](https://codecov.io/gh/pladdy/cabby2)
[![Build Status](https://travis-ci.org/pladdy/cabby2.svg?branch=master)](https://travis-ci.org/pladdy/cabby2)

# cabby
TAXII 2.0 server in Golang.

# re-org
I'm reorganizing the code base based on this article: https://medium.com/@benbjohnson/standard-package-layout-7cdbc8391fc1
So far i find it helpful and it's made me think about isolating dependencies, dependency injection, etc.

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
- Example: https://fabianlee.org/2017/05/21/golang-running-a-go-binary-as-a-systemd-service-on-ubuntu-16-04/
- Prod config for linux: https://serverfault.com/questions/413397/how-to-set-environment-variable-in-systemd-service#413408
- Debian Policy Manual: https://www.debian.org/doc/debian-policy/#debian-policy-manual
- Dependencies in Debian: https://www.debian.org/doc/debian-policy/#s-binarydeps
- How to use systemctl: https://www.digitalocean.com/community/tutorials/how-to-use-systemctl-to-manage-systemd-services-and-units
- Understanding systemd: https://www.digitalocean.com/community/tutorials/understanding-systemd-units-and-unit-files

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

#### View Collection
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
curl -sk -basic -u test@cabby.com:test -H 'Accept: application/vnd.oasis.stix+json' -X POST 'https://localhost:1234/cabby_test_root/collections/352abc04-a474-4e22-9f4d-944ca508e68c/objects/' -d @sqlite/testdata/malware_bundle.json | jq .
```

#### Check status
From the above POST, you get a status object.  You can query it from the server
```sh
curl -sk -basic -u test@cabby.com:test -H 'Accept: application/vnd.oasis.taxii+json' "https://localhost:1234/cabby_test_root/status/<your id here>/" | jq .
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
```sh
# filter on types
curl -sk -basic -u test@cabby.com:test -H 'Accept: application/vnd.oasis.stix+json' 'https://localhost:1234/cabby_test_root/collections/352abc04-a474-4e22-9f4d-944ca508e68c/objects/?match\[type\]=indicator,malware' | jq .
# filter on id
curl -sk -basic -u test@cabby.com:test -H 'Accept: application/vnd.oasis.stix+json' 'https://localhost:1234/cabby_test_root/collections/352abc04-a474-4e22-9f4d-944ca508e68c/objects/?match\[id\]=indicator--8e2e2d2b-17d4-4cbf-938f-98ee46b3cd3f' | jq .

# add objects to filter on versions
# the below bundle has objects that already exist; status will have 3 failures
curl -sk -basic -u test@cabby.com:test -H 'Accept: application/vnd.oasis.stix+json' -X POST 'https://localhost:1234/cabby_test_root/collections/352abc04-a474-4e22-9f4d-944ca508e68c/objects/' -d @sqlite/testdata/versions_bundle.json | jq .

# check status to confirm
curl -sk -basic -u test@cabby.com:test -H 'Accept: application/vnd.oasis.taxii+json' "https://localhost:1234/cabby_test_root/status/<your id here>/" | jq .

# filter on latest versions (indicator will be 2018)
curl -sk -basic -u test@cabby.com:test -H 'Accept: application/vnd.oasis.stix+json' 'https://localhost:1234/cabby_test_root/collections/352abc04-a474-4e22-9f4d-944ca508e68c/objects/?match\[version\]=last' | jq .
# filter on oldest versions (indicator will be 2016)
curl -sk -basic -u test@cabby.com:test -H 'Accept: application/vnd.oasis.stix+json' 'https://localhost:1234/cabby_test_root/collections/352abc04-a474-4e22-9f4d-944ca508e68c/objects/?match\[version\]=first' | jq .
# filter on specific versions (indicator will be 2017)
curl -sk -basic -u test@cabby.com:test -H 'Accept: application/vnd.oasis.stix+json' 'https://localhost:1234/cabby_test_root/collections/352abc04-a474-4e22-9f4d-944ca508e68c/objects/?match\[version\]=2017-01-01T12:15:12.123Z' | jq .
```

### Admin: Discovery

#### Delete TAXII Discovery
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

#### Create TAXII Discovery
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

#### Update TAXII Discovery
```sh
curl -sk -basic -u test@cabby.com:test -H 'Accept: application/vnd.oasis.taxii+json' -X PUT 'https://localhost:1234/admin/discovery/' -d '{
  "title": "a local taxii 2 server, updated",
  "description": "this is a test taxii2 server written in golang, updated",
  "contact": "github.com/pladdy",
  "default": "https://localhost:1234/taxii/"
}' | jq .
```

### Admin: API Roots

#### Create API Root
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

#### Update API Root
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

#### Delete API Root
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

### Admin: Collections

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

### Admin: Users

#### Create user
```sh
curl -sk -basic -u test@cabby.com:test -H 'Accept: application/vnd.oasis.taxii+json' -X POST 'https://localhost:1234/admin/user/' -d '{
  "email": "new_test@cabby.com",
  "password": "new_test"
}' | jq .
```

Check it:
```sh
# with headers
curl -isk -basic -u new_test@cabby.com:new_test -H 'Accept: application/vnd.oasis.taxii+json' 'https://localhost:1234/taxii/' && echo
# parsed json
curl -sk -basic -u new_test@cabby.com:new_test -H 'Accept: application/vnd.oasis.taxii+json' 'https://localhost:1234/taxii/' | jq .
```

#### Update user
```sh
curl -sk -basic -u test@cabby.com:test -H 'Accept: application/vnd.oasis.taxii+json' -X PUT 'https://localhost:1234/admin/user/' -d '{
  "email": "new_test@cabby.com",
  "can_admin": true
}' | jq .
```

Check it (by creating a collection with the new admin):
```sh
curl -sk -basic -u new_test@cabby.com:new_test -H 'Accept: application/vnd.oasis.taxii+json' -X POST 'https://localhost:1234/admin/collections/' -d '{
  "api_root_path": "new_test_cabby_collection",
  "title": "a collection, from a new admin"
}' | jq .
```

Check it:
```sh
# with headers
curl -isk -basic -u new_test@cabby.com:new_test -H 'Accept: application/vnd.oasis.taxii+json' 'https://localhost:1234/cabby_test_root/collections/' && echo
# parsed json
curl -sk -basic -u new_test@cabby.com:new_test -H 'Accept: application/vnd.oasis.taxii+json' 'https://localhost:1234/cabby_test_root/collections/' | jq .
```

#### Delete user
```sh
curl -sk -basic -u test@cabby.com:test -H 'Accept: application/vnd.oasis.taxii+json' -X DELETE 'https://localhost:1234/admin/user/' -d '{
  "email": "new_test@cabby.com"
}' | jq .
```

Check it:
```sh
# with headers
curl -isk -basic -u new_test@cabby.com:new_test -H 'Accept: application/vnd.oasis.taxii+json' 'https://localhost:1234/taxii/' && echo
# parsed json
curl -sk -basic -u new_test@cabby.com:new_test -H 'Accept: application/vnd.oasis.taxii+json' 'https://localhost:1234/taxii/' | jq .
```

#### Update user password
```sh
curl -sk -basic -u test@cabby.com:test -H 'Accept: application/vnd.oasis.taxii+json' -X PUT 'https://localhost:1234/admin/user/password/' -d '{
  "email": "test@cabby.com",
  "password": "new_test"
}' | jq .
```

Check it by accessing the root with old password (should fail)
```sh
# with headers
curl -isk -basic -u test@cabby.com:test -H 'Accept: application/vnd.oasis.taxii+json' 'https://localhost:1234/taxii/' && echo
# parsed json
curl -sk -basic -u test@cabby.com:test -H 'Accept: application/vnd.oasis.taxii+json' 'https://localhost:1234/taxii/' | jq .
```

Check it by accessing the root with new password (should respond)
```sh
# with headers
curl -isk -basic -u test@cabby.com:new_test -H 'Accept: application/vnd.oasis.taxii+json' 'https://localhost:1234/taxii/' && echo
# parsed json
curl -sk -basic -u test@cabby.com:new_test -H 'Accept: application/vnd.oasis.taxii+json' 'https://localhost:1234/taxii/' | jq .
```

Change it back:
```sh
curl -sk -basic -u test@cabby.com:test -H 'Accept: application/vnd.oasis.taxii+json' -X PUT 'https://localhost:1234/admin/user/password/' -d '{
  "email": "test@cabby.com",
  "password": "test"
}' | jq .
```

#### Add collection to user
User a new user (see create user above)

```sh
curl -sk -basic -u test@cabby.com:test -H 'Accept: application/vnd.oasis.taxii+json' -X POST 'https://localhost:1234/admin/user/collection/' -d '{
  "email": "new_test@cabby.com",
  "collection_access": {
    "id": "411abc04-a474-4e22-9f4d-944ca508e68c",
    "can_read": true,
    "can_write": false
  }
}' | jq .
```

Check it
```sh
# with headers
curl -isk -basic -u new_test@cabby.com:new_test -H 'Accept: application/vnd.oasis.taxii+json' 'https://localhost:1234/cabby_test_root/collections/' && echo
# parsed json
curl -sk -basic -u new_test@cabby.com:new_test -H 'Accept: application/vnd.oasis.taxii+json' 'https://localhost:1234/cabby_test_root/collections/' | jq .
```

#### Update collection for user
```sh
curl -sk -basic -u test@cabby.com:test -H 'Accept: application/vnd.oasis.taxii+json' -X PUT 'https://localhost:1234/admin/user/collection/' -d '{
  "email": "new_test@cabby.com",
  "collection_access": {
    "id": "411abc04-a474-4e22-9f4d-944ca508e68c",
    "can_read": true,
    "can_write": true
  }
}' | jq .
```

Check it
```sh
# with headers
curl -isk -basic -u new_test@cabby.com:new_test -H 'Accept: application/vnd.oasis.taxii+json' 'https://localhost:1234/cabby_test_root/collections/' && echo
# parsed json
curl -sk -basic -u new_test@cabby.com:new_test -H 'Accept: application/vnd.oasis.taxii+json' 'https://localhost:1234/cabby_test_root/collections/' | jq .
```

#### Delete collection from user
```sh
curl -sk -basic -u test@cabby.com:test -H 'Accept: application/vnd.oasis.taxii+json' -X DELETE 'https://localhost:1234/admin/user/collection/' -d '{
  "email": "new_test@cabby.com",
  "collection_access": {
    "id": "411abc04-a474-4e22-9f4d-944ca508e68c"
  }
}' | jq .
```

Check it
```sh
# with headers
curl -isk -basic -u new_test@cabby.com:new_test -H 'Accept: application/vnd.oasis.taxii+json' 'https://localhost:1234/cabby_test_root/collections/' && echo
# parsed json
curl -sk -basic -u new_test@cabby.com:new_test -H 'Accept: application/vnd.oasis.taxii+json' 'https://localhost:1234/cabby_test_root/collections/' | jq .
```

## Resources
- OASIS Doc: https://oasis-open.github.io/cti-documentation/resources
- TAXII 2.0 Spec: https://docs.google.com/document/d/1Jv9ICjUNZrOnwUXtenB1QcnBLO35RnjQcJLsa1mGSkI
- STIX 2.0 Spec: https://docs.oasis-open.org/cti/stix/v2.0/stix-v2.0-part1-stix-core.html
- TLS in Golang Examples: https://gist.github.com/denji/12b3a568f092ab951456
- Perfect SSL Labs Score Article: https://blog.bracebin.com/achieving-perfect-ssl-labs-score-with-go
