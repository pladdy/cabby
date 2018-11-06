PRAGMA foreign_keys = ON;

drop table if exists api_root;

create table api_root (
  id                 integer not null primary key,
  discovery_id       integer check(discovery_id = 1) default 1,
  api_root_path      text    check(api_root_path != "") not null,
  title              text    not null,
  description        text,
  versions           text,
  max_content_length integer not null,
  created_at         text,
  updated_at         text,

  unique(api_root_path) on conflict fail
);

  create trigger api_root_ai_created_at after insert on api_root
    begin
      update api_root set created_at = strftime('%Y-%m-%dT%H:%M:%fZ', 'now') where id = new.id;
      update api_root set updated_at = strftime('%Y-%m-%dT%H:%M:%fZ', 'now') where id = new.id;
    end;

  create trigger api_root_au_updated_at after update on api_root
    begin
      update api_root set updated_at = strftime('%Y-%m-%dT%H:%M:%fZ', 'now') where id = new.id;
    end;

drop table if exists collection;

create table collection (
  id            text not null primary key,
  api_root_path text not null,
  title         text,
  description   text,
  media_types   text default '',
  created_at    text,
  updated_at    text
);

  create trigger collection_ai_created_at after insert on collection
    begin
      update collection set created_at = strftime('%Y-%m-%dT%H:%M:%fZ', 'now') where id = new.id;
      update collection set updated_at = strftime('%Y-%m-%dT%H:%M:%fZ', 'now') where id = new.id;
    end;

  create trigger collection_au_updated_at after update on collection
    begin
      update collection set updated_at = strftime('%Y-%m-%dT%H:%M:%fZ', 'now') where id = new.id;
    end;

drop table if exists discovery;

create table discovery (
  id          text check(id = 1) default 1 primary key, /* can only be one, see trigger below */
  title       text not null,
  description text,
  contact     text,
  default_url text,
  created_at  text,
  updated_at  text
);

  create trigger discovery_ai_created_at after insert on discovery
    begin
      update discovery set created_at = strftime('%Y-%m-%dT%H:%M:%fZ', 'now') where id = new.id;
      update discovery set updated_at = strftime('%Y-%m-%dT%H:%M:%fZ', 'now') where id = new.id;
    end;

  create trigger discovery_au_updated_at after update on discovery
    begin
      update discovery set updated_at = strftime('%Y-%m-%dT%H:%M:%fZ', 'now') where id = new.id;
    end;

  create trigger discovery_bi_count before insert on discovery
    begin
      select
        case
          when (select count(*) from discovery) > 0
            then raise(abort, 'Only one discovery can be defined')
        end;
    end;

drop table if exists objects;

create table objects (
  id            text not null,
  type          text not null,
  created       text not null,
  modified      text not null,
  object        text not null,
  collection_id text not null,
  created_at    text,
  updated_at    text,

  constraint valid_id check(id like '%--________-____-____-____-____________'),
  constraint valid_json check(json_valid(object) = 1),

  primary key (id, modified)
);

  create trigger objects_ai_created_at after insert on objects
    begin
      update objects set created_at = strftime('%Y-%m-%dT%H:%M:%fZ', 'now') where id = new.id;
      update objects set updated_at = strftime('%Y-%m-%dT%H:%M:%fZ', 'now') where id = new.id;
    end;

  create trigger objects_au_updated_at after update on objects
    begin
      update objects set updated_at = strftime('%Y-%m-%dT%H:%M:%fZ', 'now') where id = new.id;
    end;

  create index objects_id on objects (id);
  create index objects_type on objects (type);
  create index objects_version on objects (id, type, modified);

  drop view if exists objects_id_aggregate;

  create view objects_id_aggregate as
    select rowid,
           id,
           type,
           collection_id,
           min(modified) first,
           max(modified) last
    from objects
    group by id,
             type,
             collection_id;

  drop view if exists objects_data;

  create view objects_data as
    select
      so.rowid,
      so.id,
      so.type,
      so.created,
      so.modified,
      so.object,
      so.collection_id,
      case when so.modified = sa.first and so.modified = sa.last then 'only'
           when so.modified = sa.last then 'last'
           when so.modified = sa.first then 'first'
      end version,
      so.created_at,
      so.updated_at
    from
      objects so
      left join objects_id_aggregate sa
        on so.id = sa.id
        and so.collection_id = sa.collection_id;

drop table if exists schema_version;

