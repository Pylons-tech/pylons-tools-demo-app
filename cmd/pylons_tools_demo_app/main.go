package main

import (
	"bufio"
	"bytes"
	"fmt"
	"io"
	"os"
	"os/exec"
	"strconv"
	"strings"
	"time"

	"github.com/buger/jsonparser"
)

const dualJsonRootTxSplitMagic = `{"height":`

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

func main() {
	setLocalAccount()
	generateCharacter()
	for !gameEnded {
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
				fightTroll()
			case "3\n":
				fightDragon()
			case "4\n":
				buySword()
			case "5\n":
				upgradeSword()
			case "6\n":
				rest1()
			case "7\n":
				rest2()
			case "8\n":
				rest3()
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
	curHp = retrieveLong([]byte(dat), "currentHp")
	swordLv = retrieveLong([]byte(dat), "swordLevel")
	coins = retrieveLong([]byte(dat), "coins")
	shards = retrieveLong([]byte(dat), "shards")
}

func generateCharacter() {
	println("Generating character...")
	dat := execTxCmd([]string{"tx", "pylons", "execute-recipe", "appTestCookbook", "RecipeTestAppGetCharacter", "0", "[]", "[]", "--from", localAccount})
	vs := splitDualJsonRoots(dat)
	hash, err := jsonparser.GetString([]byte(vs[1]), "txhash")
	if err != nil {
		panic(err)
	}
	dat = execQueryCmd([]string{"query", "tx", hash})
	addr, err = jsonparser.GetString([]byte(dat), "tx", "body", "messages", "[0]", "creator")
	if err != nil {
		panic(err)
	}
	dat = execQueryCmd([]string{"query", "pylons", "list-item-by-owner", addr})
	// this is a mess, we should write some helpers for queries like this
	var found *[]byte
	_, err = jsonparser.ArrayEach([]byte(dat), func(v0 []byte, dataType jsonparser.ValueType, offset int, err error) {
		_, err = jsonparser.ArrayEach(v0, func(v1 []byte, dataType jsonparser.ValueType, offset int, err error) {
			k, _ := jsonparser.GetString(v1, "key")
			if k == "entityType" {
				v, _ := jsonparser.GetString(v1, "value")
				if v == "character" {
					found = &v0 // todo: more logic to select a character, if multiple
				}
			}
		}, "strings")
	}, "items")

	characterId, err = jsonparser.GetString(*found, "id")
	if err != nil {
		panic(err)
	}
	checkCharacter()
}

func fightGoblin() {
	println("Fighting a goblin...")
	execTxCmd([]string{"tx", "pylons", "execute-recipe", "appTestCookbook", "RecipeTestAppFightGoblin", "0", fmt.Sprintf(`["%s"]`, characterId), "[]", "--from", localAccount})
	println("Victory!")
	var lastHp = curHp
	var lastCoins = coins
	checkCharacter()
	if lastHp != curHp {
		fmt.Printf("Took %s damage!\n", strconv.Itoa(lastHp-curHp))
	}
	if lastCoins != coins {
		fmt.Printf("Found %s coins!\n", strconv.Itoa(coins-lastCoins))
	}
}

func fightTroll() {
	println("Fighting a troll...")
	if swordLv < 1 {
		execTxCmd([]string{"tx", "pylons", "execute-recipe", "appTestCookbook", "RecipeTestAppFightTrollUnarmed", "0", fmt.Sprintf(`["%s"]`, characterId), "[]", "--from", localAccount})
		println("Defeat...")
		var lastHp = curHp
		var lastCoins = coins
		checkCharacter()
		if lastHp != curHp {
			fmt.Printf("Took %s damage!\n", strconv.Itoa(lastHp-curHp))
		}
		if lastCoins != coins {
			fmt.Printf("Found %s coins!\n", strconv.Itoa(coins-lastCoins))
		}
	} else {
		execTxCmd([]string{"tx", "pylons", "execute-recipe", "appTestCookbook", "RecipeTestAppFightTrollArmed", "0", fmt.Sprintf(`["%s"]`, characterId), "[]", "--from", localAccount})
		println("Victory!")
		var lastHp = curHp
		var lastCoins = coins
		checkCharacter()
		if lastHp != curHp {
			fmt.Printf("Took %s damage!\n", strconv.Itoa(lastHp-curHp))
		}
		if lastCoins != coins {
			fmt.Printf("Found %s coins!\n", strconv.Itoa(coins-lastCoins))
		}
	}
}

func fightDragon() {
	println("Fighting a dragon...")
	if swordLv < 2 {
		execTxCmd([]string{"tx", "pylons", "execute-recipe", "appTestCookbook", "RecipeTestAppFightDragonUnarmed", "0", fmt.Sprintf(`["%s"]`, characterId), "[]", "--from", localAccount})
		println("Defeat...")
		var lastHp = curHp
		var lastCoins = coins
		checkCharacter()
		if lastHp != curHp {
			fmt.Printf("Took %s damage!\n", strconv.Itoa(lastHp-curHp))
		}
		if lastCoins != coins {
			fmt.Printf("Found %s coins!\n", strconv.Itoa(coins-lastCoins))
		}
	} else {
		execTxCmd([]string{"tx", "pylons", "execute-recipe", "appTestCookbook", "RecipeTestAppFightDragonArmed", "0", fmt.Sprintf(`["%s"]`, characterId), "[]", "--from", localAccount})
		println("Victory!")
		var lastHp = curHp
		var lastCoins = coins
		checkCharacter()
		if lastHp != curHp {
			fmt.Printf("Took %s damage!\n", strconv.Itoa(lastHp-curHp))
		}
		if lastCoins != coins {
			fmt.Printf("Found %s coins!\n", strconv.Itoa(coins-lastCoins))
		}
	}
}

func buySword() {
	if swordLv > 0 {
		println("You already have a sword")
	} else if coins < 50 {
		println("You need 50 coins to buy a sword")
	} else {
		execTxCmd([]string{"tx", "pylons", "execute-recipe", "appTestCookbook", "RecipeTestAppBuySword", "0", fmt.Sprintf(`["%s"]`, characterId), "[]", "--from", localAccount})
		println("Bought a sword!")
		var lastCoins = coins
		checkCharacter()
		if lastCoins != coins {
			fmt.Printf("Spent %s coins!\n", strconv.Itoa(lastCoins-coins))
		}
	}
}

func upgradeSword() {
	if swordLv > 1 {
		println("You already have an upgraded sword")
	} else if coins < 50 {
		println("You need 5 shards to upgrade your sword")
	} else {
		execTxCmd([]string{"tx", "pylons", "execute-recipe", "appTestCookbook", "RecipeTestAppUpgradeSword", "0", fmt.Sprintf(`["%s"]`, characterId), "[]", "--from", localAccount})
		println("Upgraded your sword!")
		var lastShards = shards
		checkCharacter()
		if lastShards != shards {
			fmt.Printf("Spent %s shards!\n", strconv.Itoa(lastShards-shards))
		}
	}
}

func rest1() {
	println("Resting...")
	execDelayedTxCmd([]string{"tx", "pylons", "execute-recipe", "appTestCookbook", "RecipeTestAppRest25", "0", fmt.Sprintf(`["%s"]`, characterId), "[]", "--from", localAccount})
	println("Done!")
	var lastHp = curHp
	checkCharacter()
	if lastHp != curHp {
		fmt.Printf("Recovered %s HP!\n", strconv.Itoa(curHp-lastHp))
	}
}

func rest2() {
	println("Resting...")
	execDelayedTxCmd([]string{"tx", "pylons", "execute-recipe", "appTestCookbook", "RecipeTestAppRest50", "0", fmt.Sprintf(`["%s"]`, characterId), "[]", "--from", localAccount})
	println("Done!")
	var lastHp = curHp
	checkCharacter()
	if lastHp != curHp {
		fmt.Printf("Recovered %s HP!\n", strconv.Itoa(curHp-lastHp))
	}
}

func rest3() {
	println("Resting...")
	execDelayedTxCmd([]string{"tx", "pylons", "execute-recipe", "appTestCookbook", "RecipeTestAppRest100", "0", fmt.Sprintf(`["%s"]`, characterId), "[]", "--from", localAccount})
	println("Done!")
	var lastHp = curHp
	checkCharacter()
	if lastHp != curHp {
		fmt.Printf("Recovered %s HP!\n", strconv.Itoa(curHp-lastHp))
	}
}

func execQueryCmd(args []string) string {
	args[len(args)-1] = strings.TrimSpace(args[len(args)-1])
	cmd := exec.Command("pylonsd", append(args, "--output", "json")...)
	var outb bytes.Buffer
	cmd.Stderr = os.Stderr
	cmd.Stdout = &outb
	cmd.Run()
	return outb.String()
}

func execTxCmd(args []string) string {
	args[len(args)-1] = strings.TrimSpace(args[len(args)-1])
	cmd := exec.Command("pylonsd", append(args, "--output", "json")...)
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

func execDelayedTxCmd(args []string) string {
	args[len(args)-1] = strings.TrimSpace(args[len(args)-1])
	cmd := exec.Command("pylonsd", append(args, "--output", "json")...)
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
	dat := outb.String()
	vs := splitDualJsonRoots(dat)
	hash, err := jsonparser.GetString([]byte(vs[1]), "txhash")
	if err != nil {
		panic(err)
	}
	dat = execQueryCmd([]string{"query", "tx", hash})

	// this is a mess, we should write some helpers for queries like this
	var found *[]byte
	_, err = jsonparser.ArrayEach([]byte(dat), func(v0 []byte, dataType jsonparser.ValueType, offset int, err error) {
		k, _ := jsonparser.GetString(v0, "type")
		if k == "create_execution" {
			found = &v0
			return
		}
	}, "logs", "[0]", "events")

	execId := retrieveAttr(*found, "ID")

	var escaped = false

	for !escaped {
		dat := execQueryCmd([]string{"query", "pylons", "get-execution", execId})
		v, err := jsonparser.GetBoolean([]byte(dat), "completed")
		if err != nil {
			panic(err)
		}
		escaped = v
		println("...")
	}

	return outb.String()
}

func splitDualJsonRoots(dat string) []string {
	// this is obnoxious: we get multiple json roots off this instead of smth well-formatted, so we have to hack around that
	vs := strings.SplitN(dat, dualJsonRootTxSplitMagic, 2)
	vs[1] = dualJsonRootTxSplitMagic + vs[1]
	return vs
}

func retrieveLong(dat []byte, key string) int {
	ret := 0
	jsonparser.ArrayEach(dat, func(v []byte, dataType jsonparser.ValueType, offset int, err error) {
		k, _ := jsonparser.GetString(v, "key")
		if k == key {
			v0, _ := jsonparser.GetString(v, "value")
			ret, _ = strconv.Atoi(v0)
			return
		}
	}, "item", "longs")
	return ret
}

func retrieveString(dat []byte, key string) string {
	ret := ""
	jsonparser.ArrayEach(dat, func(v []byte, dataType jsonparser.ValueType, offset int, err error) {
		k, _ := jsonparser.GetString(v, "key")
		if k == key {
			v, _ := jsonparser.GetString(v, "value")
			ret = v
			return
		}
	}, "item", "strings")
	return ret
}

func retrieveAttr(dat []byte, key string) string {
	ret := ""
	jsonparser.ArrayEach(dat, func(v []byte, dataType jsonparser.ValueType, offset int, err error) {
		k, _ := jsonparser.GetString(v, "key")
		if k == key {
			v, _ := jsonparser.GetString(v, "value")
			ret = v
			return
		}
	}, "attributes")
	return ret
}
