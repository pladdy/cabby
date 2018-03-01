select
  c.collection_id,
  c.title,
  c.description,
  uc.can_read,
  uc.can_write,
  c.media_types
from
  taxii_collection c
  inner join taxii_user_collection uc
    on c.collection_id = uc.collection_id
where
  uc.email = ?
  and c.collection_id = ?
  and uc.can_read = 1
