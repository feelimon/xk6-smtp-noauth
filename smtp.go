package smtp

import (
	"encoding/base64"
	"fmt"
	"net/smtp"
	"os"
	"strings"
	"time"

	"github.com/google/uuid"
	"go.k6.io/k6/v2/js/modules"
)

func init() {
	modules.Register("k6/x/smtp", new(SMTP))
}

type SMTP struct{}

type options struct {
	Subject    string   `js:"subject"`
	Message    string   `js:"message"`
	UDW        []string `js:"udw"`
	Attachment string   `js:"attachment"`
}

func check(e error) {
	if e != nil {
		fmt.Println(e)
	}
}

func (*SMTP) SendMail(host string, port string, sender string, recipient string, senderDomain string, options options) {
	emailMessage := "From: " + sender + "\r\n" + "To: " + recipient + "\r\n"

	if options.Subject != "" {
		emailMessage += "Subject: " + options.Subject + "\r\n"
	}

	emailMessage += "Message-ID: <" + uuid.New().String() + "@" + senderDomain + ">\r\n"
	emailMessage += "Date: " + time.Now().Format(time.RFC1123Z) + "\r\n"

	boundary := "BOUNDARY12345"
	emailMessage += "MIME-Version: 1.0\r\n"
	emailMessage += "Content-Type: multipart/mixed; boundary=" + boundary + "\r\n\r\n"

	emailMessage += "--" + boundary + "\r\n"
	emailMessage += "Content-Type: text/plain; charset=\"utf-8\"\r\n\r\n"
	emailMessage += options.Message + "\r\n\r\n"

	if options.Attachment != "" {
		fileData, err := os.ReadFile(options.Attachment)
		check(err)

		emailMessage += "--" + boundary + "\r\n"
		emailMessage += "Content-Type: application/octet-stream\r\n"
		emailMessage += "Content-Transfer-Encoding: base64\r\n"
		emailMessage += fmt.Sprintf("Content-Disposition: attachment; filename=\"%s\"\r\n\r\n",
			getFileName(options.Attachment))

		b64 := make([]byte, base64.StdEncoding.EncodedLen(len(fileData)))
		base64.StdEncoding.Encode(b64, fileData)
		emailMessage += string(b64) + "\r\n\r\n"
	}

	emailMessage += "--" + boundary + "--"

	if len(options.UDW) == 0 {
		options.UDW = []string{recipient}
	}

	body := []byte(emailMessage)
	err := smtp.SendMail(host+":"+port, nil, sender, options.UDW, body)
	check(err)
}

func getFileName(path string) string {
	parts := strings.Split(path, string(os.PathSeparator))
	return parts[len(parts)-1]
}
