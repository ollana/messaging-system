package routes

import "github.com/gin-gonic/gin"

type Router struct {
	Users    UsersRoutes
	Groups   GroupRoutes
	Messages MessagesRoutes
}

func (r *Router) NewRouter() (engine *gin.Engine, err error) {
	engine = gin.New()
	r.Route(engine)

	engine.NoRoute(func(c *gin.Context) {
		c.Status(404)
	})
	return engine, nil

}

func (router *Router) Route(r *gin.Engine) {
	router.v1Routes(r.Group("/v1"))

}

func (router *Router) v1Routes(group *gin.RouterGroup) {

	group.POST("/users/register", router.Users.RegisterUserHandler)
	group.POST("/users/:userId", router.Users.BlockUserHandler)

	group.POST("/groups/create", router.Groups.CreateGroupHandler)
	group.POST("/groups/:groupId", router.Groups.UserToGroupHandler)

	group.POST("/messages/send", router.Messages.SendMessageHandler)
	group.GET("/messages/:userId", router.Messages.GetMessagesHandler)

}
