package ginlib

import (
	"context"
	"fmt"
	"github.com/pkg/errors"
	"github.com/tencentyun/cos-go-sdk-v5"
	"io"
	"net/http"
	"net/url"
	"os"
	"path/filepath"
)

var c *cos.Client

//InitTxCloud 初始化腾讯云
func InitTxCloud() {
	if c != nil {
		return
	}
	bucketname := Ini_Str("tx.cos_bucketname")
	appid := Ini_Str("tx.cos_appid")
	region := Ini_Str("tx.cos_region")
	secretId := Ini_Str("tx.cos_secretId")
	secretKey := Ini_Str("tx.cos_secretKey")
	if bucketname == "" || secretId == "" || secretKey == "" || appid == "" || region == "" {
		panic(errors.New("请在conf/app.conf中配置腾讯云参数,tx.cos_bucketname," +
			"tx.cos_appid,tx.cos_region,tx.cos_secretId,tx.cos_secretKey"))
	}
	u, _ := url.Parse(fmt.Sprintf("http://%s-%s.cos.%s.myqcloud.com", bucketname, appid, region))
	b := &cos.BaseURL{BucketURL: u}
	c = cos.NewClient(b, &http.Client{
		Transport: &cos.AuthorizationTransport{
			SecretID:  secretId,
			SecretKey: secretKey,
		},
	})
}

//UploadTxCloud 上传文件到腾讯云
func UploadTxCloud(name string, f io.Reader) error {
	if c == nil {
		return errors.New("请配置腾讯对象存储信息，InitTxCloud")
	}
	//对象键（Key）是对象在存储桶中的唯一标识。
	if _, err := c.Object.Put(context.Background(), name, f, nil); err != nil {
		panic(err)
	}
	return nil
}

//上传文件到本地
func UploadLocalFile(name string, f io.Reader) error {
	//检测目录是否存在，不存在则创建
	p := filepath.Dir(name)
	if err := os.MkdirAll(p, 0777); err != nil {
		return err
	}
	//保存文件
	fio, err := os.Create(name)
	if err != nil {
		return err
	}
	if _, err := io.Copy(fio, f); err != nil {
		return err
	}
	fio.Close()
	return nil
}
