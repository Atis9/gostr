/*
Copyright © 2023 NAME HERE <EMAIL ADDRESS>
*/
package cmd

import (
	"context"
	"fmt"
	"io"
	"os"
	"regexp"

	"github.com/nbd-wtf/go-nostr"
	"github.com/nbd-wtf/go-nostr/nip19"
	"github.com/spf13/cobra"
)

var nsec string

// postCmd represents the post command
var postCmd = &cobra.Command{
	Use:   "post",
	Short: "Post to Nostr",
	Run: func(cmd *cobra.Command, args []string) {
		m, err := cmd.Flags().GetString("message")
		if err != nil {
			fmt.Println(err)
			os.Exit(1)
		}

		_, v, err := nip19.Decode(nsec)
		if err != nil {
			fmt.Println(err)
			os.Exit(1)
		}
		privateKey := v.(string)

		pub, _ := nostr.GetPublicKey(privateKey)
		ev := nostr.Event{
			PubKey:    pub,
			CreatedAt: nostr.Now(),
			Kind:      nostr.KindTextNote,
			Tags:      nil,
			Content:   m,
		}

		ev.Sign(privateKey)

		ctx := context.Background()
		for _, url := range []string{"wss://yabu.me", "wss://relay-jp.nostr.wirednet.jp"} {
			relay, err := nostr.RelayConnect(ctx, url)
			if err != nil {
				fmt.Println(err)
				continue
			}
			_, err = relay.Publish(ctx, ev)
			if err != nil {
				fmt.Println(err)
				continue
			}

			fmt.Printf("published to %s\n", url)
		}
	},
}

func init() {
	cobra.OnInitialize(initConfig)
	postCmd.PersistentFlags().StringP("message", "m", "", "post message")

	rootCmd.AddCommand(postCmd)
}

func initConfig() {
	configHomePath := os.Getenv("XDG_CONFIG_HOME")
	if configHomePath == "" {
		configHomePath = os.Getenv("HOME") + "/.config"
	}

	gostrConfigDirPath := configHomePath + "/gostr"
	if _, err := os.Stat(gostrConfigDirPath); os.IsNotExist(err) {
		os.Mkdir(gostrConfigDirPath, 0755)
	}

	nsecFilePath := gostrConfigDirPath + "/nsec"
	s, err := os.Stat(nsecFilePath)
	if err != nil {
		fmt.Println(err)
		os.Exit(1)
	}
	perm := s.Mode().Perm()
	if perm != 0600 {
		fmt.Println("nsec file permission is not 600")
		os.Exit(1)
	}

	f, err := os.Open(nsecFilePath)
	if err != nil {
		fmt.Println(err)
		os.Exit(1)
	}
	defer f.Close()

	b, err := io.ReadAll(f)
	re := regexp.MustCompile(`\r?\n`)

	nsec = re.ReplaceAllString(string(b), "")
}
