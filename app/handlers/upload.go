package handlers

import (
	"bufio"
	"encoding/json"
	"fmt"
	"github.com/Xacnio/ytmp3-host-go/app/bot"
	"github.com/Xacnio/ytmp3-host-go/app/models"
	"github.com/Xacnio/ytmp3-host-go/app/utils"
	"github.com/Xacnio/ytmp3-host-go/pkg/configs"
	"github.com/Xacnio/ytmp3-host-go/platform/database"
	"github.com/gofiber/fiber/v2"
	"golang.org/x/text/cases"
	"golang.org/x/text/language"
	tele "gopkg.in/telebot.v3"
	"os"
	"os/exec"
	"regexp"
	"strings"
	"time"

	_ "github.com/paulrosania/go-charset/data"
)

func GetIPAddress(c *fiber.Ctx) string {
	IPAddress := c.Get("X-Real-Ip")
	if IPAddress == "" {
		IPAddress = c.Get("X-Forwarded-For")
	}
	if IPAddress == "" {
		IPAddress = c.IP()
	}
	return IPAddress
}

func deleteFile(filename string) {
	os.Remove(fmt.Sprintf("%s.mp3", filename))
	os.Remove(fmt.Sprintf("%s.mp3.jpg", filename))
	os.Remove(fmt.Sprintf("%s.mp3.webp", filename))
	os.Remove(fmt.Sprintf("%s.jpg", filename))
}

