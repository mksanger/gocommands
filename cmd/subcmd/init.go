package subcmd

import (
	"github.com/cyverse/gocommands/cmd/flag"
	"github.com/cyverse/gocommands/commons"
	"github.com/spf13/cobra"
	"golang.org/x/xerrors"

	irodsclient_icommands "github.com/cyverse/go-irodsclient/icommands"
	irodsclient_types "github.com/cyverse/go-irodsclient/irods/types"
)

var initCmd = &cobra.Command{
	Use:     "init",
	Aliases: []string{"iinit"},
	Short:   "Initialize gocommands",
	Long: `This sets up iRODS Host and access account for other gocommands tools. 
	Once the configuration is set, configuration files are created under ~/.irods directory.
	The configuration is fully-compatible to that of icommands`,
	RunE: processInitCommand,
	Args: cobra.NoArgs,
}

func AddInitCommand(rootCmd *cobra.Command) {
	// attach common flags
	flag.SetCommonFlags(initCmd, false)
	flag.SetInitFlags(initCmd)

	rootCmd.AddCommand(initCmd)
}

func processInitCommand(command *cobra.Command, args []string) error {
	init, err := NewInitCommand(command, args)
	if err != nil {
		return err
	}

	return init.Process()
}

type InitCommand struct {
	command *cobra.Command

	initFlagValues *flag.InitFlagValues

	environmentManager *irodsclient_icommands.ICommandsEnvironmentManager
	account            *irodsclient_types.IRODSAccount
}

func NewInitCommand(command *cobra.Command, args []string) (*InitCommand, error) {
	init := &InitCommand{
		command: command,

		initFlagValues: flag.GetInitFlagValues(),
	}

	return init, nil
}

func (init *InitCommand) Process() error {
	cont, err := flag.ProcessCommonFlags(init.command)
	if err != nil {
		return xerrors.Errorf("failed to process common flags: %w", err)
	}

	if !cont {
		return nil
	}

	init.environmentManager = commons.GetEnvironmentManager()

	// handle local flags
	updated, err := commons.ReinputFields()
	if err != nil {
		return xerrors.Errorf("failed to input fields: %w", err)
	}

	init.account, err = init.environmentManager.ToIRODSAccount()
	if err != nil {
		return xerrors.Errorf("failed to get iRODS account info from iCommands Environment: %w", err)
	}

	// update PAM TTL
	init.account.PamTTL = init.initFlagValues.PamTTL

	// test connect
	conn, err := commons.GetIRODSConnection(init.account)
	if err != nil {
		return xerrors.Errorf("failed to connect to iRODS server: %w", err)
	}
	defer conn.Disconnect()

	if init.account.AuthenticationScheme == irodsclient_types.AuthSchemePAM {
		// update pam token
		init.environmentManager.PamToken = conn.GetPAMToken()
	}

	if updated {
		// save
		err := commons.GetEnvironmentManager().SaveEnvironment()
		if err != nil {
			return xerrors.Errorf("failed to save iCommands Environment: %w", err)
		}
	} else {
		commons.Println("gocommands is already configured for following account:")
		err := commons.PrintAccount()
		if err != nil {
			return xerrors.Errorf("failed to print account info: %w", err)
		}
	}

	return nil
}
