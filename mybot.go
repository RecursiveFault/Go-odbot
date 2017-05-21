/*

mybot - Illustrative Slack bot in Go

Copyright (c) 2015 RapidLoop

Permission is hereby granted, free of charge, to any person obtaining a copy
of this software and associated documentation files (the "Software"), to deal
in the Software without restriction, including without limitation the rights
to use, copy, modify, merge, publish, distribute, sublicense, and/or sell
copies of the Software, and to permit persons to whom the Software is
furnished to do so, subject to the following conditions:

The above copyright notice and this permission notice shall be included in
all copies or substantial portions of the Software.

THE SOFTWARE IS PROVIDED "AS IS", WITHOUT WARRANTY OF ANY KIND, EXPRESS OR
IMPLIED, INCLUDING BUT NOT LIMITED TO THE WARRANTIES OF MERCHANTABILITY,
FITNESS FOR A PARTICULAR PURPOSE AND NONINFRINGEMENT. IN NO EVENT SHALL THE
AUTHORS OR COPYRIGHT HOLDERS BE LIABLE FOR ANY CLAIM, DAMAGES OR OTHER
LIABILITY, WHETHER IN AN ACTION OF CONTRACT, TORT OR OTHERWISE, ARISING FROM,
OUT OF OR IN CONNECTION WITH THE SOFTWARE OR THE USE OR OTHER DEALINGS IN
THE SOFTWARE.
*/

package main

import (
	"encoding/csv"
	"fmt"
	"log"
	"net/http"
	"os"
	"strings"
	"bufio"
	_ "text/scanner"
)

var emojifile = "emoji.csv"

func main() {
	if len(os.Args) != 2 {
		fmt.Fprintf(os.Stderr, "usage: mybot slack-bot-token\n")
		os.Exit(1)
	}

	if !FileExists(emojifile){
		CreateFile(emojifile)
	}

	// start a websocket-based Real Time API session
	ws, id := slackConnect(os.Args[1])
	fmt.Println("mybot ready, ^C exits")

	for { // main loop, read each incoming message
		m, err := getMessage(ws)
		if err != nil {
			log.Fatal(err)
		}

		// see if bot is mentioned
		if m.Type == "message" && strings.HasPrefix(m.Text, "<@"+id+">") {
			// if so try to parse if
			parts := strings.Fields(m.Text)
			if len(parts) == 4 && parts[1] == "emoji" {
				// looks good, get the quote and reply with the result
				go func(m Message) {
					if(m.Channel =="G5GSTPRPZ") {
						i := AddEmojiToCSV(parts[2], parts[3])
						if i {
							print(parts)
							m.Text = fmt.Sprintf("Added " + parts[2])
							postMessage(ws, m)
						} else {
							m.Text = fmt.Sprintf("Failed to add.")
							postMessage(ws, m)
						}
					}
				}(m)
				// NOTE: the Message object is copied, this is intentional
			} else if len(parts) ==3 && parts[1] =="gbf" { //gbf stickers
				go func(m Message) {
					if (m.Channel == "C09HBS03F"){ //hardcode mobile game channel ID
						m.Text = getEmoji(parts[2])
						postMessage(ws, m)}

				}(m)
			} else { //unrecognized command
				m.Text = fmt.Sprintf("Sorry, that's not a recognized command.\n")
				postMessage(ws, m)
			}
		}
	}
}

// Get the quote via Yahoo. You should replace this method to something
// relevant to your team!
func getQuote(sym string) string {
	sym = strings.ToUpper(sym)
	url := fmt.Sprintf("http://download.finance.yahoo.com/d/quotes.csv?s=%s&f=nsl1op&e=.csv", sym)
	resp, err := http.Get(url)
	if err != nil {
		return fmt.Sprintf("error: %v", err)
	}
	rows, err := csv.NewReader(resp.Body).ReadAll()
	if err != nil {
		return fmt.Sprintf("error: %v", err)
	}
	if len(rows) >= 1 && len(rows[0]) == 5 {
		return fmt.Sprintf("%s (%s) is trading at $%s", rows[0][0], rows[0][1], rows[0][2])
	}
	return fmt.Sprintf("unknown response format (symbol was \"%s\")", sym)
}

func getEmoji(sym string) string {
	sym = strings.ToUpper(sym)
	if !FileExists(emojifile){
		return fmt.Sprintf("error: No emoji file")
	}
	file,err:= os.Open(emojifile)
	if err != nil {
		return fmt.Sprintf("error: %v", err)
	}
	defer file.Close()
	rows, err := csv.NewReader(file).ReadAll()
	if err != nil {
		return fmt.Sprintf("error: %v", err)
	}
	for i:= 0; i < len(rows); i++ {
		s := fmt.Sprintf("%s", rows[i][0])
		if len(rows) >= 1 && s == sym {
			return fmt.Sprintf("%s", rows[i][1]) //return emoji URL for Slack to embed
		}
	}
	file.Close()
	return fmt.Sprintf("unknown response format (symbol was \"%s\")", sym)
}

func AddEmojiToCSV(key string, value string) bool { //TODO: Permission system? Password?
	ret := true
	ins := false
	//insert into file, change to true if applicable
	key = strings.ToUpper(key)
	value = strings.ToLower(value)
	value = value[1:(len(value)-1)] //trim slack formatting for URLs //TODO: Input validation
	if !FileExists(emojifile){
		return false
	}
	file,err:= os.OpenFile("emoji.csv",os.O_WRONLY|os.O_CREATE|os.O_APPEND,0644)
	if err != nil {
		return false
	}
	defer file.Close()

	//TODO: Look for duplicate emoji names, set ret accordingly
	if (ret==true) {
		b :=[]string{key,value}
		fmt.Println(b)
		w := csv.NewWriter(file)
		err := w.Write([]string{key,value})
		if err != nil {
			println(fmt.Sprintf("error: %v", err))
			return false
		} else{ins = true}
		w.Flush()
	}
	return ins
}

func GetStringFromFile(name string, key string) string {
	ret := "String not found"
	if FileExists(name) {
		file, err := os.Open(name)
		if err != nil {
			log.Fatal(err)
		}
		defer file.Close()

		scanner := bufio.NewScanner(file)
		for scanner.Scan() {
			i := scanner.Text()
			if strings.HasPrefix(i, key) {
				ret = i
			}
		}
		if err := scanner.Err(); err != nil {
			log.Fatal(err)
		}
	}
	return ret
}

func FileExists(name string) bool {
	if _, err := os.Stat(name); err != nil {
		if os.IsNotExist(err) {
			return false
		}
	}
	return true
}

func CreateFile(name string) error {
	fo, err := os.Create(name)
	if err != nil {
		return err
	}
	defer func() {
		fo.Close()
	}()
	return nil
}