package main

import (
  "fmt"
  rethink "gopkg.in/gorethink/gorethink.v3"
)

func main() {
  /* connect */
  session, err := rethink.Connect(rethink.ConnectOpts{
    Address: "localhost",
  })
  if err != nil {
    fmt.Println("An error!", err)
  }
  defer session.Close()

  /* add a user */
  wr, err := rethink.DB("rethinkdb").Table("users").Insert(map[string]string{
   "id": "pladdy",
   "password": "pants",
  }).RunWrite(session)
  if err != nil {
    fmt.Println("An insert error!", err)
  }

  fmt.Printf("Insert response: %#v\n", wr)
  fmt.Println(wr.Errors)

  /* check a user */
  res, err := rethink.DB("rethinkdb").Table("users").Get("pladdy").Run(session)
  if err != nil {
    fmt.Println(err)
  }
  defer res.Close()

  var user map[string]interface{}

  for res.Next(&user) {
    fmt.Printf("Get response: %#v\n", user)
  }
  if res.Err() != nil {
    fmt.Println("A get error!", err)
  }
}
