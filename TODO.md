Handle migrations with a tool?
  Ref: https://blog.codinghorror.com/get-your-database-under-version-control/
  Tool?: https://github.com/pressly/goose
         https://sqitch.org/

Should we open port 443 in vagrant? https://help.ubuntu.com/community/UFW#UFW_-_Uncomplicated_Firewall

#### Admin section
What about these use cases?  Right now, data is left in place...
  - if i delete a user, should i delete their collection associations?
  - if i delete discovery, do i auto delete roots and collections?
    - my gut says no...just leave the stuff there, there's only ever one discovery so the id is '1'
  - if i delete an api_root, do i auto delete collections?
    - then delete collection association for all users?
  - if i delete a collection, do i auto delete stix objects associated to them?
  - if i delete a collection, do i delete users/collection combos associated with it?
update cli
  - setup admin user on install with randomly generated pw

#### tech debt
- should backend return some generic object with it's data and take a generic input?
  - ServiceRequest{} & ServiceResponse{]?
- sql tables are using a uuid for an id, use 'id integer not null primary key' <- auto incrementing id
- for any error handler (unauthorized, etc.) send the request object and get the context into the log
- should i move api roots and collections back to config file? https://gist.github.com/pladdy/954ea6f01794e51c1c7d8d217e6f9fdd

#### test schema.sql?
schema.sql should have schema_test.sql too...like do inserts and updates to test triggers, sql syntax, etc.
- this should be handled in the migration?  like to verify if the thing exists (probably shouldn't insert/update data...
  but how do you test a schema without data?)

#### Structured logging
Log the right things
  - http://www.miyagijournal.com/articles/five-steps-application-logging/
  - https://www.owasp.org/index.php/Logging_Cheat_Sheet
  - http://arctecgroup.net/pdf/howtoapplogging.pdf
  Who - process and/or user
  What
    - happened – Was the event a failed login or a crash? Use plain English (or your local language) and just write in the log what happened.
    - component is impacted by this event – Was it the main executable, a dll, or a 3rd party tool?
  Where did it happen – Was it a network issue, a local server issue, or a failure to access the application database? This will help you know where to start your troubleshooting efforts.
  When did it happen – If you have your timestamp in the event, this will ready help you understand the sequence of events and how they relate to events outside of your program.
  Why did it happen – If your program knows why you got an error or why some processing step couldn’t complete, write it in the log.
  How did it happen – Logging what button the user clicked or the text of the entry that caused the error might abbreviate the troubleshooting.

Create a config validator/linter/test?

#### Basic Auth + JWT?
  - https://jwt.io/introduction/
  - https://github.com/SermoDigital/jose
