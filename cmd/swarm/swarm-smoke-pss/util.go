// Copyright 2018 The go-ethereum Authors
// This file is part of go-ethereum.
//
// go-ethereum is free software: you can redistribute it and/or modify
// it under the terms of the GNU General Public License as published by
// the Free Software Foundation, either version 3 of the License, or
// (at your option) any later version.
//
// go-ethereum is distributed in the hope that it will be useful,
// but WITHOUT ANY WARRANTY; without even the implied warranty of
// MERCHANTABILITY or FITNESS FOR A PARTICULAR PURPOSE. See the
// GNU General Public License for more details.
//
// You should have received a copy of the GNU General Public License
// along with go-ethereum. If not, see <http://www.gnu.org/licenses/>.

package main

import (
	"fmt"
	"math/rand"
	"os"
	"strings"
	"time"

	"github.com/ethereum/go-ethereum/log"
	"github.com/ethereum/go-ethereum/metrics"
	"github.com/pborman/uuid"
	cli "gopkg.in/urfave/cli.v1"
)

var (
	commandName = ""
	seed        = int(time.Now().UTC().UnixNano())
)

func init() {
	rand.Seed(int64(seed))
}

func wsEndpoint(host string) string {
	return fmt.Sprintf("ws://%s:%d", host, wsPort)
}

func wrapCliCommand(name string, command func(*cli.Context, string) error) func(*cli.Context) error {
	return func(ctx *cli.Context) error {
		log.PrintOrigins(true)
		log.Root().SetHandler(log.LvlFilterHandler(log.Lvl(verbosity), log.StreamHandler(os.Stdout, log.TerminalFormat(false))))

		// test uuid
		tuid := uuid.New()[:8]

		commandName = name

		hosts = strings.Split(allhosts, ",")

		defer func(now time.Time) {
			totalTime := time.Since(now)
			log.Info("total time", "tuid", tuid, "time", totalTime)
			metrics.GetOrRegisterResettingTimer(name+".total-time", nil).Update(totalTime)
		}(time.Now())

		log.Info("smoke test starting", "tuid", tuid, "task", name, "timeout", timeout)
		metrics.GetOrRegisterCounter(name, nil).Inc(1)

		return command(ctx, tuid)
	}
}