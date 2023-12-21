package main

import (
	"log"
	"os"

	"github.com/google/uuid"
	"github.com/labstack/echo/v4"
	"github.com/labstack/echo/v4/middleware"

	"github.com/tebeka/selenium"
)

type SeleniumSession struct {
	ID   string `json:"id"`
	Name string `json:"name"`
}

type SeleniumNavigation struct {
	URL        string `json:"url"`
	PageSource string `json:"page_source"`
}

type SeleniumScreenshot struct {
	ImageURL string `json:"image_url"`
}

type SeleniumPageSource struct {
	PageSource string `json:"page_source"`
}

type SeleniumElementSelection struct {
	By    string `json:"by"`
	Value string `json:"value"`
}

var SeleniumSessionList []SeleniumSession

func main() {
	// Echo instance
	e := echo.New()

	// Middleware
	e.Use(middleware.Logger())
	e.Use(middleware.Recover())

	// init
	SeleniumSessionList = []SeleniumSession{}
	service, err := selenium.NewChromeDriverService("/usr/bin/chromedriver", 4444)
	if err != nil {
		panic(err)
	}
	defer service.Stop()

	// configure the browser options
	caps := selenium.Capabilities{}
	// caps.AddChrome(chrome.Capabilities{Args: []string{
	// 	"--headless-new", // comment out this line for testing
	// }})
	driver, err := selenium.NewRemote(caps, "")
	if err != nil {
		log.Fatal("Error:", err)
	}

	// Route => handler
	e.GET("/session", func(c echo.Context) error {
		return c.JSON(200, SeleniumSessionList)
	})
	e.POST("/session", func(c echo.Context) error {
		s := new(SeleniumSession)
		if err := c.Bind(s); err != nil {
			return err
		}

		// create a new session
		id, err := driver.NewSession()
		if err != nil {
			log.Print("Error:", err)
			return echo.NewHTTPError(500, err)
		}
		s.ID = id

		SeleniumSessionList = append(SeleniumSessionList, *s)
		return c.JSON(200, s)
	})

	e.GET("/session/:id", func(c echo.Context) error {
		id := c.Param("id")
		for _, s := range SeleniumSessionList {
			if s.ID == id {
				return c.JSON(200, s)
			}
		}

		return echo.NewHTTPError(404, "Not Found")
	})

	e.DELETE("/session/:id", func(c echo.Context) error {
		id := c.Param("id")
		for i, s := range SeleniumSessionList {
			if s.ID == id {
				driver.SwitchSession(s.ID)
				driver.Quit()

				SeleniumSessionList = append(SeleniumSessionList[:i], SeleniumSessionList[i+1:]...)
				return c.JSON(200, s)
			}
		}

		return echo.NewHTTPError(404, "Not Found")
	})

	e.GET("/navigation/:id", func(c echo.Context) error {
		var sess SeleniumSession
		id := c.Param("id")
		for _, s := range SeleniumSessionList {
			if s.ID == id {
				sess = s
			}
		}

		if sess.ID == "" {
			return echo.NewHTTPError(404, "Not Found")
		}

		if err := driver.SwitchSession(sess.ID); err != nil {
			log.Print("Error:", err)
			return echo.NewHTTPError(500, err)
		}

		// get current url
		url, err := driver.CurrentURL()
		if err != nil {
			log.Print("Error:", err)
			return echo.NewHTTPError(500, err)
		}

		r := new(SeleniumNavigation)
		r.URL = url

		// get page source
		pageSource, err := driver.PageSource()
		if err != nil {
			log.Print("Error:", err)
			return echo.NewHTTPError(500, err)
		}

		r.PageSource = pageSource

		return c.JSON(200, r)
	})

	e.POST("/navigation/:id/to", func(c echo.Context) error {
		var sess SeleniumSession

		id := c.Param("id")
		for _, s := range SeleniumSessionList {
			if s.ID == id {
				sess = s
			}
		}

		if sess.ID == "" {
			return echo.NewHTTPError(404, "Not Found")
		}

		if err := driver.SwitchSession(sess.ID); err != nil {
			log.Print("Error:", err)
			return echo.NewHTTPError(500, err)
		}

		n := new(SeleniumNavigation)
		if err := c.Bind(n); err != nil {
			return err
		}

		if err := driver.Get(n.URL); err != nil {
			log.Print("Error:", err)
			return echo.NewHTTPError(500, err)
		}

		// get page source
		pageSource, err := driver.PageSource()
		if err != nil {
			log.Print("Error:", err)
			return echo.NewHTTPError(500, err)
		}

		n.PageSource = pageSource
		return c.JSON(200, n)
	})

	e.POST("/navigation/:id/back", func(c echo.Context) error {
		var sess SeleniumSession
		id := c.Param("id")
		for _, s := range SeleniumSessionList {
			if s.ID == id {
				sess = s
			}
		}

		if sess.ID == "" {
			return echo.NewHTTPError(404, "Not Found")
		}

		if err := driver.SwitchSession(sess.ID); err != nil {
			log.Print("Error:", err)
			return echo.NewHTTPError(500, err)
		}

		if err := driver.Back(); err != nil {
			log.Print("Error:", err)
			return echo.NewHTTPError(500, err)
		}

		// get current url
		url, err := driver.CurrentURL()
		if err != nil {
			log.Print("Error:", err)
			return echo.NewHTTPError(500, err)
		}

		r := new(SeleniumNavigation)
		r.URL = url
		return c.JSON(200, r)
	})

	e.GET("/document/:id/screenshot", func(c echo.Context) error {
		var sess SeleniumSession
		id := c.Param("id")
		for _, s := range SeleniumSessionList {
			if s.ID == id {
				sess = s
			}
		}

		if sess.ID == "" {
			return echo.NewHTTPError(404, "Not Found")
		}

		if err := driver.SwitchSession(sess.ID); err != nil {
			log.Print("Error:", err)
			return echo.NewHTTPError(500, err)
		}

		// take a screenshot
		screenshot, err := driver.Screenshot()
		if err != nil {
			log.Print("Error:", err)
			return echo.NewHTTPError(500, err)
		}

		// save to result/screenshots
		imgId := genUUID()
		err = os.WriteFile("result/screenshots/"+imgId+".png", screenshot, 0644)
		if err != nil {
			log.Print("Error:", err)
			return echo.NewHTTPError(500, err)
		}

		r := new(SeleniumScreenshot)
		r.ImageURL = "https://2ad566c4516e-6989911764160887308.ngrok-free.app/screenshots/" + imgId
		return c.JSON(200, r)
	})

	e.GET("/document/:id/page_source", func(c echo.Context) error {
		var sess SeleniumSession
		id := c.Param("id")
		for _, s := range SeleniumSessionList {
			if s.ID == id {
				sess = s
			}
		}

		if sess.ID == "" {
			return echo.NewHTTPError(404, "Not Found")
		}

		if err := driver.SwitchSession(sess.ID); err != nil {
			log.Print("Error:", err)
			return echo.NewHTTPError(500, err)
		}

		// get page source
		pageSource, err := driver.PageSource()
		if err != nil {
			log.Print("Error:", err)
			return echo.NewHTTPError(500, err)
		}

		r := new(SeleniumPageSource)
		r.PageSource = pageSource
		return c.JSON(200, r)
	})

	e.GET("/screenshots/:id", func(c echo.Context) error {
		id := c.Param("id")
		return c.File("result/screenshots/" + id + ".png")
	})

	e.POST("/element/:id/click", func(c echo.Context) error {
		var sess SeleniumSession
		id := c.Param("id")
		for _, s := range SeleniumSessionList {
			if s.ID == id {
				sess = s
			}
		}

		if sess.ID == "" {
			return echo.NewHTTPError(404, "Not Found")
		}

		if err := driver.SwitchSession(sess.ID); err != nil {
			log.Print("Error:", err)
			return echo.NewHTTPError(500, err)
		}

		e := new(SeleniumElementSelection)
		if err := c.Bind(e); err != nil {
			return err
		}

		var selectBy string
		switch e.By {
		case "id":
			selectBy = selenium.ByID
		case "xpath":
			selectBy = selenium.ByXPATH
		case "link_text":
			selectBy = selenium.ByLinkText
		case "partial_link_text":
			selectBy = selenium.ByPartialLinkText
		case "tag_name":
			selectBy = selenium.ByTagName
		case "class_name":
			selectBy = selenium.ByClassName
		case "css_selector":
			selectBy = selenium.ByCSSSelector
		default:
			selectBy = selenium.ByID
		}

		log.Print("selectBy:", selectBy)
		log.Print("value:", e.Value)

		elem, err := driver.FindElement(selectBy, e.Value)
		if err != nil {
			log.Print("Error:", err)
			return echo.NewHTTPError(500, err)
		}

		if err := elem.Click(); err != nil {
			log.Print("Error:", err)
			return echo.NewHTTPError(500, err)
		}

		return c.NoContent(204)
	})

	// Start server
	e.Logger.Fatal(e.Start(":18080"))
}

func genUUID() string {
	uuid := uuid.New()
	return uuid.String()
}
