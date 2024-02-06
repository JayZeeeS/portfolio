package main

import (
	//"encoding/json"
	"fmt"
	"html/template"
	"net/http"
)

type Account struct {
	Name       string
	Percentage float64
	Amount     float64
}

type User struct {
	Name     string
	isActive bool
	Accounts []Account
}

func createUser(name string) *User {
	return &User{Name: name, isActive: true}
}

var users []User

//var accounts User

func main() {
	/*accounts = User{
	        Name: "Fulano, el hijo de Mengano",
			Accounts: []Account{{Name: "Savings", Percentage: 15, Amount: 12343832},
				{Name: "Investments", Percentage: 35, Amount: 133399999920},
				{Name: "Happy Hour", Percentage: 50, Amount: 10}},
		}*/
	handleRequests()
}

func handleRequests() {
	http.Handle("/styles/", http.StripPrefix("/styles/", http.FileServer(http.Dir("styles"))))
	http.Handle("/scripts/", http.StripPrefix("/scripts/", http.FileServer(http.Dir("scripts"))))
	http.HandleFunc("/", home)
	http.HandleFunc("/getAccounts", getUserData)
	http.HandleFunc("/createUser", create)

	http.ListenAndServe(":8080", nil)
}

func home(w http.ResponseWriter, r *http.Request) {
	tmpl, err := template.ParseFiles("./layouts/layout.html",
		"layouts/accounts.html",
		"layouts/form.html",
	)
	if err != nil {
		w.WriteHeader(500)
		return
	}
	tmpl.Execute(w, users)
}

func create(w http.ResponseWriter, r *http.Request) {
	r.ParseForm()
	name := r.FormValue("Name")
	users = append(users, *createUser(name))
    http.Redirect(w, r, "/getAccounts", 300)
}

func getUserData(w http.ResponseWriter, r *http.Request) {
	/*	enc := json.NewEncoder(w)
		enc.SetIndent(" ", " ")
		enc.Encode(Accounts)
	*/
	var user User
	r.ParseForm()
	usrName := r.FormValue("name")
	for _, v := range users {
		if usrName == v.Name {
			user = v
			break
		}
	}
	if user.Name == "" {
		user = *createUser(usrName)
	}
	tmpl, err := template.ParseFiles("./layouts/layout.html",
		"layouts/accounts.html",
		"layouts/form.html",
	)
	if err != nil {
		w.WriteHeader(500)
		return
	}
	tmpl.Execute(w, users)
	fmt.Println(user)
}
