package main

import (
	"bufio"
	"bytes"
	"fmt"
	"io"
	"os"
	"os/exec"
	"regexp"
	"strconv"
	"strings"
	"time"
)

const menu = "1) Fight a goblin!\n2) Fight a troll!\n3) Fight a dragon!\n4) Buy a sword!\n" +
	"5) Upgrade your sword!\n6) Rest for a moment\n7) Rest for a bit\n8) Rest for a while\n9) Quit"

var reader = bufio.NewReader(os.Stdin)
var swordLv = 0
var shards = 0
var coins = 0
var curHp = 20
var characterId = ""
var localAccount = ""
var addr = ""
var gameEnded = false
var addrRegex = regexp.MustCompile(`{"key":"creator","value":"(.*?)"}`)

var curHpRegex = regexp.MustCompile(`  - key: currentHp\n    value: "(.*?)"\n`)

var swordLvRegex = regexp.MustCompile(`  - key: swordLevel\n    value: "(.*?)"\n`)

var coinRegex = regexp.MustCompile(`  - key: coins\n    value: "(.*?)"\n`)

var shardRegex = regexp.MustCompile(`  - key: shards\n    value: "(.*?)"\n`)

var itemIdRegex = regexp.MustCompile(`id: (.*?)\n`)

func main() {
	setLocalAccount()
	generateCharacter()
	for !gameEnded {
		checkCharacter()
		if swordLv == 0 {
			fmt.Printf("You have %s/20 HP remaining. You are unarmed.\n\n", strconv.Itoa(curHp))
		} else {
			fmt.Printf("You have %s/20 HP remaining. You have a sword of level %s.\n\n", strconv.Itoa(curHp), strconv.Itoa(swordLv))
		}
		fmt.Printf("Coins: %s; Shards: %s\n\n", strconv.Itoa(coins), strconv.Itoa(shards))

		if curHp < 1 {
			println(("You have died."))
			gameEnded = true
		} else {
			println(menu)
			str, err := reader.ReadString('\n')
			if err != nil {
				panic(err)
			}
			switch str {
			case "1\n":
				fightGoblin()
			case "2\n":
				//fightTroll()
			case "3\n":
				//fightDragon()
			case "4\n":
				//buySword()
			case "5\n":
				//upgradeSword()
			case "6\n":
				//rest1()
			case "7\n":
				//rest2()
			case "8\n":
				//rest3()
			case "9\n":
				gameEnded = true
			}
		}
	}
}

func setLocalAccount() {
	for localAccount == "" {
		println("Please provide the name of a local keypair corresponding to an extant Pylons account.\nThis will be used for the remainder of the session.")
		str, err := reader.ReadString('\n')
		if err != nil {
			panic(err)
		}
		localAccount = str
	}
}

func checkCharacter() {
	println("Checking character...")
	dat := execQueryCmd([]string{"query", "pylons", "get-item", "appTestCookbook", characterId})
	var err error
	curHp, err = strconv.Atoi(string(curHpRegex.FindStringSubmatch(dat)[1]))
	if err != nil {
		panic(err)
	}
	swordLv, err = strconv.Atoi(string(swordLvRegex.FindStringSubmatch(dat)[1]))
	if err != nil {
		panic(err)
	}
	coins, err = strconv.Atoi(string(coinRegex.FindStringSubmatch(dat)[1]))
	if err != nil {
		panic(err)
	}
	shards, err = strconv.Atoi(string(shardRegex.FindStringSubmatch(dat)[1]))
	if err != nil {
		panic(err)
	}
}

func generateCharacter() {
	println("Generating character...")
	dat := execTxCmd([]string{"tx", "pylons", "execute-recipe", "appTestCookbook", "RecipeTestAppGetCharacter", "0", "[]", "[]", "--from", localAccount})
	hash := dat[len(dat)-65 : len(dat)-1]
	dat = execQueryCmd([]string{"query", "tx", hash})
	addr = addrRegex.FindStringSubmatch(dat)[1]
	dat = execQueryCmd([]string{"query", "pylons", "list-item-by-owner", addr})
	characterId = itemIdRegex.FindAllStringSubmatch(dat, -1)[1][1]
}

func fightGoblin() {
	execTxCmd([]string{"tx", "pylons", "execute-recipe", "appTestCookbook", "RecipeTestAppFightGoblin", "0", fmt.Sprintf(`["%s"]`, characterId), "[]", "--from", localAccount})
}

func execQueryCmd(args []string) string {
	args[len(args)-1] = strings.TrimSpace(args[len(args)-1])
	cmd := exec.Command("pylonsd", args...)
	var outb bytes.Buffer
	cmd.Stderr = os.Stderr
	cmd.Stdout = &outb
	cmd.Run()
	return outb.String()
}

func execTxCmd(args []string) string {
	args[len(args)-1] = strings.TrimSpace(args[len(args)-1])
	cmd := exec.Command("pylonsd", args...)
	var outb bytes.Buffer
	cmd.Stderr = os.Stderr
	cmd.Stdout = &outb
	stdin, err := cmd.StdinPipe()
	if err != nil {
		fmt.Println(err)
	}
	cmd.Start()
	io.WriteString(stdin, "y\n")
	cmd.Wait()
	time.Sleep(time.Second * 5)
	return outb.String()
}
