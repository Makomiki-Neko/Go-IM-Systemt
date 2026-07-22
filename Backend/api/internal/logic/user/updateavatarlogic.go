// Code scaffolded by goctl. Safe to edit.
// goctl 1.10.1

package user

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"net/http"
	"path/filepath"
	"strings"
	"time"

	"IMM/api/internal/svc"
	"IMM/api/internal/types"
	"IMM/rpc/user/user"

	"github.com/google/uuid"
	"github.com/zeromicro/go-zero/core/logx"
)

type UpdateAvatarLogic struct {
	logx.Logger
	ctx     context.Context
	svcCtx  *svc.ServiceContext
	httpReq *http.Request
}

func NewUpdateAvatarLogic(ctx context.Context, svcCtx *svc.ServiceContext, r *http.Request) *UpdateAvatarLogic {
	return &UpdateAvatarLogic{
		Logger:  logx.WithContext(ctx),
		ctx:     ctx,
		svcCtx:  svcCtx,
		httpReq: r,
	}
}

func (l *UpdateAvatarLogic) UpdateAvatar() (resp *types.UpdateAvatarResponse, err error) {
	// 取得Token中Uid
	uidVal := l.ctx.Value("uid")
	if uidVal == nil {
		return nil, errors.New("uid not found in context")
	}

	uidNum, ok := uidVal.(json.Number)
	if !ok {
		return nil, errors.New("invalid uid type in context")
	}

	uid, err := uidNum.Int64()
	if err != nil {
		return nil, errors.New("invalid uid value")
	}

	// 解析Form表单	--最大文件限制
	const maxFileSize = 5 << 20 // 5MB
	if err := l.httpReq.ParseMultipartForm(maxFileSize); err != nil {
		return nil, errors.New("解析表单失败")
	}

	// 获取上传的文件
	file, fileHeader, err := l.httpReq.FormFile("photo") // 前端 form-data 中的 key
	if err != nil {
		if err == http.ErrMissingFile {
			return nil, errors.New("请上传图片文件")
		}
		return nil, errors.New("获取文件失败")
	}
	defer file.Close()

	// 校验文件大小
	if fileHeader.Size > maxFileSize {
		return nil, fmt.Errorf("图片大小不能超过 %dMB", maxFileSize>>20)
	}

	// 校验文件 MIME 类型
	buffer := make([]byte, 512)
	if _, err := file.Read(buffer); err != nil {
		return nil, errors.New("文件内容读取失败")
	}
	contentType := http.DetectContentType(buffer)
	if !strings.HasPrefix(contentType, "image/") {
		return nil, errors.New("仅支持图片格式（jpeg, png, gif, webp 等）")
	}
	// 重置文件指针
	if _, err := file.Seek(0, 0); err != nil {
		return nil, errors.New("文件指针重置失败")
	}

	// 上传到 SeaweedFS Filer
	ext := filepath.Ext(fileHeader.Filename)
	if ext == "" {
		// 若没有扩展名，根据 content-type 补充
		switch contentType {
		case "image/jpeg":
			ext = ".jpg"
		case "image/png":
			ext = ".png"
		case "image/gif":
			ext = ".gif"
		case "image/webp":
			ext = ".webp"
		default:
			ext = ".bin"
		}
	}
	uniqueName := uuid.New().String() + ext
	// Filer 路径：/avatars/{uid}/{uniqueName}
	path := fmt.Sprintf("avatars/%d/%s", uid, uniqueName)

	filerURL := l.svcCtx.Config.SeaweedFS.Filers[0] + "/" + path
	req, err := http.NewRequestWithContext(context.Background(), "PUT", filerURL, file) // filer支持直接通过put上传
	if err != nil {
		return nil, errors.New("构建上传请求失败")
	}
	req.Header.Set("Content-Type", contentType)

	// 构建临时http put 请求，执行上传
	httpClient := &http.Client{
		Timeout: 30 * time.Second,
	}
	putResp, err := httpClient.Do(req)
	if err != nil {
		logx.Errorf("Filer 上传失败: %v", err)
		return nil, errors.New("图片存储服务异常")
	}
	defer putResp.Body.Close()

	if putResp.StatusCode != http.StatusCreated && putResp.StatusCode != http.StatusOK {
		body, _ := io.ReadAll(putResp.Body)
		logx.Errorf("Filer 返回错误: status=%d, body=%s", putResp.StatusCode, string(body))
		return nil, errors.New("上传图片失败")
	}

	// 删除旧头像
	getResp, getErr := l.svcCtx.UserRpc.GetUserInfo(l.ctx, &user.UserInfoReq{Uid: uint64(uid)})
	if getErr != nil {
		logx.Errorf("获取旧头像信息失败: %v", getErr)
		// 不影响主流程，继续更新，但跳过删除旧文件
	}
	var oldPath string
	if getResp != nil {
		oldPath = getResp.Photo // 头像文件相对路径，如 "avatars/123/old.jpg"
	}

	// 更新数据库记录
	r, err := l.svcCtx.UserRpc.UpdateUserInfo(l.ctx, &user.InfoUpdateReq{Uid: uint64(uid), Photo: path})
	if err != nil {
		// 尝试回滚：删除刚上传的新文件（可选）
		if delErr := l.deleteFilerFile(filerURL, httpClient); delErr != nil {
			logx.Errorf("回滚删除新文件失败: %v", delErr)
		}
		return nil, errors.New("UpLoad Finish, But Updata Failed, " + err.Error())
	}

	// 删除旧头像（如果存在且不同于新路径）
	if oldPath != "" && oldPath != filerURL {
		if delErr := l.deleteFilerFile(l.svcCtx.Config.SeaweedFS.Filers[0]+"/"+oldPath, httpClient); delErr != nil {
			// 删除失败仅记录日志，不影响响应
			logx.Errorf("删除旧头像失败: %v, path: %s", delErr, oldPath)
		}
	}

	return &types.UpdateAvatarResponse{Code: int(r.Code), PhotoFileID: path}, nil
}

// deleteFilerFile 删除 Filer 上的文件
func (l *UpdateAvatarLogic) deleteFilerFile(url string, HTTPClient *http.Client) error {
	req, err := http.NewRequestWithContext(context.Background(), "DELETE", url, nil)
	if err != nil {
		return err
	}
	resp, err := HTTPClient.Do(req)
	if err != nil {
		return err
	}
	defer resp.Body.Close()
	// 204 No Content 或 404 Not Found 都表示成功（已删除或不存在）
	if resp.StatusCode != http.StatusNoContent && resp.StatusCode != http.StatusNotFound {
		body, _ := io.ReadAll(resp.Body)
		return fmt.Errorf("unexpected status %d: %s", resp.StatusCode, string(body))
	}
	return nil
}
