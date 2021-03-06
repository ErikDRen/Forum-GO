package main

import (
	"database/sql"
	"fmt"
	"html/template"
	"log"
	"net/http"
	"regexp"

	"github.com/gorilla/sessions"
	_ "github.com/mattn/go-sqlite3"
	"golang.org/x/crypto/bcrypt"
)

type Login struct {
	Username string
	Password string
}

type Err struct {
	Nope         string
	WrongPass    string
	UserNotFound string
}

type Post struct {
	Comment string
	Tag     string
}

type PrintPost struct {
	Username string
	Id       int
	Comment  string
	Tag      string
}

type Comments struct {
	id       string
	username string
	text     string
}

var allData []PrintPost
var commentData []Comments
var errToSend Err
var store = sessions.NewCookieStore([]byte("mysession"))
var dataLogin Login
var tag Post
var database, _ = sql.Open("sqlite3", "data.db")

func createTablePeople() { // create table for account in database
	statement, _ :=
		database.Prepare("CREATE TABLE IF NOT EXISTS people (id INTEGER PRIMARY KEY, username TEXT, email TEXT, password TEXT)")
	defer statement.Close()
	statement.Exec()
}

func createTablePost() { // create table for post in database
	statement, _ :=
		database.Prepare("CREATE TABLE IF NOT EXISTS post (id INTEGER PRIMARY KEY, username TEXT, comment TEXT, tag TEXT)")
	defer statement.Close()
	statement.Exec()
}

func createTableComment() { // create table for comment in database
	statement, _ :=
		database.Prepare("CREATE TABLE IF NOT EXISTS comments (id TEXT , username TEXT, newComment TEXT)")
	defer statement.Close()
	statement.Exec()
}

func addUser(username string, email string, password string) { // add user information in database

	statement, _ :=
		database.Prepare("INSERT INTO people (username, email, password) VALUES (?, ?, ?)")
	// defer statement.Close()
	statement.Exec(username, email, password)

	lastRow := database.QueryRow("SELECT * FROM user WHERE id=(SELECT max(id) FROM user)")
	var id int
	lastRow.Scan(&id, &email, &username, &password)
	fmt.Println("NEW USER : ", username)

}

func addPost(username string) { // add post information in database
	fmt.Println(username)
	fmt.Println(tag.Comment)

	statement, _ :=
		database.Prepare("INSERT INTO post (username, comment, tag) VALUES (?, ?, ?)")
	statement.Exec(username, tag.Comment, tag.Tag)

}

func addComment(id string, username string, newComment string) { // add comment information in database

	statement, _ :=
		database.Prepare("INSERT INTO comments (id, username, newComment) VALUES (?, ?, ?)")
	defer statement.Close()
	statement.Exec(id, username, newComment)
}

func HashPassword(password string) (string, error) {
	bytes, err := bcrypt.GenerateFromPassword([]byte(password), 14)
	return string(bytes), err
}

func CheckPasswordHash(password, hash string) bool {
	err := bcrypt.CompareHashAndPassword([]byte(hash), []byte(password))
	return err == nil
}

func verifRegisterData(regUser string, regEmail string) bool { // check if user information is already use or not
	var username string
	var email string
	var password string
	var id int
	var result = true

	rows, _ :=
		database.Query("SELECT id, username, email, password FROM people")
	for rows.Next() {
		rows.Scan(&id, &username, &email, &password)
		if regEmail == email || regUser == username {
			fmt.Println("nope")
			result = false
		}
	}
	return result
}

func login(logUser string, logPassword string) int { // check if user can connect to his account or not
	var username string
	var email string
	var password string
	var id int
	var result = 0

	rows, _ :=
		database.Query("SELECT id, username, email, password FROM people")
	for rows.Next() {
		rows.Scan(&id, &username, &email, &password)
		// fmt.Println(strconv.Itoa(id) + ": " + username + " " + email + " " + password)
		if logUser == username && CheckPasswordHash(logPassword, password) {
			result = 1
		} else if logUser == username && !(CheckPasswordHash(logPassword, password)) {
			result = 2
		}
	}
	return result
}

func logoutHandle(w http.ResponseWriter, r *http.Request) { // for the user to disconnect to his account
	session, _ := store.Get(r, "mysession")
	session.Options.MaxAge = -1
	session.Save(r, w)
	http.Redirect(w, r, "/login", http.StatusSeeOther)
}

func LoginHandle(w http.ResponseWriter, r *http.Request) { // for the user to connect to his account
	fmt.Println("home")
	dataLogin.Username = r.FormValue("username")
	dataLogin.Password = r.FormValue("password")
	errToSend.WrongPass = ""
	errToSend.UserNotFound = ""
	tpl := template.Must(template.ParseFiles("assets/login.html"))

	if dataLogin.Password != "" && dataLogin.Username != "" {
		if login(dataLogin.Username, dataLogin.Password) == 1 {
			fmt.Println("correct password")
			session, _ := store.Get(r, "mysession")
			session.Values["username"] = dataLogin.Username
			session.Save(r, w)
			http.Redirect(w, r, "/account", http.StatusSeeOther)
		} else if login(dataLogin.Username, dataLogin.Password) == 2 {
			fmt.Println("Wrong password")
			errToSend.WrongPass = "Wrong password"
			tpl.Execute(w, errToSend)
			return
		} else {
			fmt.Println("user doesn't exist")
			errToSend.UserNotFound = "user doesn't exist"
			tpl.Execute(w, errToSend)
			return
		}

	}

	err := tpl.Execute(w, nil)
	if err != nil {
		log.Fatal(err)
	}
}

