package main

import (
	"context"
	"fmt"
	"log"
	"net/http"
	"sort"
	"time"

	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/aws/credentials"
	"github.com/aws/aws-sdk-go/aws/session"
	"github.com/aws/aws-sdk-go/service/s3"
)

// ==================== 配置项 ====================
const (
	// SeaweedFS S3 服务地址
	S3Endpoint = "http://localhost:8333"
	// 访问密钥（SeaweedFS 默认未启用认证时可填任意值）
	AccessKey = "SEAWEEDFS_S3_ACCESS_KEY"
	SecretKey = "SEAWEEDFS_S3_SECRET_KEY"
	// Region  SeaweedFS 不校验，但 AWS SDK 要求必填
	Region = "ap-northeast-2"
	// 存储桶名称
	BucketName = "my-bucket"
	// 预签名 URL 有效期（秒）
	PresignExpireSeconds = 6000
	// 轮询检查间隔
	PollInterval = 2 * time.Second
	// 轮询超时时间
	PollTimeout = 10 * time.Minute
)

// ==================== 初始化 S3 客户端 ====================
func newS3Client() (*s3.S3, error) {
	log.Println("[INFO] 初始化 SeaweedFS S3 客户端...")
	log.Printf("[INFO] Endpoint: %s", S3Endpoint)
	log.Printf("[INFO] Region:   %s", Region)
	log.Printf("[INFO] AccessKey: %s", AccessKey)

	sess, err := session.NewSession(&aws.Config{
		Endpoint:         aws.String(S3Endpoint),
		Region:           aws.String(Region),
		Credentials:      credentials.NewStaticCredentials(AccessKey, SecretKey, ""),
		DisableSSL:       aws.Bool(true), // SeaweedFS 默认 HTTP
		S3ForcePathStyle: aws.Bool(true), // 必须启用路径风格，否则会用子域名方式
	})
	if err != nil {
		return nil, fmt.Errorf("创建 session 失败: %w", err)
	}

	client := s3.New(sess)
	log.Println("[INFO] S3 客户端初始化完成")
	return client, nil
}

// ==================== 生成 PUT 预签名 URL ====================
func generatePutPresignedURL(svc *s3.S3, objectKey string) (string, error) {
	log.Printf("[INFO] 开始生成 PUT 预签名 URL，对象键: %s", objectKey)
	log.Printf("[INFO] 存储桶: %s", BucketName)
	log.Printf("[INFO] 有效期: %d 秒", PresignExpireSeconds)

	req, _ := svc.PutObjectRequest(&s3.PutObjectInput{
		Bucket: aws.String(BucketName),
		Key:    aws.String(objectKey),
	})

	urlStr, err := req.Presign(time.Duration(PresignExpireSeconds) * time.Second)
	if err != nil {
		return "", fmt.Errorf("生成预签名 URL 失败: %w", err)
	}

	log.Println("[SUCCESS] 预签名 URL 生成成功")
	return urlStr, nil
}

// ==================== 检查文件是否已上传 ====================
func checkObjectExists(svc *s3.S3, objectKey string) (bool, *s3.HeadObjectOutput, error) {
	log.Printf("[DEBUG] 检查对象是否存在: %s", objectKey)

	input := &s3.HeadObjectInput{
		Bucket: aws.String(BucketName),
		Key:    aws.String(objectKey),
	}

	result, err := svc.HeadObject(input)
	if err != nil {
		// 404 表示文件不存在，属于正常情况
		if awsErr, ok := err.(s3.RequestFailure); ok {
			if awsErr.StatusCode() == http.StatusNotFound {
				log.Printf("[DEBUG] 对象 %s 尚未上传（404）", objectKey)
				return false, nil, nil
			}
		}
		return false, nil, fmt.Errorf("HEAD 请求出错: %w", err)
	}

	return true, result, nil
}

// ==================== 轮询等待文件上传 ====================
func waitForUpload(svc *s3.S3, objectKey string) error {
	log.Println("==================================================")
	log.Println("[INFO] 开始轮询检测文件上传状态...")
	log.Printf("[INFO] 轮询间隔: %v", PollInterval)
	log.Printf("[INFO] 超时时间: %v", PollTimeout)
	log.Println("==================================================")

	ctx, cancel := context.WithTimeout(context.Background(), PollTimeout)
	defer cancel()

	ticker := time.NewTicker(PollInterval)
	defer ticker.Stop()

	for {
		select {
		case <-ctx.Done():
			return fmt.Errorf("轮询超时（%v），未检测到文件上传", PollTimeout)

		case <-ticker.C:
			exists, meta, err := checkObjectExists(svc, objectKey)
			if err != nil {
				log.Printf("[WARN] 检查失败: %v，继续轮询...", err)
				continue
			}

			if exists {
				log.Println("==================================================")
				log.Println("[SUCCESS] ✅ 检测到文件已上传成功！")
				log.Printf("[INFO] 文件大小: %d 字节", aws.Int64Value(meta.ContentLength))
				log.Printf("[INFO] ContentType: %s", aws.StringValue(meta.ContentType))
				log.Printf("[INFO] ETag: %s", aws.StringValue(meta.ETag))
				log.Printf("[INFO] 最后修改: %v", aws.TimeValue(meta.LastModified))
				log.Println("==================================================")
				return nil
			}
		}
	}
}

// ==================== 主函数 ====================
func main2() {
	// 从命令行参数获取文件名，默认使用 test-file.txt
	objectKey := "ChatFile/picture/123.jpg"

	log.SetFlags(log.LstdFlags | log.Lmicroseconds)
	log.Println("==================================================")
	log.Println("SeaweedFS S3 预签名 URL 生成 & 上传检测工具")
	log.Println("==================================================")

	// 1. 初始化客户端
	svc, err := newS3Client()
	if err != nil {
		log.Fatalf("[FATAL] 初始化 S3 客户端失败: %v", err)
	}

	// 2. 生成预签名 URL
	presignedURL, err := generatePutPresignedURL(svc, objectKey)
	if err != nil {
		log.Fatalf("[FATAL] 生成预签名 URL 失败: %v", err)
	}

	// 3. 打印 URL 和 Postman 使用说明
	fmt.Println()
	log.Println("================ 预签名 URL ================")
	fmt.Println(presignedURL)
	log.Println("============================================")
	fmt.Println()
	log.Println("[INFO] Postman 使用方法:")
	log.Println("  1. 请求方式: PUT")
	log.Println("  2. URL: 粘贴上面的预签名URL")
	log.Println("  3. Body -> binary -> 选择文件上传")
	log.Println("  4. 点击 Send 即可上传")
	fmt.Println()

	// 4. 开始轮询检测
	err = waitForUpload(svc, objectKey)
	if err != nil {
		log.Fatalf("[FATAL] %v", err)
	}

	log.Println("[INFO] 程序结束")
}

func main3() {
	intervals := [][]int{{1, 4}, {3, 6}, {2, 8}}
	sort.Slice(intervals, func(i, j int) bool {
		if intervals[i][0] < intervals[j][0] {
			return intervals[i][0] < intervals[j][0]
		}
		return intervals[i][1] > intervals[j][1]
	})
	fmt.Print(intervals)
	ans, m := len(intervals), intervals[0][1]
	for i := 1; i < len(intervals); i++ {
		if intervals[i][1] <= m {
			ans--
		}
		m = max(m, intervals[i][1])
	}
	fmt.Print(ans)
}
