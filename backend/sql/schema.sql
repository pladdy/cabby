drop table if exists taxii_user;

create table taxii_user (
  id    text not null primary key,
  email text not null check (email != "")
);

drop table if exists taxii_user_pass;

create table taxii_user_pass (
  user_id text not null primary key,
  pass    text not null check (pass != "")
);

drop table if exists taxi_user_collection;

create table taxi_user_collection (
  user_id       text not null,
  collection_id text not null,
  can_read      bool   not null,
  can_write     bool   not null,

  primary key (user_id, collection_id)
);

drop table if exists taxii_collection;

create table taxii_collection (
  id          text not null primary key,
  title       text,
  description text
);
