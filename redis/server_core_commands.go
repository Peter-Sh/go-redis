// Copyright (C) 2022 Satoshi Konno All rights reserved.
//
// Licensed under the Apache License, Version 2.0 (the "License");
// you may not use this file except in compliance with the License.
// You may obtain a copy of the License at
//
//    http://www.apache.org/licenses/LICENSE-2.0
//
// Unless required by applicable law or agreed to in writing, software
// distributed under the License is distributed on an "AS IS" BASIS,
// WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
// See the License for the specific language governing permissions and
// limitations under the License.

package redis

import (
	"errors"
	"time"

	"github.com/cybergarage/go-redis/redis/proto"
)

// handleCommand handles a client command message.
// nolint: gocyclo, maintidx
func (server *Server) initCoreCommandExecutors() {
	// Sets connection management commands.

	server.commandExecutors["PING"] = func(ctx *DBContext, cmd string, args commandArgs) (*Message, error) {
		arg := ""
		var err error
		if msg, _ := args.Next(); msg != nil {
			arg, err = msg.String()
			if err != nil {
				return nil, err
			}
		}
		return server.systemCommandHandler.Ping(ctx, arg)
	}

	server.commandExecutors["ECHO"] = func(ctx *DBContext, cmd string, args commandArgs) (*Message, error) {
		msg, err := args.NextString()
		if err != nil {
			return nil, newMissingArgumentError(cmd, "msg", err)
		}
		return server.systemCommandHandler.Echo(ctx, msg)
	}

	server.commandExecutors["SELECT"] = func(ctx *DBContext, cmd string, args commandArgs) (*Message, error) {
		id, err := args.NextInteger()
		if err != nil {
			return nil, newMissingArgumentError(cmd, "id", err)
		}
		return server.systemCommandHandler.Select(ctx, id)
	}

	server.commandExecutors["QUIT"] = func(ctx *DBContext, cmd string, args commandArgs) (*Message, error) {
		return server.systemCommandHandler.Quit(ctx)
	}

	// Sets string commands.

	parseHashArg := func(cmd string, args commandArgs) (string, error) {
		hash, err := args.NextString()
		if err != nil {
			return "", newMissingArgumentError(cmd, "hash", err)
		}
		return hash, nil
	}

	parseKeyArg := func(cmd string, args commandArgs) (string, error) {
		key, err := args.NextString()
		if err != nil {
			return "", newMissingArgumentError(cmd, "key", err)
		}
		return key, nil
	}

	parseSetArgs := func(cmd string, args commandArgs) (string, string, error) {
		key, err := args.NextString()
		if err != nil {
			return "", "", newMissingArgumentError(cmd, "key", err)
		}
		val, err := args.NextString()
		if err != nil {
			return "", "", newMissingArgumentError(cmd, "value", err)
		}
		return key, val, err
	}

	parseMSetArgs := func(cmd string, args commandArgs) (map[string]string, error) {
		var key, val string
		var err error
		dir := map[string]string{}
		key, err = args.NextString()
		for err == nil {
			val, err = args.NextString()
			if err != nil {
				newMissingArgumentError(cmd, key, err)
			}
			dir[key] = val
			key, err = args.NextString()
		}
		if !errors.Is(err, proto.ErrEOM) {
			return nil, err
		}
		return dir, nil
	}

	parseMGetArgs := func(cmd string, args commandArgs) ([]string, error) {
		var key string
		var err error
		keys := []string{}
		key, err = args.NextString()
		for err == nil {
			keys = append(keys, key)
			key, err = args.NextString()
		}
		if !errors.Is(err, proto.ErrEOM) {
			return nil, err
		}
		return keys, nil
	}

	server.commandExecutors["SET"] = func(ctx *DBContext, cmd string, args commandArgs) (*Message, error) {
		now := time.Now()
		opt := SetOption{
			NX:      false,
			XX:      false,
			EX:      0,
			PX:      0,
			EXAT:    now,
			PXAT:    now,
			KEEPTTL: false,
			GET:     false,
		}
		key, val, err := parseSetArgs(cmd, args)
		if err != nil {
			return nil, err
		}
		return server.userCommandHandler.Set(ctx, key, val, opt)
	}

	server.commandExecutors["SETNX"] = func(ctx *DBContext, cmd string, args commandArgs) (*Message, error) {
		now := time.Now()
		opt := SetOption{
			NX:      true,
			XX:      false,
			EX:      0,
			PX:      0,
			EXAT:    now,
			PXAT:    now,
			KEEPTTL: false,
			GET:     false,
		}
		key, val, err := parseSetArgs(cmd, args)
		if err != nil {
			return nil, err
		}
		return server.userCommandHandler.Set(ctx, key, val, opt)
	}
	server.commandExecutors["GET"] = func(ctx *DBContext, cmd string, args commandArgs) (*Message, error) {
		opt := GetOption{}
		key, err := parseKeyArg(cmd, args)
		if err != nil {
			return nil, err
		}
		return server.userCommandHandler.Get(ctx, key, opt)
	}
	server.commandExecutors["GETSET"] = func(ctx *DBContext, cmd string, args commandArgs) (*Message, error) {
		now := time.Now()
		opt := SetOption{
			NX:      false,
			XX:      false,
			EX:      0,
			PX:      0,
			EXAT:    now,
			PXAT:    now,
			KEEPTTL: false,
			GET:     true,
		}
		key, val, err := parseSetArgs(cmd, args)
		if err != nil {
			return nil, err
		}
		return server.userCommandHandler.Set(ctx, key, val, opt)
	}

	// Sets hash commands.

	server.commandExecutors["HSET"] = func(ctx *DBContext, cmd string, args commandArgs) (*Message, error) {
		opt := HSetOption{}
		hash, err := parseHashArg(cmd, args)
		if err != nil {
			return nil, err
		}
		key, val, err := parseSetArgs(cmd, args)
		if err != nil {
			return nil, err
		}
		return server.userCommandHandler.HSet(ctx, hash, key, val, opt)
	}
	server.commandExecutors["HGET"] = func(ctx *DBContext, cmd string, args commandArgs) (*Message, error) {
		opt := HGetOption{}
		hash, err := parseHashArg(cmd, args)
		if err != nil {
			return nil, err
		}
		key, err := parseKeyArg(cmd, args)
		if err != nil {
			return nil, err
		}
		return server.userCommandHandler.HGet(ctx, hash, key, opt)
	}
	server.commandExecutors["HGETALL"] = func(ctx *DBContext, cmd string, args commandArgs) (*Message, error) {
		hash, err := parseHashArg(cmd, args)
		if err != nil {
			return nil, err
		}
		return server.userCommandHandler.HGetAll(ctx, hash)
	}
	server.commandExecutors["HMSET"] = func(ctx *DBContext, cmd string, args commandArgs) (*Message, error) {
		opt := HMSetOption{}
		hash, err := parseHashArg(cmd, args)
		if err != nil {
			return nil, err
		}
		dir, err := parseMSetArgs(cmd, args)
		if err != nil {
			return nil, err
		}
		return server.userCommandHandler.HMSet(ctx, hash, dir, opt)
	}
	server.commandExecutors["HMGET"] = func(ctx *DBContext, cmd string, args commandArgs) (*Message, error) {
		opt := HMGetOption{}
		hash, err := parseHashArg(cmd, args)
		if err != nil {
			return nil, err
		}
		keys, err := parseMGetArgs(cmd, args)
		if err != nil {
			return nil, err
		}
		return server.userCommandHandler.HMGet(ctx, hash, keys, opt)
	}
	server.commandExecutors["MSET"] = func(ctx *DBContext, cmd string, args commandArgs) (*Message, error) {
		opt := MSetOption{
			NX: false,
		}
		dir, err := parseMSetArgs(cmd, args)
		if err != nil {
			return nil, err
		}
		return server.userCommandHandler.MSet(ctx, dir, opt)
	}
	server.commandExecutors["MSETNX"] = func(ctx *DBContext, cmd string, args commandArgs) (*Message, error) {
		opt := MSetOption{
			NX: true,
		}
		dir, err := parseMSetArgs(cmd, args)
		if err != nil {
			return nil, err
		}
		return server.userCommandHandler.MSet(ctx, dir, opt)
	}
	server.commandExecutors["MGET"] = func(ctx *DBContext, cmd string, args commandArgs) (*Message, error) {
		opt := MGetOption{}
		keys, err := parseMGetArgs(cmd, args)
		if err != nil {
			return nil, err
		}
		return server.userCommandHandler.MGet(ctx, keys, opt)
	}
}
