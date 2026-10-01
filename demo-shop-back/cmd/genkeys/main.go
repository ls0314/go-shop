// 生成 RS256 密钥对(仅本地开发用)。用法: go run ./cmd/genkeys <输出目录>
package main

import (
	"crypto/rand"
	"crypto/rsa"
	"crypto/x509"
	"encoding/pem"
	"fmt"
	"os"
	"path/filepath"
)

func main() {
	outDir := "jwt_keys"
	if len(os.Args) > 1 {
		outDir = os.Args[1]
	}
	if err := os.MkdirAll(outDir, 0o755); err != nil {
		panic(err)
	}

	priv, err := rsa.GenerateKey(rand.Reader, 2048)
	if err != nil {
		panic(err)
	}

	privPath := filepath.Join(outDir, "dev_private.pem")
	privFile, err := os.Create(privPath)
	if err != nil {
		panic(err)
	}
	if err := pem.Encode(privFile, &pem.Block{
		Type:  "RSA PRIVATE KEY",
		Bytes: x509.MarshalPKCS1PrivateKey(priv),
	}); err != nil {
		panic(err)
	}
	privFile.Close()

	pubBytes, err := x509.MarshalPKIXPublicKey(&priv.PublicKey)
	if err != nil {
		panic(err)
	}
	pubPath := filepath.Join(outDir, "dev_public.pem")
	pubFile, err := os.Create(pubPath)
	if err != nil {
		panic(err)
	}
	if err := pem.Encode(pubFile, &pem.Block{
		Type:  "PUBLIC KEY",
		Bytes: pubBytes,
	}); err != nil {
		panic(err)
	}
	pubFile.Close()

	fmt.Printf("已生成:\n  %s\n  %s\n", privPath, pubPath)
}
