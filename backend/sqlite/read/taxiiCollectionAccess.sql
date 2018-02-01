select
  tuc.collection_id,
  tuc.can_read,
  tuc.can_write
from
  taxii_user tu
  inner join taxii_user_collection tuc
    on tu.email = tuc.email
where
  tu.email = ?
