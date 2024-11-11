package main

import (
	"log"
	"net/http"
	"os"
	"path/filepath"
	"strings"
	"text/template"

	"golang.org/x/crypto/bcrypt"
)

func hashPassword(password string) (string, error) {
	bytes, err := bcrypt.GenerateFromPassword([]byte(password), 14)
	return string(bytes), err
}

func checkPasswordHash(password, hash string) bool {
	err := bcrypt.CompareHashAndPassword([]byte(hash), []byte(password))
	return err == nil
}

func registerHandler(w http.ResponseWriter, r *http.Request) {
	switch r.Method {
	case http.MethodGet:
		ts, err := template.ParseFiles("themes/register.html")
		if err != nil {
			log.Print(err.Error())
			http.Error(w, "Internal Server Error", http.StatusInternalServerError)
			return
		}

		err = ts.Execute(w, nil)
		if err != nil {
			log.Print(err.Error())
			http.Error(w, "Internal Server Error", http.StatusInternalServerError)
			return
		}

	case http.MethodPost:
		username := r.FormValue("username")
		password := r.FormValue("password")

		hashedPassword, err := hashPassword(password)
		if err != nil {
			http.Error(w, "Server error", http.StatusInternalServerError)
			return
		}

		_, err = db.Exec("INSERT INTO users (username, password) VALUES (?, ?)", username, hashedPassword)
		if err != nil {
			http.Error(w, "Username already taken", http.StatusBadRequest)
			return
		}

		http.Redirect(w, r, "/login", http.StatusSeeOther)

	default:
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
	}
}

func loginHandler(w http.ResponseWriter, r *http.Request) {
	switch r.Method {
	case http.MethodGet:
		ts, err := template.ParseFiles("themes/login.html")
		if err != nil {
			log.Print(err.Error())
			http.Error(w, "Internal Server Error", http.StatusInternalServerError)
			return
		}

		err = ts.Execute(w, nil)
		if err != nil {
			log.Print(err.Error())
			http.Error(w, "Internal Server Error", http.StatusInternalServerError)
			return
		}

	case http.MethodPost:
		username := r.FormValue("username")
		password := r.FormValue("password")

		var hashedPassword string
		err := db.QueryRow("SELECT password FROM users WHERE username = ?", username).Scan(&hashedPassword)
		if err != nil || !checkPasswordHash(password, hashedPassword) {
			http.Error(w, "Invalid username or password", http.StatusUnauthorized)
			return
		}

		session, _ := store.Get(r, "session")
		session.Values["authenticated"] = true
		session.Save(r, w)

		http.Redirect(w, r, "/", http.StatusSeeOther)

	default:
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
	}
}

func homeHandler(w http.ResponseWriter, r *http.Request) {
	if r.URL.Path != "/" {
		http.NotFound(w, r)
		return
	}

	session, _ := store.Get(r, "session")
	auth, ok := session.Values["authenticated"].(bool)

	data := struct {
		Authenticated bool
	}{
		Authenticated: ok && auth,
	}

	ts, err := template.ParseFiles("themes/home.html")
	if err != nil {
		log.Print(err.Error())
		http.Error(w, "Internal Server Error", http.StatusInternalServerError)
		return
	}

	err = ts.Execute(w, data)
	if err != nil {
		log.Print(err.Error())
		http.Error(w, "Internal Server Error", http.StatusInternalServerError)
	}
}

func pageHandler(w http.ResponseWriter, r *http.Request) {
	path := strings.TrimPrefix(r.URL.Path, "/")

	if path == "" {
		path = "home.html"
	}

	if !strings.HasSuffix(path, ".html") {
		path += ".html"
	}

	fullPath := filepath.Join("themes", path)

	if _, err := os.Stat(fullPath); os.IsNotExist(err) {
		http.NotFound(w, r)
		return
	}

	protectedPages := map[string]bool{
		"monitoring.html": true,
		"incident.html":   true,
		"scan.html":       true,
	}

	if protectedPages[path] {
		session, _ := store.Get(r, "session")
		auth, ok := session.Values["authenticated"].(bool)
		if !ok || !auth {
			http.Redirect(w, r, "/login", http.StatusSeeOther)
			return
		}
	}

	ts, err := template.ParseFiles(fullPath)
	if err != nil {
		log.Printf("Error parsing template %s: %v", path, err)
		http.Error(w, "Internal Server Error", http.StatusInternalServerError)
		return
	}

	session, _ := store.Get(r, "session")
	auth, ok := session.Values["authenticated"].(bool)

	data := struct {
		Authenticated bool
	}{
		Authenticated: ok && auth,
	}

	err = ts.Execute(w, data)
	if err != nil {
		log.Printf("Error executing template %s: %v", path, err)
		http.Error(w, "Internal Server Error", http.StatusInternalServerError)
	}
}

