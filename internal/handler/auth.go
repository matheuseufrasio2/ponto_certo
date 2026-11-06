package handler

import (
	"net/http"

	"github.com/matheuseufrasio2/ponto_certo/internal/service"
)

type AuthHandler struct {
	authService *service.AuthService
}

func NewAuthHandler(authService *service.AuthService) *AuthHandler {
	return &AuthHandler{authService: authService}
}

// Login authenticates a user and returns a signed JWT access token.
// @Summary Realizar autenticação
// @Description Autentica um usuário via e-mail e senha, retornando um JWT com seu papel e escopo de tenant.
// @Tags Auth
// @Accept json
// @Produce json
// @Param request body service.LoginInput true "Credenciais de acesso"
// @Success 200 {object} service.LoginOutput "Token de acesso e dados do usuário"
// @Failure 400 {object} ErrorResponse "Requisição malformada ou inválida"
// @Failure 401 {object} ErrorResponse "Credenciais incorretas"
// @Failure 403 {object} ErrorResponse "Usuário desativado"
// @Failure 500 {object} ErrorResponse "Erro interno do servidor"
// @Router /auth/login [post]
func (h *AuthHandler) Login(w http.ResponseWriter, r *http.Request) {
	var input service.LoginInput

	if err := ReadJSON(w, r, &input); err != nil {
		WriteError(w, http.StatusBadRequest, err.Error())
		return
	}

	output, err := h.authService.Login(r.Context(), input)
	if err != nil {
		HandleError(w, err)
		return
	}

	_ = WriteJSON(w, http.StatusOK, output)

}
