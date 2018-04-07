select
  id,
  min(created) date_added,
  group_concat(modified) versions
  -- media_types omitted for now
from
  stix_objects
where
  collection_id = ?
group by
  id
