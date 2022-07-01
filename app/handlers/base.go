package handlers

import (
	"encoding/json"
	"fmt"
	"github.com/Xacnio/ytmp3-host-go/app/bot"
	"github.com/Xacnio/ytmp3-host-go/app/models"
	"github.com/Xacnio/ytmp3-host-go/app/utils"
	"github.com/Xacnio/ytmp3-host-go/pkg/configs"
	"github.com/Xacnio/ytmp3-host-go/platform/database"
	"github.com/gofiber/fiber/v2"
	"github.com/gofiber/fiber/v2/middleware/proxy"
	"golang.org/x/text/cases"
	"golang.org/x/text/language"
	"os/exec"
	"regexp"
	"strings"
	"time"
)

func ProxyHandler(c *fiber.Ctx) error {
	filename := c.Params("filename")
	rd := database.NewRConnection()
	defer rd.RClose()

	data, err := rd.RGet(filename)
	if err != nil {
		return c.SendStatus(fiber.StatusInternalServerError)
	}
	if data == "" {
		return c.SendStatus(fiber.StatusNotFound)
	}
	dataS := models.Music{}
	err2 := json.Unmarshal([]byte(data), &dataS)
	if err2 != nil {
		return c.SendStatus(fiber.StatusInternalServerError)
	}

	url := fmt.Sprintf("https://api.telegram.org/file/bot%s/%s", configs.Get("TG_BOT_TOKEN"), dataS.FilePath)
	if err := proxy.Do(c, url); err != nil {
		return err
	}
	// Telegram file paths is expiring after a while, renew file path with fileID
	if string(c.Response().Header.ContentType()) == "application/json" {
		foundAgain := false
		result := new(struct {
			Ok          bool   `json:"ok"`
			Error       uint16 `json:"error_code"`
			Description string `json:"description"`
		})
		if err := json.Unmarshal(c.Response().Body(), result); err == nil {
			if !result.Ok && result.Error == 404 {
				file, err2 := bot.Bot.FileByID(dataS.FileID)
				if err2 == nil {
					foundAgain = true
					dataS.FilePath = file.FilePath
					jsonData, _ := json.Marshal(dataS)
					e := rd.RSet(strings.Split(filename, ".")[0], string(jsonData))
					if e == nil {
						url := fmt.Sprintf("https://api.telegram.org/file/bot%s/%s", configs.Get("TG_BOT_TOKEN"), file.FilePath)
						if err := proxy.Do(c, url); err != nil {
							return err
						}
					}
				} else {
					return c.SendStatus(fiber.StatusNotFound)
				}
			}
		}
		if !foundAgain {
			return c.SendStatus(fiber.StatusNotFound)
		}
	}
	// Remove Server headers from response
	c.Response().Header.Del(fiber.HeaderServer)
	c.Response().Header.Del(fiber.HeaderContentDisposition)
	c.Response().Header.Set("Content-type", "audio/mpeg")
	c.Set("Content-type", "audio/mpeg")
	return nil
}

func GetMusic(c *fiber.Ctx) error {
	c.Set("Content-type", "text/plain; charset=windows-1254")
	rdb := database.NewRConnection()
	defer rdb.RClose()
	payload := struct {
		URL string `form:"url"`
	}{}
	if err := c.BodyParser(&payload); err == nil {
		if !strings.HasPrefix(payload.URL, "http") {
			r, _ := regexp.Compile("([a-z]+):([a-zA-Z0-9_-]+)")
			if r.MatchString(payload.URL) {
				fmt.Println("matched")
				matches := r.FindStringSubmatch(payload.URL)
				search := fmt.Sprintf("%s_%s", matches[2], matches[1])
				fmt.Println(search)
				xx, _ := rdb.RGetSearchOne(search)
				fmt.Println("aa", xx)
				if xx != "" {
					dataS := models.Music{}
					json.Unmarshal([]byte(xx), &dataS)
					c.Set("Content-type", "text/plain; charset=windows-1254")
					t, _ := utils.ToISO88599(dataS.Title)
					return c.SendString(fmt.Sprintf("%s/%s*%s*%s", configs.Get("FIBER_URL"), dataS.Filename, t, cases.Title(language.English).String(dataS.Extractor)))
				}
			}
			payload.URL = "ytsearch:" + payload.URL
		}
		cachedKey, _ := rdb.RGet(payload.URL)
		if len(cachedKey) > 2 {
			dataSz, _ := rdb.RGet(cachedKey)
			if len(dataSz) > 2 {
				dataS := models.Music{}
				json.Unmarshal([]byte(dataSz), &dataS)
				c.Set("Content-type", "text/plain; charset=windows-1254")
				t, _ := utils.ToISO88599(dataS.Title)
				return c.SendString(fmt.Sprintf("%s/%s*%s*%s", configs.Get("FIBER_URL"), dataS.Filename, t, cases.Title(language.English).String(dataS.Extractor)))
			}
		} else if cachedKey == "no" {
			return c.SendString("yok")
		}
		out, _ := exec.Command("yt-dlp", "-j", payload.URL).Output()
		videoX := struct {
			ID        string `json:"id"`
			Title     string `json:"title"`
			Filename  string `json:"filename"`
			Extractor string `json:"extractor"`
		}{}
		json.Unmarshal(out, &videoX)
		if videoX.ID == "" {
			rdb.RSetTTL(payload.URL, "no", time.Hour*24*10)
			return c.SendString("yok")
		}
		videoX.Filename = strings.ReplaceAll(videoX.Title, " ", "_")
		videoX.Filename = utils.TurkishToEnglish(videoX.Filename)
		sampleRegexp := regexp.MustCompile(`[^a-zA-Z0-9-_]`)
		videoX.Filename = string(sampleRegexp.ReplaceAllString(videoX.Filename, ""))
		if len(videoX.Filename) > 64 {
			videoX.Filename = videoX.Filename[0:55]
		}
		videoX.Filename = fmt.Sprintf("%s_%s_%s.mp3", videoX.Filename, videoX.ID, videoX.Extractor)
		dataS, _ := rdb.RGet(videoX.Filename)
		if len(dataS) > 2 {
			c.Set("Content-type", "text/plain; charset=windows-1254")
			t, _ := utils.ToISO88599(videoX.Title)
			return c.SendString(fmt.Sprintf("%s/%s*%s*%s", configs.Get("FIBER_URL"), videoX.Filename, t, cases.Title(language.English).String(videoX.Extractor)))
		} else {
			return c.SendString("yok")
		}
	}
	return c.SendString("yok")
}

func HttpsForwarder(c *fiber.Ctx) error {
	https := c.Get(fiber.HeaderXForwardedProto)
	if https == "http" && strings.Contains(c.Request().URI().String(), ".mp3") {
		return c.Redirect(strings.Replace(c.Request().URI().String(), "http://", "https://", 1))
	}
	return c.Next()
}