create table if not exists schema_version (
  id         text check(id = 1) default 1 primary key, /* can only be one, see trigger below */
  version    integer not null,
  created_at text,
  updated_at text
);

  create trigger schema_version_ai_created_at after insert on schema_version
    begin
      update schema_version set created_at = strftime('%Y-%m-%dT%H:%M:%fZ', 'now') where id = new.id;
      update schema_version set updated_at = strftime('%Y-%m-%dT%H:%M:%fZ', 'now') where id = new.id;
    end;

  create trigger schema_version_au_updated_at after update on schema_version
    begin
      update schema_version set updated_at = strftime('%Y-%m-%dT%H:%M:%fZ', 'now') where id = new.id;
    end;

  create trigger schema_version_bi_count before insert on schema_version
    begin
      select
        case
          when (select count(*) from schema_version) > 0
            then raise(abort, 'Only one version can be set')
        end;
    end;

insert into schema_version(version) values(1);

drop table if exists status;

create table status (
  id                text not null,
  status            text not null,
  request_timestamp text,
  total_count       integer not null,
  success_count     integer not null,
  successes         text,
  failure_count     integer not null,
  failures          text,
  pending_count     integer not null,
  pendings          text,
  /* internal */
  created_at    text,
  updated_at    text
);

  create trigger status_ai_created_at after insert on status
    begin
      update status set created_at = strftime('%Y-%m-%dT%H:%M:%fZ', 'now') where id = new.id;
      update status set updated_at = strftime('%Y-%m-%dT%H:%M:%fZ', 'now') where id = new.id;
    end;

  create trigger status_au_updated_at after update on status
    begin
      update status set updated_at = strftime('%Y-%m-%dT%H:%M:%fZ', 'now') where id = new.id;
    end;

  create index status_id on status (id);

drop table if exists user;

create table user (
  email      text not null primary key,
  can_admin  integer check(can_admin in (1, 0)) default 0 not null,
  created_at text,
  updated_at text
);

  create trigger user_ai_created_at after insert on user
    begin
      update user set created_at = strftime('%Y-%m-%dT%H:%M:%fZ', 'now') where email = new.email;
      update user set updated_at = strftime('%Y-%m-%dT%H:%M:%fZ', 'now') where email = new.email;
    end;

  create trigger user_bi_email before insert on user
    begin
      select case when new.email not like '%_@__%.__%' then raise(abort, 'Invalid email address, expecting <username>@<domain>.<tld>') end;
    end;

  create trigger user_au_updated_at after update on user
    begin
      update user set updated_at = strftime('%Y-%m-%dT%H:%M:%fZ', 'now') where email = new.email;
    end;

drop table if exists user_collection;

create table user_collection (
  id            integer primary key not null,
  email         text    not null,
  collection_id text    not null,
  can_read      integer check(can_read in (1, 0)) not null,
  can_write     integer check(can_read in (1, 0)) not null,
  created_at    text,
  updated_at    text,

  unique (email, collection_id) on conflict ignore,
  foreign key (email) references user(email) on delete cascade
);

  create trigger user_collection_ai_created_at after insert on user_collection
    begin
      update user_collection set created_at = strftime('%Y-%m-%dT%H:%M:%fZ', 'now') where email = new.email;
      update user_collection set updated_at = strftime('%Y-%m-%dT%H:%M:%fZ', 'now') where email = new.email;
    end;

  create trigger user_collection_au_updated_at after update on user_collection
    begin
      update user_collection set updated_at = strftime('%Y-%m-%dT%H:%M:%fZ', 'now') where email = new.email;
    end;

drop table if exists user_pass;

create table user_pass (
  id         integer not null primary key,
  email      text not null,
  -- check password is not empty string or sha256 of empty string
  pass       text not null check (
               pass not in ("", "e3b0c44298fc1c149afbf4c8996fb92427ae41e4649b934ca495991b7852b855")
               and length(pass) == 64
             ),
  created_at text,
  updated_at text,

  unique(email) on conflict ignore,
  foreign key (email) references user(email) on delete cascade
);

  create trigger user_pass_ai_created_at after insert on user_pass
    begin
      update user_pass set created_at = strftime('%Y-%m-%dT%H:%M:%fZ', 'now') where email = new.email;
      update user_pass set updated_at = strftime('%Y-%m-%dT%H:%M:%fZ', 'now') where email = new.email;
    end;

  create trigger user_pass_au_updated_at after update on user_pass
    begin
      update user_pass set updated_at = strftime('%Y-%m-%dT%H:%M:%fZ', 'now') where email = new.email;
    end;
