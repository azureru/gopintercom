package main

import (
    "fmt"
    tgbotapi "github.com/go-telegram-bot-api/telegram-bot-api/v5"
    cmdchain "github.com/rainu/go-command-chain"
    "io"
    "net/http"
    "os"
    "os/exec"
    "strconv"
    "strings"
    "sync"
)

func main() {
    bot, err := tgbotapi.NewBotAPI(os.Getenv("TELEGRAM_APITOKEN"))
    if err != nil {
        fmt.Println("Fill TELEGRAM_APITOKEN with telegram token")
        panic(err)
    }
    chatIds := os.Getenv("TELEGRAM_CHATID")
    ids := strings.Split(chatIds, ",")
    var idnt []int64
    for _, id := range ids {
        idi, _ := strconv.ParseInt(id, 10, 64)
        idnt = append(idnt, idi)
    }
    if os.Getenv("VERBOSE") != ""  {
        bot.Debug = true
    }

    // Create a new UpdateConfig struct with an offset of 0. Offsets are used
    // to make sure Telegram knows we've handled previous values and we don't
    // need them repeated.
    updateConfig := tgbotapi.NewUpdate(0)

    // Tell Telegram we should wait up to 30 seconds on each request for an
    // update. This way we can get information just as quickly as making many
    // frequent requests without having to send nearly as many.
    updateConfig.Timeout = 30

    // Start polling Telegram for updates.
    updates := bot.GetUpdatesChan(updateConfig)

    // Let's go through each update that we're getting from Telegram.
    for update := range updates {
        // Telegram can send many types of up dates depending on what your Bot
        // is up to. We only want to look at messages for now, so we can
        // discard any other updates.
        if update.Message == nil {
            continue
        }

        // chatIds ENV config is defined - ignore users not listed in the list
        if chatIds != "" {
            found := false
            for i := 0; i < len(idnt) ; i++ {
                if idnt[i] == update.Message.Chat.ID {
                    found = true
                }
            }
            if !found {
                continue
            }
        }

        var msg tgbotapi.MessageConfig
        msg = tgbotapi.NewMessage(update.Message.Chat.ID, "")
        msg.ReplyToMessageID = update.Message.MessageID

        // specific command handler
        if update.Message.IsCommand() {
           cmd := update.Message.Command()
           fmt.Println("Receive command " + cmd)
           if cmd == "record" {
               // reply that we are attempting to record
               msg.Text = "Recording..."
               args := update.Message.CommandArguments()
               if args == "" {
                   args = "5"
               }
               if err := Spawn("arecord", []string{"arecord","-d",args,"record.wav"}); err != nil {
                   fmt.Println("Failed to record", err.Error())
               }
               audio := tgbotapi.NewAudio(msg.ChatID, tgbotapi.FilePath("./record.wav"))
               mg := tgbotapi.NewMediaGroup(msg.ChatID, []interface{}{
                   audio,
               })
               if _, err := bot.SendMediaGroup(mg); err != nil {
                   fmt.Println("Failed to send media-group", err.Error())
               }
           } else if cmd == "myid" {
               // return current chatid
               msg.Text = fmt.Sprintf("Your chat.ID is %d", update.Message.Chat.ID)
           }
        } else if update.Message.Audio != nil {
            url, err := bot.GetFileDirectURL(update.Message.Audio.FileID)
            if err != nil {
                fmt.Println("Error on getting Audio file" + update.Message.Audio.FileID)
            }
            // download the file and save to temp - then play it
            msg.Text = "Played"
            if err := DownloadFile(url, "./audio.ogg"); err != nil {
                msg.Text = "Failed to Download" + url +":" + err.Error()
            }
            // spawn process to play
            if err := PlayAudio("./audio.ogg"); err != nil {
                msg.Text = "Failed to play " + ":" + err.Error()
            }
        } else if update.Message.Voice != nil {
            url, err := bot.GetFileDirectURL(update.Message.Voice.FileID)
            if err != nil {
                fmt.Println("Error on getting Voice file" + update.Message.Voice.FileID)
            }
            // download the file and save to temp - then play it
            msg.Text = "Played"
            if err := DownloadFile(url, "./voice.ogg"); err != nil {
                msg.Text = "Failed to Download" + url +":" + err.Error()
            }
            // spawn process to play
            if err := PlayAudio("./voice.ogg"); err != nil {
                msg.Text = "Failed to play " + ":" + err.Error()
            }
        } else {
            continue
        }

        if _, err := bot.Send(msg); err != nil {
            // just panic here - let systemd restart whole service
            panic(err)
        }
        fmt.Println(msg.Text)
    }

    // infinite wait until terminated
    wg := sync.WaitGroup{}
    wg.Add(1)
    wg.Wait()
}

// DownloadFile will download a url to a local file.
func DownloadFile(url string, targetPath string) error {

	// Get the data
	resp, err := http.Get(url)
	if err != nil {
		return err
	}
	defer resp.Body.Close()

	// Create the file
	out, err := os.Create(targetPath)
	if err != nil {
		return err
	}
	defer out.Close()

	// Write the body to file
	_, err = io.Copy(out, resp.Body)
	return err
}

// Spawn - simply spawn a process and wait
func Spawn(executable string, params []string) error {
    cmd := exec.Command(executable, params...)
    return cmd.Run()
}

func PlayAudio(file string) error {
 return cmdchain.Builder().
        Join("opusdec", "--force-wav", "--quiet", file, "-").
        Join("aplay", "-q").
        Finalize().
        Run()
}
