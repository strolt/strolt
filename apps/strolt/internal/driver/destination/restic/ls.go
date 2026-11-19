package restic

import (
	"encoding/json"
	"errors"
	"fmt"
	"os"
	"strings"
	"time"

	"github.com/strolt/strolt/shared/process"

	"github.com/strolt/strolt/apps/strolt/internal/context"
)

type lsSnapshot struct {
	ID      string `json:"id"`
	ShortID string `json:"short_id"`
	// StructType string `json:"struct_type"` // "snapshot"
}

type lsNodeType string

const (
	lsNodeTypeDir  lsNodeType = "dir"
	lsNodeTypeFile lsNodeType = "file"
)

type lsNode struct {
	Name        string      `json:"name"`
	Type        lsNodeType  `json:"type"`
	Path        string      `json:"path"`
	UID         uint32      `json:"uid"`
	GID         uint32      `json:"gid"`
	Size        *uint64     `json:"size,omitempty"`
	Mode        os.FileMode `json:"mode,omitempty"`
	Permissions string      `json:"permissions,omitempty"`
	ModTime     time.Time   `json:"mtime,omitzero"`
	AccessTime  time.Time   `json:"atime,omitzero"`
	ChangeTime  time.Time   `json:"ctime,omitzero"`
	// StructType  string      `json:"struct_type"` // "node"
}

type lsType struct {
	StructType string `json:"struct_type"`
}

type lsItemType string

const (
	lsItemTypeNode     lsItemType = "node"
	lsItemTypeSnapshot lsItemType = "snapshot"
)

type lsItem struct {
	Node     *lsNode
	Snapshot *lsSnapshot
	Type     lsItemType
}

func parseLsItem(line string) (lsItem, error) {
	var itemType lsType
	if err := json.Unmarshal([]byte(line), &itemType); err != nil {
		return lsItem{Node: nil, Snapshot: nil, Type: ""}, fmt.Errorf("unmarshal ls item type: %w", err)
	}

	switch itemType.StructType {
	case "snapshot":
		var snapshot lsSnapshot

		if err := json.Unmarshal([]byte(line), &snapshot); err != nil {
			return lsItem{Node: nil, Snapshot: nil, Type: ""}, fmt.Errorf("unmarshal ls snapshot: %w", err)
		}

		return lsItem{
			Node:     nil,
			Snapshot: &snapshot,
			Type:     lsItemTypeSnapshot,
		}, nil

	case "node":
		var node lsNode

		if err := json.Unmarshal([]byte(line), &node); err != nil {
			return lsItem{Node: nil, Snapshot: nil, Type: ""}, fmt.Errorf("unmarshal ls node: %w", err)
		}

		return lsItem{
			Node:     &node,
			Snapshot: nil,
			Type:     lsItemTypeNode,
		}, nil

	default:
		return lsItem{Node: nil, Snapshot: nil, Type: ""}, errors.New("unknown item type")
	}
}

func (i *Restic) ls(ctx context.Context, snapshotID string, path string) ([]lsItem, error) {
	globalFlags := i.getGlobalFlags()

	args := make([]string, 0, len(globalFlags)+3)
	args = append(args, globalFlags...)
	args = append(args, "ls", snapshotID, path)

	cmd := process.CommandContext(ctx.ProcessContext(), i.getBin(), args...)

	env, err := i.getEnv()
	if err != nil {
		return []lsItem{}, err
	}

	cmd.Env = env

	output, err := cmd.Output()
	if err != nil {
		return []lsItem{}, fmt.Errorf("restic ls: %w", err)
	}

	list := []lsItem{}

	for line := range strings.SplitSeq(string(output), "\n") {
		if line == "" {
			continue
		}

		item, err := parseLsItem(line)
		if err != nil {
			return []lsItem{}, err
		}

		list = append(list, item)
	}

	return list, nil
}

func (i *Restic) getFilenameForRestorePipe(ctx context.Context, snapshotID string) (string, string, uint64, bool, error) {
	list, err := i.ls(ctx, snapshotID, "/")
	if err != nil {
		return "", "", 0, false, err
	}

	for _, item := range list {
		if item.Type == lsItemTypeNode && item.Node.Type == lsNodeTypeFile {
			var size uint64
			hasSize := false
			if item.Node.Size != nil {
				size, hasSize = *item.Node.Size, true
			}
			return item.Node.Name, item.Node.Path, size, hasSize, nil
		}
	}

	return "", "", 0, false, errors.New("not found file")
}
