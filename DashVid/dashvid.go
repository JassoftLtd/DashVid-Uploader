package main // import "github.com/JassoftLtd/DashVid-Uploader"

import (
	"fmt"
	"github.com/urfave/cli" // imports as package "cli"
	"os"
)

func main() {
	app := cli.NewApp()
	app.Name = "dashvid"
	app.Usage = "DashVid.io Client"

	app.Commands = []cli.Command{
		{
			Name:    "upload",
			Usage:   "upload video files using CameraKey. eg. dashvid upload [CameraKey] [VideoFolder]",
			Action:  func(c *cli.Context) error {
				fmt.Println("Uploading videos for CameraKey", c.Args().Get(0))
				fmt.Println("Uploading videos in folder", c.Args().Get(1))
				if c.String("delete") == "yes" {
					fmt.Println("Deleting files after uplaod")
				}
				return nil
			},
			Flags: []cli.Flag{
				cli.StringFlag{
					Name: "delete",
					Value: "yes",
					Usage: "delete files after upload",
				},
			},

		},
	}


	app.Run(os.Args)
}
