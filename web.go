package main

import (
	"dobrikov91/shubert/model"
	"encoding/json"
	"fmt"
	"html/template"
	"log"
	"net/http"
	"net/url"
	"os"
	"strconv"
	"sync"

	"github.com/gorilla/websocket"
)

// basicAuth wraps a handler with HTTP Basic Authentication when SHUBERT_USER
// and SHUBERT_PASS environment variables are set. If either variable is empty,
// the handler is returned unchanged (auth disabled).
func basicAuth(h http.HandlerFunc) http.HandlerFunc {
	user := os.Getenv("SHUBERT_USER")
	pass := os.Getenv("SHUBERT_PASS")
	if user == "" || pass == "" {
		return h
	}
	return func(w http.ResponseWriter, r *http.Request) {
		u, p, ok := r.BasicAuth()
		if !ok || u != user || p != pass {
			w.Header().Set("WWW-Authenticate", `Basic realm="shubert"`)
			http.Error(w, "Unauthorized", http.StatusUnauthorized)
			return
		}
		h(w, r)
	}
}

type WebServer struct {
	c         *Controller
	broadcast chan model.Commands
}

var clients = make(map[*websocket.Conn]bool)
var mu sync.Mutex

func NewWebserver(c *Controller) *WebServer {
	return &WebServer{
		c,
		make(chan model.Commands),
	}
}

func (web *WebServer) Run() {
	http.HandleFunc("/", basicAuth(web.handleHome))
	http.HandleFunc("/help", basicAuth(web.handleHelp))
	http.HandleFunc("/contact", basicAuth(web.handleContact))

	http.HandleFunc("/save", basicAuth(web.handleSave))
	http.HandleFunc("/delete", basicAuth(web.handleDelete))

	http.HandleFunc("/modeCommands", basicAuth(web.handleModeCommands))
	http.HandleFunc("/modeEdit", basicAuth(web.handleModeConfig))

	http.HandleFunc("/ws", basicAuth(web.handleConnections))
	http.HandleFunc("/init", basicAuth(web.handleInitialData))

	http.Handle("/css/", http.StripPrefix("/css/", http.FileServer(http.Dir("./templates/css"))))
	http.Handle("/pics/", http.StripPrefix("/pics/", http.FileServer(http.Dir("./templates/pics"))))
	http.Handle("/scripts/", http.StripPrefix("/scripts/", http.FileServer(http.Dir("./templates/scripts"))))

	hostname, err := os.Hostname()
	if err != nil {
		log.Fatal("Cant get hostname", err)
	}
	addr := fmt.Sprintf("%s:%s", hostname, web.c.Port)

	go web.handleMessages()

	log.Printf("Server is running at http://%s", addr)
	log.Fatal(http.ListenAndServe(fmt.Sprintf(":%s", web.c.Port), nil))
}

var upgrader = websocket.Upgrader{
	CheckOrigin: func(r *http.Request) bool {
		return true // Allow all origins (update as needed for security)
	},
}

func (web *WebServer) handleInitialData(w http.ResponseWriter, r *http.Request) {
	mu.Lock()
	defer mu.Unlock()
	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(web.c.Config.Data)
}

// WebSocket endpoint
func (web *WebServer) handleConnections(w http.ResponseWriter, r *http.Request) {
	ws, err := upgrader.Upgrade(w, r, nil)
	if err != nil {
		log.Printf("Error upgrading to WebSocket: %v\n", err)
		return
	}
	defer ws.Close()

	clients[ws] = true

	for {
		// Keep the connection open for receiving messages (if needed)
		if _, _, err := ws.ReadMessage(); err != nil {
			log.Printf("Error reading WebSocket message: %v\n", err)
			delete(clients, ws)
			break
		}
	}
}

