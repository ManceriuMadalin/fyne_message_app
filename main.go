package main

import (
	"database/sql"
	"fmt"
	"log"
	"strings"

	"fyne.io/fyne/v2"
	"fyne.io/fyne/v2/app"
	"fyne.io/fyne/v2/container"
	"fyne.io/fyne/v2/widget"
	_ "github.com/go-sql-driver/mysql"
)

var db *sql.DB
var currentUserID int
var currentUsername string
var currentWindow fyne.Window

var friendsListContainer fyne.CanvasObject
var requestsListContainer fyne.CanvasObject

type Friend struct {
	ID       int
	Username string
	FullName string
}

type FriendRequest struct {
	ID       int
	Username string
	FullName string
}

func main() {
	var err error
	db, err = sql.Open("mysql", "root:password@tcp(127.0.0.1:3306)/user_auth")
	if err != nil {
		log.Fatal("Eroare la deschiderea conexiunii:", err)
	}
	err = db.Ping()
	if err != nil {
		log.Fatal("Nu mă pot conecta la baza de date:", err)
	}

	defer db.Close()

	a := app.New()
	w := a.NewWindow("Autentificare")
	currentWindow = w

	showMainPage(w)

	w.Resize(fyne.NewSize(500, 600))
	w.ShowAndRun()
}

func showMainPage(w fyne.Window) {
	loginBtn := widget.NewButton("Log In", func() {
		showLoginPage(w)
	})

	signUpBtn := widget.NewButton("Sign Up", func() {
		showSignUpPage(w)
	})

	w.SetContent(container.NewVBox(
		widget.NewLabel("Bine ai venit!"),
		loginBtn,
		signUpBtn,
	))
}

func showLoginPage(w fyne.Window) {
	username := widget.NewEntry()
	username.SetPlaceHolder("Username")

	password := widget.NewPasswordEntry()
	password.SetPlaceHolder("Parola")

	msg := widget.NewLabel("")

	loginBtn := widget.NewButton("Loghează-te", func() {
		var userID int
		var dbUsername string
		err := db.QueryRow("SELECT id, username FROM users WHERE username=? AND password=?",
			username.Text, password.Text).Scan(&userID, &dbUsername)

		if err != nil {
			if err == sql.ErrNoRows {
				msg.SetText("Date incorecte!")
				return
			}
			log.Println("Eroare la query:", err)
			msg.SetText("Eroare la interogare.")
			return
		}

		currentUserID = userID
		currentUsername = dbUsername
		msg.SetText("Autentificare reușită!")
		log.Println("Autentificare reușită pentru username:", username.Text)

		showFriendsPage(w)
	})

	backBtn := widget.NewButton("Înapoi", func() {
		showMainPage(w)
	})

	w.SetContent(container.NewVBox(
		widget.NewLabel("Log In"),
		username,
		password,
		loginBtn,
		msg,
		backBtn,
	))
}

func showSignUpPage(w fyne.Window) {
	fullName := widget.NewEntry()
	fullName.SetPlaceHolder("Nume complet")

	username := widget.NewEntry()
	username.SetPlaceHolder("Username")

	password := widget.NewPasswordEntry()
	password.SetPlaceHolder("Parolă")

	msg := widget.NewLabel("")

	registerBtn := widget.NewButton("Creează cont", func() {
		result, err := db.Exec("INSERT INTO users (full_name, username, password) VALUES (?, ?, ?)",
			fullName.Text, username.Text, password.Text)
		if err != nil {
			msg.SetText("Eroare: username deja folosit?")
			log.Println(err)
			return
		}

		userID, _ := result.LastInsertId()
		currentUserID = int(userID)
		currentUsername = username.Text

		msg.SetText("Cont creat cu succes!")

		showFriendsPage(w)
	})

	backBtn := widget.NewButton("Înapoi", func() {
		showMainPage(w)
	})

	w.SetContent(container.NewVBox(
		widget.NewLabel("Sign Up"),
		fullName,
		username,
		password,
		registerBtn,
		msg,
		backBtn,
	))
}

