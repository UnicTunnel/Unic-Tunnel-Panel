package server

import (
	"context"
	"errors"
	"html/template"
	"io/fs"
	"log"
	"net/http"
	"strconv"
	"strings"
	"time"

	"github.com/UnicTunnel/Unic-Tunnel-Panel/internal/accounts"
	"github.com/UnicTunnel/Unic-Tunnel-Panel/internal/auth"
	"github.com/UnicTunnel/Unic-Tunnel-Panel/internal/links"
	"github.com/UnicTunnel/Unic-Tunnel-Panel/internal/store"
	"github.com/UnicTunnel/Unic-Tunnel-Panel/web"
)

const (
	sessionCookie = "unic_session"
	sessionTTL    = 7 * 24 * time.Hour
)

type Deps struct {
	Store       *store.Store
	Provisioner accounts.Provisioner
	SSHHost     string
	SSHPort     int
}

type Server struct {
	deps Deps
	mux  *http.ServeMux
	tmpl map[string]*template.Template
}

type pageData struct {
	Error      string
	ShowLogout bool
	Data       any
}

func New(deps Deps) *Server {
	s := &Server{deps: deps, mux: http.NewServeMux()}
	s.loadTemplates()
	s.routes()
	return s
}

func (s *Server) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	s.mux.ServeHTTP(w, r)
}

func (s *Server) loadTemplates() {
	funcs := template.FuncMap{
		"fmtTime": func(t time.Time) string { return t.Format("2006-01-02 15:04") },
	}
	pages := []string{"login", "dashboard", "user"}
	s.tmpl = make(map[string]*template.Template, len(pages))
	for _, p := range pages {
		t, err := template.New("").Funcs(funcs).ParseFS(web.FS,
			"templates/layout.html", "templates/"+p+".html")
		if err != nil {
			log.Fatalf("template %s: %v", p, err)
		}
		s.tmpl[p] = t
	}
}

func (s *Server) routes() {
	staticFS, err := fs.Sub(web.FS, "static")
	if err != nil {
		log.Fatalf("static fs: %v", err)
	}
	s.mux.Handle("GET /static/", http.StripPrefix("/static/", http.FileServer(http.FS(staticFS))))

	s.mux.HandleFunc("GET /login", s.loginGet)
	s.mux.HandleFunc("POST /login", s.loginPost)
	s.mux.HandleFunc("POST /logout", s.logout)

	s.mux.HandleFunc("GET /{$}", s.requireAuth(s.dashboard))
	s.mux.HandleFunc("POST /users", s.requireAuth(s.createUser))
	s.mux.HandleFunc("GET /users/{id}", s.requireAuth(s.userDetail))
	s.mux.HandleFunc("POST /users/{id}/revoke", s.requireAuth(s.revokeUser))
}

func (s *Server) render(w http.ResponseWriter, page string, pd pageData) {
	t, ok := s.tmpl[page]
	if !ok {
		http.Error(w, "no template", http.StatusInternalServerError)
		return
	}
	w.Header().Set("Content-Type", "text/html; charset=utf-8")
	if err := t.ExecuteTemplate(w, "layout", pd); err != nil {
		log.Printf("template exec %s: %v", page, err)
	}
}

// --- auth middleware ---

type ctxKey int

const adminCtxKey ctxKey = 1

func (s *Server) requireAuth(next http.HandlerFunc) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		c, err := r.Cookie(sessionCookie)
		if err != nil {
			http.Redirect(w, r, "/login", http.StatusSeeOther)
			return
		}
		sess, err := s.deps.Store.GetSession(r.Context(), c.Value)
		if err != nil {
			s.clearCookie(w)
			http.Redirect(w, r, "/login", http.StatusSeeOther)
			return
		}
		ctx := context.WithValue(r.Context(), adminCtxKey, sess.AdminID)
		next(w, r.WithContext(ctx))
	}
}

func (s *Server) setCookie(w http.ResponseWriter, id string, expires time.Time) {
	http.SetCookie(w, &http.Cookie{
		Name:     sessionCookie,
		Value:    id,
		Path:     "/",
		Expires:  expires,
		HttpOnly: true,
		SameSite: http.SameSiteStrictMode,
	})
}

func (s *Server) clearCookie(w http.ResponseWriter) {
	http.SetCookie(w, &http.Cookie{
		Name:     sessionCookie,
		Value:    "",
		Path:     "/",
		MaxAge:   -1,
		HttpOnly: true,
		SameSite: http.SameSiteStrictMode,
	})
}

// --- handlers ---

func (s *Server) loginGet(w http.ResponseWriter, r *http.Request) {
	s.render(w, "login", pageData{})
}

