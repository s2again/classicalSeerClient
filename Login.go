package snet

import (
	"crypto/md5"
	"fmt"
	"log"

	"github.com/fanliao/go-promise"

	"github.com/s2again/snet/core"
)

type LoginResponse struct {
	SessionID   [16]byte
	RoleCreated bool
}

// unimplemented
func (c *GuideServerConnection) LoginByEmail(email string, password string) (prom *promise.Promise) {
	const channel uint32 = 0
	const gameType uint32 = 1
	prom = promise.NewPromise()
	c.SetSession(0, [16]byte{})
	c.SendInPromise(Command_MAIN_LOGIN_IN, emailBytes(email), pwdHashBytes(password), channel, gameType, uint32(0)).
		OnSuccess(func(v interface{}) {
			body := v.(core.PacketBody)
			resp, err := parseLoginResponseFromGuide(body)
			if err == nil {
				log.Printf("LoginResponse %X\n", v.(core.PacketBody).Bytes())
				// todo: 从head获取uid
				// c.SetSession(uid, resp.SessionID)
				prom.Resolve(resp)
			} else {
				prom.Reject(fmt.Errorf("LoginByEmail Rejected, reason: %v", err))
			}
		}).
		OnFailure(func(v interface{}) {
			prom.Reject(fmt.Errorf("登录失败，可能密码错误。 %v", v))
		})
	return prom
}

func (c *GuideServerConnection) Login(uid uint32, password string) (prom *promise.Promise) {
	const channel uint32 = 0
	const gameType uint32 = 1
	prom = promise.NewPromise()
	c.SetSession(uid, [16]byte{})
	c.SendInPromise(Command_MAIN_LOGIN_IN, pwdHashBytes(password), channel, gameType, uint32(0)).
		OnSuccess(func(v interface{}) {
			body := v.(core.PacketBody)
			resp, err := parseLoginResponseFromGuide(body)
			if err == nil {
				log.Printf("LoginResponse %X\n", v.(core.PacketBody).Bytes())
				c.SetSession(uid, resp.SessionID)
				prom.Resolve(resp)
			} else {
				prom.Reject(fmt.Errorf("Login Rejected, reason: %v", err))
			}
		}).
		OnFailure(func(v interface{}) {
			prom.Reject(fmt.Errorf("登录失败，可能密码错误。 %v", v))
		})
	return prom
}

func emailBytes(email string) (emailBytes [64]byte) {
	copy(emailBytes[:], []byte(email)[:64])
	return
}

func pwdHashBytes(password string) (hashBytes [32]byte) {
	t := fmt.Sprintf("%x", md5.Sum([]byte(password)))
	copy(hashBytes[:], []byte(t)[:32])
	return
}

func parseLoginResponseFromGuide(buffer core.PacketBody) (info LoginResponse, err error) {
	defer func() {
		if x := recover(); x != nil {
			err = x.(error)
			return
		}
	}()
	var t struct {
		SessionID   [16]byte
		RoleCreated uint32
	}
	core.MustBinaryRead(buffer, &t)
	info = LoginResponse{
		SessionID:   t.SessionID,
		RoleCreated: t.RoleCreated != 0,
	}
	return
}
