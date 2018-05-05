package main

import "net/http"

func ListenAndServe(addr string) {
	http.HandleFunc("/api/v1/user/add", AddUser)
	http.HandleFunc("/api/v1/user/remove", RemoveUser)

	http.ListenAndServe(addr, nil)
}

func AddUser(w http.ResponseWriter, r *http.Request) {

}

func RemoveUser(w http.ResponseWriter, r *http.Request) {

}