func (s *Server) loginPost(w http.ResponseWriter, r *http.Request) {
	if err := r.ParseForm(); err != nil {
		http.Error(w, "bad form", http.StatusBadRequest)
		return
	}
	username := r.FormValue("username")
	password := r.FormValue("password")
	admin, err := s.deps.Store.GetAdmin(r.Context(), username)
	if err != nil {
		s.render(w, "login", pageData{Error: "Wrong username or password."})
		return
	}
	if err := auth.VerifyPassword(admin.PasswordHash, password); err != nil {
		s.render(w, "login", pageData{Error: "Wrong username or password."})
		return
	}
	sess, err := s.deps.Store.CreateSession(r.Context(), admin.ID, sessionTTL)
	if err != nil {
		http.Error(w, "session error", http.StatusInternalServerError)
		return
	}
	s.setCookie(w, sess.ID, sess.ExpiresAt)
	http.Redirect(w, r, "/", http.StatusSeeOther)
}

func (s *Server) logout(w http.ResponseWriter, r *http.Request) {
	if c, err := r.Cookie(sessionCookie); err == nil {
		_ = s.deps.Store.DeleteSession(r.Context(), c.Value)
	}
	s.clearCookie(w)
	http.Redirect(w, r, "/login", http.StatusSeeOther)
}

type dashboardData struct {
	Users   []store.TunnelUser
	SSHHost string
	SSHPort int
}

func (s *Server) dashboard(w http.ResponseWriter, r *http.Request) {
	users, err := s.deps.Store.ListTunnelUsers(r.Context())
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
	s.render(w, "dashboard", pageData{
		ShowLogout: true,
		Data:       dashboardData{Users: users, SSHHost: s.deps.SSHHost, SSHPort: s.deps.SSHPort},
	})
}

func (s *Server) createUser(w http.ResponseWriter, r *http.Request) {
	if err := r.ParseForm(); err != nil {
		http.Error(w, "bad form", http.StatusBadRequest)
		return
	}
	name := strings.TrimSpace(r.FormValue("name"))
	if name == "" {
		name = "Untitled"
	}
	username, err := accounts.GenerateUsername()
	if err != nil {
		http.Error(w, "name gen", http.StatusInternalServerError)
		return
	}
	password, err := accounts.GeneratePassword()
	if err != nil {
		http.Error(w, "pw gen", http.StatusInternalServerError)
		return
	}
	if err := s.deps.Provisioner.Create(r.Context(), username, password); err != nil {
		http.Error(w, "provisioner: "+err.Error(), http.StatusInternalServerError)
		return
	}
	user, err := s.deps.Store.CreateTunnelUser(r.Context(), name, username, password)
	if err != nil {
		// Best-effort rollback: undo the OS account so it doesn't drift from the DB.
		_ = s.deps.Provisioner.Delete(r.Context(), username)
		http.Error(w, "store: "+err.Error(), http.StatusInternalServerError)
		return
	}
	http.Redirect(w, r, "/users/"+strconv.FormatInt(user.ID, 10), http.StatusSeeOther)
}

type userDetailData struct {
	User store.TunnelUser
	Link string
}

func (s *Server) userDetail(w http.ResponseWriter, r *http.Request) {
	id, err := strconv.ParseInt(r.PathValue("id"), 10, 64)
	if err != nil {
		http.NotFound(w, r)
		return
	}
	user, err := s.deps.Store.GetTunnelUser(r.Context(), id)
	if err != nil {
		if errors.Is(err, store.ErrNotFound) {
			http.NotFound(w, r)
			return
		}
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
	link, err := links.Build(links.Payload{
		Name:     user.Name,
		Host:     s.deps.SSHHost,
		Port:     s.deps.SSHPort,
		User:     user.Username,
		Password: user.Password,
	})
	if err != nil {
		http.Error(w, "link: "+err.Error(), http.StatusInternalServerError)
		return
	}
	s.render(w, "user", pageData{
		ShowLogout: true,
		Data:       userDetailData{User: *user, Link: link},
	})
}

func (s *Server) revokeUser(w http.ResponseWriter, r *http.Request) {
	id, err := strconv.ParseInt(r.PathValue("id"), 10, 64)
	if err != nil {
		http.NotFound(w, r)
		return
	}
	user, err := s.deps.Store.GetTunnelUser(r.Context(), id)
	if err != nil {
		http.NotFound(w, r)
		return
	}
	if err := s.deps.Provisioner.Lock(r.Context(), user.Username); err != nil {
		http.Error(w, "provisioner: "+err.Error(), http.StatusInternalServerError)
		return
	}
	if err := s.deps.Store.RevokeTunnelUser(r.Context(), id); err != nil {
		http.Error(w, "store: "+err.Error(), http.StatusInternalServerError)
		return
	}
	http.Redirect(w, r, "/", http.StatusSeeOther)
}
