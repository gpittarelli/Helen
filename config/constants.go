// Copyright (C) 2015  TF2Stadium
// Use of this source code is governed by the GPLv3
// that can be found in the COPYING file.

package config

import (
	"flag"
	"os"

	"github.com/TF2Stadium/Helen/helpers"
	"github.com/bitly/go-simplejson"
)

var jsonConfig *simplejson.Json

type constants struct {
	GlobalChatRoom     string
	Port               string
	Domain             string
	OpenIDRealm        string
	CookieDomain       string
	LoginRedirectPath  string
	CookieStoreSecret  string
	StaticFileLocation string
	ChatLogsDir        string
	SessionName        string
	PaulingPort        string
	SocketMockUp       bool
	ServerMockUp       bool
	ChatLogsEnabled    bool
	MockupAuth         bool
	AllowedCorsOrigins []string

	// database
	DbHost     string
	DbPort     string
	DbDatabase string
	DbUsername string
	DbPassword string

	SteamDevApiKey string
	SteamApiMockUp bool
}

func overrideStringFromConfig(constant *string, name string) {
	val, _ := jsonConfig.Get(name).String()
	if "" != val {
		*constant = val
	}

}

func overrideBoolFromConfig(constant *bool, name string) {
	val, err := jsonConfig.Get(name).Bool()
	if err == nil {
		*constant = val
	}
}

var Constants constants

func SetupConstants() {
	fileName := flag.String("config", "config.json", "Configuration file")
	deploy := flag.String("deploy", "development", "Deployment mode")
	configFile, pathErr := os.Open(*fileName)
	if pathErr != nil {
		helpers.Logger.Warning("No config file found. Using %s defaults.", *deploy)
	}

	var err error
	jsonConfig, err = simplejson.NewFromReader(configFile)
	if err != nil && pathErr == nil {
		helpers.Logger.Fatal(err.Error())
	}

	Constants = constants{}

	setupDevelopmentConstants()
	switch *deploy {
	case "production":
		setupProductionConstants()
	case "test":
		setupTestConstants()
	case "travis_test":
		setupTravisTestConstants()
	}

	overrideStringFromConfig(&Constants.Port, "PORT")
	overrideStringFromConfig(&Constants.ChatLogsDir, "CHAT_LOG_DIR")
	overrideStringFromConfig(&Constants.CookieStoreSecret, "COOKIE_STORE_SECRET")
	overrideStringFromConfig(&Constants.SteamDevApiKey, "STEAM_API_KEY")
	overrideStringFromConfig(&Constants.DbHost, "DATABASE_HOST")
	overrideStringFromConfig(&Constants.DbPort, "DATABASE_PORT")
	overrideStringFromConfig(&Constants.DbUsername, "DATABASE_USERNAME")
	overrideStringFromConfig(&Constants.DbPassword, "DATABASE_PASSWORD")
	overrideStringFromConfig(&Constants.PaulingPort, "PAULING_PORT")
	overrideStringFromConfig(&Constants.Domain, "SERVER_DOMAIN")
	overrideStringFromConfig(&Constants.OpenIDRealm, "SERVER_OPENID_REALM")
	overrideStringFromConfig(&Constants.CookieDomain, "SERVER_COOKIE_DOMAIN")
	overrideBoolFromConfig(&Constants.ChatLogsEnabled, "LOG_CHAT")
	overrideBoolFromConfig(&Constants.ServerMockUp, "PAULING_ENABLE")
	overrideBoolFromConfig(&Constants.MockupAuth, "MOCKUP_AUTH")
	overrideStringFromConfig(&Constants.LoginRedirectPath, "SERVER_REDIRECT_PATH")
	// conditional assignments

	if Constants.SteamDevApiKey == "your steam dev api key" && !Constants.SteamApiMockUp {
		helpers.Logger.Warning("Steam api key not provided, setting SteamApiMockUp to true")
		Constants.SteamApiMockUp = true
	}

}

func setupDevelopmentConstants() {
	Constants.GlobalChatRoom = "0"
	Constants.Port = "8080"
	Constants.Domain = "http://localhost:8080"
	Constants.OpenIDRealm = "http://localhost:8080"
	Constants.CookieDomain = ""
	Constants.LoginRedirectPath = "http://localhost:8080/"
	Constants.CookieStoreSecret = "dev secret is very secret"
	Constants.SessionName = "defaultSession"
	Constants.StaticFileLocation = os.Getenv("GOPATH") + "/src/github.com/TF2Stadium/Helen/static"
	Constants.PaulingPort = "1234"
	Constants.ChatLogsDir = "."
	Constants.SocketMockUp = false
	Constants.ServerMockUp = true
	Constants.ChatLogsEnabled = false
	Constants.AllowedCorsOrigins = []string{"*"}

	Constants.DbHost = "127.0.0.1"
	Constants.DbPort = "5724"
	Constants.DbDatabase = "tf2stadium"
	Constants.DbUsername = "tf2stadium"
	Constants.DbPassword = "dickbutt" // change this

	Constants.SteamDevApiKey = "your steam dev api key"
	Constants.SteamApiMockUp = false
}

func setupProductionConstants() {
	// override production stuff here
	Constants.Port = "5555"
	Constants.ChatLogsDir = "."
	Constants.CookieDomain = ".tf2stadium.com"
	Constants.ServerMockUp = false
	Constants.ChatLogsEnabled = true
	Constants.SocketMockUp = false
}

func setupTestConstants() {
	Constants.DbHost = "127.0.0.1"
	Constants.DbDatabase = "TESTtf2stadium"
	Constants.DbUsername = "TESTtf2stadium"
	Constants.DbPassword = "dickbutt"

	Constants.ServerMockUp = true
	Constants.SteamApiMockUp = true
}

func setupTravisTestConstants() {
	Constants.DbHost = "127.0.0.1"
	Constants.DbDatabase = "tf2stadium"
	Constants.DbUsername = "postgres"
	Constants.DbPassword = ""

	Constants.ServerMockUp = true
	Constants.SteamApiMockUp = true
}
