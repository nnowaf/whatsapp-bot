package authlogs

import (
	"context"
	"fmt"
	"os"
	"os/signal"
	"strings"
	"syscall"

	"github.com/golang/protobuf/proto"
	_ "github.com/mattn/go-sqlite3"
	"github.com/mdp/qrterminal"
	"go.mau.fi/whatsmeow"
	waProto "go.mau.fi/whatsmeow/binary/proto"
	"go.mau.fi/whatsmeow/store/sqlstore"
	"go.mau.fi/whatsmeow/types"
	"go.mau.fi/whatsmeow/types/events"
	waLog "go.mau.fi/whatsmeow/util/log"
)

var client *whatsmeow.Client

func myEventHandler(evt interface{}) {
	switch v := evt.(type) {
	case *events.Message:
		fmt.Println("GetConversation : ", v.Message.GetConversation())
		fmt.Println("Sender : ", v.Info.Sender)
		fmt.Println("Sender Number : ", v.Info.Sender.User)
		fmt.Println("IsGroup : ", v.Info.IsGroup)
		fmt.Println("MessageSource : ", v.Info.MessageSource)
		fmt.Println("ID : ", v.Info.ID)
		fmt.Println("PushName : ", v.Info.PushName)
		fmt.Println("BroadcastListOwner : ", v.Info.BroadcastListOwner)
		fmt.Println("Category : ", v.Info.Category)
		fmt.Println("Chat1 : ", v.Info.Chat)
		fmt.Println("Chat2 Server : ", v.Info.Chat.Server)
		fmt.Println("Chat3 User : ", v.Info.Chat.User)
		fmt.Println("DeviceSentMeta : ", v.Info.DeviceSentMeta)
		fmt.Println("IsFromMe : ", v.Info.IsFromMe)
		fmt.Println("MediaType : ", v.Info.MediaType)
		fmt.Println("Multicast : ", v.Info.Multicast)
		fmt.Println("Info.Chat.Server : ", v.Info.Chat.Server)
		// fmt.Println("livelocation : ", v.Message.LiveLocationMessage.DegreesLatitude)
		msg := &waProto.Message{
			Conversation: proto.String("Hello World WJAIODHWAIDHWAHDUIWAHDUIWAHDUIWAHDUIWA\nDWUAHDUIWAHDUIAW\nUIDHWAUIDAWH"),
		}
		sendTo, ok := parseJID(v.Info.Chat.User)
		if !ok {
			return
		}

		if v.Message.GetConversation() == "nowaf" {
			respon, err := client.SendMessage(context.Background(), sendTo, "", msg)
			if err != nil {
				fmt.Println(err.Error())
			}
			fmt.Println(respon)
		}
	case *events.Receipt:

	}
}

func parseJID(arg string) (types.JID, bool) {
	if arg[0] == '+' {
		arg = arg[1:]
	}
	if !strings.ContainsRune(arg, '@') {
		return types.NewJID(arg, types.DefaultUserServer), true
	} else {
		recipient, err := types.ParseJID(arg)
		if err != nil {
			log.Errorf("Invalid JID %s: %v", arg, err)
			return recipient, false
		} else if recipient.User == "" {
			log.Errorf("Invalid JID %s: no server specified", arg)
			return recipient, false
		}
		return recipient, true
	}
}
func LoginAuth() {

	dbLog := waLog.Stdout("Database", "DEBUG", true)
	// Make sure you add appropriate DB connector imports, e.g. github.com/mattn/go-sqlite3 for SQLite
	container, err := sqlstore.New("sqlite3", "file:wapp.db?_foreign_keys=on", dbLog)
	if err != nil {
		panic(err)
	}
	// If you want multiple sessions, remember their JIDs and use .GetDevice(jid) or .GetAllDevices() instead.
	deviceStore, err := container.GetFirstDevice()
	if err != nil {
		panic(err)
	}
	clientLog := waLog.Stdout("Client", "DEBUG", true)
	client = whatsmeow.NewClient(deviceStore, clientLog)
	client.AddEventHandler(myEventHandler)

	if client.Store.ID == nil {
		// No ID stored, new login
		qrChan, _ := client.GetQRChannel(context.Background())
		err = client.Connect()
		if err != nil {
			panic(err)
		}
		for evt := range qrChan {
			if evt.Event == "code" {
				// Render the QR code here
				// e.g. qrterminal.GenerateHalfBlock(evt.Code, qrterminal.L, os.Stdout)
				// or just manually `echo 2@... | qrencode -t ansiutf8` in a terminal
				qrterminal.GenerateHalfBlock(evt.Code, qrterminal.L, os.Stdout)
				fmt.Println("QR code:", evt.Code)
			} else {
				fmt.Println("Login event:", evt.Event)
			}
		}
	} else {
		// Already logged in, just connect
		err = client.Connect()
		if err != nil {
			panic(err)
		}
	}

	// Listen to Ctrl+C (you can also do something else that prevents the program from exiting)
	c := make(chan os.Signal)
	signal.Notify(c, os.Interrupt, syscall.SIGTERM)
	<-c

	client.Disconnect()
}
