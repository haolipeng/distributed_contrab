package common

import "encoding/json"

//任务信息
type Job struct {
	Name     string `json:"name"`    //任务名
	Command  string `json:"command"` //shell命令
	CronExpr string `json:"cronExpr"`
}

type Response struct {
	Errno int         `json:"errno"`
	Msg   string      `json:"msg"`
	Data  interface{} `json:"data"`
}

func BuildResponse(errno int, msg string, data interface{}) ([]byte, error) {
	var (
		resp    Response
		err     error
		content []byte
	)

	//变量赋值
	resp.Errno = errno
	resp.Msg = msg
	resp.Data = data

	//json序列化
	content, err = json.Marshal(resp)

	return content, err
}