// Broadcast updates to all clients
func (web *WebServer) handleMessages() {
	for {
		updatedConfig := <-web.broadcast

		mu.Lock()
		message, err := json.Marshal(updatedConfig)
		mu.Unlock()

		if err != nil {
			log.Printf("Error marshalling JSON: %v\n", err)
			continue
		}

		for client := range clients {
			if err := client.WriteMessage(websocket.TextMessage, message); err != nil {
				log.Printf("Error sending WebSocket message: %v\n", err)
				client.Close()
				delete(clients, client)
			}
		}
	}
}

func (web *WebServer) handleModeCommands(w http.ResponseWriter, r *http.Request) {
	web.c.EditMode = false
	log.Print("Commands mode")
}

func (web *WebServer) handleModeConfig(w http.ResponseWriter, r *http.Request) {
	web.c.EditMode = true
	log.Print("Config mode")
}

func (web *WebServer) handleHome(w http.ResponseWriter, r *http.Request) {
	tmpl, err := template.ParseFiles("./templates/editor.html")
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	tmpl.Execute(w, web.c)
}

func (web *WebServer) handleHelp(w http.ResponseWriter, r *http.Request) {
	tmpl, err := template.ParseFiles("./templates/help-en.html")
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	tmpl.Execute(w, web.c)
}

func (web *WebServer) handleContact(w http.ResponseWriter, r *http.Request) {
	tmpl, err := template.ParseFiles("./templates/contact.html")
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	tmpl.Execute(w, web.c)
}

func commandFromForm(form url.Values, i int) (model.Command, error) {
	channel, err := atoi(form["Channel"][i])
	if err != nil {
		return model.Command{}, fmt.Errorf("invalid Channel: %w", err)
	}
	key, err := atoi(form["Key"][i])
	if err != nil {
		return model.Command{}, fmt.Errorf("invalid Key: %w", err)
	}
	timeout, err := atoi(form["Timeout"][i])
	if err != nil {
		return model.Command{}, fmt.Errorf("invalid Timeout: %w", err)
	}
	return model.Command{
		Event: model.Event{
			Device:  form["Device"][i],
			Channel: channel,
			Key:     key,
			Value:   0,
		},
		Alias:     form["Alias"][i],
		Trigger:   form["Trigger"][i],
		Command:   form["Command"][i],
		TimeoutMs: timeout,
	}, nil
}

func (web *WebServer) handleSave(w http.ResponseWriter, r *http.Request) {
	if err := r.ParseForm(); err != nil {
		http.Error(w, "invalid form data", http.StatusBadRequest)
		return
	}

	var commands []model.Command
	for i := range len(r.Form["Command"]) {
		cmd, err := commandFromForm(r.Form, i)
		if err != nil {
			http.Error(w, err.Error(), http.StatusBadRequest)
			return
		}
		commands = append(commands, cmd)
	}

	web.c.configMu.Lock()
	web.c.Config.ClearConfig()
	for _, cmd := range commands {
		web.c.Config.AddCommand(cmd)
	}
	web.c.configMu.Unlock()

	web.saveConfig(w, r)
}

func (web *WebServer) saveConfig(w http.ResponseWriter, r *http.Request) {
	err := web.c.Config.Save()
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	http.Redirect(w, r, "/", http.StatusSeeOther)
}

func (web *WebServer) handleDelete(w http.ResponseWriter, r *http.Request) {
	r.ParseForm()

	index, err := atoi(r.FormValue("index"))
	if err != nil {
		http.Error(w, "Invalid index", http.StatusBadRequest)
		return
	}

	web.c.configMu.Lock()
	if index < 0 || index >= len(web.c.Config.Data.Commands) {
		web.c.configMu.Unlock()
		http.Error(w, "Invalid index", http.StatusBadRequest)
		return
	}
	web.c.Config.DeleteCommand(index)
	web.c.configMu.Unlock()

	web.broadcast <- web.c.Config.Data
	web.saveConfig(w, r)
}

func atoi(s string) (int, error) {
	return strconv.Atoi(s)
}
