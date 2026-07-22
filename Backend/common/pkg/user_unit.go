package pkg

import "golang.org/x/crypto/bcrypt"

// 加密明文密码，返回哈希串
func EncryptPwd(rawPwd string) (string, error) {
	hash, err := bcrypt.GenerateFromPassword([]byte(rawPwd), 12)
	return string(hash), err
}

// 校验输入密码和库中哈希是否匹配
func CheckPwd(hashPwd, rawPwd string) bool {
	err := bcrypt.CompareHashAndPassword([]byte(hashPwd), []byte(rawPwd))
	return err == nil
}