func UploadProcess(c *fiber.Ctx, export bool) error {
	rdb := database.NewRConnection()
	defer rdb.RClose()
	payload := struct {
		URL string `form:"url"`
	}{}
	if err := c.BodyParser(&payload); err == nil {
		if !strings.HasPrefix(payload.URL, "http") {
			payload.URL = "ytsearch:" + payload.URL
		}
		cachedKey, _ := rdb.RGet(payload.URL)
		if len(cachedKey) > 3 {
			dataSz, _ := rdb.RGet(cachedKey)
			if len(dataSz) > 2 {
				dataS := models.Music{}
				json.Unmarshal([]byte(dataSz), &dataS)
				if export {
					t := dataS.Title
					return c.SendString(fmt.Sprintf("%s/%s*%s*%s", configs.Get("FIBER_URL"), dataS.Filename, t, cases.Title(language.English).String(dataS.Extractor)))
				} else {
					return c.Redirect(fmt.Sprintf("/%s", dataS.Filename))
				}
			}
		} else if cachedKey == "no" {
			return c.SendString("Error: Invalid download.")
		}
		cmd := exec.Command("yt-dlp", "-j", payload.URL)
		out, _ := cmd.Output()
		videoX := struct {
			ID         string              `json:"id"`
			Title      string              `json:"title"`
			Uploader   string              `json:"uploader"`
			Filename   string              `json:"filename"`
			Thumbnails []map[string]string `json:"thumbnails"`
			Duration   int64               `json:"duration"`
			Extractor  string              `json:"extractor"`
			Artist     string              `json:"artist"`
			Track      string              `json:"track"`
		}{}
		json.Unmarshal(out, &videoX)
		if videoX.ID == "" {
			rdb.RSetTTL(payload.URL, "no", time.Hour*24*10)
			return c.SendString("Error: Invalid download.")
		}
		if videoX.Uploader != "" {
			videoX.Title = videoX.Title + " - " + videoX.Uploader
		}
		if videoX.Artist != "" && videoX.Track != "" && videoX.Artist != "Çeşitli Sanatçılar" && videoX.Artist != "Various Artists" {
			videoX.Title = videoX.Artist + " - " + videoX.Track
		}
		videoX.Filename = strings.ReplaceAll(videoX.Title, " ", "_")
		videoX.Filename = utils.TurkishToEnglish(videoX.Filename)
		sampleRegexp := regexp.MustCompile(`[^a-zA-Z\d-_]`)
		videoX.Filename = sampleRegexp.ReplaceAllString(videoX.Filename, "")
		if len(videoX.Filename) > 64 {
			videoX.Filename = videoX.Filename[0:55]
		}
		filename := fmt.Sprintf("web/data/%s_%s_%s", videoX.Filename, videoX.ID, videoX.Extractor)
		videoX.Filename = fmt.Sprintf("%s_%s_%s.mp3", videoX.Filename, videoX.ID, videoX.Extractor)
		dataS, _ := rdb.RGet(videoX.Filename)
		videoX.Title = strings.ReplaceAll(videoX.Title, "*", "")
		if len(dataS) > 2 {
			if export {
				t := videoX.Title
				return c.SendString(fmt.Sprintf("%s/%s*%s*%s", configs.Get("FIBER_URL"), videoX.Filename, t, strings.ToTitle(videoX.Extractor)))
			} else {
				return c.Redirect(fmt.Sprintf("/%s", videoX.Filename))
			}
		}
		if videoX.Duration > 60*10 {
			return c.SendString("Error: Maximum 10-minute videos are accepting!")
		}
		output := fmt.Sprintf("web/data/%s", videoX.Filename)
		fmt.Println(exec.Command("yt-dlp", "-i", "--extract-audio", "-x", "--audio-format mp3", "--audio-quality 0", payload.URL).String())
		exec.Command("yt-dlp", "-i", "--add-metadata", "--write-thumbnail", "--embed-thumbnail", "--extract-audio", "--audio-format", "mp3", "--audio-quality", "0", "--output", output, payload.URL).Output()
		if _, err := os.Stat(output); err == nil {
			fData, err := os.Open(output)
			if err != nil {
				deleteFile(filename)
				return c.SendString("Error: Download failed!")
			}
			reader := bufio.NewReader(fData)
			url := fmt.Sprintf("%s/%s", configs.Get("FIBER_URL"), videoX.Filename)
			document := &tele.Audio{File: tele.FromReader(reader), Caption: url, FileName: videoX.Filename}
			result, err := bot.Bot.Send(&tele.Chat{ID: configs.GetInt64("TG_BOT_CHANNEL_ID")}, document, &tele.SendOptions{ParseMode: tele.ModeHTML})
			if err != nil {
				fmt.Println(err)
				fData.Close()
				deleteFile(filename)
				return c.SendString("Error: File save failed!")
			}
			if result.Audio == nil {
				fData.Close()
				deleteFile(filename)
				return c.SendString("Error: File save failed!")
			} else {
				keyT := strings.Split(result.Caption, "/")
				if len(keyT) != 4 {
					fData.Close()
					deleteFile(filename)
					return c.SendString("Error: File save failed!")
				}
				key := keyT[3]
				filef, errf := bot.Bot.FileByID(result.Audio.FileID)
				if errf != nil {
					fData.Close()
					deleteFile(filename)
					return c.SendString("Error: File save failed!")
				}
				data := models.Music{
					UploadDate: time.Now().Unix(),
					FileID:     result.Audio.FileID,
					FilePath:   filef.FilePath,
					FileSize:   result.Audio.FileSize,
					IpAddress:  GetIPAddress(c),
					Type:       result.Audio.MIME,
					Title:      videoX.Title,
					Filename:   videoX.Filename,
					Extractor:  videoX.Extractor,
				}
				jsonData, _ := json.Marshal(data)
				errQ := rdb.RSet(key, string(jsonData))
				if errQ != nil {
					fData.Close()
					deleteFile(filename)
					return c.SendString("Error: File save failed!")
				}
				fData.Close()
				rdb.RSetTTL(payload.URL, key, time.Hour*24*10)
				deleteFile(filename)
				if export {
					t := videoX.Title
					return c.SendString(fmt.Sprintf("%s/%s*%s*%s", configs.Get("FIBER_URL"), videoX.Filename, t, cases.Title(language.English).String(videoX.Extractor)))
				} else {
					return c.Redirect(fmt.Sprintf("/%s", videoX.Filename))
				}
			}
		} else {
			return c.SendString("Error: Download failed!")
		}
	}
	return c.SendString("Error: Invalid request!")
}
