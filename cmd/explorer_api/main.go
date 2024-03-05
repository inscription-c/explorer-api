package main

import (
	"fmt"
	"github.com/btcsuite/btcd/btcutil"
	"github.com/btcsuite/btcwallet/netparams"
	"github.com/inscription-c/cins/btcd/rpcclient"
	"github.com/inscription-c/cins/pkg/signal"
	"github.com/inscription-c/cins/pkg/util"
	"github.com/inscription-c/explorer-api/config"
	"github.com/inscription-c/explorer-api/dao"
	"github.com/inscription-c/explorer-api/dao/indexer"
	"github.com/inscription-c/explorer-api/handle"
	"github.com/inscription-c/explorer-api/log"
	"github.com/inscription-c/explorer-api/runner"
	"github.com/inscription-c/explorer-api/tables"
	"github.com/spf13/cobra"
	"os"
	"path/filepath"
)

var Cmd = &cobra.Command{
	Use:   "explorer-api",
	Short: "explorer-api of the inscription",
	Run: func(cmd *cobra.Command, args []string) {
		if err := ExplorerApi(); err != nil {
			fmt.Println(err)
			os.Exit(1)
		}
	},
}

var configFilePath string

func init() {
	Cmd.Flags().StringVarP(&configFilePath, "config", "c", "./config/config.yaml", "config file path")
}

func main() {
	if err := Cmd.Execute(); err != nil {
		os.Exit(1)
	}
}

func ExplorerApi() error {
	if err := config.Init(configFilePath); err != nil {
		return err
	}
	if config.Cfg.Server.Testnet {
		util.ActiveNet = &netparams.TestNet3Params
	}

	logDir := filepath.Join(config.Cfg.Server.Name, "explorer.log")
	logFile := btcutil.AppDataDir(logDir, false)
	log.InitLogRotator(logFile)

	db, err := dao.NewDB(
		dao.WithAddr(config.Cfg.DB.Mysql.Addr),
		dao.WithUser(config.Cfg.DB.Mysql.User),
		dao.WithPassword(config.Cfg.DB.Mysql.Password),
		dao.WithDBName(config.Cfg.DB.Mysql.DB),
		dao.WithAutoMigrateTables(tables.Tables...),
	)
	if err != nil {
		return err
	}

	indexerDB, err := indexer.NewDB(
		indexer.WithAddr(config.Cfg.DB.Indexer.Addr),
		indexer.WithUser(config.Cfg.DB.Indexer.User),
		indexer.WithPassword(config.Cfg.DB.Indexer.Password),
		indexer.WithDBName(config.Cfg.DB.Indexer.DB),
	)
	if err != nil {
		return err
	}

	cli, err := rpcclient.NewClient(
		rpcclient.WithClientHost(config.Cfg.Chain.Url),
		rpcclient.WithClientUser(config.Cfg.Chain.Username),
		rpcclient.WithClientPassword(config.Cfg.Chain.Password),
	)
	if err != nil {
		return err
	}

	// runner
	blockRunner := runner.NewRunner(
		runner.WithClient(cli),
		runner.WithDB(db),
		runner.WithIndexerDB(indexerDB),
		runner.WithStartHeight(config.Cfg.Chain.StartHeight),
	)
	blockRunner.Start()
	signal.AddInterruptHandler(func() {
		_ = blockRunner.Wait()
	})

	// http handler
	handler, err := handle.New(
		handle.WithAddr(config.Cfg.Server.RpcListen),
		handle.WithTestNet(config.Cfg.Server.Testnet),
		handle.WithClient(cli),
		handle.WithDB(db),
		handle.WithIndexerDB(indexerDB),
	)
	if err != nil {
		return err
	}
	if err := handler.Run(); err != nil {
		return err
	}
	return nil
}
