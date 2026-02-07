package service

import (
	"encoding/hex"
	"errors"

	"github.com/funkymotions/go-ya-practicum-gophkeeper/internal/apperror"
	"github.com/funkymotions/go-ya-practicum-gophkeeper/internal/model"
	"github.com/funkymotions/go-ya-practicum-gophkeeper/internal/ports"
	"github.com/funkymotions/go-ya-practicum-gophkeeper/internal/utils"
	"go.uber.org/zap"
)

type authService struct {
	userRepository ports.UserRepository
	logger         *zap.SugaredLogger
	jwtSecret      []byte
}

type AuthServiceArgs struct {
	UserRepository ports.UserRepository
	Logger         *zap.SugaredLogger
	JWTSecret      []byte
}

var _ ports.AuthService = (*authService)(nil)

func NewAuthService(args AuthServiceArgs) *authService {
	return &authService{
		userRepository: args.UserRepository,
		logger:         args.Logger,
		jwtSecret:      args.JWTSecret,
	}
}

func (s *authService) Register(username, password string) (string, error) {
	_, err := s.userRepository.ReadByUsername(username)
	if err != nil {
		if errors.Is(err, apperror.DBErrorNoRows) {
			// TODO: app pepper from config for paswword hasing
			passwordHash, err := utils.HashPassword([]byte(password))
			if err != nil {
				s.logger.Errorw("failed to hash password", "error", err)

				return "", apperror.AuthErrorGeneric
			}

			user := &model.User{
				Username:     username,
				PasswordHash: hex.EncodeToString(passwordHash),
			}

			user, err = s.userRepository.Create(user)
			if err != nil {
				s.logger.Errorw("failed to create user", "error", err)

				return "", apperror.AuthCreateUserError
			}

			token, err := utils.IssueJWTToken(user.ID, s.jwtSecret)
			if err != nil {
				s.logger.Errorw("failed to issue JWT token", "error", err)

				return "", apperror.AuthErrorGeneric
			}

			s.logger.Infow("User registered successfully", "username", username)

			return string(token), nil
		}

		return "", apperror.AuthErrorGeneric
	}

	return "", apperror.AuthUserExistsError
}

func (s *authService) Authenticate(username, password string) (string, error) {
	user, err := s.userRepository.ReadByUsername(username)
	if errors.Is(err, apperror.DBErrorNoRows) {
		return "", apperror.AuthUserNotExistsError
	}
	if err != nil {
		return "", err
	}

	decoded, err := hex.DecodeString(user.PasswordHash)
	if err != nil {
		return "", err
	}
	if err := utils.ComparePassword(decoded, []byte(password)); err != nil {
		return "", apperror.AuthInvalidCredentialsError
	}

	token, err := utils.IssueJWTToken(user.ID, s.jwtSecret)
	if err != nil {
		s.logger.Errorw("failed to issue JWT token", "error", err)

		return "", err
	}

	s.logger.Infow("User authenticated successfully", "username", username)

	return string(token), nil
}
