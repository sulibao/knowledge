package middleware

import (
	"net/http"

	"github.com/gorilla/sessions"
)

var ( // TODO: Use a more secure key in production
	key   = []byte("super-secret-key")
	Store *sessions.CookieStore
)

func init() {
	Store = sessions.NewCookieStore(key)
	Store.Options.HttpOnly = true
	Store.Options.Secure = false                      // Set to true in production with HTTPS
	Store.Options.SameSite = http.SameSiteDefaultMode // Adjust as needed for production
}

func AuthRequired(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		session, _ := Store.Get(r, "session-name")

		if auth, ok := session.Values["authenticated"].(bool); !ok || !auth {
			http.Redirect(w, r, "/login", http.StatusFound)
			return
		}
		next.ServeHTTP(w, r)
	})
}
