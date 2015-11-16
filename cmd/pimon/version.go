// Copyright (c) 2015, Ben Morgan. All rights reserved.
// Use of this source code is governed by an MIT license
// that can be found in the LICENSE file.

package main

import (
	"io"
	"os"
	"text/template"
	"time"

	"github.com/spf13/cobra"
)

var VersionCmd = &cobra.Command{
	Use:   "version",
	Short: "show version and date information",
	Long:  "Show the official version number of pimon, as well as the release date.",
	Run: func(cmd *cobra.Command, args []string) {
		writeVersion(os.Stdout)
	},
}

type programInfo struct {
	Name      string
	Author    string
	Email     string
	Version   string
	Date      string
	Homepage  string
	Copyright string
	License   string
}

var versionTmpl = template.Must(template.New("version").Parse(
	`{{.Name}} version {{.Version}} ({{.Date}})
Copyright {{.Copyright}}, {{.Author}} <{{.Email}}>

You may find {{.Name}} on the Internet at

    {{.Homepage}}

Please report any bugs you may encounter.

The source code of {{.Name}} is licensed under the {{.License}} license.
`))

var progInfo = programInfo{
	Name:      "pimon",
	Author:    "Ben Morgan",
	Email:     "neembi@gmail.com",
	Version:   "0.1",
	Date:      time.Now().Format("2 January 2006"),
	Copyright: time.Now().Format("2006"),
	Homepage:  "https://github.com/cassava/pillr",
	License:   "MIT",
}

func writeVersion(w io.Writer) {
	versionTmpl.Execute(w, progInfo)
}
