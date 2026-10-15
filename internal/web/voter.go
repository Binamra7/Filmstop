package web

import (
	"crypto/hmac"
	"crypto/rand"
	"crypto/sha256"
	"encoding/hex"
	"net/http"
	"strings"

	"github.com/gin-gonic/gin"
)

const voterCookie = "fs_voter"

// Voters are anonymous: a random id in an HMAC-signed cookie. One cookie, one
// ballot; casting again replaces it.
type voters struct{ secret []byte }

func (v voters) sign(id string) string {
	m := hmac.New(sha256.New, v.secret)
	m.Write([]byte(id))
	return hex.EncodeToString(m.Sum(nil)[:16])
}

func (v voters) verify(cookie string) (string, bool) {
	id, sig, ok := strings.Cut(cookie, ".")
	if !ok || len(id) != 32 {
		return "", false
	}
	return id, hmac.Equal([]byte(sig), []byte(v.sign(id)))
}

// middleware puts the voter id in the context, issuing a cookie on first visit.
func (v voters) middleware(c *gin.Context) {
	if raw, err := c.Cookie(voterCookie); err == nil {
		if id, ok := v.verify(raw); ok {
			c.Set(voterCookie, id)
			c.Next()
			return
		}
	}
	buf := make([]byte, 16)
	rand.Read(buf)
	id := hex.EncodeToString(buf)
	secure := c.Request.TLS != nil || c.GetHeader("X-Forwarded-Proto") == "https"
	http.SetCookie(c.Writer, &http.Cookie{
		Name: voterCookie, Value: id + "." + v.sign(id), Path: "/",
		MaxAge: 2 * 365 * 24 * 3600, HttpOnly: true, Secure: secure, SameSite: http.SameSiteLaxMode,
	})
	c.Set(voterCookie, id)
	c.Next()
}

func voterID(c *gin.Context) string { return c.GetString(voterCookie) }

// sameOrigin rejects cross-site form posts. SameSite=Lax already keeps the cookie
// off them; this also covers browsers that ignore it.
func sameOrigin(c *gin.Context) {
	if c.Request.Method == http.MethodPost {
		if o := c.GetHeader("Origin"); o != "" && !strings.HasSuffix(o, "://"+c.Request.Host) {
			c.AbortWithStatus(http.StatusForbidden)
			return
		}
	}
	c.Next()
}
