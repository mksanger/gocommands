package subcmd

import (
	"fmt"

	irodsclient_fs "github.com/cyverse/go-irodsclient/fs"
	"github.com/cyverse/gocommands/commons"
	log "github.com/sirupsen/logrus"
	"github.com/spf13/cobra"
)

var cdCmd = &cobra.Command{
	Use:   "cd [collection1]",
	Short: "Change current working iRODS collection",
	Long:  `This changes current working iRODS collection.`,
	RunE:  processCdCommand,
}

func AddCdCommand(rootCmd *cobra.Command) {
	// attach common flags
	commons.SetCommonFlags(cdCmd)

	rootCmd.AddCommand(cdCmd)
}

func processCdCommand(command *cobra.Command, args []string) error {
	logger := log.WithFields(log.Fields{
		"package":  "main",
		"function": "processCdCommand",
	})

	cont, err := commons.ProcessCommonFlags(command)
	if err != nil {
		logger.Error(err)
	}

	if !cont {
		return err
	}

	// handle local flags
	_, err = commons.InputMissingFields()
	if err != nil {
		logger.Error(err)
		return err
	}

	// Create a file system
	account := commons.GetAccount()
	filesystem, err := commons.GetIRODSFSClient(account)
	if err != nil {
		return err
	}

	defer filesystem.Release()

	if len(args) == 0 {
		// do nothing
		return nil
	} else if len(args) >= 2 {
		return fmt.Errorf("too many arguments (%d) are given", len(args))
	} else {
		// cd
		err = changeWorkingDir(filesystem, args[0])
		if err != nil {
			return err
		}
	}
	return nil
}

func changeWorkingDir(filesystem *irodsclient_fs.FileSystem, collectionPath string) error {
	logger := log.WithFields(log.Fields{
		"package":  "main",
		"function": "changeWorkingDir",
	})

	cwd := commons.GetCWD()
	collectionPath = commons.MakeIRODSPath(cwd, collectionPath)

	logger.Debugf("changing working dir: %s\n", collectionPath)

	if !filesystem.ExistsDir(collectionPath) {
		return fmt.Errorf("collection %s does not exist", collectionPath)
	}

	commons.SetCWD(collectionPath)
	return nil
}