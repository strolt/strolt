package restic

import (
	"encoding/json"
	"errors"
	"os"
	"os/exec"
	"strings"
	"time"

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
	ModTime     time.Time   `json:"mtime,omitempty"`
	AccessTime  time.Time   `json:"atime,omitempty"`
	ChangeTime  time.Time   `json:"ctime,omitempty"`
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
		return lsItem{}, err
	}

	switch itemType.StructType {
	case "snapshot":
		var snapshot lsSnapshot

		if err := json.Unmarshal([]byte(line), &snapshot); err != nil {
			return lsItem{}, err
		}

		return lsItem{
			Snapshot: &snapshot,
			Type:     lsItemTypeSnapshot,
		}, nil

	case "node":
		var node lsNode

		if err := json.Unmarshal([]byte(line), &node); err != nil {
			return lsItem{}, err
		}

		return lsItem{
			Node: &node,
			Type: lsItemTypeNode,
		}, nil

	default:
		return lsItem{}, errors.New("unknown item type")
	}
}

func (i *Restic) ls(_ context.Context, snapshotID string, path string) ([]lsItem, error) {
	var args []string
	args = append(args, i.getGlobalFlags()...)
	args = append(args, "ls", snapshotID, path)

	cmd := exec.Command(i.getBin(), args...)

	env, err := i.getEnv()
	if err != nil {
		return []lsItem{}, err
	}

	cmd.Env = env

	output, err := cmd.Output()
	if err != nil {
		return []lsItem{}, err
	}

	list := []lsItem{}

	for _, line := range strings.Split(string(output), "\n") {
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

func (i *Restic) getFilenameForRestorePipe(ctx context.Context, snapshotID string) (string, string, error) {
	list, err := i.ls(ctx, snapshotID, "/")
	if err != nil {
		return "", "", err
	}

	for _, item := range list {
		if item.Type == lsItemTypeNode && item.Node.Type == lsNodeTypeFile {
			return item.Node.Name, item.Node.Path, nil
		}
	}

	return "", "", errors.New("not found file")
}
