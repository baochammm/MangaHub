# **Manga Hub**

---

## Prerequisites

- Go 1.22 or newer  
- (Optional) C compiler 

---

## Getting  started



1 Clone the repository
```bash
git clone https://github.com/baochammm/mangahub.git
cd mangahub
```



2 Install the 
```
go mod tidy
```



3 Initialize database
```
go run ./data/init.go
```
a database named "mangahub.db" will be created in your /data directory



4 Run the server
```
go run ./cmd/api-server
```

5 .env 
 Create a constant "JWT_SECRET" that stores your Jwt Secret Key

# **How to run CLI**
1. Run this in your terminal (T MỚI MOVE MANGAHUB RA NGOÀI NHA)
```
go build ./mangahub 
```
Lệnh này dùng mỗi lần có chỉnh sửa gì trong file cmd/mangahub/main.go

2. How to run commands
Available commands:
```
mangahub library list
mangahub auth
mangahub notify 
```
and
```
mangahub auth login
```
