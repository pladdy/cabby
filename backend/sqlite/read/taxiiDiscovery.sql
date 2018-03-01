select
  td.title,
  td.description,
  td.contact,
  td.default_url,
  case
    when tar.api_root_path is null then 'No API Roots defined' else tar.api_root_path
  end api_root_path
from
  taxii_discovery td
  left join taxii_api_root tar
    on td.discovery_id = tar.discovery_id