func showFriendsPage(w fyne.Window) {
	title := widget.NewLabel(fmt.Sprintf("Bun venit, %s!", currentUsername))
	title.TextStyle = fyne.TextStyle{Bold: true}

	searchEntry := widget.NewEntry()
	searchEntry.SetPlaceHolder("Caută după username...")

	searchResults := container.NewVBox()

	searchBtn := widget.NewButton("Caută", func() {
		if strings.TrimSpace(searchEntry.Text) == "" {
			searchResults.RemoveAll()
			return
		}
		updateSearchResults(searchResults, searchEntry.Text)
	})

	searchContainer := container.NewBorder(nil, nil, nil, searchBtn, searchEntry)

	friendsList := container.NewVBox()
	friendsListContainer = friendsList
	updateFriendsList()

	requestsList := container.NewVBox()
	requestsListContainer = requestsList
	updateFriendRequests()

	logoutBtn := widget.NewButton("Logout", func() {
		currentUserID = 0
		currentUsername = ""
		friendsListContainer = nil
		requestsListContainer = nil
		showMainPage(w)
	})

	friendsScroll := container.NewScroll(friendsList)
	friendsScroll.SetMinSize(fyne.NewSize(450, 150))

	requestsScroll := container.NewScroll(requestsList)
	requestsScroll.SetMinSize(fyne.NewSize(450, 100))

	searchScroll := container.NewScroll(searchResults)
	searchScroll.SetMinSize(fyne.NewSize(450, 100))

	content := container.NewVBox(
		title,
		widget.NewSeparator(),

		widget.NewLabel("Caută utilizatori:"),
		searchContainer,
		searchScroll,

		widget.NewSeparator(),
		widget.NewLabel("Cererile tale de prietenie:"),
		requestsScroll,

		widget.NewSeparator(),
		widget.NewLabel("Prietenii tăi:"),
		friendsScroll,

		widget.NewSeparator(),
		logoutBtn,
	)

	w.SetContent(content)
}

func messagesPage(w fyne.Window, friend string, friendID int) {
	title := widget.NewLabel(fmt.Sprintf("Conversație cu %s", friend))
	title.TextStyle = fyne.TextStyle{Bold: true}

	messagesContainer := container.NewVBox()
	messagesScroll := container.NewScroll(messagesContainer)
	messagesScroll.SetMinSize(fyne.NewSize(450, 500))

	messageEntry := widget.NewEntry()
	messageEntry.SetPlaceHolder("Scrie un mesaj...")

	sendBtn := widget.NewButton("Trimite", func() {
		if strings.TrimSpace(messageEntry.Text) == "" {
			return
		}
		sendMessage(friendID, messageEntry.Text, messagesContainer)
		messageEntry.SetText("")
	})

	inputContainer := container.NewBorder(nil, nil, nil, sendBtn, messageEntry)

	backBtn := widget.NewButton("Înapoi", func() {
		showFriendsPage(w)
	})

	loadMessages(currentUserID, friendID, messagesContainer)

	content := container.NewVBox(
		title,
		widget.NewSeparator(),
		messagesScroll,
		widget.NewSeparator(),
		inputContainer,
		widget.NewSeparator(),
		backBtn,
	)

	w.SetContent(content)
}

func loadMessages(userID1, userID2 int, container *fyne.Container) {
	container.RemoveAll()

	rows, err := db.Query(`
		SELECT sender_id, message, created_at 
		FROM messages 
		WHERE (sender_id = ? AND receiver_id = ?) OR (sender_id = ? AND receiver_id = ?)
		ORDER BY created_at ASC
	`, userID1, userID2, userID2, userID1)

	if err != nil {
		log.Println("Eroare la încărcarea mesajelor:", err)
		container.Add(widget.NewLabel("Eroare la încărcarea mesajelor"))
		return
	}
	defer rows.Close()

	for rows.Next() {
		var senderID int
		var message string
		var createdAt string

		err := rows.Scan(&senderID, &message, &createdAt)
		if err != nil {
			log.Println("Eroare la scanarea mesajului:", err)
			continue
		}

		var senderName string
		if senderID == currentUserID {
			senderName = "Tu"
		} else {
			senderName = getUsernameByID(senderID)
		}

		messageWidget := widget.NewRichTextFromMarkdown(fmt.Sprintf("**%s**: %s", senderName, message))
		messageWidget.Wrapping = fyne.TextWrapWord

		container.Add(messageWidget)
	}

	if len(container.Objects) == 0 {
		container.Add(widget.NewLabel("Nu există mesaje încă. Începe conversația!"))
	}
}

