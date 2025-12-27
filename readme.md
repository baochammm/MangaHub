# **Manga Hub**


## Prerequisites

- Go 1.22 or newer  
- (Optional) C compiler 
- Sqlite

---

## 📦 Step 1: Get the Code

Open your terminal and run:

```bash
git clone https://github.com/baochammm/mangahub.git
cd mangahub
```

You should now be in the `mangahub` directory.

---

## 🔐 Step 2: Configure Environment

Create a `.env` file with your JWT secret:

```powershell
@"
JWT_SECRET=my_super_secret_jwt_key_change_in_production
DB_PATH=./data/mangahub.db

```

## 🗄️ Step 3: Install dependencies

```powershell
@"
go mod tidy

```

## 🗄️ Step 4: Start the server
```powershell
@"
go run ./cmd/api-server

```
The server will now be running! Port used: 8080, 9090, 9091, 9092
a database named "mangahub.db" will be created in your /data directory


## 🗄️ Step 5: Initialize the database
On another terminal, run
```powershell
@"
go run ./data/init.go

```
This will populate the database with mangas & account:

**Expected output:**
```
🗄️  Initializing database...
✅ Manga seed import completed
✅ User seed import completed
✅ Database initialized
```

## 👤 Default User Accounts

| Username | Password | Role | Description |
|----------|----------|------|-------------|
| `admin` | `admin123` | admin | Full access |
| `user` | `user123` | user | Regular user |

⚠️ **Remember to change these passwords for production!**

---


# **How to run CLI - FOR TESTING PURPOSE ONLY, FOR ACTUAL CLIENT SIDE WITH FULL SERVICES, PLEASE CHECKOUT OUR DESKTOP APP**
1. Run this in your terminal 
```
go build ./mangahub 
```

2. How to run commands
Available commands:
cd the project folder
```
./mangahub library list
./mangahub auth
./mangahub notify 
./mangahub sync
./mangahub progress
```
and
```
mangahub auth login
```
