package main

import (
	"fmt"
	"os"

	"github.com/spf13/viper"
)

func main() {
	cwd, _ := os.Getwd()
	fmt.Println("CWD:", cwd)

	viper.SetConfigName("config")
	viper.AddConfigPath("./config")
	viper.SetConfigType("yaml")

	if err := viper.ReadInConfig(); err != nil {
		fmt.Printf("Error reading config: %v\n", err)
		return
	}

	fmt.Println("Config file used:", viper.ConfigFileUsed())
	fmt.Println("All keys:", viper.AllKeys())
	fmt.Println("database.url:", viper.GetString("database.url"))
}
