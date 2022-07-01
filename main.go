package main

import (
	"fmt"
	"github.com/Xacnio/ytmp3-host-go/app/bot"
	"github.com/Xacnio/ytmp3-host-go/app/handlers"
	"github.com/Xacnio/ytmp3-host-go/pkg/configs"
	"github.com/Xacnio/ytmp3-host-go/platform/database"
	"github.com/gofiber/fiber/v2"
	"github.com/gofiber/fiber/v2/middleware/csrf"
	"github.com/gofiber/fiber/v2/middleware/logger"
	"github.com/gofiber/template/html"
	"os"
	"os/signal"
	"strings"
	"syscall"
)

var AllowedIps map[string]bool

func SetupFiber() {
	AllowedIps = make(map[string]bool)
	allowedIpList := strings.Split(os.Getenv("ALLOWED_IPS"), ",")
	for i := 0; i < len(allowedIpList); i++ {
		if allowedIpList[i] != "" {
			AllowedIps[allowedIpList[i]] = true
		}
	}

	engine := html.New("./web/views", ".html")

	app := fiber.New(fiber.Config{
		Views:       engine,
		BodyLimit:   30 * 1024 * 1024,
		ProxyHeader: fiber.HeaderXForwardedFor,
	})
	app.Use(logger.New())
	app.Use(csrf.New(csrf.Config{
		KeyLookup:      "form:csrf",
		ContextKey:     "csrf",
		CookieName:     "csrf_",
		CookieSameSite: "Strict",
	}))
	//app.Use(handlers.HttpsForwarder)

	app.Static("/", "./web/static")
	app.Static("/", "./web/data")

	app.Get("/", func(c *fiber.Ctx) error {
		c.Set("cache-control", "no-store, no-cache, must-revalidate")
		tk := c.Locals("csrf")
		return c.Render("index", fiber.Map{"CSRF": tk})
	})

	app.Get("/:filename", func(c *fiber.Ctx) error {
		return handlers.ProxyHandler(c)
	})
	app.Post("/upload", func(c *fiber.Ctx) error {
		return handlers.UploadProcess(c, false)
	})

	app.Post("/get-custom", func(c *fiber.Ctx) error {
		if _, ok := AllowedIps[handlers.GetIPAddress(c)]; !ok {
			return c.SendStatus(fiber.StatusForbidden)
		}
		return c.SendStatus(fiber.StatusForbidden)
	})

	app.Post("/download-custom", func(c *fiber.Ctx) error {
		if _, ok := AllowedIps[handlers.GetIPAddress(c)]; !ok {
			return c.SendStatus(fiber.StatusForbidden)
		}
		return c.SendStatus(fiber.StatusForbidden)
	})

	err := app.Listen(fmt.Sprintf("%s:%s", configs.Get("FIBER_HOSTNAME"), configs.Get("FIBER_PORT")))
	if err != nil {
		panic(err)
	}
}

func main() {
	c := make(chan os.Signal)
	signal.Notify(c, os.Interrupt, syscall.SIGTERM)

	go func() {
		<-c
		os.Exit(1)
	}()

	rd := database.NewRConnection()
	if err := rd.RPing(); err != nil {
		rd.RClose()
		panic(err)
	}
	rd.RClose()
	err := bot.Create()
	if err != nil {
		panic(err)
	}
	if os.Getenv("CONTAINER_TYPE") == "mixed" {
		go bot.Init()
		SetupFiber()
	} else if os.Getenv("CONTAINER_TYPE") == "bot" {
		bot.Init()
	} else {
		SetupFiber()
	}
}
