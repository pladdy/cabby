#!/usr/bin/env bash

set -euo pipefail
# ref: http://redsymbol.net/articles/unofficial-bash-strict-mode/

function usage() {
  echo "Usage: $0 -c [config path] -u [user name] -p [password]"
  exit 1
}

SCRIPT_DIR=$(cd $(dirname $0) && pwd)
CABBY_ENV="${CABBY_ENVIRONMENT:-development}"
USER_ADMIN=0

CABBY_SCHEMA="$SCRIPT_DIR/../sqlite/schema.sql"
if [ "$CABBY_ENV" == "production" ]; then
  CABBY_SCHEMA="/var/cabby/schema.sql"
fi

CONFIG_PATH="config/cabby.json" # default dev config
CABBY_USER=""
PASS=""

while getopts ":c:hp:u:a" opt; do
  case $opt in
    a)
      USER_ADMIN=1
      ;;
    c)
      CONFIG_PATH="$OPTARG"
      ;;
    h)
      usage
      ;;
    p)
      PASS=$(printf "$OPTARG" | shasum -a 256 | cut -d '-' -f 1 | xargs)
      ;;
    u)
      CABBY_USER="$OPTARG"
      ;;
    \?)
      echo "Invalid option: -$OPTARG" >&2
      exit 1
      ;;
    :)
      echo "Option -$OPTARG requires an argument" >&2
      exit 1
      ;;
  esac
done

if [ ! -f "$CONFIG_PATH" ]; then
  echo "Config path $CONFIG_PATH does not exist"
  exit 1
fi

if [ -z $CABBY_USER ]; then
  echo "User not set"
  exit 1
fi

if [ -z $PASS ]; then
  echo "Password not set"
  exit 1
fi

DB_PATH="$(jq .data_store.path $CONFIG_PATH | sed 's/\"//g')"
DB_DIR="$(dirname $DB_PATH)"

mkdir -p "$DB_DIR"
sqlite3 "$DB_PATH" ".read $CABBY_SCHEMA"

# set up user
echo "Creating a user"
sqlite3 $DB_PATH "insert into taxii_user (email) values('$CABBY_USER')"
if [ $USER_ADMIN -eq 1 ]; then
  echo "  Making user admin"
  sqlite3 $DB_PATH "update taxii_user set can_admin = 1 where email = '$CABBY_USER'"
fi

sqlite3 $DB_PATH "insert into taxii_user_pass (email, pass) values('$CABBY_USER', '$PASS')"

# set up discovery
sqlite3 $DB_PATH "
insert into taxii_discovery (title, description, contact, default_url) values(
  'a local taxii 2 server',
  'this is a test taxii2 server written in golang',
  'github.com/pladdy',
  'https://localhost/taxii/'
)"

# set up api root
echo "Creating an API Root"
API_ROOT=cabby_test_root

sqlite3 $DB_PATH "
insert into taxii_api_root (api_root_path, title, description, versions, max_content_length) values (
  '$API_ROOT',
  'a title',
  'a description',
  'application/vnd.oasis.stix+json; version=2.0',
  8388608 /* 8 MB */
)"

# set up collections
echo "Creating a collections"
COLLECTION_IDS="352abc04-a474-4e22-9f4d-944ca508e68c 40aabc04-a474-4e02-cf4d-a44ca901e68c"

for id in $COLLECTION_IDS; do
  echo "  Creating collection $id"
  sqlite3 $DB_PATH "
  insert into taxii_collection (id, api_root_path, title, description, media_types) values (
    '$id',
    '$API_ROOT',
    'a test collection',
    'a test collection description',
    ''
  )"

  echo "  Associate user to collection $id"
  sqlite3 $DB_PATH "
  insert into taxii_user_collection (email, collection_id, can_write, can_read) values (
    '$CABBY_USER',
    '$id',
    1,
    1
  )"
done

if [ "$CABBY_ENV" == "production" ]; then
  echo "CABBY_ENVIRONMENT is production, changing ownership of $DB_PATH to cabby"
  chown cabby:cabby $DB_PATH
fi
