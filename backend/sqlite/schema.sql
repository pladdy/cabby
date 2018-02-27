/* stix */

drop table if exists stix_object;

create table stix_object (
  stix_id       text not null,
  stix_type     text not null,
  stix_modified text not null,
  collection_id text not null,
  stix_object   text not null,
  created_at    text,
  updated_at    text
);

  create trigger stix_object_ai_created_at after insert on stix_object
    begin
      update stix_object set created_at = datetime('now') where stix_id = new.stix_id;
      update stix_object set updated_at = datetime('now') where stix_id = new.stix_id;
    end;

  create trigger stix_object_au_updated_at after update on stix_object
    begin
      update stix_object set updated_at = datetime('now') where stix_id = new.stix_id;
    end;

  create index stix_object_stix_id on stix_object (stix_id);
  create index stix_object_stix_type on stix_object (stix_type);
  create index stix_object_stix_version on stix_object (stix_id, stix_modified);

/* taxii */

drop table if exists taxii_api_root;

create table taxii_api_root (
  api_root_id        text    not null primary key,
  discovery_id       integer check(discovery_id = 1) default 1,
  api_root_path      text    not null,
  title              text    not null,
  description        text,
  versions           text,
  max_content_length integer not null,
  created_at         text,
  updated_at         text
);

  create trigger taxii_api_root_ai_created_at after insert on taxii_api_root
    begin
      update taxii_api_root set created_at = datetime('now') where api_root_id = new.api_root_id;
      update taxii_api_root set updated_at = datetime('now') where api_root_id = new.api_root_id;
    end;

  create trigger taxii_api_root_au_updated_at after update on taxii_api_root
    begin
      update taxii_api_root set updated_at = datetime('now') where api_root_id = new.api_root_id;
    end;

drop table if exists taxii_collection;

create table taxii_collection (
  collection_id text not null primary key,
  title        text,
  description  text,
  media_types  text,
  created_at   text,
  updated_at   text
);

  create trigger taxii_collection_ai_created_at after insert on taxii_collection
    begin
      update taxii_collection set created_at = datetime('now') where collection_id = new.collection_id;
      update taxii_collection set updated_at = datetime('now') where collection_id = new.collection_id;
    end;

  create trigger taxii_collection_au_updated_at after update on taxii_collection
    begin
      update taxii_collection set updated_at = datetime('now') where collection_id = new.collection_id;
    end;

drop table if exists taxii_collection_api_root;

create table taxii_collection_api_root (
  collection_id text not null,
  api_root_id   text not null,
  created_at    text,
  updated_at    text,

  primary key (collection_id, api_root_id)
);

  create trigger taxii_collection_api_root_ai_created_at after insert on taxii_collection_api_root
    begin
      update taxii_collection_api_root set created_at = datetime('now')
        where collection_id = new.collection_id and api_root_id = new.api_root_id;
      update taxii_collection_api_root set updated_at = datetime('now')
        where collection_id = new.collection_id and api_root_id = new.api_root_id;
    end;

  create trigger taxii_collection_api_root_au_updated_at after update on taxii_collection_api_root
    begin
      update taxii_collection_api_root set updated_at = datetime('now')
        where collection_id = new.collection_id and api_root_id = new.api_root_id;
    end;

drop table if exists taxii_discovery;

create table taxii_discovery (
  discovery_id text check(discovery_id = 1) default 1 primary key, /* can only be one, see trigger below */
  title        text not null,
  description  text,
  contact      text,
  default_url  text,
  created_at   text,
  updated_at   text
);

  create trigger taxii_discovery_ai_created_at after insert on taxii_discovery
    begin
      update taxii_discovery set created_at = datetime('now') where discovery_id = new.discovery_id;
      update taxii_discovery set updated_at = datetime('now') where discovery_id = new.discovery_id;
    end;

  create trigger taxii_discovery_au_updated_at after update on taxii_discovery
    begin
      update taxii_discovery set updated_at = datetime('now') where discovery_id = new.discovery_id;
    end;

  create trigger taxii_discovery_bi_count before insert on taxii_discovery
    begin
      select
        case
          when (select count(*) from taxii_discovery) > 0
            then raise(abort, 'Only one discovery can be defined')
        end;
    end;

drop table if exists taxii_status;

create table taxii_status (
  status_id         text not null,
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

  create trigger taxii_status_ai_created_at after insert on taxii_status
    begin
      update taxii_status set created_at = datetime('now') where status_id = new.status_id;
      update taxii_status set updated_at = datetime('now') where status_id = new.status_id;
    end;

  create trigger taxii_status_au_updated_at after update on taxii_status
    begin
      update taxii_status set updated_at = datetime('now') where status_id = new.status_id;
    end;

  create index taxii_status_taxii_id on taxii_status (status_id);

drop table if exists taxii_user;

create table taxii_user (
  email      text not null primary key,
  created_at text,
  updated_at text
);

  create trigger taxii_user_ai_created_at after insert on taxii_user
    begin
      update taxii_user set created_at = datetime('now') where email = new.email;
      update taxii_user set updated_at = datetime('now') where email = new.email;
    end;

  create trigger taxii_user_bi_email before insert on taxii_user
    begin
      select case when new.email not like '%_@__%.__%' then raise(abort, 'Invalid email address') end;
    end;

  create trigger taxii_user_au_updated_at after update on taxii_user
    begin
      update taxii_user set updated_at = datetime('now') where email = new.email;
    end;

drop table if exists taxii_user_collection;

create table taxii_user_collection (
  email         text    not null,
  collection_id text    not null,
  can_read      integer check(can_read in (1, 0)) not null,
  can_write     integer check(can_read in (1, 0)) not null,
  created_at    text,
  updated_at    text,

  primary key (email, collection_id)
);

  create trigger taxii_user_collection_ai_created_at after insert on taxii_user_collection
    begin
      update taxii_user_collection set created_at = datetime('now') where email = new.email;
      update taxii_user_collection set updated_at = datetime('now') where email = new.email;
    end;

  create trigger taxii_user_collection_au_updated_at after update on taxii_user_collection
    begin
      update taxii_user_collection set updated_at = datetime('now') where email = new.email;
    end;

drop table if exists taxii_user_pass;

create table taxii_user_pass (
  email      text not null primary key,
  pass       text not null check (pass != ""),
  created_at text,
  updated_at text
);

  create trigger taxii_user_pass_ai_created_at after insert on taxii_user_pass
    begin
      update taxii_user_pass set created_at = datetime('now') where email = new.email;
      update taxii_user_pass set updated_at = datetime('now') where email = new.email;
    end;

  create trigger taxii_user_pass_au_updated_at after update on taxii_user_pass
    begin
      update taxii_user_pass set updated_at = datetime('now') where email = new.email;
    end;
