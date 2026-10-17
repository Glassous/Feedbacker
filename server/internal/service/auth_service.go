package service

import (
	"context"
	"crypto/rand"
	"errors"
	"fmt"
	"io"
	"log/slog"
	"math/big"
	"strings"
	"time"

	"github.com/golang-jwt/jwt/v5"
	"go.mongodb.org/mongo-driver/v2/bson"
	"golang.org/x/crypto/bcrypt"

	"feedbacker/internal/config"
	"feedbacker/internal/model"
	"feedbacker/internal/repository"
)

// ErrInvalidCredentials 表示账号或密码不正确。
var ErrInvalidCredentials = errors.New("invalid credentials")

// ErrUnauthorized 表示令牌无效或已过期。
var ErrUnauthorized = errors.New("unauthorized")

const (
	bcryptCost       = 12
	minPasswordLen   = 8
	maxPasswordLen   = 128
	tokenIssuer      = "feedbacker"
	passwordAlphabet = "abcdefghijkmnopqrstuvwxyzABCDEFGHJKLMNPQRSTUVWXYZ23456789"
)

// 用于在账号不存在时执行一次等价耗时的对比，避免通过响应时间枚举用户名。
const dummyPasswordHash = "$2a$12$R9h/cIPz0gi.URNNX3kh2OPST9/PgBkqquzi.Ss7KIUgO2t0jWMUW"

// AuthService 负责管理员认证、密码与令牌管理。
type AuthService struct {
	cfg      *config.Config
	admins   *repository.AdminRepository
	settings *repository.SettingsRepository
	logger   *slog.Logger
	secret   []byte
}

// LoginResult 是登录与改密成功后的返回结构。
type LoginResult struct {
	Token              string `json:"token"`
	Username           string `json:"username"`
	MustChangePassword bool   `json:"mustChangePassword"`
}

// ResetResult 是命令行重置密码的结果。
type ResetResult struct {
	Username  string
	Password  string
	Generated bool
	Created   bool
}

// Claims 是 JWT 载荷。
type Claims struct {
	Username     string `json:"username"`
	TokenVersion int    `json:"ver"`
	jwt.RegisteredClaims
}

// NewAuthService 创建认证服务。
func NewAuthService(cfg *config.Config, admins *repository.AdminRepository, settings *repository.SettingsRepository, logger *slog.Logger) *AuthService {
	return &AuthService{cfg: cfg, admins: admins, settings: settings, logger: logger}
}

// Bootstrap 加载签名密钥，并在首次启动时生成一次性管理员密码。
func (s *AuthService) Bootstrap(ctx context.Context, out io.Writer) error {
	if s.cfg.JWTSecret != "" {
		s.secret = []byte(s.cfg.JWTSecret)
	} else {
		secret, err := s.settings.GetOrCreateJWTSecret(ctx)
		if err != nil {
			return fmt.Errorf("初始化 JWT 密钥失败: %w", err)
		}
		s.secret = []byte(secret)
	}

	count, err := s.admins.Count(ctx)
	if err != nil {
		return fmt.Errorf("读取管理员账号失败: %w", err)
	}
	if count > 0 {
		return nil
	}

	password, err := GeneratePassword(16)
	if err != nil {
		return err
	}
	hash, err := bcrypt.GenerateFromPassword([]byte(password), bcryptCost)
	if err != nil {
		return err
	}

	admin := &model.Admin{
		Username:           s.cfg.AdminUsername,
		PasswordHash:       string(hash),
		MustChangePassword: true,
	}
	if err := s.admins.Create(ctx, admin); err != nil {
		return fmt.Errorf("创建管理员账号失败: %w", err)
	}

	PrintCredentialBanner(out, "已生成初始管理员账号", admin.Username, password, true)
	return nil
}

// Login 校验密码并签发访问令牌。
func (s *AuthService) Login(ctx context.Context, username, password string) (*LoginResult, error) {
	username = strings.TrimSpace(username)
	if username == "" {
		username = s.cfg.AdminUsername
	}

	admin, err := s.admins.FindByUsername(ctx, username)
	if errors.Is(err, repository.ErrNotFound) {
		_ = bcrypt.CompareHashAndPassword([]byte(dummyPasswordHash), []byte(password))
		return nil, ErrInvalidCredentials
	}
	if err != nil {
		return nil, err
	}
	if err := bcrypt.CompareHashAndPassword([]byte(admin.PasswordHash), []byte(password)); err != nil {
		return nil, ErrInvalidCredentials
	}

	token, err := s.issueToken(admin)
	if err != nil {
		return nil, err
	}
	return &LoginResult{Token: token, Username: admin.Username, MustChangePassword: admin.MustChangePassword}, nil
}

// Authenticate 校验令牌并返回当前管理员。
func (s *AuthService) Authenticate(ctx context.Context, rawToken string) (*model.Admin, error) {
	rawToken = strings.TrimSpace(rawToken)
	if rawToken == "" {
		return nil, ErrUnauthorized
	}

	claims := &Claims{}
	token, err := jwt.ParseWithClaims(rawToken, claims, func(t *jwt.Token) (any, error) {
		if _, ok := t.Method.(*jwt.SigningMethodHMAC); !ok {
			return nil, ErrUnauthorized
		}
		return s.secret, nil
	}, jwt.WithIssuer(tokenIssuer), jwt.WithValidMethods([]string{jwt.SigningMethodHS256.Alg()}))
	if err != nil || !token.Valid {
		return nil, ErrUnauthorized
	}

	id, err := bson.ObjectIDFromHex(claims.Subject)
	if err != nil {
		return nil, ErrUnauthorized
	}
	admin, err := s.admins.FindByID(ctx, id)
	if errors.Is(err, repository.ErrNotFound) {
		return nil, ErrUnauthorized
	}
	if err != nil {
		return nil, err
	}
	if admin.TokenVersion != claims.TokenVersion {
		return nil, ErrUnauthorized
	}
	return admin, nil
}

