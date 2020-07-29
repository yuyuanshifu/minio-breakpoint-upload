package rsa

import (
	"crypto/rand"
	"crypto/rsa"
	"crypto/x509"
	"encoding/pem"
	"errors"
)

//openssl genrsa -out rsa_private_key.pem 1024
var privateKey = []byte(`  
-----BEGIN RSA PRIVATE KEY-----
MIICXQIBAAKBgQCmXWddKgE04A0hoIYDJaFcwk6e94MseTBmU5UhNP9IQweXNKDj
hkkRgM3KE3Nx/FX1SthUAQPqhSYHicle8CQEmE8Z1+Ay+GnmJf0rNbTF1WdIFn6z
CIxSwyGUdqgB7o8o0422q75aIWJNp/2JJMEa3hP5j8EQjEiXVbZ3avzRhwIDAQAB
AoGAF6qT6LnwAACbfZwjVFFHGjX+DzkwrOP0kan0PgwXAMWgT89Jz/QrctT68XNA
6gc2hzWFfrXCQ9t7JHlibyIETXQ7fpQHQZLjXXoBCdsm6Whw/OwDu3Sg5R/B1QxW
ZGu6aZsBztynczJrQ99iKYvhr7Vp8ANIdhAc/GKMupAE07kCQQDZ/JYuVcAnsvCi
+7boNx2UGA8p7j0zjMYKa+evqdPWrd0YL9NyCUzP7l/6LjbmvTax6RJEB1C59feu
8dlbCDa7AkEAw2BNzQW5Bh/CiaoAVpYNiqQ5sV9Luj5a66Wy+ndj/e+XwRglOV6B
ZA6t37qzlUX/DVdD9rofzWgAIKfBRYtxpQJBAMxyJACNGE2jfCHAZ0nf93PgJMi0
0t24WD2J+qA8bZxZMJXwtSWtJ0eVUJr6IS/Dorq12BXJrqLa2FRSLAM+7uUCQGYu
pAIqkA5n5fLh+rNOX163bYUa9hw+KIc+blEYyC8zdAcFfdJ3XuzZ0I5Gs03LAg4U
KfOMfL2NOyPZGPgqahECQQCHAa2vPpgEo1MNrdWo/qyPvd2pUGZSLwcvuHkdk+9+
08DhOfLsKPpsSw9jAaOl55plJ/5uhhcUEUVsCBA9UDsV
-----END RSA PRIVATE KEY-----
`)

//openssl
//openssl rsa -in rsa_private_key.pem -pubout -out rsa_public_key.pem
var publicKey = []byte(`  
-----BEGIN PUBLIC KEY-----
MIGfMA0GCSqGSIb3DQEBAQUAA4GNADCBiQKBgQCmXWddKgE04A0hoIYDJaFcwk6e
94MseTBmU5UhNP9IQweXNKDjhkkRgM3KE3Nx/FX1SthUAQPqhSYHicle8CQEmE8Z
1+Ay+GnmJf0rNbTF1WdIFn6zCIxSwyGUdqgB7o8o0422q75aIWJNp/2JJMEa3hP5
j8EQjEiXVbZ3avzRhwIDAQAB
-----END PUBLIC KEY-----    
`)

// 加密
func RsaEncrypt(origData []byte) ([]byte, error) {
	//解密pem格式的公钥
	block, _ := pem.Decode(publicKey)
	if block == nil {
		return nil, errors.New("public key error")
	}
	// 解析公钥
	pubInterface, err := x509.ParsePKIXPublicKey(block.Bytes)
	if err != nil {
		return nil, err
	}
	// 类型断言
	pub := pubInterface.(*rsa.PublicKey)
	//加密
	return rsa.EncryptPKCS1v15(rand.Reader, pub, origData)
}

// 解密
func RsaDecrypt(ciphertext []byte) ([]byte, error) {
	//解密
	block, _ := pem.Decode(privateKey)
	if block == nil {
		return nil, errors.New("private key error!")
	}
	//解析PKCS1格式的私钥
	priv, err := x509.ParsePKCS1PrivateKey(block.Bytes)
	if err != nil {
		return nil, err
	}
	// 解密
	return rsa.DecryptPKCS1v15(rand.Reader, priv, ciphertext)
}
