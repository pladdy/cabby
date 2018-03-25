[![Go Report Card](https://goreportcard.com/badge/github.com/pladdy/cabby)](https://goreportcard.com/report/github.com/pladdy/cabby)
[![Code Coverage](https://codecov.io/gh/pladdy/cabby/branch/master/graph/badge.svg)](https://codecov.io/gh/pladdy/cabby)
[![Build Status](https://travis-ci.org/pladdy/cabby.svg?branch=master)](https://travis-ci.org/pladdy/cabby)

# cabby
TAXII 2.0 server in Golang.

## Dependencies
- Golang 1.9.x
- Sqlite

## Setup
`make`

## Testing
To run all tests: `make test`

"Helper" functions are in `test_helper_test.go`.  The goal with this file was to put repetitive code that make the
tests verbose into a DRY'er format.

## Configuration
The `make` task will generate certs and a default config file.  Edit the `config/cabby.json` file to adjust things like
- discovery
- api root definitions
- data store file path

## DB Setup
Using Sqlite as a light-weight data store to run this in development mode.  Goal is to move to some kind of JSON store
(rethinkdb or elasticsearch) in the future.
`make sqlite`

## API Examples with a test user
The examples below require
- sqlite
- jq

On a mac you can install via `brew`:
```sh
brew install sqlite
brew install jq
```

Set up the DB:
```sh
make sqlite

# set up user
pass=`printf test | sha256sum | cut -d '-' -f 1 | xargs`
sqlite3 db/cabby.db "insert into taxii_user (email) values('test@cabby.com')"
sqlite3 db/cabby.db "insert into taxii_user_pass (email, pass) values('test@cabby.com', '${pass}')"

# set up discovery
sqlite3 db/cabby.db '
insert into taxii_discovery (title, description, contact, default_url) values(
  "a local taxii 2 server",
  "this is a test taxii2 server written in golang",
  "github.com/pladdy",
  "https://localhost/taxii/"
)'

# set up api root
sqlite3 db/cabby.db '
insert into taxii_api_root (id, api_root_path, title, description, versions, max_content_length) values (
  "testId",
  "cabby_test_root",
  "a title",
  "a description",
  "application/vnd.oasis.stix+json; version=2.0",
  8388608 /* 8 MB */
)'
```

In another terminal, run a server:
```sh
make run
```

##### View TAXII Root
```sh
curl -sk -basic -u test@cabby.com:test -H 'Accept: application/vnd.oasis.taxii+json' 'https://localhost:1234/taxii/' | jq .
# without a trailing slash
curl -sk --location-trusted -basic -u test@cabby.com:test -H 'Accept: application/vnd.oasis.taxii+json' 'https://localhost:1234/taxii' | jq .
```

##### View API Root
```sh
curl -sk -basic -u test@cabby.com:test -H 'Accept: application/vnd.oasis.taxii+json' 'https://localhost:1234/cabby_test_root/' | jq .
```

##### Create a collection in API Root
Let the server assign an ID:
```sh
curl -sk -basic -u test@cabby.com:test -H 'Accept: application/vnd.oasis.taxii+json' -X POST 'https://localhost:1234/cabby_test_root/collections/' -d '{
  "title": "a collection"
}' | jq .
```

Check it:
```sh
curl -sk -basic -u test@cabby.com:test -H 'Accept: application/vnd.oasis.taxii+json' 'https://localhost:1234/cabby_test_root/collections/' | jq .
```

##### Create a collection with an ID in API Root
```sh
curl -sk -basic -u test@cabby.com:test -H 'Accept: application/vnd.oasis.taxii+json' -X POST 'https://localhost:1234/cabby_test_root/collections/' -d '{
  "title": "a collection",
  "id": "352abc04-a474-4e22-9f4d-944ca508e68c"
}' | jq .
```

Check it:
```sh
curl -sk -basic -u test@cabby.com:test -H 'Accept: application/vnd.oasis.taxii+json' 'https://localhost:1234/cabby_test_root/collections/352abc04-a474-4e22-9f4d-944ca508e68c' | jq .
```

## Resources
- OASIS Doc: https://oasis-open.github.io/cti-documentation/resources
- TAXII 2.0 Spec: https://docs.google.com/document/d/1Jv9ICjUNZrOnwUXtenB1QcnBLO35RnjQcJLsa1mGSkI
- STIX 2.0 Spec: https://docs.oasis-open.org/cti/stix/v2.0/stix-v2.0-part1-stix-core.html
- TLS in Golang Examples: https://gist.github.com/denji/12b3a568f092ab951456
- Perfect SSL Labs Score Article: https://blog.bracebin.com/achieving-perfect-ssl-labs-score-with-go