// func contactusHandler(w http.ResponseWriter, r *http.Request) {
// 	// Debug logging
// 	log.Printf("Attempting to serve contactus page")

// 	// Check if file exists before trying to parse it
// 	_, err := os.Stat("themes/contactus.html")
// 	if os.IsNotExist(err) {
// 		log.Printf("File not found: themes/contactus.html")
// 		http.Error(w, "Page not found", http.StatusNotFound)
// 		return
// 	}

// 	ts, err := template.ParseFiles("themes/contactus.html")
// 	if err != nil {
// 		log.Printf("Error parsing template: %v", err)
// 		http.Error(w, "Internal Server Error", http.StatusInternalServerError)
// 		return
// 	}

// 	session, _ := store.Get(r, "session")
// 	auth, ok := session.Values["authenticated"].(bool)

// 	data := struct {
// 		Authenticated bool
// 	}{
// 		Authenticated: ok && auth,
// 	}

// 	err = ts.Execute(w, data)
// 	if err != nil {
// 		log.Printf("Error executing template: %v", err)
// 		http.Error(w, "Internal Server Error", http.StatusInternalServerError)
// 	}
// }

// func monitoringHandler(w http.ResponseWriter, r *http.Request) {
// 	session, _ := store.Get(r, "session")
// 	auth, ok := session.Values["authenticated"].(bool)
// 	if !ok || !auth {
// 		http.Redirect(w, r, "/login", http.StatusSeeOther)
// 		return
// 	}

// 	ts, err := template.ParseFiles("themes/monitoring.html")
// 	if err != nil {
// 		log.Print(err.Error())
// 		http.Error(w, "Internal Server Error", http.StatusInternalServerError)
// 		return
// 	}

// 	data := struct {
// 		Authenticated bool
// 	}{
// 		Authenticated: true,
// 	}

// 	err = ts.Execute(w, data)
// 	if err != nil {
// 		log.Print(err.Error())
// 		http.Error(w, "Internal Server Error", http.StatusInternalServerError)
// 	}
// }
// func auditHandler(w http.ResponseWriter, r *http.Request) {
//     session, _ := store.Get(r, "session")
//     auth, ok := session.Values["authenticated"].(bool)
//     if !ok || !auth {
//         http.Redirect(w, r, "/login", http.StatusSeeOther)
//         return
//     }

//     ts, err := template.ParseFiles("themes/audit.html")
//     if err != nil {
//         log.Print(err.Error())
//         http.Error(w, "Internal Server Error", http.StatusInternalServerError)
//         return
//     }

//     data := struct {
//         Authenticated bool
//     }{
//         Authenticated: true,
//     }

//     err = ts.Execute(w, data)
//     if err != nil {
//         log.Print(err.Error())
//         http.Error(w, "Internal Server Error", http.StatusInternalServerError)
//     }
// }

// func contactUsHandler(w http.ResponseWriter, r *http.Request) {
//     ts, err := template.ParseFiles("themes/contactus.html")
//     if err != nil {
//         log.Print(err.Error())
//         http.Error(w, "Internal Server Error", http.StatusInternalServerError)
//         return
//     }

//     session, _ := store.Get(r, "session")
//     auth, ok := session.Values["authenticated"].(bool)

//     data := struct {
//         Authenticated bool
//     }{
//         Authenticated: ok && auth,
//     }

//     err = ts.Execute(w, data)
//     if err != nil {
//         log.Print(err.Error())
//         http.Error(w, "Internal Server Error", http.StatusInternalServerError)
//     }
// }

