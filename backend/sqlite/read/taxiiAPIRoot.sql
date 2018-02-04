select
  title,
  description,
  versions,
  max_content_length
from
  taxii_api_root
where
  api_root_path = ?
