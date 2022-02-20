# Gopintercom

### TL:DR
The purpose of this simple executable is to act as Telegram-client that listen for specific commands and play audio message sent to it on speaker

# Config 
You will need to define these ENV variables before run the executable
```sh 
# ask botfather for api-token
TELEGRAM_APITOKEN="yourtelegramapitoken"
# if set, will check chatid of sender and only respond to those specified
TELEGRAM_CHATID="yourchatid,groupchatid"
```

# Command
- `/record {duration}` : ask the client to record for {duration} seconds and then send the Audio file to us
- `/listen` : `!!!TODO!!!` - enable persistent listen mode where it will only send Audio within some certain level of threshold
- `/myid` : will return your `id` to be used on `TELEGRAM_CHATID`

# My Setup
I am using OrangePIPC+ that have internal 3.5mm Line Out and also internal Mic (albeit not that accurate)

# Some Research Point
- Using default `alsa` since this is the easiest and kind of default on a lot of SBCs
- Still looking for the best way to persistent-record-for-audio and only save if the threshold level is higher than certain value (ideally this can be triggered by GPIO from external Arduino?)