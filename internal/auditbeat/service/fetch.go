// Copyright 2024 Emory.Du <orangeduxiaocheng@gmail.com>. All rights reserved.
// Use of this source code is governed by a MIT style
// license that can be found in the LICENSE file.

package service

import (
	"bytes"
	"context"
	"fmt"
	"github.com/emorydu/dbaudit/internal/auditbeat/repository"
	"github.com/emorydu/dbaudit/internal/common"
	"github.com/emorydu/dbaudit/internal/common/utils"
	"path/filepath"
)

const (
	stopped = iota
	startup
)

type ParserType int

const (
	RegexParser ParserType = iota
	Unknown
	JSONParser
)

func (t ParserType) String() string {
	switch t {
	case RegexParser:
		return "regex"
	case JSONParser:
		return "json"
	default:
		return ""
	}
}

type FetchService interface {
	TODO()

	Download(
		context.Context,
		common.OperatingSystemType,
	) ([]byte, error)

	QueryConfigInfo(context.Context, string) ([]byte, error)
	CreateOrModUsage(ctx context.Context, ip string, cpuUse, memUse float64, status, updated int) error
	QueryMonitorInfo(context.Context, string) (int, error)
	Updated(context.Context, string) error
}

type fetchService struct {
	// todo
	ctx  context.Context
	repo repository.Repository
}

var _ FetchService = (*fetchService)(nil)

func NewFetchService(ctx context.Context, repo repository.Repository) FetchService {
	return &fetchService{
		ctx:  ctx,
		repo: repo,
	}
}

func (f *fetchService) Updated(ctx context.Context, ip string) error {
	return f.repo.Update(ctx, ip)
}

func (f *fetchService) QueryMonitorInfo(ctx context.Context, ip string) (int, error) {
	return f.repo.QueryMonitorInfo(ctx, ip)
}

func (f *fetchService) CreateOrModUsage(ctx context.Context, ip string, cpuUse, memUse float64, status, updated int) error {
	return f.repo.InsertOrUpdateMonitor(ctx, ip, cpuUse, memUse, status, updated)
}

func (f *fetchService) QueryConfigInfo(ctx context.Context, ip string) ([]byte, error) {
	info, err := f.repo.FetchConfInfo(ctx, ip)
	if err != nil {
		return nil, err
	}
	if len(info) == 0 {
		return nil, fmt.Errorf("ip: %s not fetch config infos", ip)
	}

	inoutBuffer := new(bytes.Buffer)
	parserBuffer := new(bytes.Buffer)
	for _, v := range info {
		if v.Check == stopped {
			continue
		}
		inoutBuffer.Write(builderSingleConf(v.AgentPath, v.IndexName, "kafka:9092", v.MultiParse))
		name := v.IndexName
		if v.MultiParse == startup {
			name = "multiline"
		}
		parserBuffer.Write(builderSingleParserConf(name, ParserType(v.ParseType), v.RegexParamValue))
	}

	inoutBuffer.Write([]byte(common.InParserConn))
	inoutBuffer.Write(parserBuffer.Bytes())

	return inoutBuffer.Bytes(), nil
}

func builderSingleParserConf(name string, parserType ParserType, regexValue string) []byte {
	parser := ""
	if parserType == RegexParser {
		parser = fmt.Sprintf(`
[PARSER]
	Name %s
	Format %s
	Regex %s
`, name, parserType.String(), regexValue)
	} else {
		parser = fmt.Sprintf(`
[PARSER]
	Name %s
	Format %s
`, name, parserType.String())
	}

	return []byte(parser)
}

func builderSingleConf(path, indexName, others string, multipart int8) []byte {
	inputBlock := ""
	if multipart == stopped {
		inputBlock = fmt.Sprintf(`
[INPUT]
	Name tail
	Path %s
	Tag %s
	Read_From_Head true
	(insert) %s
`, path, indexName, indexName)
	} else {
		inputBlock = fmt.Sprintf(`
[INPUT]
	Name tail
	Multiline On
	Path %s
	Parser_Firstline multiline
	Skip_Empty_Lines on
	Tag %s
	Read_From_Head true
	(insert) %s 
`, path, indexName, indexName)
	}

	filterBlock := fmt.Sprintf(`
[FILTER]
	Name parser
	Match %s
	Key_Name log
	Parser %s
	Reserve_Data on
`, indexName, indexName)

	outputBlock := fmt.Sprintf(`
[OUTPUT]
	Name kafka
	Match %s
	Brokers %s
	Topics data_%s
`, indexName, others, indexName)

	return []byte(inputBlock + filterBlock + outputBlock)
}

func (f *fetchService) TODO() {}

func (f *fetchService) Download(_ context.Context, systemType common.OperatingSystemType) ([]byte, error) {
	// TODO
	path := ""
	switch systemType {
	case common.Linux:
		path = "linux agent install path"
	case common.Windows:
		path = "windows agent install path"
	default:
		return nil, ErrSupportPlatform
	}

	data, err := utils.ReadFromDisk(filepath.Join(path, "updates", "", ""))
	if err != nil {
		return nil, ErrPathExists
	}

	return data, nil
}
