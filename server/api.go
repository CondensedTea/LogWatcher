package main

import (
	"net/http"

	"github.com/gin-gonic/gin"
)

type CreateMessage struct {
	Server   int    `json:"Server"`
	Region   string `json:"Region"`
	PickupID int    `json:"pickup_id"`
}

func NewRouter(server *Server) *gin.Engine {
	gin.SetMode(gin.ReleaseMode)
	router := gin.Default()

	//router.POST("/game", server.putPickupInfo)
	return router
}

func (s *Server) putPickupInfo(context *gin.Context) {
	var cm CreateMessage
	if err := context.BindJSON(&cm); err != nil {
		context.JSON(http.StatusBadRequest, gin.H{
			"error": err,
		})
		return
	}
	for _, v := range s.addressMap {
		if v.Region == cm.Region && v.Server == cm.Server {
			v.PickupID = cm.PickupID
			context.JSON(http.StatusAccepted, gin.H{
				"status": "accepted",
			})
			return
		}
	}
	context.JSON(http.StatusNotFound, gin.H{
		"error": "gameserver was not found",
	})
	return
}
