select
  1
from
  taxii_user tu
  inner join taxii_user_pass tup
    on tu.email = tup.email
where
  tu.email = ?
  and tup.pass = ?