// func incidentHandler(w http.ResponseWriter, r *http.Request) {
//     session, _ := store.Get(r, "session")
//     auth, ok := session.Values["authenticated"].(bool)
//     if !ok || !auth {
//         http.Redirect(w, r, "/login", http.StatusSeeOther)
//         return
//     }

//     ts, err := template.ParseFiles("themes/incident.html")
//     if err != nil {
//         log.Print(err.Error())
//         http.Error(w, "Internal Server Error", http.StatusInternalServerError)
//         return
//     }

//     data := struct {
//         Authenticated bool
//     }{
//         Authenticated: true,
//     }

//     err = ts.Execute(w, data)
//     if err != nil {
//         log.Print(err.Error())
//         http.Error(w, "Internal Server Error", http.StatusInternalServerError)
//     }
// }

// func consultationHandler(w http.ResponseWriter, r *http.Request) {
//     ts, err := template.ParseFiles("themes/consultation.html")
//     if err != nil {
//         log.Print(err.Error())
//         http.Error(w, "Internal Server Error", http.StatusInternalServerError)
//         return
//     }

//     session, _ := store.Get(r, "session")
//     auth, ok := session.Values["authenticated"].(bool)

//     data := struct {
//         Authenticated bool
//     }{
//         Authenticated: ok && auth,
//     }

//     err = ts.Execute(w, data)
//     if err != nil {
//         log.Print(err.Error())
//         http.Error(w, "Internal Server Error", http.StatusInternalServerError)
//     }
// }

// func penetrationHandler(w http.ResponseWriter, r *http.Request) {
//     session, _ := store.Get(r, "session")
//     auth, ok := session.Values["authenticated"].(bool)
//     if !ok || !auth {
//         http.Redirect(w, r, "/login", http.StatusSeeOther)
//         return
//     }

//     ts, err := template.ParseFiles("themes/penetration.html")
//     if err != nil {
//         log.Print(err.Error())
//         http.Error(w, "Internal Server Error", http.StatusInternalServerError)
//         return
//     }

//     data := struct {
//         Authenticated bool
//     }{
//         Authenticated: true,
//     }

//     err = ts.Execute(w, data)
//     if err != nil {
//         log.Print(err.Error())
//         http.Error(w, "Internal Server Error", http.StatusInternalServerError)
//     }
// }

// func educationHandler(w http.ResponseWriter, r *http.Request) {
//     ts, err := template.ParseFiles("themes/education.html")
//     if err != nil {
//         log.Print(err.Error())
//         http.Error(w, "Internal Server Error", http.StatusInternalServerError)
//         return
//     }

//     session, _ := store.Get(r, "session")
//     auth, ok := session.Values["authenticated"].(bool)

//     data := struct {
//         Authenticated bool
//     }{
//         Authenticated: ok && auth,
//     }

//     err = ts.Execute(w, data)
//     if err != nil {
//         log.Print(err.Error())
//         http.Error(w, "Internal Server Error", http.StatusInternalServerError)
//     }
// }

// func trainingHandler(w http.ResponseWriter, r *http.Request) {
//     ts, err := template.ParseFiles("themes/training.html")
//     if err != nil {
//         log.Print(err.Error())
//         http.Error(w, "Internal Server Error", http.StatusInternalServerError)
//         return
//     }

//     session, _ := store.Get(r, "session")
//     auth, ok := session.Values["authenticated"].(bool)

//     data := struct {
//         Authenticated bool
//     }{
//         Authenticated: ok && auth,
//     }

//     err = ts.Execute(w, data)
//     if err != nil {
//         log.Print(err.Error())
//         http.Error(w, "Internal Server Error", http.StatusInternalServerError)
//     }
// }

// func subHandler(w http.ResponseWriter, r *http.Request) {
//     ts, err := template.ParseFiles("themes/sub.html")
//     if err != nil {
//         log.Print(err.Error())
//         http.Error(w, "Internal Server Error", http.StatusInternalServerError)
//         return
//     }

//     session, _ := store.Get(r, "session")
//     auth, ok := session.Values["authenticated"].(bool)

//     data := struct {
//         Authenticated bool
//     }{
//         Authenticated: ok && auth,
//     }

//     err = ts.Execute(w, data)
//     if err != nil {
//         log.Print(err.Error())
//         http.Error(w, "Internal Server Error", http.StatusInternalServerError)
//     }
// }
