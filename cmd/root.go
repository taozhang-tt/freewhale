/*
Copyright © 2022 NAME HERE <EMAIL ADDRESS>
*/
package cmd

import (
	"encoding/json"
	"fmt"
	"io/ioutil"
	"net/http"
	"net/url"
	"os"
	"strconv"
	"strings"
	"time"

	"github.com/spf13/cobra"
)

var (
	Email  string
	Passwd string
	Domain string
)

var rootCmd = &cobra.Command{
	Use:   "freewhale",
	Short: "freewhale checkin",
	Run:   FreeWhalefunc,
}

// Execute adds all child commands to the root command and sets flags appropriately.
// This is called by main.main(). It only needs to happen once to the rootCmd.
func Execute() {
	err := rootCmd.Execute()
	if err != nil {
		os.Exit(1)
	}
}

func init() {
	rootCmd.Flags().StringVarP(&Email, "email", "e", "", "email")
	rootCmd.Flags().StringVarP(&Passwd, "passwd", "p", "", "password")
	rootCmd.Flags().StringVarP(&Domain, "domain", "d", "", "freewhale.xyz")
}

func FreeWhalefunc(cmd *cobra.Command, args []string) {
	if Passwd == "" || Email == "" {
		fmt.Println("please input email and password, do't be shy, just to try...")
		return
	}

	fmt.Printf("%v --- start login\n", time.Now().Format("2006-01-02 15:04:05"))
	cookies, err := Login(Email, Passwd)
	if err != nil {
		fmt.Printf("login with error(%v)", err)
		return
	}

	fmt.Printf("%v --- start checkin\n", time.Now().Format("2006-01-02 15:04:05"))
	if err := Checkin(cookies); err != nil {
		fmt.Printf("checkin with error(%v)", err)
		return
	}

	fmt.Printf("%v --- checkin success\n", time.Now().Format("2006-01-02 15:04:05"))
}

func Login(email, passwd string) ([]*http.Cookie, error) {
	form := url.Values{}
	form.Add("email", email)
	form.Add("passwd", passwd)
	url := "https://" + Domain + "/auth/login"
	resp, err := http.DefaultClient.PostForm(url, form)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	bs, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		return nil, err
	}

	if !json.Valid(bs) {
		return nil, fmt.Errorf("wrong email or password")
	}

	response := new(Response)
	if err := json.Unmarshal(bs, response); err != nil {
		return nil, err
	}

	if response.Ret != 1 {
		ret, err := unicode2zh(bs)
		if err != nil {
			return nil, fmt.Errorf("unicode2zh with error(%w)", err)
		}
		return nil, fmt.Errorf("Login response body: %s", ret)
	}

	return resp.Cookies(), nil
}

func Checkin(cookies []*http.Cookie) error {
	url := "https://" + Domain + "/user/checkin"
	request, err := http.NewRequest("POST", url, nil)
	if err != nil {
		return err
	}
	for _, cookie := range cookies {
		request.AddCookie(cookie)
	}
	resp, err := http.DefaultClient.Do(request)
	if err != nil {
		return err
	}
	defer resp.Body.Close()

	bs, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		return err
	}
	ret, err := unicode2zh(bs)
	if err != nil {
		return fmt.Errorf("unicode2zh with error(%w)", err)
	}
	fmt.Printf("Checkin response body: %s\n", ret)
	return nil
}

func unicode2zh(raw []byte) ([]byte, error) {
	str, err := strconv.Unquote(strings.Replace(strconv.Quote(string(raw)), `\\u`, `\u`, -1))
	if err != nil {
		return nil, err
	}
	return []byte(str), nil
}

type Response struct {
	Ret int    `json:"ret"`
	Msg string `json:"msg"`
}
