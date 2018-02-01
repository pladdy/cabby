[![Go Report Card](https://goreportcard.com/badge/github.com/pladdy/cabby)](https://goreportcard.com/report/github.com/pladdy/cabby)
[![Coverage Status](https://coveralls.io/repos/github/pladdy/cabby/badge.svg)](https://coveralls.io/github/pladdy/cabby)
[![Build Status](https://travis-ci.org/pladdy/cabby.svg?branch=master)](https://travis-ci.org/pladdy/cabby)

# cabby
TAXII 2.0 server in Golang.

## Dependencies
- Golang 1.9.x
- Sqlite

## Setup
`make`

## Test
`make test`

## Run
`make run`

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

On a mac you can install via `brew`

Run a server and set up a user first:
```sh
make
make run
```

In a new terminal:
```sh
make sqlite
pass=`printf test | sha256sum | cut -d '-' -f 1 | xargs`
  sqlite3 db/cabby.db "insert into taxii_user (email) values('test@cabby.com')"
sqlite3 db/cabby.db "insert into taxii_user_pass (email, pass) values('test@cabby.com', '${pass}')"
```

##### View TAXII Root
```sh
curl -sk -basic -u test@cabby.com:test 'https://localhost:1234/taxii/' | jq .
# without a trailing slash
curl -sk --location-trusted -basic -u test@cabby.com:test 'https://localhost:1234/taxii' | jq .
```

##### View API Root
```sh
curl -sk -basic -u test@cabby.com:test 'https://localhost:1234/api_root/' | jq .`
```

##### Create a collection
Let the server assign an ID:
```sh
curl -sk -basic -u test@cabby.com:test -X POST 'https://localhost:1234/api_root/collections/' -d '{
  "title": "a collection"
}' | jq .
```

Check it:
```sh
curl -sk -basic -u test@cabby.com:test 'https://localhost:1234/api_root/collections/' | jq .
```

Assign an ID in request:
```sh
curl -sk -basic -u test@cabby.com:test -X POST 'https://localhost:1234/api_root/collections/' -d '{
  "title": "a collection",
  "id": "352abc04-a474-4e22-9f4d-944ca508e68c"
}' | jq .
```

Check it:
```sh
curl -sk -basic -u test@cabby.com:test 'https://localhost:1234/api_root/collections/352abc04-a474-4e22-9f4d-944ca508e68c' | jq .
```

## Resources
- Oasis Docs: https://oasis-open.github.io/cti-documentation/resources.html#taxii-20-specification
- TLS in Golang Examples: https://gist.github.com/denji/12b3a568f092ab951456
- Perfect SSL Labs Score Article: https://blog.bracebin.com/achieving-perfect-ssl-labs-score-with-go
