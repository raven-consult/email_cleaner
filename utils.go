package main

import (
	"fmt"
	"io"
	"log"

	"github.com/emersion/go-imap/v2"
	"github.com/emersion/go-imap/v2/imapclient"
)

func getClient(email, password, server string) (*imapclient.Client, error) {
	c, err := imapclient.DialTLS(server, nil)
	if err != nil {
		log.Fatalf("failed to dial IMAP server: %v", err)
		return nil, err
	}

	if err := c.Login(email, password).Wait(); err != nil {
		log.Fatalf("failed to login: %v", err)
		return nil, err
	}

	return c, nil
}

// func getMailBoxes(c *imapclient.Client) ([]*imap.ListData, error) {
// 	mailboxes, err := c.List("", "%", nil).Collect()
// 	if err != nil {
// 		log.Fatalf("failed to list mailboxes: %v", err)
// 		return nil, err
// 	}
// 	log.Printf("Found %v mailboxes", len(mailboxes))
// 	return mailboxes, nil
// 	// for _, mbox := range mailboxes {
// 	// 	log.Printf(" - %v", mbox.Mailbox)
// 	// }
// }

func search(c *imapclient.Client, query *imap.SearchCriteria) ([]imap.UID, error) {
	data, err := c.UIDSearch(query, nil).Wait()
	if err != nil {
		log.Fatalf("UID SEARCH command failed: %v", err)
		return nil, err
	}
	// log.Fatalf("UIDs matching the search criteria: %v", data.AllUIDs())
	return data.AllUIDs(), nil
}

func selectMailBox(c *imapclient.Client, mbox string) {
	selectedMbox, err := c.Select(mbox, nil).Wait()
	if err != nil {
		log.Fatalf("failed to select %s: %v", mbox, err)
	}
	log.Printf("%s contains %v messages", mbox, selectedMbox.NumMessages)
}

type OnMessage func(uid imap.UID, subject string, message string) error

func iterateOverMessages(c *imapclient.Client, uids []imap.UID, onMessage OnMessage) error {
	seqset := imap.UIDSetNum(uids...)
	fetchOptions := &imap.FetchOptions{
		UID:         true,
		Envelope:    true,
		BodySection: []*imap.FetchItemBodySection{{}},
	}
	fetchCmd := c.Fetch(seqset, fetchOptions)
	defer fetchCmd.Close()

	for {
		msg := fetchCmd.Next()
		if msg == nil {
			break
		}

		var uid imap.UID
		var subject string
		var message string

		for {
			item := msg.Next()
			if item == nil {
				break
			}

			switch item := item.(type) {
			case imapclient.FetchItemDataEnvelope:
				log.Printf("Subject: %v", item.Envelope.Subject)
				subject = item.Envelope.Subject

			case imapclient.FetchItemDataUID:
				log.Printf("UID: %v", item.UID)
				uid = item.UID

			case imapclient.FetchItemDataBodySection:
				b, err := io.ReadAll(item.Literal)
				if err != nil {
					log.Fatalf("failed to read body section: %v", err)
				}
				message = fmt.Sprintln(string(b))
			}
		}

		if err := onMessage(uid, subject, message); err != nil {
			log.Fatalf("failed to process message: %v", err)
			return err
		}

	}

	if err := fetchCmd.Close(); err != nil {
		log.Fatalf("FETCH command failed: %v", err)
	}

	return nil
}
