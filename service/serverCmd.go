package service

import (
	"errors"
	"fmt"
	"github.com/lejianwen/rustdesk-api/v2/model"
	"io"
	"net"
	"strings"
	"time"
)

const (
	serverCmdMaxLength       = 64
	serverCmdOptionMaxLength = 512
	serverCmdDialTimeout     = 5 * time.Second
	serverCmdReadBufSize     = 64 * 1024
)

type ServerCmdService struct{}

// List
func (is *ServerCmdService) List(page, pageSize uint) (res *model.ServerCmdList) {
	res = &model.ServerCmdList{}
	res.Page = int64(page)
	res.PageSize = int64(pageSize)
	tx := DB.Model(&model.ServerCmd{})
	tx.Count(&res.Total)
	tx.Scopes(Paginate(page, pageSize))
	tx.Find(&res.ServerCmds)
	return
}

// Info
func (is *ServerCmdService) Info(id uint) *model.ServerCmd {
	u := &model.ServerCmd{}
	DB.Where("id = ?", id).First(u)
	return u
}

// Delete
func (is *ServerCmdService) Delete(u *model.ServerCmd) error {
	return DB.Delete(u).Error
}

// Create
func (is *ServerCmdService) Create(u *model.ServerCmd) error {
	res := DB.Create(u).Error
	return res
}

// ValidateCmdName 校验命令名本身是否合法：只允许字母、数字、下划线、点、短横线
func ValidateCmdName(cmd string) error {
	if cmd == "" {
		return errors.New("cmd cannot be empty")
	}
	if len(cmd) > serverCmdMaxLength {
		return errors.New("cmd is too long")
	}
	for _, r := range cmd {
		ok := (r >= 'a' && r <= 'z') || (r >= 'A' && r <= 'Z') || (r >= '0' && r <= '9') ||
			r == '-' || r == '_' || r == '.'
		if !ok {
			return fmt.Errorf("cmd contains invalid character: %q", r)
		}
	}
	return nil
}

// ValidateOption 校验命令参数：禁止回车、换行与 NUL，长度受限
func ValidateOption(option string) error {
	if len(option) > serverCmdOptionMaxLength {
		return errors.New("option is too long")
	}
	for _, r := range option {
		if r == '\r' || r == '\n' || r == 0 {
			return fmt.Errorf("option contains invalid character: %q", r)
		}
	}
	return nil
}

// IsCmdAllowed 判断命令是否属于该目标的白名单（系统内置命令或数据库中的自定义命令）
func (is *ServerCmdService) IsCmdAllowed(target string, cmd string) bool {
	sysList := model.SysIdServerCmds
	if target == model.ServerCmdTargetRelayServer {
		sysList = model.SysRelayServerCmds
	}
	for _, item := range sysList {
		if item.Target == target && (item.Cmd == cmd || (item.Alias != "" && item.Alias == cmd)) {
			return true
		}
	}

	// 数据库中管理员创建的自定义命令
	var count int64
	DB.Model(&model.ServerCmd{}).
		Where("target = ? and (cmd = ? or (alias <> '' and alias = ?))", target, cmd, cmd).
		Count(&count)
	return count > 0
}

// SendCmd 发送命令
func (is *ServerCmdService) SendCmd(port int, cmd string, arg string) (string, error) {
	if err := ValidateCmdName(cmd); err != nil {
		return "", err
	}
	if err := ValidateOption(arg); err != nil {
		return "", err
	}
	//组装命令
	cmd = strings.TrimSpace(cmd + " " + arg)
	res, err := is.SendSocketCmd("v6", port, cmd)
	if err == nil {
		return res, nil
	}
	//v6连接失败，尝试v4
	res, err = is.SendSocketCmd("v4", port, cmd)
	if err == nil {
		return res, nil
	}
	return "", err
}

// SendSocketCmd
func (is *ServerCmdService) SendSocketCmd(ty string, port int, cmd string) (string, error) {
	addr := "[::1]"
	tcp := "tcp6"
	if ty == "v4" {
		tcp = "tcp"
		addr = "127.0.0.1"
	}
	conn, err := net.DialTimeout(tcp, fmt.Sprintf("%s:%v", addr, port), serverCmdDialTimeout)
	if err != nil {
		Logger.Debugf("%s connect to id server failed: %v", ty, err)
		return "", err
	}
	defer conn.Close()

	if err = conn.SetDeadline(time.Now().Add(serverCmdDialTimeout)); err != nil {
		Logger.Debugf("%s set deadline failed: %v", ty, err)
		return "", err
	}

	//发送命令
	_, err = conn.Write([]byte(cmd))
	if err != nil {
		Logger.Debugf("%s send cmd failed: %v", ty, err)
		return "", err
	}
	//读取返回
	buf := make([]byte, serverCmdReadBufSize)
	n, err := conn.Read(buf)
	if err != nil && !errors.Is(err, io.EOF) {
		Logger.Debugf("%s read response failed: %v", ty, err)
		return "", err
	}
	return string(buf[:n]), nil
}

func (is *ServerCmdService) Update(f *model.ServerCmd) error {
	return DB.Model(f).Updates(f).Error
}
