#  💬Fyne Message App

This is a desktop application written in Go using the [Fyne](https://fyne.io/) framework. It provides:

- Account creation (Sign Up)
- User login
- Searching for other users
- Sending friend requests
- Accepting or rejecting friend requests
- Managing your friends list
- Chatting with friends (messaging)

## 📖Requirements

- Go 1.18+ installed
- A running MySQL server
- The `Fyne` library
- MySQL driver for Go (`github.com/go-sql-driver/mysql`)

## 💿Database Setup

Create a MySQL database named `user_auth` and the required tables:

```sql
CREATE DATABASE user_auth;

USE user_auth;

CREATE TABLE users (
    id INT AUTO_INCREMENT PRIMARY KEY,
    full_name VARCHAR(100),
    username VARCHAR(50) UNIQUE,
    password VARCHAR(100)
);

CREATE TABLE friendships (
    id INT AUTO_INCREMENT PRIMARY KEY,
    user_id INT,
    friend_id INT,
    status ENUM('pending','accepted'),
    FOREIGN KEY (user_id) REFERENCES users(id),
    FOREIGN KEY (friend_id) REFERENCES users(id)
);

CREATE TABLE messages (
    id INT AUTO_INCREMENT PRIMARY KEY,
    sender_id INT,
    receiver_id INT,
    message TEXT,
    created_at DATETIME,
    FOREIGN KEY (sender_id) REFERENCES users(id),
    FOREIGN KEY (receiver_id) REFERENCES users(id)
);
```

> **Note:** In main.go, the database connection is defined like this:
> db, err = sql.Open("mysql", "root:password@tcp(127.0.0.1:3306)/user_auth")
> > **Important:** You should change the user and password to your own credentials, or use environment variables for security.

## 📥Install Dependencies

 ```bat
   go mod tidy
   ```
This will download all the necessary modules.

## 📥Build and Run

Run directly:
 ```bat
   go run main.go
   ```
Or build a binary:
```bat
   go build -o fyne_app
./fyne_app
   ```

## 📁Project Structure

- main.go – the main application source code
- go.mod / go.sum – Go module and dependency definitions
- Graphical interface created with Fyne
- MySQL connection for storing users, friendships, and messages

## 🔐Security Notice

Never commit real database credentials to a public repository.
It's recommended to use environment variables to configure your database connection.

## ✍️ Author

- Created with ❤️ by Manceriu Mădălin
