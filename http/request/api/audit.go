package api

import (
	"encoding/json"
	"github.com/lejianwen/rustdesk-api/v2/global"
	"github.com/lejianwen/rustdesk-api/v2/model"
	"strconv"
)

type AuditConnForm struct {
	Action    string   `json:"action" validate:"omitempty,lte=32"`
	ConnId    int64    `json:"conn_id"`
	Id        string   `json:"id" validate:"omitempty,lte=128"`
	Peer      []string `json:"peer" validate:"omitempty,lte=2,dive,lte=255"`
	Ip        string   `json:"ip" validate:"omitempty,lte=64"`
	SessionId float64  `json:"session_id"`
	Type      int      `json:"type"`
	Uuid      string   `json:"uuid" validate:"omitempty,lte=128"`
}

func (a *AuditConnForm) ToAuditConn() *model.AuditConn {
	fp := ""
	fn := ""
	if len(a.Peer) >= 1 {
		fp = a.Peer[0]
		if len(a.Peer) == 2 {
			fn = a.Peer[1]
		}
	}
	ssid := strconv.FormatFloat(a.SessionId, 'f', -1, 64)
	return &model.AuditConn{
		Action:    a.Action,
		ConnId:    a.ConnId,
		PeerId:    a.Id,
		FromPeer:  fp,
		FromName:  fn,
		Ip:        a.Ip,
		SessionId: ssid,
		Type:      a.Type,
		Uuid:      a.Uuid,
	}
}

type AuditFileForm struct {
	Id     string `json:"id" validate:"omitempty,lte=128"`
	Info   string `json:"info" validate:"omitempty,lte=2048"`
	IsFile bool   `json:"is_file"`
	Path   string `json:"path" validate:"omitempty,lte=4096"`
	PeerId string `json:"peer_id" validate:"omitempty,lte=128"`
	Type   int    `json:"type"`
	Uuid   string `json:"uuid" validate:"omitempty,lte=128"`
}
type AuditFileInfo struct {
	Ip   string `json:"ip"`
	Name string `json:"name"`
	Num  int    `json:"num"`
}

func (a *AuditFileForm) ToAuditFile() *model.AuditFile {
	fi := &AuditFileInfo{}
	err := json.Unmarshal([]byte(a.Info), fi)
	if err != nil {
		global.Logger.Warn("ToAuditFile", err)
	}

	return &model.AuditFile{
		PeerId:   a.Id,
		Info:     a.Info,
		IsFile:   a.IsFile,
		FromPeer: a.PeerId,
		Path:     a.Path,
		Type:     a.Type,
		Uuid:     a.Uuid,
		FromName: fi.Name,
		Ip:       fi.Ip,
		Num:      fi.Num,
	}
}
