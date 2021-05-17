package main

import "strings"

var cfg config

type config struct {
	Groups              map[string][]string // group -> tables
	SkipLinkMDExtension bool
	LinkPrefix          string
}

func tableNameAndGroup(name string) (string, string) {
	id := strings.ToLower(name)
	for group, tables := range cfg.Groups {
		for _, table := range tables {
			if strings.ToLower(table) == id {
				return table, group
			}
		}
	}
	return name, ""
}

func groupTables(data Data) map[string]Data {
	isGrouped := map[string]bool{}
	groupedData := map[string]Data{}
	for group, tables := range cfg.Groups {
		groupData := Data{}
		for _, tableName := range tables {
			if table, ok := data.tableByName(tableName); ok {
				isGrouped[tableName] = true
				// table.Group could be wrong if table is in multiple groups
				table.Group = group
				table.deriveLinks()
				groupData.Tables = append(groupData.Tables, table)
			}
		}
		groupedData[group] = groupData
	}
	ungroupedData := Data{}
	for _, table := range data.Tables {
		if !isGrouped[table.Name] {
			table.deriveLinks()
			ungroupedData.Tables = append(ungroupedData.Tables, table)
		}
	}
	if len(ungroupedData.Tables) > 0 {
		groupedData["Ungrouped"] = ungroupedData
	}
	return groupedData
}
