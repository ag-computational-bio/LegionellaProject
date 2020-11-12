package server

import "github.com/gin-gonic/gin"

func addCORS(c *gin.Context) {
	headers := c.Writer.Header()
	headers["Access-Control-Allow-Origin"] = []string{"*"}
	headers["Access-Control-Allow-Methods"] = []string{"GET,POST,PUT,OPTIONS"}
}
