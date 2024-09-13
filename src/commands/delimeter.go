package commands

import (
	"fmt"
	"strings"

	"github.com/costaluu/flag/constants"
	"github.com/costaluu/flag/core"
	"github.com/costaluu/flag/logger"
	"github.com/urfave/cli/v2"
)

var DelimeterListCommand *cli.Command = &cli.Command{
	Name:  "list",
	Usage: "list all delimeters",
	Action: func(ctx *cli.Context) error {
		core.ListDelimeters()
		return nil
	},
}

var DelimeterSetCommand *cli.Command = &cli.Command{
	Name:  "set",
	Usage: "creates or updates an existing delimeter",
	ArgsUsage: `<file_extension> <delimeter_start> <delimeter_end>`,
	Action: func(ctx *cli.Context) error {
		args := ctx.Args().Slice()
		
		if len(args) < 3 {
			logger.Info[string](fmt.Sprintf("usage: %s delimeter %s", constants.COMMAND, ctx.Command.ArgsUsage))
			
			return nil
		}

		extension := args[0]

		if !strings.HasPrefix(extension, ".") {
			logger.Info[string]("invalid extension")
		}

		core.SetDelimeter(extension, args[1], args[2])

		logger.Success[string](fmt.Sprintf("delimeter for file extension %s seted", extension))

		return nil
	},
}

var DelimeterDeleteCommand *cli.Command = &cli.Command{
	Name:  "delete",
	Usage: "deletes an existing delimeter",
	ArgsUsage: `<file_extension>`,
	Action: func(ctx *cli.Context) error {
		args := ctx.Args().Slice()
		
		if len(args) != 3 {
			logger.Info[string](fmt.Sprintf("usage: %s delimeter %s", constants.COMMAND, ctx.Command.ArgsUsage))
			
			return nil
		}

		extension := args[0]

		if !strings.HasPrefix(extension, ".") {
			logger.Result[string]("invalid extension")
		}

		if extension == "default" {
			logger.Result[string]("invalid extension")
		}

		core.DeleteDelimeter(extension)

		logger.Success[string](fmt.Sprintf("delimeter for file extension %s deleted", extension))

		return nil
	},
}

var DelimeterCommand *cli.Command = &cli.Command{
	Name: "delimeters",
	Usage: "operations for delimeters",
	Subcommands: []*cli.Command{
		DelimeterListCommand,
		DelimeterSetCommand,
		DelimeterDeleteCommand,
	},
}