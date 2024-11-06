# File Security Scanner Web Application

## Overview
This web application provides a secure file scanning service with user authentication. It features a Windows Defender integration for malware scanning, a modern dark-themed UI, and protected routes requiring user authentication.

## Features
- User Authentication System
- Protected Routes
- File Upload and Scanning
- Windows Defender Integration
- Real-time Scan Status Updates
- Debug Information Display
- Modern Dark Theme UI
- Session Management
- Secure Password Handling

## Prerequisites
- Go 1.16 or higher
- Windows OS (for Windows Defender integration)
- Git (for cloning the repository)
- SQLite3

## Required Go Packages
```bash
go get github.com/gorilla/sessions
go get golang.org/x/crypto/bcrypt
```

## Installation

1. Clone the repository:
```bash
git clone [your-repository-url]
cd [project-directory]
```

2. Run the application:
```bash
go run ./cmd
```

## Security Features

### Password Hashing
- Uses bcrypt for password hashing
- Configurable work factor (currently set to 14)

### Session Management
- Secure cookie-based sessions
- Configurable session timeout
- HTTP-only flag enabled
- Secure flag available for HTTPS

### Authentication
- Required for protected routes
- Session-based authentication
- Secure password storage
- Login/logout functionality

### Authentication Routes
- `GET /login` - Login page
- `POST /login` - Login submission
- `GET /register` - Registration page
- `POST /register` - Registration submission

### Scanner Routes
- `GET /scan` - Scanner interface (protected)
- `POST /scan` - File scanning endpoint (protected)


## License
This project is licensed under the MIT License - see the LICENSE file for details

## Support
For support, please open an issue in the repository or contact [your-contact-information]