// ChangePassword 校验旧密码后更新密码，并使旧令牌全部失效。
func (s *AuthService) ChangePassword(ctx context.Context, adminID bson.ObjectID, oldPassword, newPassword string) (*LoginResult, error) {
	if len(newPassword) < minPasswordLen {
		return nil, &ValidationError{Message: "新密码至少需要 8 个字符"}
	}
	if len(newPassword) > maxPasswordLen {
		return nil, &ValidationError{Message: "新密码不能超过 128 个字符"}
	}

	admin, err := s.admins.FindByID(ctx, adminID)
	if err != nil {
		return nil, err
	}
	if err := bcrypt.CompareHashAndPassword([]byte(admin.PasswordHash), []byte(oldPassword)); err != nil {
		return nil, ErrInvalidCredentials
	}
	if oldPassword == newPassword {
		return nil, &ValidationError{Message: "新密码不能与当前密码相同"}
	}

	hash, err := bcrypt.GenerateFromPassword([]byte(newPassword), bcryptCost)
	if err != nil {
		return nil, err
	}
	if err := s.admins.UpdatePassword(ctx, admin.ID, string(hash), false); err != nil {
		return nil, err
	}

	admin.PasswordHash = string(hash)
	admin.MustChangePassword = false
	admin.TokenVersion++

	token, err := s.issueToken(admin)
	if err != nil {
		return nil, err
	}
	return &LoginResult{Token: token, Username: admin.Username, MustChangePassword: false}, nil
}

// ResetPassword 供命令行使用：重置密码，空值则生成一次性密码。
func (s *AuthService) ResetPassword(ctx context.Context, username, password string) (*ResetResult, error) {
	username = strings.TrimSpace(username)
	if username == "" {
		username = s.cfg.AdminUsername
	}

	generated := strings.TrimSpace(password) == ""
	if generated {
		pwd, err := GeneratePassword(16)
		if err != nil {
			return nil, err
		}
		password = pwd
	}
	if len(password) < minPasswordLen {
		return nil, &ValidationError{Message: "密码至少需要 8 个字符"}
	}
	if len(password) > maxPasswordLen {
		return nil, &ValidationError{Message: "密码不能超过 128 个字符"}
	}

	hash, err := bcrypt.GenerateFromPassword([]byte(password), bcryptCost)
	if err != nil {
		return nil, err
	}

	result := &ResetResult{Username: username, Password: password, Generated: generated}

	admin, err := s.admins.FindByUsername(ctx, username)
	if errors.Is(err, repository.ErrNotFound) {
		admin = &model.Admin{
			Username:           username,
			PasswordHash:       string(hash),
			MustChangePassword: generated,
		}
		if err := s.admins.Create(ctx, admin); err != nil {
			return nil, err
		}
		result.Created = true
		return result, nil
	}
	if err != nil {
		return nil, err
	}

	if err := s.admins.UpdatePassword(ctx, admin.ID, string(hash), generated); err != nil {
		return nil, err
	}
	return result, nil
}

func (s *AuthService) issueToken(admin *model.Admin) (string, error) {
	now := time.Now()
	claims := &Claims{
		Username:     admin.Username,
		TokenVersion: admin.TokenVersion,
		RegisteredClaims: jwt.RegisteredClaims{
			Subject:   admin.ID.Hex(),
			Issuer:    tokenIssuer,
			IssuedAt:  jwt.NewNumericDate(now),
			ExpiresAt: jwt.NewNumericDate(now.Add(s.cfg.JWTTTL)),
		},
	}
	return jwt.NewWithClaims(jwt.SigningMethodHS256, claims).SignedString(s.secret)
}

// GeneratePassword 生成指定长度的随机密码，剔除易混淆字符。
func GeneratePassword(length int) (string, error) {
	if length < 1 {
		length = 16
	}
	buf := make([]byte, length)
	maxIndex := big.NewInt(int64(len(passwordAlphabet)))
	for i := range buf {
		n, err := rand.Int(rand.Reader, maxIndex)
		if err != nil {
			return "", err
		}
		buf[i] = passwordAlphabet[n.Int64()]
	}
	return string(buf), nil
}

// PrintCredentialBanner 在标准输出打印账号信息，便于通过容器日志获取。
func PrintCredentialBanner(out io.Writer, title, username, password string, oneTime bool) {
	line := strings.Repeat("=", 68)
	fmt.Fprintf(out, "\n%s\n", line)
	fmt.Fprintf(out, "  %s\n", title)
	fmt.Fprintf(out, "  用户名:     %s\n", username)
	fmt.Fprintf(out, "  密码:       %s\n", password)
	if oneTime {
		fmt.Fprintln(out, "  该密码为一次性密码，登录后必须立即修改。")
	}
	fmt.Fprintln(out, "  忘记密码时可执行: docker compose exec app ./feedbacker admin reset-password")
	fmt.Fprintf(out, "%s\n\n", line)
}
