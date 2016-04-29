package main

import (
  "html"
  "os"
  "log"
  "strings"
  "regexp"
  "net/http"
  "gopkg.in/mgo.v2"
  "gopkg.in/mgo.v2/bson"
)

var c * mgo.Collection

type Link struct {
  Id bson.ObjectId `bson:"_id,omitempty"`
  Src string
  Dst string
}

type Body struct {
  Dst string
}

func sendError (w http.ResponseWriter, t int) {
  http.Error(w, http.StatusText(t), t)
}

/**
 * Link retrieve with redirect
 */
func get (w http.ResponseWriter, r * http.Request) {
  var id string = strings.TrimLeft(html.EscapeString(r.URL.Path), "/")
  match, _ := regexp.MatchString("^[A-Za-z0-9]*$", id)
  if (!match) {
    sendError(w, http.StatusBadRequest)
    return
  }

  result := Link{}
  err := c.Find(bson.M{ "src": id }).One(&result)
  if err != nil {
    sendError(w, http.StatusNotFound)
    return
  }
  err = c.Remove(bson.M{ "src": id })
  http.Redirect(w, r, result.Dst, 301)
}

/**
 * Link creation
 */
func post (w http.ResponseWriter, r * http.Request) {
  dst := r.FormValue("dst")
  if (strings.Index(dst, "http://") != 0 && strings.Index(dst, "https://") != 0) {
    sendError(w, http.StatusBadRequest)
    return
  }

  var id bson.ObjectId = bson.NewObjectId()
  var hex string = id.Hex()
  hex = hex[len(hex) - 7:]
  err := c.Insert(&Link{ Src: hex, Dst: dst, Id: id })
  if (err != nil) { panic(err) }
  w.Write([]byte(hex))
}

/**
 * Return the port to listen to according to the PORT env variable, or 3000 by default
 */
func port () string {
  var p string = os.Getenv("PORT")
  if (p != "") { return ":" + p }
  return ":3000"
}

/**
 * Connect to mongo and set global collection variable
 */
func main () {
  mongo, err := mgo.Dial(os.Getenv("MONGO"))
  if (err != nil) { panic(err) }

  defer mongo.Close()
  mongo.SetMode(mgo.Monotonic, true)

  c = mongo.DB("fugitive").C("links")

  http.HandleFunc("/", func (w http.ResponseWriter, r * http.Request) {
    w.Header().Set("Access-Control-Allow-Origin", "*")
    w.Header().Set("Access-Control-Allow-Methods", "GET, POST")

    if (r.Method == "GET") { get(w, r) }
    if (r.Method == "POST") { post(w, r) }
  })

  log.Fatal(http.ListenAndServe(port(), nil))
}
