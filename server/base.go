package server

import (
	"net/http"

	"github.com/gin-gonic/gin"
)

func index(c *gin.Context) {
	c.Redirect(http.StatusPermanentRedirect, "/browser")
}

func base(c *gin.Context) {
	c.Redirect(http.StatusPermanentRedirect, "/browser")
}
