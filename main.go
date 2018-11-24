// mautrix - A Matrix client-server library intended for bots.
// Copyright (C) 2017 Tulir Asokan
//
// This program is free software: you can redistribute it and/or modify
// it under the terms of the GNU General Public License as published by
// the Free Software Foundation, either version 3 of the License, or
// (at your option) any later version.
//
// This program is distributed in the hope that it will be useful,
// but WITHOUT ANY WARRANTY; without even the implied warranty of
// MERCHANTABILITY or FITNESS FOR A PARTICULAR PURPOSE.  See the
// GNU General Public License for more details.
//
// You should have received a copy of the GNU General Public License
// along with this program.  If not, see <http://www.gnu.org/licenses/>.

package main

import (
	"flag"
	"fmt"
	"os"
	"os/signal"
	"regexp"
	"syscall"

	"maunium.net/go/mautrix"
)

var homeserver = flag.String("homeserver", "https://matrix.org", "Matrix homeserver")
var username = flag.String("username", "", "Matrix username localpart")
var password = flag.String("password", "", "Matrix password")

func main() {
	flag.Parse()
	fmt.Println("Logging to", *homeserver, "as", *username)
	mxbot := mautrix.Create(*homeserver)

	err := mxbot.PasswordLogin(*username, *password)
	if err != nil {
		fmt.Println(err)
		return
	}
	fmt.Println("Login successful")

	stop := make(chan bool, 1)

	shortcutmap := map[string]string{
		"pma#": "https://gitlab.com/postmarketOS/pmaports/issues/",
		"pma!": "https://gitlab.com/postmarketOS/pmaports/merge_requests/",
		"pmb#": "https://gitlab.com/postmarketOS/pmbootstrap/issues/",
		"pmb!": "https://gitlab.com/postmarketOS/pmbootstrap/merge_requests/",
		"org#": "https://gitlab.com/postmarketOS/postmarketos.org/issues/",
		"org!": "https://gitlab.com/postmarketOS/postmarketos.org/merge_requests/",
	}

	go mxbot.Listen()
	go func() {
	Loop:
		for {
			select {
			case <-stop:
				break Loop
			case evt := <-mxbot.Timeline:
				evt.MarkRead()
				switch evt.Type {
				case mautrix.EvtRoomMessage:
					if evt.Sender != *username &&
						(evt.Room.ID == "!clcCCNrLZYwdfNqkkR:disroot.org" || // #postmarketos:disroot.org
							evt.Room.ID == "!MxNOnZlZaurAGfcxFy:matrix.org" || // #postmarketos-lowlevel:disroot.org
							evt.Room.ID == "!VTQfOrQIBniIdCuMOq:matrix.org" || // #postmarketos-offtopic:disroot.org
							evt.Room.ID == "!NBvxopLbDoLCDlqKkL:z3ntu.xyz") { // #test2:z3ntu.xyz
						r := regexp.MustCompile("(pma[#!]|pmb[#!]|org[#!])(\\d+)")
						matches := r.FindStringSubmatch(evt.Content["body"].(string))
						if matches != nil {
							fmt.Println(matches[1] + matches[2] + " matched!")
							fmt.Printf("<%[1]s> %[4]s (%[2]s/%[3]s)\n", evt.Sender, evt.Type, evt.ID, evt.Content["body"])
							evt.Room.Send(shortcutmap[matches[1]] + matches[2])
						}
					}
				default:
					fmt.Println("Unidentified event of type", evt.Type)
				}
			case roomID := <-mxbot.InviteChan:
				invite := mxbot.Invites[roomID]
				fmt.Printf("%s invited me to %s (%s)\n", invite.Sender, invite.Name, invite.ID)
				fmt.Println(invite.Accept())
			}
		}
		mxbot.Stop()
	}()

	c := make(chan os.Signal, 1)
	signal.Notify(c, os.Interrupt, syscall.SIGTERM)
	<-c
	stop <- true

}
