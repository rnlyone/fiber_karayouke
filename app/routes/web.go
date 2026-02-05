package routes

import (
	"GoFiberMVC/app/controllers"
	ws "GoFiberMVC/app/websocket"

	"github.com/gofiber/fiber/v2"
	"github.com/gofiber/fiber/v2/middleware/cors"
	"github.com/gofiber/websocket/v2"
)

// RegisterWebRoutes registers web routes
func RegisterWebRoutes(app *fiber.App) {
	// Enable CORS for frontend
	app.Use(cors.New(cors.Config{
		AllowOrigins:     "http://localhost:5173,http://127.0.0.1:5173,http://karayouke.com,http://www.karayouke.com,https://karayouke.com,https://www.karayouke.com",
		AllowHeaders:     "Origin, Content-Type, Accept, Authorization",
		AllowMethods:     "GET, POST, PUT, DELETE, OPTIONS",
		AllowCredentials: true,
	}))

	userController := &controllers.UserController{}
	authController := &controllers.AuthController{}
	roomController := &controllers.RoomController{}
	adminController := &controllers.AdminController{}
	packageController := &controllers.PackageController{}

	app.Get("", userController.Index)

	// Auth routes
	app.Post("/api/auth/register", authController.Register)
	app.Post("/api/auth/login", authController.Login)
	app.Get("/api/auth/me", authController.Me)
	app.Post("/api/auth/logout", authController.Logout)

	// Room routes
	app.Post("/api/rooms", roomController.Create)
	app.Get("/api/rooms", roomController.List)
	app.Get("/api/rooms/:roomKey", roomController.Get)
	app.Get("/api/rooms/:roomKey/access", roomController.CheckAccess)

	// Admin check (no middleware - returns is_admin status)
	app.Get("/api/admin/check", adminController.CheckAdmin)

	// Admin routes (protected)
	admin := app.Group("/api/admin", controllers.AdminMiddleware)
	admin.Get("/dashboard", adminController.GetDashboardStats)
	admin.Get("/configs", adminController.GetConfigs)
	admin.Post("/configs", adminController.CreateConfig)
	admin.Put("/configs/:key", adminController.UpdateConfig)
	admin.Delete("/configs/:key", adminController.DeleteConfig)
	admin.Get("/packages", adminController.ListPackages)
	admin.Post("/packages", adminController.CreatePackage)
	admin.Put("/packages/:id", adminController.UpdatePackage)
	admin.Delete("/packages/:id", adminController.DeletePackage)
	admin.Get("/users", adminController.ListUsers)
	admin.Get("/users/:id", adminController.GetUser)
	admin.Post("/credits/award", adminController.AwardCredits)
	admin.Get("/transactions", adminController.ListTransactions)
	admin.Put("/transactions/:id/status", adminController.UpdateTransactionStatus)
	admin.Get("/rooms", adminController.ListRooms)

	// Public package routes (for users to see available packages)
	app.Get("/api/packages", packageController.ListPublic)
	app.Post("/api/packages/:id/purchase", packageController.Purchase)
	app.Get("/api/transactions", packageController.MyTransactions)
	app.Get("/api/transactions/:id", packageController.GetTransaction)
	app.Get("/api/credits", packageController.GetMyCredits)
	app.Post("/api/payment/callback", packageController.PaymentCallback)

	// Stripe payment routes
	stripeController := &controllers.StripeController{}
	app.Get("/api/stripe/config", stripeController.GetPublishableKey)
	app.Post("/api/stripe/create-payment-intent", stripeController.CreatePaymentIntent)
	app.Post("/api/stripe/confirm-payment", stripeController.ConfirmPayment)
	app.Post("/api/stripe/webhook", stripeController.HandleWebhook)

	// TV connection routes
	tvController := &controllers.TVController{}
	app.Post("/api/tv/token", tvController.GenerateToken)          // Generate new TV token (no auth - TV device)
	app.Get("/api/tv/status/:token", tvController.GetStatus)       // Check TV connection status (no auth - TV polls)
	app.Post("/api/tv/connect", tvController.Connect)              // Connect TV to room (requires auth - room master)
	app.Post("/api/tv/disconnect/:token", tvController.Disconnect) // Disconnect TV from room

	// WebSocket routes for karaoke rooms
	// Support both /ws/:roomKey and /parties/main/:roomKey for compatibility
	app.Use("/ws", ws.WebSocketUpgrade)
	app.Get("/ws/:roomKey", websocket.New(ws.HandleWebSocket))

	// PartyKit-compatible route
	app.Use("/parties/main", ws.WebSocketUpgrade)
	app.Get("/parties/main/:roomKey", websocket.New(ws.HandleWebSocket))

	// HTTP API for room state
	app.Get("/api/room/:roomKey", ws.GetRoomState)
}
