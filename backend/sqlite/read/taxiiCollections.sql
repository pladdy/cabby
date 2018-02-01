select
  c.id,
  c.title,
  c.description,
  uc.can_read,
  uc.can_write,
  c.media_types
from
  taxii_collection c
  inner join taxii_user_collection uc
    on c.id = uc.collection_id
where
  uc.email = ?
  and uc.can_read = 1 or uc.can_write = 1
