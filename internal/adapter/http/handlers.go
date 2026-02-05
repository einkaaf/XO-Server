package http

import (
    "encoding/json"
    "net/http"
    "os"
    "path/filepath"

    "xo-server/internal/domain"
    "xo-server/internal/usecase"
)

type Handler struct {
    auth usecase.AuthService
}

func NewHandler(auth usecase.AuthService) *Handler {
    return &Handler{auth: auth}
}

func (h *Handler) RegisterRoutes(mux *http.ServeMux) {
    mux.HandleFunc("/health", h.handleHealth)
    mux.HandleFunc("/api/register", h.handleRegister)
    mux.HandleFunc("/api/login", h.handleLogin)
    mux.HandleFunc("/docs", h.handleSwaggerUI)
    mux.HandleFunc("/openapi.yaml", h.handleOpenAPI)
    swaggerDir := filepath.Join("docs", "swagger-ui")
    mux.Handle("/swagger-ui/", http.StripPrefix("/swagger-ui/", http.FileServer(http.Dir(swaggerDir))))
}

func (h *Handler) handleHealth(w http.ResponseWriter, r *http.Request) {
    if r.Method != http.MethodGet {
        w.WriteHeader(http.StatusMethodNotAllowed)
        return
    }
    writeJSON(w, http.StatusOK, map[string]string{"status": "ok"})
}

func (h *Handler) handleRegister(w http.ResponseWriter, r *http.Request) {
    if r.Method != http.MethodPost {
        w.WriteHeader(http.StatusMethodNotAllowed)
        return
    }

    var req struct {
        Username string `json:"username"`
        Password string `json:"password"`
    }
    if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
        writeError(w, http.StatusBadRequest, "invalid json")
        return
    }

    user, err := h.auth.Register(r.Context(), req.Username, req.Password)
    if err != nil {
        mapDomainError(w, err)
        return
    }

    writeJSON(w, http.StatusCreated, map[string]string{"user_id": user.ID.String()})
}

func (h *Handler) handleLogin(w http.ResponseWriter, r *http.Request) {
    if r.Method != http.MethodPost {
        w.WriteHeader(http.StatusMethodNotAllowed)
        return
    }

    var req struct {
        Username string `json:"username"`
        Password string `json:"password"`
    }
    if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
        writeError(w, http.StatusBadRequest, "invalid json")
        return
    }

    token, user, err := h.auth.Login(r.Context(), req.Username, req.Password)
    if err != nil {
        mapDomainError(w, err)
        return
    }

    writeJSON(w, http.StatusOK, map[string]string{
        "token":  token,
        "user_id": user.ID.String(),
        "username": user.Username,
    })
}

func mapDomainError(w http.ResponseWriter, err error) {
    switch err {
    case domain.ErrInvalidInput:
        writeError(w, http.StatusBadRequest, err.Error())
    case domain.ErrUnauthorized:
        writeError(w, http.StatusUnauthorized, err.Error())
    case domain.ErrForbidden:
        writeError(w, http.StatusForbidden, err.Error())
    case domain.ErrNotFound:
        writeError(w, http.StatusNotFound, err.Error())
    default:
        writeError(w, http.StatusInternalServerError, "internal error")
    }
}

func writeError(w http.ResponseWriter, status int, message string) {
    writeJSON(w, status, map[string]string{"error": message})
}

func writeJSON(w http.ResponseWriter, status int, v interface{}) {
    w.Header().Set("Content-Type", "application/json")
    w.WriteHeader(status)
    _ = json.NewEncoder(w).Encode(v)
}

func (h *Handler) handleSwaggerUI(w http.ResponseWriter, r *http.Request) {
    if r.Method != http.MethodGet {
        w.WriteHeader(http.StatusMethodNotAllowed)
        return
    }
    serveDocsFile(w, "swagger.html", "text/html; charset=utf-8")
}

func (h *Handler) handleOpenAPI(w http.ResponseWriter, r *http.Request) {
    if r.Method != http.MethodGet {
        w.WriteHeader(http.StatusMethodNotAllowed)
        return
    }
    serveDocsFile(w, "openapi.yaml", "application/yaml")
}

func serveDocsFile(w http.ResponseWriter, name, contentType string) {
    path := filepath.Join("docs", name)
    b, err := os.ReadFile(path)
    if err != nil {
        writeError(w, http.StatusNotFound, "docs file not found")
        return
    }
    w.Header().Set("Content-Type", contentType)
    w.WriteHeader(http.StatusOK)
    _, _ = w.Write(b)
}
