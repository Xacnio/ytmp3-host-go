package bot

import (
	"crypto/md5"
	"encoding/hex"
	"encoding/json"
	"fmt"
	"github.com/Xacnio/ytmp3-host-go/app/models"
	"github.com/Xacnio/ytmp3-host-go/app/utils"
	"github.com/Xacnio/ytmp3-host-go/pkg/configs"
	"github.com/Xacnio/ytmp3-host-go/platform/database"
	tele "gopkg.in/telebot.v3"
	"regexp"
	"strings"
	"time"
)

var (
	Bot *tele.Bot
)

func Create() error {
	var err error
	Bot, err = tele.NewBot(tele.Settings{
		Token:  configs.Get("TG_BOT_TOKEN"),
		Poller: &tele.LongPoller{Timeout: 10 * time.Second},
	})
	return err
}
func Init() {
	Bot.Handle("/upload", func(c tele.Context) error {
		rdb := database.NewRConnection()
		defer rdb.RClose()
		cmsg := c.Message()
		if cmsg.ReplyTo != nil && cmsg.ReplyTo.Audio != nil {
			var audio *tele.Audio
			audio = cmsg.ReplyTo.Audio
			if audio.FileSize > (1024 * 1024 * 20) {
				return c.Reply("`MP3 couldn't be saved! Maximum file size: 20 MB!`", &tele.SendOptions{ParseMode: tele.ModeMarkdown})
			}
			filename := audio.FileName
			title := audio.FileName
			if audio.Title != "" {
				title = audio.Title
				if audio.Performer != "" {
					title = fmt.Sprintf("%s - %s", audio.Title, audio.Performer)
				}
				filename = strings.ReplaceAll(title, " ", "_")
				filename = utils.TurkishToEnglish(filename)
				sampleRegexp := regexp.MustCompile(`[^a-zA-Z0-9-_]`)
				filename = string(sampleRegexp.ReplaceAllString(filename, ""))
			}
			if len(filename) > 64 {
				filename = filename[0:55]
			}
			title = strings.ReplaceAll(title, "*", "")
			hash := md5.Sum([]byte(fmt.Sprintf("%s:%s", filename, audio.FileID)))
			id := hex.EncodeToString(hash[:])
			filename2 := fmt.Sprintf("%s_%s_%s.mp3", filename, id[0:10], "telegram")
			bs, _ := rdb.RGet(filename2)
			if len(bs) > 3 {
				dataS := models.Music{}
				json.Unmarshal([]byte(bs), &dataS)
				url := fmt.Sprintf("%s/%s", configs.Get("FIBER_URL"), filename2)
				message := fmt.Sprintf("MP3 saved: %s\r\n\r\ntelegram:%s", url, id[0:10])
				return c.Reply(message, &tele.SendOptions{DisableWebPagePreview: true})
			}
			url := fmt.Sprintf("%s/%s", configs.Get("FIBER_URL"), filename2)
			cmsg.ReplyTo.Audio.Caption = url
			a := &tele.Audio{}
			*a = *audio
			a.Caption = string(url)
			msg, err := c.Bot().Send(&tele.Chat{ID: configs.GetInt64("TG_BOT_CHANNEL_ID")}, a, &tele.SendOptions{ParseMode: tele.ModeHTML})
			if err != nil {
				fmt.Println(err)
				return c.Reply("`Error (2)`", &tele.SendOptions{ParseMode: tele.ModeMarkdown})
			} else {
				keyT := strings.Split(msg.Caption, "/")
				if len(keyT) != 4 {
					return c.Reply("`Error (3)`", &tele.SendOptions{ParseMode: tele.ModeMarkdown})
				}
				key := keyT[3]
				data := models.Music{
					IpAddress:  fmt.Sprintf("%d:%s:%s:%s", cmsg.Sender.ID, cmsg.Sender.Username, cmsg.Sender.FirstName, cmsg.Sender.LastName),
					UploadDate: time.Now().Unix(),
					FileID:     msg.Audio.FileID,
					FilePath:   msg.Audio.FilePath,
					FileSize:   msg.Audio.FileSize,
					Type:       msg.Audio.MIME,
					Title:      title,
					Filename:   filename2,
					Extractor:  "telegram",
				}
				jsonData, _ := json.Marshal(data)
				errQ := rdb.RSet(key, string(jsonData))
				if errQ != nil {
					return c.Reply("`Error (4)`", &tele.SendOptions{ParseMode: tele.ModeMarkdown})
				}
				message := fmt.Sprintf("MP3 saved: %s\r\n\r\ntelegram:%s", url, id[0:10])
				return c.Reply(message, &tele.SendOptions{DisableWebPagePreview: true})
			}
		} else {
			return c.Reply("`Reply to a music!`", &tele.SendOptions{ParseMode: tele.ModeMarkdown})
		}
		return nil
	})
	Bot.Start()
}
