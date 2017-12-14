[![Go Report Card](https://goreportcard.com/badge/github.com/pladdy/cabby)](https://goreportcard.com/report/github.com/pladdy/cabby)
[![Coverage Status](https://coveralls.io/repos/github/pladdy/cabby/badge.svg)](https://coveralls.io/github/pladdy/cabby)
[![Build Status](https://travis-ci.org/pladdy/cabby.svg?branch=master)](https://travis-ci.org/pladdy/cabby)

# cabby
TAXII 2.0 server in Golang.

## Test
`make test`

## Run
`make run`

## Configure
- `make cert` to generate self signed certs
- `make config`
- Edit the `config/cabby.json` file to adjust settings

## Resources
- Oasis Docs: https://oasis-open.github.io/cti-documentation/resources.html#taxii-20-specification
- TLS in Golang Examples: https://gist.github.com/denji/12b3a568f092ab951456
- Perfect SSL Labs Score Article: https://blog.bracebin.com/achieving-perfect-ssl-labs-score-with-go
