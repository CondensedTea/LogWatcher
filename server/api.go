package main

import (
	"github.com/gin-gonic/gin"
)

type CreateMessage struct {
	server   int
	region   string
	pickupID int
}

func NewRouter(server *Server) *gin.Engine {
	router := gin.Default()

	router.PUT("/game", server.putPickupInfo)
	return router
}

func (s *Server) putPickupInfo(context *gin.Context) {
	var cm CreateMessage
	if err := context.BindJSON(&cm); err != nil {
		context.AbortWithStatus(400)
	}
	for _, v := range s.addressMap {
		if v.region == cm.region && v.server == cm.server {
			v.pickupID = cm.pickupID
		}
	}
	context.JSON(201, gin.H{
		"status": "accepted",
	})
}