func sendMessage(receiverID int, message string, container *fyne.Container) {
	_, err := db.Exec(`
		INSERT INTO messages (sender_id, receiver_id, message, created_at) 
		VALUES (?, ?, ?, NOW())
	`, currentUserID, receiverID, message)

	if err != nil {
		log.Println("Eroare la trimiterea mesajului:", err)
		return
	}

	loadMessages(currentUserID, receiverID, container)
}

func getUsernameByID(userID int) string {
	var username string
	err := db.QueryRow("SELECT username FROM users WHERE id = ?", userID).Scan(&username)
	if err != nil {
		log.Println("Eroare la obținerea username-ului:", err)
		return "Utilizator necunoscut"
	}
	return username
}

func updateFriendsList() {
	if friendsListContainer == nil {
		return
	}

	if vbox, ok := friendsListContainer.(interface{ RemoveAll() }); ok {
		vbox.RemoveAll()
	}

	rows, err := db.Query(`
		SELECT u.id, u.username, u.full_name
		FROM users u
		INNER JOIN friendships f ON u.id = f.friend_id
		WHERE f.user_id = ? AND f.status = 'accepted'
		ORDER BY u.username
	`, currentUserID)

	if err != nil {
		log.Println("Eroare la încărcarea prietenilor:", err)
		if vbox, ok := friendsListContainer.(interface{ Add(fyne.CanvasObject) }); ok {
			vbox.Add(widget.NewLabel("Eroare la încărcarea prietenilor"))
		}
		return
	}
	defer rows.Close()

	friends := []Friend{}
	for rows.Next() {
		var friend Friend
		err := rows.Scan(&friend.ID, &friend.Username, &friend.FullName)
		if err != nil {
			log.Println("Eroare la scanarea prietenului:", err)
			continue
		}
		friends = append(friends, friend)
	}

	if len(friends) == 0 {
		if vbox, ok := friendsListContainer.(interface{ Add(fyne.CanvasObject) }); ok {
			vbox.Add(widget.NewLabel("Momentan nu ai prieteni"))
		}
		return
	}

	for _, friend := range friends {
		friendCard := container.NewHBox(
			widget.NewLabel(fmt.Sprintf("%s (%s)", friend.Username, friend.FullName)),
			widget.NewButton("Șterge", func(friendID int) func() {
				return func() {
					removeFriend(friendID)
				}
			}(friend.ID)),
			widget.NewButton("Mesaje", func(friendID int, friendUsername string) func() {
				return func() {
					messagesPage(currentWindow, friendUsername, friendID)
				}
			}(friend.ID, friend.Username)),
		)
		if vbox, ok := friendsListContainer.(interface{ Add(fyne.CanvasObject) }); ok {
			vbox.Add(friendCard)
		}
	}
}

func updateFriendRequests() {
	if requestsListContainer == nil {
		return
	}

	if vbox, ok := requestsListContainer.(interface{ RemoveAll() }); ok {
		vbox.RemoveAll()
	}

	rows, err := db.Query(`
		SELECT u.id, u.username, u.full_name
		FROM users u
		INNER JOIN friendships f ON u.id = f.user_id
		WHERE f.friend_id = ? AND f.status = 'pending'
		ORDER BY u.username
	`, currentUserID)

	if err != nil {
		log.Println("Eroare la încărcarea cererilor:", err)
		return
	}
	defer rows.Close()

	requests := []FriendRequest{}
	for rows.Next() {
		var request FriendRequest
		err := rows.Scan(&request.ID, &request.Username, &request.FullName)
		if err != nil {
			log.Println("Eroare la scanarea cererii:", err)
			continue
		}
		requests = append(requests, request)
	}

	if len(requests) == 0 {
		if vbox, ok := requestsListContainer.(interface{ Add(fyne.CanvasObject) }); ok {
			vbox.Add(widget.NewLabel("Nu ai cereri de prietenie"))
		}
		return
	}

	for _, request := range requests {
		requestCard := container.NewHBox(
			widget.NewLabel(fmt.Sprintf("%s (%s)", request.Username, request.FullName)),
			widget.NewButton("✓", func(reqID int) func() {
				return func() {
					acceptFriendRequest(reqID)
				}
			}(request.ID)),
			widget.NewButton("✗", func(reqID int) func() {
				return func() {
					rejectFriendRequest(reqID)
				}
			}(request.ID)),
		)
		if vbox, ok := requestsListContainer.(interface{ Add(fyne.CanvasObject) }); ok {
			vbox.Add(requestCard)
		}
	}
}

