package service

import (
	"context"
	"fmt"

	"github.com/pkg/errors"
	"github.com/urfave/cli"
)

func backupCommands(client *Client) []cli.Command {
	return []cli.Command{
		cli.Command{
			Name:  "backup",
			Usage: "Backup",
			Subcommands: []cli.Command{
				cli.Command{
					Name:  "export",
					Usage: "Export",
					Flags: []cli.Flag{
						cli.StringFlag{Name: "uri, u", Usage: "where to export to (file://)"},
						cli.StringFlag{Name: "password, p", Usage: "password"},
					},
					Action: func(c *cli.Context) error {
						if c.String("uri") == "" {
							return errors.Errorf("specify --uri")
						}

						password := c.String("password")
						if len(password) == 0 {
							p, err := readVerifyPassword("Enter a backup password:")
							if err != nil {
								return err
							}
							password = p
						}

						req := &BackupExportRequest{
							URI:      c.String("uri"),
							Password: password,
						}
						resp, err := client.KeysClient().BackupExport(context.TODO(), req)
						if err != nil {
							return err
						}
						fmt.Printf("Saved to %s\n", resp.URI)
						return nil
					},
				},
				cli.Command{
					Name:  "import",
					Usage: "Import",
					Flags: []cli.Flag{
						cli.StringFlag{Name: "uri, u", Usage: "where to export to (file://)"},
					},
					Action: func(c *cli.Context) error {
						password := c.String("password")
						if len(password) == 0 {
							p, err := readPassword("Enter the password:")
							if err != nil {
								return err
							}
							password = p
						}

						req := &BackupImportRequest{
							URI:      c.String("uri"),
							Password: password,
						}
						_, err := client.KeysClient().BackupImport(context.TODO(), req)
						if err != nil {
							return err
						}
						return nil
					},
				},
			},
		},
	}
}
