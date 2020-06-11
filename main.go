package main

import (
	"encoding/json"
	"fmt"
	"github.com/tebeka/selenium"
	"github.com/thoas/go-funk"
	"io/ioutil"
	"log"
	"os"
	"strings"
	"time"
)

type config struct {
	Username string `json:"username"`
	Password string `json:"password"`
	Message  string `json:"message"`
}

var (
	user          string
	lastFollowers []string
)

func main() {

	config := getConfig()

	user = config.Username

	const (
		seleniumPath     = "selenium-server-standalone-3.141.59.jar"
		chromeDriverPath = "chromedriver_linux"
		port             = 8101
	)

	opts := []selenium.ServiceOption{
		selenium.StartFrameBuffer(),             // Start an X frame buffer for the browser to run in.
		selenium.ChromeDriver(chromeDriverPath), // Specify the path to GeckoDriver in order to use Firefox.
		//selenium.Output(os.Stderr),              // Output debug information to STDERR.
	}

	service, err := selenium.NewSeleniumService(seleniumPath, port, opts...)
	if err != nil {
		log.Println(err)
	}
	defer func() {
		_ = service.Stop()
	}()

	caps := selenium.Capabilities{"browserName": "chrome"}
	wd, err := selenium.NewRemote(caps, fmt.Sprintf("http://localhost:%d/wd/hub", port))
	if err != nil {
		panic(err)
	}
	defer func() {
		_ = wd.Quit()
	}()

	log.Println("logging...")
	err = login(config.Username, config.Password, wd)
	if err != nil {
		log.Println("logging failed, bye")
		panic(err)
	}
	log.Println("logging success")

	lastFollowers, err = getCurrentFollowers(wd)
	if err != nil {
		panic(err)
	}

	ticker := time.NewTicker(time.Minute)
	for {
		select {
		case <-ticker.C:
			log.Println("check new followers")
			newFollowers, err := getNewFollowers(wd)
			if err != nil {
				log.Println(err)
			}

			lastFollowers = append(lastFollowers, newFollowers...)

			for _, follower := range newFollowers {
				err = sendMessage(follower, config.Message, wd)
				if err != nil {
					log.Println(err)
				}
			}
		}
	}

}

func login(username, password string, wd selenium.WebDriver) error {
	if err := wd.Get("https://www.instagram.com/accounts/login/?next=%2Flogin%2F&source=desktop_nav"); err != nil {
		return err
	}

	time.Sleep(time.Second)

	usernameInput, err := wd.FindElement(selenium.ByCSSSelector, `input[name="username"]`)
	if err != nil {
		return err
	}
	err = usernameInput.SendKeys(username)
	if err != nil {
		return err
	}
	passwordInput, err := wd.FindElement(selenium.ByCSSSelector, `input[name="password"]`)
	if err != nil {
		return err
	}
	err = passwordInput.SendKeys(password)

	submitButton, err := wd.FindElement(selenium.ByCSSSelector, `button[type="submit"]`)
	if err != nil {
		return err
	}
	err = submitButton.Click()
	if err != nil {
		return err
	}

	time.Sleep(time.Second * 3)

	_, err = wd.FindElements(selenium.ByCSSSelector, "html.logged-in")
	if err != nil {
		return fmt.Errorf("login failed")
	}

	return nil
}

func sendMessage(recipient, message string, wd selenium.WebDriver) error {

	if err := wd.Get(fmt.Sprintf("https://www.instagram.com/%s/", recipient)); err != nil {
		return err
	}

	messageButton, err := wd.FindElement(selenium.ByXPATH, `//button[text()="Message"]`)
	if err != nil {
		return err
	}
	err = messageButton.Click()
	if err != nil {
		return err
	}

	time.Sleep(time.Second)

	textarea, err := wd.FindElement(selenium.ByCSSSelector, `textarea`)
	if err != nil {
		return err
	}
	err = textarea.SendKeys(message)
	if err != nil {
		return err
	}

	time.Sleep(time.Second)

	sendButton, err := wd.FindElement(selenium.ByXPATH, `//button[text()="Send"]`)
	if err != nil {
		return err
	}
	classes, err := sendButton.GetAttribute("class")
	if err != nil {
		return err
	}
	_, err = wd.ExecuteScriptRaw(
		fmt.Sprintf("return document.getElementsByClassName(\"%s\")[1].click();", classes),
		[]interface{}{},
	)
	if err != nil {
		return err
	}

	return nil

}

func getNewFollowers(wd selenium.WebDriver) ([]string, error) {
	currentFollowers, err := getCurrentFollowers(wd)
	if err != nil {
		return nil, err
	}
	diff, _ := funk.DifferenceString(currentFollowers, lastFollowers)
	return diff, nil
}

func getCurrentFollowers(wd selenium.WebDriver) ([]string, error) {
	if err := wd.Get(fmt.Sprintf("https://www.instagram.com/%s/", user)); err != nil {
		return nil, err
	}

	time.Sleep(time.Second)

	followersButton, err := wd.FindElement(selenium.ByCSSSelector, fmt.Sprintf(`a[href="/%s/followers/"]`, user))
	if err != nil {
		return nil, err
	}
	err = followersButton.Click()
	if err != nil {
		return nil, err
	}

	time.Sleep(time.Second)

	followersDialog, err := wd.FindElement(selenium.ByCSSSelector, `div[role="dialog"]`)
	if err != nil {
		return nil, err
	}

	time.Sleep(time.Second)

	ul, err := followersDialog.FindElement(selenium.ByCSSSelector, "ul")
	if err != nil {
		return nil, err
	}

	time.Sleep(time.Second)

	lis, err := ul.FindElements(selenium.ByCSSSelector, "li")
	if err != nil {
		return nil, err
	}

	var result []string
	for _, li := range lis {
		text, _ := li.Text()
		result = append(result, strings.Split(text, "\n")[0])
	}

	return result, nil
}

func getConfig() config {
	jsonFile, err := os.Open("config.json")
	defer func() {
		_ = jsonFile.Close()
	}()
	if err != nil {
		panic(err)
	}
	bytes, _ := ioutil.ReadAll(jsonFile)

	c := config{}
	err = json.Unmarshal(bytes, &c)
	if err != nil {
		panic(err)
	}
	return c
}
