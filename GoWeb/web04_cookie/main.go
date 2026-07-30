package main

import "net/http"

func setCookie(w http.ResponseWriter, r *http.Request) {
	cookie := http.Cookie{
		Name:     "user",
		Value:    "admin",
		HttpOnly: true,
	}

	cookie2 := http.Cookie{
		Name:     "user",
		Value:    "admin2",
		HttpOnly: true,
	}
	w.Header().Set("Set-Cookie", cookie.String())
	w.Header().Add("Set-Cookie", cookie2.String())
}

func main() {
	http.HandleFunc("/setCookie", setCookie)
	http.ListenAndServe(":8080", nil)
}
