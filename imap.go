package main

import (
	"fmt"
	"log"
	"os"

	"github.com/emersion/go-imap/v2"
)

func main() {
	imapServer := os.Getenv("IMAP_SERVER")
	username := os.Getenv("IMAP_USERNAME")
	password := os.Getenv("IMAP_PASSWORD")

	c, err := getClient(username, password, imapServer)
	if err != nil {
		log.Fatalf("failed to dial IMAP server: %v", err)
	}

	defer c.Close()

	selectMailBox(c, "INBOX")

	uids, err := search(c, &imap.SearchCriteria{
		NotFlag: []imap.Flag{"\\Seen"},
	})
	if err != nil {
		log.Fatalf("failed to select INBOX: %v", err)
	}

	slice := uids[0:4]

	iterateOverMessages(c, slice, func(uid imap.UID, subject, message string) error {
		fmt.Printf("UID: %v, Subject: %s\n", uid, subject)
		return nil
	})

	if err := c.Logout().Wait(); err != nil {
		log.Fatalf("failed to logout: %v", err)
	}
}