func updateSearchResults(searchContainer fyne.CanvasObject, searchTerm string) {
	if vbox, ok := searchContainer.(interface{ RemoveAll() }); ok {
		vbox.RemoveAll()
	}

	rows, err := db.Query(`
		SELECT u.id, u.username, u.full_name
		FROM users u
		WHERE u.username LIKE ? AND u.id != ?
		AND u.id NOT IN (
			SELECT friend_id FROM friendships WHERE user_id = ?
		)
		AND u.id NOT IN (
			SELECT user_id FROM friendships WHERE friend_id = ?
		)
		LIMIT 10
	`, "%"+searchTerm+"%", currentUserID, currentUserID, currentUserID)

	if err != nil {
		log.Println("Eroare la căutare:", err)
		return
	}
	defer rows.Close()

	found := false
	for rows.Next() {
		found = true
		var userID int
		var username, fullName string
		err := rows.Scan(&userID, &username, &fullName)
		if err != nil {
			continue
		}

		userCard := container.NewHBox(
			widget.NewLabel(fmt.Sprintf("%s (%s)", username, fullName)),
			widget.NewButton("Adaugă prieten", func(uid int) func() {
				return func() {
					sendFriendRequest(uid, searchContainer, searchTerm)
				}
			}(userID)),
		)
		if vbox, ok := searchContainer.(interface{ Add(fyne.CanvasObject) }); ok {
			vbox.Add(userCard)
		}
	}

	if !found {
		if vbox, ok := searchContainer.(interface{ Add(fyne.CanvasObject) }); ok {
			vbox.Add(widget.NewLabel("Nu s-au găsit utilizatori"))
		}
	}
}

func sendFriendRequest(friendID int, searchContainer fyne.CanvasObject, searchTerm string) {
	_, err := db.Exec("INSERT INTO friendships (user_id, friend_id, status) VALUES (?, ?, 'pending')",
		currentUserID, friendID)
	if err != nil {
		log.Println("Eroare la trimiterea cererii:", err)
		return
	}

	updateSearchResults(searchContainer, searchTerm)
}

func acceptFriendRequest(requesterID int) {
	_, err := db.Exec("UPDATE friendships SET status = 'accepted' WHERE user_id = ? AND friend_id = ?",
		requesterID, currentUserID)
	if err != nil {
		log.Println("Eroare la acceptarea cererii:", err)
		return
	}

	_, err = db.Exec("INSERT INTO friendships (user_id, friend_id, status) VALUES (?, ?, 'accepted')",
		currentUserID, requesterID)
	if err != nil {
		log.Println("Eroare la adăugarea relației inverse:", err)
	}

	updateFriendRequests()
	updateFriendsList()
}

func rejectFriendRequest(requesterID int) {
	_, err := db.Exec("DELETE FROM friendships WHERE user_id = ? AND friend_id = ?",
		requesterID, currentUserID)
	if err != nil {
		log.Println("Eroare la respingerea cererii:", err)
		return
	}

	updateFriendRequests()
}

func removeFriend(friendID int) {
	_, err := db.Exec("DELETE FROM friendships WHERE (user_id = ? AND friend_id = ?) OR (user_id = ? AND friend_id = ?)",
		currentUserID, friendID, friendID, currentUserID)
	if err != nil {
		log.Println("Eroare la ștergerea prietenului:", err)
		return
	}

	updateFriendsList()
}
