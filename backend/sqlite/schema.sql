drop table if exists taxii_collection;

create table taxii_collection (
  id          text not null primary key,
  title       text,
  description text,
  media_types text,
  created_at  text,
  updated_at  text
);

  create trigger taxii_collection_ai_created_at after insert on taxii_collection
    begin
      update taxii_collection set created_at = datetime('now') where id = new.id;
      update taxii_collection set updated_at = datetime('now') where id = new.id;
    end;

  create trigger taxii_collection_au_updated_at after update on taxii_collection
    begin
      update taxii_collection set updated_at = datetime('now') where id = new.id;
    end;

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
  can_read      integer not null,
  can_write     integer not null,
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
