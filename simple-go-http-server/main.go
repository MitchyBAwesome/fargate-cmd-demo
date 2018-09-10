package main

import (
 "log"
 "net/http"

 "github.com/gorilla/mux"
)

func YourHandler(w http.ResponseWriter, r *http.Request) {
 w.Write([]byte("<h1>Hello Mitch Beaumont!</h1>\n"))
}

func main() {
 r := mux.NewRouter()

 r.HandleFunc("/", YourHandler)

 log.Fatal(http.ListenAndServe(":8000", r))
}