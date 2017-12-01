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

  /* test niner */
  res, err := rethink.Expr("Hello World!").Run(session)
  if err != nil {
    fmt.Println("An error!", err)
  }
  defer res.Close()

  var response string
  err = res.One(&response)
  if err != nil {
    fmt.Println(err)
  }

  fmt.Println(response)

  /* add a user */
  err = rethink.DB("rethinkdb").Table("users").Insert(map[string]string{
   "id": "pladdy",
   "password": "pants",
  }).Exec(session)
  if err != nil {
    fmt.Println("An insert error!", err)
  }

  res, err = rethink.DB("rethinkdb").Table("users").Get("pladdy").Run(session)
  if err != nil {
    fmt.Println(err)
  }
  defer res.Close()

  var row interface{}
  for res.Next(&row) {
    fmt.Println(row)
  }
  if res.Err() != nil {
    fmt.Println("A get error!", err)
  }
}