func getAllData() []PrintPost { // get post information from database
	var temp PrintPost

	rows, _ :=
		database.Query("SELECT id, username, comment, tag FROM post")
	allData = nil
	for rows.Next() {
		rows.Scan(&temp.Id, &temp.Username, &temp.Comment, &temp.Tag)
		allData = append(allData, temp)
	}
	return allData
}

func getAllDataComment() []Comments { // get comment information from database
	var temp Comments

	rows, _ :=
		database.Query("SELECT id, username,  newComment FROM comments")
	commentData = nil
	for rows.Next() {
		rows.Scan(&temp.id, &temp.username, &temp.text)
		commentData = append(commentData, temp)
	}
	return commentData
}

func RegisterHandle(w http.ResponseWriter, r *http.Request) { // for user to register
	fmt.Println("Register")

	var regData Err
	userName := r.FormValue("username")
	email := r.FormValue("email")
	password, _ := HashPassword(r.FormValue("password"))
	tpl := template.Must(template.ParseFiles("assets/register.html"))
	regData.Nope = ""
	// fmt.Println(userName)
	// fmt.Println(email)
	// fmt.Println(password)
	regexp, _ := regexp.Match(`[0-9]`, []byte(userName))
	fmt.Println(regexp)
	if regexp {
		regData.Nope = "your name must contain single characters (A-Z a-z)"
		tpl.Execute(w, regData)
		return
	}
	if userName != "" && email != "" && password != "" {
		if verifRegisterData(userName, email) {
			addUser(userName, email, password)
			http.Redirect(w, r, "/login", http.StatusSeeOther)
		} else {
			regData.Nope = "your email or your username are already used !"
		}

	}

	err := tpl.Execute(w, regData)
	if err != nil {
		log.Fatal(err)
	}
}

func PostHandle(w http.ResponseWriter, r *http.Request) { // for user to see all post (the main page)
	fmt.Println("Connect")

	session, _ := store.Get(r, "mysession")
	userName := fmt.Sprintf("%v", session.Values["username"])

	data := map[string]interface{}{
		"username": &userName,
		"post":     getAllData(),
	}
	// fmt.Print(getAllData())

	tpl, _ := template.ParseFiles("assets/signIn.html")
	tpl.Execute(w, data)
	// if err != nil {
	// 	log.Fatal(err)
	// }
}

func GuestHandle(w http.ResponseWriter, r *http.Request) { // for guest(without account) to see all post (the main page)

	data := map[string]interface{}{
		"post": getAllData(),
	}
	// fmt.Print(getAllData())

	tpl, _ := template.ParseFiles("assets/guest.html")
	tpl.Execute(w, data)
	// if err != nil {
	// 	log.Fatal(err)
	// }
}

func CreatePostHandle(w http.ResponseWriter, r *http.Request) { // for user to add a post
	fmt.Println("POST")

	session, _ := store.Get(r, "mysession")
	username := fmt.Sprintf("%v", session.Values["username"])

	tag.Comment = r.FormValue("comment")
	tag.Tag = r.FormValue("Tag")

	if tag.Comment != "" {
		fmt.Printf("psot en cour\n")
		addPost(username)
		http.Redirect(w, r, "/account", http.StatusSeeOther)
	}

	tpl := template.Must(template.ParseFiles("assets/post.html"))
	tpl.Execute(w, nil)
}

func CommentHandle(w http.ResponseWriter, r *http.Request) { // for user to see and add a comment
	getAllDataComment()
	var comments Comments
	session, _ := store.Get(r, "mysession")
	comments.id = r.FormValue("id")
	comments.username = fmt.Sprintf("%v", session.Values["username"])
	comments.text = r.FormValue("comment")

	if comments.text != "" {
		addComment(comments.id, comments.username, comments.text)
	}

	data := map[string]interface{}{
		"comments": getAllDataComment(),
		"post":     getAllData(),
	}

	tpl := template.Must(template.ParseFiles("assets/comment.html"))
	tpl.Execute(w, data)
}

func GuestCommentHandle(w http.ResponseWriter, r *http.Request) { // for guest(without account) to see all comment

	data := map[string]interface{}{
		"comments": getAllDataComment(),
		"post":     getAllData(),
	}

	tpl := template.Must(template.ParseFiles("assets/guestComment.html"))
	tpl.Execute(w, data)
}

func main() {
	createTablePeople()
	createTablePost()
	createTableComment()
	fs := http.FileServer(http.Dir("assets"))
	http.Handle("/assets/", http.StripPrefix("/assets/", fs))
	http.Handle("/", fs)
	http.HandleFunc("/login", LoginHandle)
	http.HandleFunc("/logout", logoutHandle)
	http.HandleFunc("/register", RegisterHandle)
	http.HandleFunc("/account", PostHandle)
	http.HandleFunc("/CreateNewPost", CreatePostHandle)
	http.HandleFunc("/comment", CommentHandle)
	http.HandleFunc("/Guest", GuestHandle)
	http.HandleFunc("/GuestComment", GuestCommentHandle)
	http.ListenAndServe(":8080", nil)
}
