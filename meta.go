package runner

import (
	"encoding/json"
	"io/ioutil"
	"os"
	"path"
)

// MetaInfo 是当前环境的信息，由 CI 系统自动生成。
type MetaInfo struct {
	Project     string `json:"project"`
	Namespace   string `json:"namespace"`
	Env         string `json:"env"`
	Type        string `json:"type"`
	GitRefName  string `json:"git_ref_name"`
	GitRevision string `json:"git_revision"`
}

var metaInfo MetaInfo

const metaPath = ".meta.json"

func parseMetaInfo() {
	// 设置默认值。
	metaInfo.Project = path.Base(os.Args[0])
	metaInfo.Env = "development"
	metaInfo.Type = "debug"
	metaInfo.GitRefName = "00000000"
	metaInfo.GitRevision = "00000000"

	f, err := os.Open(metaPath)

	if err != nil {
		return
	}

	defer f.Close()
	data, err := ioutil.ReadAll(f)

	if err != nil {
		return
	}

	json.Unmarshal(data, &metaInfo)
}

// Meta 返回项目的 meta 信息。
func Meta() *MetaInfo {
	return &metaInfo
}
